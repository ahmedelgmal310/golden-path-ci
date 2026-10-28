#!/usr/bin/env bash
# Meta-test: prove that each golden-path gate fails on its negative fixture,
# at the expected step, for the expected reason.
#
#  1. Create a throwaway commit with the same tree as the commit under test and
#     point a meta-test/<run> branch at it. The negative run's red status lands
#     on that commit, not on main.
#  2. Dispatch negative-fixtures.yml on that branch and wait for it.
#  3. For each expected failure, assert:
#       - the job concluded "failure"
#       - the first failed step is the gate step (every earlier step passed)
#       - the job log contains the evidence (rule ID, CVE ID, file name)
#  4. Assert no other job failed, then delete the branch.
#
# Env: GH_TOKEN (contents:write, actions:write), GITHUB_REPOSITORY, GITHUB_SHA,
#      GITHUB_RUN_ID, GITHUB_RUN_ATTEMPT, optional GITHUB_STEP_SUMMARY.
set -euo pipefail

repo="${GITHUB_REPOSITORY:?}"
src="${GITHUB_SHA:?}"
branch="meta-test/${GITHUB_RUN_ID:?}-${GITHUB_RUN_ATTEMPT:-1}"
summary="${GITHUB_STEP_SUMMARY:-/dev/null}"
workdir="$(mktemp -d)"

cleanup() {
  gh api -X DELETE "repos/${repo}/git/refs/heads/${branch}" >/dev/null 2>&1 || true
  rm -rf "$workdir"
}
trap cleanup EXIT

echo "::group::Create ${branch} (same tree as ${src})"
tree="$(gh api "repos/${repo}/git/commits/${src}" --jq .tree.sha)"
commit="$(gh api "repos/${repo}/git/commits" \
  -f message="meta-test: negative fixtures for ${src}" \
  -f tree="$tree" -f "parents[]=${src}" --jq .sha)"
gh api "repos/${repo}/git/refs" -f ref="refs/heads/${branch}" -f sha="$commit" >/dev/null
echo "commit ${commit} tree ${tree}"
echo "::endgroup::"

gh workflow run negative-fixtures.yml --repo "$repo" --ref "$branch" -f source-sha="$src"

run_id=""
for _ in $(seq 1 40); do
  run_id="$(gh run list --repo "$repo" --workflow negative-fixtures.yml --branch "$branch" \
    --json databaseId --jq '.[0].databaseId // empty')"
  [ -n "$run_id" ] && break
  sleep 3
done
if [ -z "$run_id" ]; then
  echo "::error::negative-fixtures run did not start on ${branch}"
  exit 1
fi
run_url="https://github.com/${repo}/actions/runs/${run_id}"
echo "Negative run: ${run_url}"

# The run is supposed to fail, so the exit status of watch is not the verdict.
gh run watch "$run_id" --repo "$repo" --interval 10 >/dev/null 2>&1 || true
while [ "$(gh run view "$run_id" --repo "$repo" --json status --jq .status)" != "completed" ]; do
  sleep 10
done

gh api "repos/${repo}/actions/runs/${run_id}/jobs?per_page=100" > "$workdir/jobs.json"
run_conclusion="$(gh run view "$run_id" --repo "$repo" --json conclusion --jq .conclusion)"

failures=0
rows=()

# assert_failure <job name> <gate step name> <regex> [<regex> ...]
assert_failure() {
  local job="$1" step="$2"
  shift 2
  local id conclusion first_failed earlier_bad log reason=""
  id="$(jq -r --arg n "$job" '.jobs[] | select(.name == $n) | .id' "$workdir/jobs.json")"
  if [ -z "$id" ]; then
    reason="job not found"
  else
    conclusion="$(jq -r --arg n "$job" '.jobs[] | select(.name == $n) | .conclusion' "$workdir/jobs.json")"
    first_failed="$(jq -r --arg n "$job" \
      '[.jobs[] | select(.name == $n) | .steps[] | select(.conclusion == "failure")][0].name // ""' "$workdir/jobs.json")"
    earlier_bad="$(jq -r --arg n "$job" --arg s "$step" '
      [.jobs[] | select(.name == $n) | .steps] | first
      | (map(.name) | index($s)) as $i
      | if $i == null then "gate step missing"
        else [.[:$i][] | select(.conclusion != "success" and .conclusion != "skipped") | .name] | join(", ")
        end' "$workdir/jobs.json")"
    if [ "$conclusion" != "failure" ]; then
      reason="conclusion is ${conclusion}, expected failure"
    elif [ "$first_failed" != "$step" ]; then
      reason="first failed step is '${first_failed}', expected '${step}'"
    elif [ -n "$earlier_bad" ]; then
      reason="steps before the gate did not pass: ${earlier_bad}"
    else
      gh api "repos/${repo}/actions/jobs/${id}/logs" > "$workdir/${id}.log"
      for pattern in "$@"; do
        if ! grep -Eq -- "$pattern" "$workdir/${id}.log"; then
          reason="log does not match /${pattern}/"
          break
        fi
      done
    fi
  fi
  if [ -z "$reason" ]; then
    echo "PASS  ${job}: failed at '${step}' with evidence: $*"
    rows+=("| \`${job}\` | ${step} | $(printf '`%s` ' "$@")| expected failure |")
  else
    echo "::error title=meta-test::${job}: ${reason}"
    rows+=("| \`${job}\` | ${step} | $(printf '`%s` ' "$@")| **WRONG: ${reason}** |")
    failures=$((failures + 1))
  fi
}

assert_failure "leaked-secret / Gitleaks" "Gitleaks" \
  'RuleID:[[:space:]]+private-key' 'leaks found: 1'
assert_failure "critical-cve-image / Build and scan" "Trivy vulnerability gate" \
  'CVE-2021-36159' 'Trivy gate failed: [1-9][0-9]* '
assert_failure "unformatted-terraform / Terraform checks" "terraform fmt" \
  'main\.tf'

# Nothing else may fail, and the gated publish job must not have run.
unexpected="$(jq -r '[.jobs[] | select(.conclusion == "failure") | .name
  | select(. != "leaked-secret / Gitleaks" and . != "critical-cve-image / Build and scan"
           and . != "unformatted-terraform / Terraform checks")] | join(", ")' "$workdir/jobs.json")"
if [ -n "$unexpected" ]; then
  echo "::error title=meta-test::unexpected failed jobs: ${unexpected}"
  failures=$((failures + 1))
fi
publish="$(jq -r '.jobs[] | select(.name == "critical-cve-image / Publish, sign, attest") | .conclusion' "$workdir/jobs.json")"
if [ "$publish" != "skipped" ]; then
  echo "::error title=meta-test::publish job for the vulnerable image concluded '${publish}', expected skipped"
  failures=$((failures + 1))
fi
if [ "$run_conclusion" != "failure" ]; then
  echo "::error title=meta-test::negative run concluded '${run_conclusion}', expected failure"
  failures=$((failures + 1))
fi

{
  echo "### Meta-test: negative fixtures"
  echo
  echo "Negative run: [${run_id}](${run_url}) on throwaway commit \`${commit:0:7}\` (tree of \`${src:0:7}\`), conclusion **${run_conclusion}**."
  echo
  echo "| Job | Gate step | Evidence in log | Verdict |"
  echo "|---|---|---|---|"
  printf '%s\n' "${rows[@]}"
  echo
  echo "Publish job for the vulnerable image: **${publish:-missing}** (must be skipped)."
} >> "$summary"

if [ "$failures" -gt 0 ]; then
  echo "meta-test: ${failures} assertion(s) failed"
  exit 1
fi
echo "meta-test: all negative fixtures failed for the right reason"
