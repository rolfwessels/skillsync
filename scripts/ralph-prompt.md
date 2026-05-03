# Ralph loop iteration prompt

You are running inside the Ralph loop driven by `scripts/ralph.sh`.
Do exactly one issue per invocation, then exit. The bash loop will
re-spawn you with a fresh context for the next iteration.

## How to work this iteration

1. If `.scratch/ralph-steering.md` exists and is non-empty, read it first
   and treat its contents as the highest-priority instructions for this
   iteration.
2. Follow the implement-issues agent at
   `.cursor/agents/agent-implement-issues.md` — pick the next unblocked
   `Status: ready-for-agent` issue and implement it with TDD.
3. **Skip the per-issue user confirmation.** This invocation is
   non-interactive; just go.
4. Run `make test` (fall back to `go test ./...` if the Make target is
   empty) before marking the issue done.
5. When all acceptance criteria pass, update the issue file:
   `Status: ready-for-agent` -> `Status: done` and append the
   `## Implementation notes` section described in the agent skill.
6. **Commit per finished issue.** Stage all changes and commit with a
   conventional message: `feat(<slug>): <issue title>` (or `fix:` /
   `chore:` / `refactor:` as appropriate). The user invoked
   `scripts/ralph.sh`, which is the explicit ask required by
   `.cursor/rules/no-auto-commit.mdc`.

## Stop conditions

- If you find no `ready-for-agent` issue without unresolved blockers,
  output the literal token `RALPH_DONE` on its own line and exit.
- Otherwise exit normally after finishing exactly one issue. Do NOT
  loop yourself — the bash script drives the next iteration.

## Constraints

- Only touch the one issue you are implementing.
- Never modify other issues' `Status:` line.
- Never start an issue whose `## Blocked by` references a non-`done`
  issue.
- Tests must pass before the commit.
