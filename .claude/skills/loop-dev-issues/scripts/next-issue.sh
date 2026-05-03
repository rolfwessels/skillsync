#!/usr/bin/env bash
# Print the path of the next unblocked ready-for-agent issue, sorted by path.
# Exits 1 if there is no eligible issue.
#
# An issue is eligible when:
#   - Status: ready-for-agent
#   - Every path mentioned in its "## Blocked by" section refers to an
#     issue file whose Status: is "done".
#
# Blocker paths are read from backtick-quoted *.md references in the
# "## Blocked by" section. Free-text ranges like "01-foo.md through 11-bar.md"
# are not expanded — only the explicitly backticked paths are checked.
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

declare -A status_map
while IFS=$'\t' read -r path status; do
  status_map["$path"]="$status"
done < <("$script_dir/list-issues.sh")

extract_blockers() {
  awk '
    /^## Blocked by/ { in_section = 1; next }
    in_section && /^## / { in_section = 0 }
    in_section {
      while (match($0, /`[^`]+\.md`/)) {
        print substr($0, RSTART + 1, RLENGTH - 2)
        $0 = substr($0, RSTART + RLENGTH)
      }
    }
  ' "$1"
}

is_blocked() {
  local issue="$1" blocker status
  while IFS= read -r blocker; do
    [[ -z "$blocker" ]] && continue
    status="${status_map[$blocker]:-unknown}"
    if [[ "$status" != "done" ]]; then
      return 0
    fi
  done < <(extract_blockers "$issue")
  return 1
}

result=""
while IFS= read -r path; do
  if ! is_blocked "$path"; then
    result="$path"
    break
  fi
done < <(
  for p in "${!status_map[@]}"; do
    [[ "${status_map[$p]}" == "ready-for-agent" ]] && printf '%s\n' "$p"
  done | sort
)

if [[ -z "$result" ]]; then
  exit 1
fi
printf '%s\n' "$result"
