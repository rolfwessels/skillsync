---
name: loop-dev-issues
description: Loop through local .scratch/ issues with status ready-for-agent and implement them one by one using TDD. Use when the user says "implement the next issue", "work through the backlog", "pick up an issue", or asks to implement outstanding work.
---

# Implement Issues

Reads the local issue tracker under `.scratch/`, picks the next unblocked `ready-for-agent` issue, implements it using TDD, then marks it done and moves to the next.

## Issue file conventions

- Issues live at `.scratch/<slug>/issues/<NN>-<slug>.md`
- `Status:` line near the top controls workflow state
- `## Blocked by` section lists blocking issue paths (or "None")

## Step 1 — Find the next issue

1. Find all files matching `.scratch/**/issues/*.md`
2. Read each file and collect those with `Status: ready-for-agent`
3. Filter out any whose `## Blocked by` references an issue that is NOT `Status: done`
4. Sort by path (lowest number first within a slug)
5. Pick the first — that's your issue

If none found: report current statuses and stop. Don't invent work.

## Step 2 — Read and confirm

Read the full issue file. Present to the user:
- **Issue**: file path
- **What to build**: the `## What to build` section
- **Acceptance criteria**: the checklist

Ask: "Ready to implement this? Any context I should know first?"

## Step 3 — Implement with TDD

Follow the TDD workflow (red-green-refactor, vertical slices):

1. Read `CONTEXT.md` and `docs/adr/` — use project domain vocabulary throughout
2. For each acceptance criterion:
   - Write a failing test that captures the behavior
   - Write minimal code to make it pass
   - Refactor if needed
   - Confirm test passes before moving to next criterion
3. Tests must use public interfaces only — not internal implementation details
4. Run the full test suite before declaring done

If the codebase has a Makefile with a `test` target, use `make test`. Otherwise infer the test command from the project (`go test ./...`, `cargo test`, etc.).

## Step 4 — Mark done

When all acceptance criteria pass:

1. Update the issue file: change `Status: ready-for-agent` → `Status: done`
2. Append to the bottom of the issue file:

```
## Implementation notes

<brief summary of what was built and any notable decisions>
```

## Step 5 — Continue or stop

Ask: "Issue done. Continue to next issue, or stop here?"

- If continue: repeat from Step 1
- If stop: summarise what was completed this session


