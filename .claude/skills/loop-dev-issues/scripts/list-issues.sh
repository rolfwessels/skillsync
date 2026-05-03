#!/usr/bin/env bash
# Print "<path>\t<status>" for every issue file under .scratch/<slug>/issues/.
# Usage: list-issues.sh [--status STATUS]
#   --status   restrict output to issues whose Status: line matches STATUS
set -euo pipefail

filter=""
if [[ "${1:-}" == "--status" ]]; then
  filter="${2:-}"
  if [[ -z "$filter" ]]; then
    echo "list-issues.sh: --status requires a value" >&2
    exit 2
  fi
fi

if [[ ! -d .scratch ]]; then
  echo "list-issues.sh: no .scratch/ directory in $(pwd)" >&2
  exit 2
fi

while IFS= read -r -d '' f; do
  status=$(grep -m1 '^Status:' "$f" 2>/dev/null | sed -E 's/^Status:[[:space:]]*//; s/[[:space:]]+$//' | tr -d '\r' || true)
  if [[ -z "$filter" || "$status" == "$filter" ]]; then
    printf '%s\t%s\n' "$f" "$status"
  fi
done < <(find .scratch -path '*/issues/*.md' -type f -print0 | sort -z)
