---
name: commit
description: Runs tests, checks documentation currency, then stages and commits all changes with a well-formed message. Use when the user says "commit", "ship it", "check and commit", or finishes implementing a feature and wants to lock in the work.
---

# Commit

## Step 1 — Run tests

```bash
docker compose exec dev go test ./...
```

If tests fail, stop and report the failures. Do not proceed to commit.

## Step 2 — Update CONTEXT.md (when relevant)

Read `CONTEXT.md` and compare it against the current changes. Update it **only** if the changes:

- Introduce a new domain term (e.g. a new concept the codebase now relies on)
- Rename or redefine an existing term
- Add, remove, or change a relationship between domain concepts
- Resolve a flagged ambiguity in the `## Flagged ambiguities` section
- Change the `Kind→path map` or any other normative table

If none of the above apply, **do not touch CONTEXT.md**. Keep any updates minimal and precise — one or two lines, not a rewrite.

## Step 3 — Check other docs

Review these files for staleness:

- `docs/adr/` — any architectural decisions made implicitly during implementation
- `README.md` — usage instructions or feature list

Update only if the changes affect them. Keep updates minimal and precise.

## Step 4 — Commit

1. Run `git status` and `git diff` to review all staged and unstaged changes
2. Run `git log -5 --oneline` to match the existing commit style
3. Stage all relevant files: `git add -A` (or selectively if some files should be excluded)
4. Write a commit message:
   - Subject line: imperative mood, ≤72 chars, no period
   - Body (if needed): explain *why*, not *what*
5. Commit using a HEREDOC to preserve formatting:

```bash
git commit -m "$(cat <<'EOF'
subject line here

Optional body here.
EOF
)"
```

6. Confirm with `git log -1` that the commit landed cleanly.

## Rules

- Never push
- Never use `--no-verify`
- Never amend a commit that has already been pushed
