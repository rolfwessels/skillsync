---
name: to-issues-local
description: Break a plan, spec, or PRD into independently-grabbable issues on the project issue tracker using tracer-bullet vertical slices. Use when user wants to convert a plan into issues, create implementation tickets, or break down work into issues.
---

# To Issues

Break a plan into independently-grabbable issues using vertical slices (tracer bullets).

Issues for this project live as markdown files in `.scratch/`. See the [local-issues-local skill](../local-issues-local/SKILL.md) for file conventions.

## Process

### 1. Gather context

Work from whatever is already in the conversation context. If the user passes an issue reference (path like `.scratch/<slug>/PRD.md`), read its full body.

### 2. Explore the codebase (optional)

If you haven't already, explore the codebase to understand the current state. Issue titles and descriptions should use the project's domain glossary vocabulary, and respect ADRs in the area you're touching.

### 3. Draft vertical slices

Break the plan into **tracer bullet** issues. Each issue is a thin vertical slice that cuts through ALL integration layers end-to-end, NOT a horizontal slice of one layer.

Slices may be **HITL** (requires human interaction) or **AFK** (can be implemented by an agent with no human context). Prefer AFK over HITL where possible.

Good slices:
- Deliver a narrow but COMPLETE path through every layer (schema, API, UI, tests)
- Are demoable or verifiable on their own
- Are thin rather than thick

### 4. Quiz the user

Present the proposed breakdown as a numbered list. For each slice, show:

- **Title**: short descriptive name
- **Type**: HITL / AFK
- **Blocked by**: which other slices must complete first
- **User stories covered**: which user stories this addresses (if source had them)

Ask the user:

- Does the granularity feel right?
- Are the dependency relationships correct?
- Should any slices be merged or split further?
- Are the correct slices marked as HITL and AFK?

Iterate until the user approves the breakdown.

### 5. Write the issue files

For each approved slice, create `.scratch/<slug>/issues/<NN>-<slug>.md` (numbered from `01`, in dependency order so blockers are created first).

Use this template:

```markdown
Status: needs-triage

## Parent

<path to parent PRD, e.g. `.scratch/<slug>/PRD.md`> (omit if no parent)

## What to build

A concise description of this vertical slice. Describe the end-to-end behavior, not layer-by-layer implementation.

## Acceptance criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## Blocked by

- `.scratch/<slug>/issues/<NN>-<slug>.md` (or "None - can start immediately")
```

Do NOT close or modify any parent PRD.
