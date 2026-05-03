# Testing Transforms

Manual testing guide for the transform layer. Run these scenarios after changing any transform logic to verify end-to-end behaviour.

Prerequisites: `make import` and `make publish` have been run successfully.

---

## Setup

```bash
rm -rf /tmp/skillsync-test
mkdir /tmp/skillsync-test
cd /tmp/skillsync-test
git init
echo "readme" > README.md
git add README.md
git commit -m "Initial commit"
skillsync init
```

---

## Skill: cursor push → registry

Skills have no frontmatter — content is copied unchanged in both directions.

```bash
# Confirm registry does not yet contain marker
cat ~/.skillsync/registry/skills/loop-dev-issues/SKILL.md | grep 'testing >>>'
# Should print nothing

echo 'testing >>>' >> .cursor/skills/loop-dev-issues/SKILL.md
skillsync push

cat ~/.skillsync/registry/skills/loop-dev-issues/SKILL.md | grep 'testing >>>'
# Should print: testing >>>

# Cleanup
sed -i 's/testing >>>//' ~/.skillsync/registry/skills/loop-dev-issues/SKILL.md
skillsync pull

cat .cursor/skills/loop-dev-issues/SKILL.md | grep 'testing >>>'
# Should print nothing
```

---

## Rule: claude → cursor field mapping

`paths` (YAML array) and `alwaysApply` are the canonical Claude registry fields. When syncing to cursor format, `paths` is collapsed to a comma-separated `globs` scalar.

```bash
# Inspect a rule in the registry (claude canonical format)
cat ~/.skillsync/registry/rules/go-naming/go-naming.md | head -6
# Expected frontmatter: alwaysApply, paths (claude field names)

# Sync and inspect the cursor output
skillsync pull
cat .cursor/rules/go-naming.mdc | head -5
# Expected frontmatter: alwaysApply, globs (cursor field names)
```

### Rule: cursor push → registry (reverse mapping)

```bash
echo '# extra content' >> .cursor/rules/no-auto-commit.mdc
skillsync push

cat ~/.skillsync/registry/rules/no-auto-commit/no-auto-commit.md | grep 'extra content'
# Should print: # extra content

# Cleanup
sed -i 's/# extra content//' ~/.skillsync/registry/rules/no-auto-commit/no-auto-commit.md
skillsync pull
```

---

## Agent: claude → cursor field mapping

`read_only` is renamed to `readonly` when syncing to cursor format.

```bash
# Inspect an agent in the registry (claude format)
cat ~/.skillsync/registry/agents/<agent-name>/<agent-name>.md | head -5
# Expected frontmatter: read_only (claude field name)

skillsync pull
cat .cursor/agents/<agent-name>.md | head -5
# Expected frontmatter: readonly (cursor field name)
```

### Agent: cursor push → registry (reverse mapping)

```bash
echo '# marker' >> .cursor/agents/<agent-name>.md
skillsync push

cat ~/.skillsync/registry/agents/<agent-name>/<agent-name>.md | grep 'marker'
# Should print: # marker

# Cleanup
sed -i 's/# marker//' ~/.skillsync/registry/agents/<agent-name>/<agent-name>.md
skillsync pull
```

---

## Command: cursor push → registry

Commands have no frontmatter — content is copied unchanged in both directions.

```bash
echo '# marker' >> .cursor/commands/<command-name>.md
skillsync push

cat ~/.skillsync/registry/commands/<command-name>/<command-name>.md | grep 'marker'
# Should print: # marker

# Cleanup
sed -i 's/# marker//' ~/.skillsync/registry/commands/<command-name>/<command-name>.md
skillsync pull
```

---
