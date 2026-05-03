# Matt Pocock Skills to Try

From [mattpocock/skills](https://github.com/mattpocock/skills) — engineering skills worth experimenting with during development.

---

## Top picks

### 1. `grill-me` / `grill-with-docs`
**Category**: Productivity / Engineering  
**When**: Before starting any non-trivial feature.

The most impactful skill in the repo. Instead of diving in, the agent interviews you relentlessly about the plan — one question at a time — until every branch of the decision tree is resolved. It surfaces misalignment *before* you write code rather than after.

`grill-with-docs` is the enhanced version: it also challenges your plan against existing domain terminology in `CONTEXT.md` and will propose ADRs when you make a hard-to-reverse decision.

**Try it when**: You're about to build a new feature, design a new module, or make an architectural decision.

---

### 2. `tdd`
**Category**: Engineering  
**When**: Implementing any new feature or fixing a bug.

Red-green-refactor, but done correctly. Key insight: avoid "horizontal slicing" (write all tests first, then all code). Instead it enforces **vertical slices** — one test, one implementation, repeat. Tests must verify behavior through public interfaces, not implementation details.

Also includes solid guidance on mocking (only at true system boundaries), deep modules (small interface, lots of implementation), and what makes a good vs bad test.

**Try it when**: Starting a new feature or when existing tests feel brittle.

---

### 3. `diagnose`
**Category**: Engineering  
**When**: Debugging a hard bug or performance regression.

Structured 6-phase debugging loop: build a feedback loop → reproduce → hypothesise → instrument → fix → cleanup. The key insight is that **building a fast, deterministic pass/fail signal is the real skill** — everything else is mechanical.

Forces you to generate 3–5 ranked hypotheses *before* testing any of them, and to show them to the user before instrumenting.

**Try it when**: A bug is taking more than 15 minutes to track down, or a performance regression appeared.

---

### 4. `zoom-out`
**Category**: Engineering  
**When**: Navigating unfamiliar code.

One-liner skill: tells the agent to go up a layer of abstraction and give a map of all relevant modules and callers in domain vocabulary. Great antidote to the agent getting lost in the weeds.

**Try it when**: The agent (or you) doesn't have context on how a piece of code fits into the bigger picture.

---

### 5. `improve-codebase-architecture`
**Category**: Engineering  
**When**: Periodic codebase health check (every few days / after a sprint).

Finds "deepening opportunities" — places where shallow modules (large interface, thin implementation) can be collapsed into deep ones (small interface, lots of hidden complexity). Uses a consistent vocabulary: **module**, **interface**, **seam**, **adapter**, **leverage**, **locality**.

Includes a grilling loop once you pick a candidate, and will update `CONTEXT.md` and offer ADRs inline as decisions crystallise.

**Try it when**: The codebase feels hard to navigate, tests are brittle, or you just shipped a chunk of features and want to consolidate.

---

## Lower priority (situational)

| Skill | When useful |
|-------|-------------|
| `to-prd` | Turn a conversation into a structured PRD submitted as a GitHub issue |
| `to-issues` | Break a PRD into independently-grabbable vertical-slice GitHub issues |
| `triage` | Systematic issue triage with state machine roles |
| `caveman` | Cut token usage ~75% by switching to ultra-compressed communication |

---

## Setup note

The skills expect a `CONTEXT.md` at the repo root (shared domain glossary) and optionally `docs/adr/` for architectural decision records. Both are created lazily the first time they're needed — no manual setup required.
