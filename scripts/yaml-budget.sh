#!/usr/bin/env bash
# Consumer YAML budget: for each sample consumer workflow in examples/, count
# the lines the service team writes and the lines of library YAML (reusable
# workflows + the composite actions they call) that those lines replace.
# A "line" is a non-blank, non-comment line.
set -euo pipefail
cd "$(dirname "$0")/.."

loc() { grep -cvE '^[[:space:]]*(#|$)' "$@" | awk -F: '{s += $NF} END {print s + 0}'; }

echo "### Consumer YAML budget"
echo
echo "| Consumer workflow | Lines the team writes | Library YAML it reuses | Reused workflows and actions |"
echo "|---|---:|---:|---|"

total_consumer=0
total_library=0
for f in examples/*/.github/workflows/*.yml; do
  consumer=$(loc "$f")
  mapfile -t workflows < <(grep -oE 'golden-path-ci/\.github/workflows/[a-z-]+\.yml' "$f" | sed 's#golden-path-ci/##' | sort -u)
  files=("${workflows[@]}")
  for w in "${workflows[@]}"; do
    while read -r a; do
      [ -n "$a" ] && files+=("actions/${a}/action.yml")
    done < <(grep -oE '\$/actions/[a-z-]+' "$w" | sed 's#.*/##' | sort -u)
  done
  mapfile -t files < <(printf '%s\n' "${files[@]}" | sort -u)
  library=$(loc "${files[@]}")
  names=$(printf '%s\n' "${files[@]}" | sed -e 's#\.github/workflows/##' -e 's#actions/\(.*\)/action.yml#\1 (action)#' | paste -sd, - | sed 's/,/, /g')
  echo "| \`${f}\` | ${consumer} | ${library} | ${names} |"
  total_consumer=$((total_consumer + consumer))
  total_library=$((total_library + library))
done
echo
echo "Across all examples: **${total_consumer}** consumer lines reuse **${total_library}** lines of maintained library YAML."
