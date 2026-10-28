#!/usr/bin/env bash
# Usage: scripts/bench-report.sh <cold-run-id> <warm-run-id> [repo]
# Prints a Markdown table of job durations (started_at -> completed_at) for a
# cold and a warm benchmark run. Requires gh and jq.
set -euo pipefail
cold="${1:?cold run id}"
warm="${2:?warm run id}"
repo="${3:-${GITHUB_REPOSITORY:-ahmedelgmal310/golden-path-ci}}"

durations() {
  gh api "repos/${repo}/actions/runs/$1/jobs?per_page=100" --jq '
    .jobs[] | select(.conclusion == "success")
    | [.name, ((.completed_at | fromdateiso8601) - (.started_at | fromdateiso8601))] | @tsv'
}

declare -A c w
while IFS=$'\t' read -r name secs; do c["$name"]=$secs; done < <(durations "$cold")
while IFS=$'\t' read -r name secs; do w["$name"]=$secs; done < <(durations "$warm")

fmt() { printf '%dm %02ds' $(($1 / 60)) $(($1 % 60)); }

echo "| Job | Cold cache | Warm cache | Saved |"
echo "|---|---:|---:|---:|"
while IFS= read -r name; do
  [ -n "${w[$name]:-}" ] || continue
  cs=${c[$name]}
  ws=${w[$name]}
  pct=$(( (cs - ws) * 100 / (cs > 0 ? cs : 1) ))
  echo "| \`${name}\` | $(fmt "$cs") | $(fmt "$ws") | ${pct}% |"
done < <(printf '%s\n' "${!c[@]}" | sort)
echo
echo "Cold: https://github.com/${repo}/actions/runs/${cold} · Warm: https://github.com/${repo}/actions/runs/${warm}"
