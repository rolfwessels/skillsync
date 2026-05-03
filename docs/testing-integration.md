# Integration Testing — Multi-Format Golden Path

Manual scenarios for verifying `init`, `pull`, and `push` when a project declares both `cursor` and `claude` formats.

## Prerequisites

- `skillsync` binary built (`make publish` or `go build ./cmd/skillsync`)
- A registry on disk (the `testdata/registry` folder works, or point at your real registry)

```bash
REGISTRY=$(pwd)/testdata/registry
PROJECT=$(mktemp -d)
```

---

## Scenario A — init with multiple formats

```bash
cd "$PROJECT"
skillsync init
# when prompted:
#   registry: $REGISTRY
#   formats:  cursor, claude
#   bundles:  rules/naming-convention, skills/tdd
```

**Verify:**

- `.skillsync/config.toml` exists and contains `formats = ["cursor", "claude"]`
- `.skillsync/sync.lock` exists (init triggers a pull)
- Both format outputs are present (see Scenario B)

---

## Scenario B — pull writes both formats

Starting from an initialised project (or run `skillsync pull` in the project from Scenario A):

```bash
skillsync pull
```

**Verify:**

| Path | Expected |
|------|----------|
| `.claude/rules/naming-convention.md` | exists; frontmatter contains `alwaysApply:` and `paths:` |
| `.cursor/rules/naming-convention.mdc` | exists; frontmatter contains `alwaysApply:` |
| `.claude/skills/tdd/SKILL.md` | exists |
| `.cursor/skills/tdd/SKILL.md` | exists |
| `.skillsync/sync.lock` | contains 4 entries (2 bundles × 2 formats) |

Check the lockfile entry count:

```bash
grep 'bundle =' "$PROJECT/.skillsync/sync.lock" | wc -l
# expect: 4
```

---

## Scenario C — pull after registry update

```bash
# add a line to the registry bundle
echo "" >> "$REGISTRY/rules/naming-convention/rule.md"
echo "## Registry Update" >> "$REGISTRY/rules/naming-convention/rule.md"

skillsync pull
```

**Verify:**

- Both `.claude/rules/naming-convention.md` and `.cursor/rules/naming-convention.mdc` contain `Registry Update`
- The `hash` values in `.skillsync/sync.lock` differ from before the update

```bash
grep -A3 'bundle = "rules/naming-convention"' "$PROJECT/.skillsync/sync.lock"
```

---

## Scenario D — push from cursor format

```bash
# edit the cursor output
echo "" >> "$PROJECT/.cursor/rules/naming-convention.mdc"
echo "## Pushed Section" >> "$PROJECT/.cursor/rules/naming-convention.mdc"

skillsync push
```

**Verify — registry updated with canonical Claude format:**

```bash
cat "$REGISTRY/rules/naming-convention/rule.md"
# expect: contains "Pushed Section"
# expect: frontmatter uses alwaysApply: and paths: (canonical Claude fields)
```

**Verify — next pull is consistent in both formats:**

```bash
FRESH=$(mktemp -d)
cp -r "$PROJECT/.skillsync" "$FRESH/"
skillsync pull   # run from $FRESH, or: skillsync pull --project "$FRESH"

grep "Pushed Section" "$FRESH/.claude/rules/naming-convention.md"
grep "Pushed Section" "$FRESH/.cursor/rules/naming-convention.mdc"
```

---

## Automated tests

All four scenarios above are covered by the automated integration tests in
`internal/integration/multiformat_test.go` and run as part of the standard test suite:

```bash
go test ./...
# or, to run only integration tests:
go test ./internal/integration/...
```
