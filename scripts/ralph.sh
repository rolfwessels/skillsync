#!/usr/bin/env bash
# Ralph loop: re-pipe scripts/ralph-prompt.md to cursor-agent until the
# .scratch/ issue queue is empty, RALPH_DONE is signalled, or the iteration
# cap is hit. Each iteration runs in a fresh cursor-agent session.
set -uo pipefail

MAX_ITERATIONS=10
while [[ $# -gt 0 ]]; do
  case "$1" in
    -n) MAX_ITERATIONS="$2"; shift 2 ;;
    -h|--help)
      echo "usage: $0 [-n MAX_ITERATIONS]"
      exit 0
      ;;
    *)
      echo "usage: $0 [-n MAX_ITERATIONS]" >&2
      exit 1
      ;;
  esac
done

PROMPT_FILE="scripts/ralph-prompt.md"
LOG_DIR=".scratch/ralph-logs"

if [[ ! -f "$PROMPT_FILE" ]]; then
  echo "missing $PROMPT_FILE" >&2
  exit 1
fi

mkdir -p "$LOG_DIR"

for i in $(seq 1 "$MAX_ITERATIONS"); do
  echo "=== Ralph iteration $i / $MAX_ITERATIONS ==="

  if ! grep -rl "Status: ready-for-agent" .scratch/ >/dev/null 2>&1; then
    echo "No ready-for-agent issues remain. Done."
    exit 0
  fi

  ts="$(date +%Y%m%d-%H%M%S)"
  log="$LOG_DIR/iter-$(printf '%02d' "$i")-${ts}.log"

  cursor-agent -p --force --output-format text < "$PROMPT_FILE" | tee "$log"

  if grep -q "RALPH_DONE" "$log"; then
    echo "Agent signalled completion."
    exit 0
  fi
done

echo "Hit max iterations ($MAX_ITERATIONS) without completion."
exit 1
