# skillsync

A CLI tool that syncs AI assistant bundles (skills, rules, coding standards, agents) from a central registry into developer projects.

## Language

**Bundle**:
A named unit in the registry — one `bundle.toml` metadata file plus the content files (skill, rule, coding standard, agent, etc.) it describes. Identified by `<kind>/<name>`.
_Avoid_: package, plugin, module

**Kind**:
The category of a bundle, used as the top-level folder in the registry (e.g. `skills/`, `rules/`, `agents/`). Determines where the bundle gets synced on the target machine.
_Avoid_: type, category

**Registry**:
A folder (local path or git URL) containing bundles organised as `<kind>/<name>/`. The source of truth for all bundles.
_Avoid_: store, repo (use "git repo" only when referring to the underlying VCS)

**Project**:
A developer's working directory that has a `.skillsync/config.toml` declaring which bundles it wants.
_Avoid_: workspace, target

**Tag**:
A label on a bundle (declared in `bundle.toml`) used for discovery and grouping in the TUI. Not used in project config for v1 — config references bundles by `kind/name` only.
_Avoid_: label, category

**Format**:
The AI assistant platform a project targets (e.g. `claude`, `cursor`). Determines the target path and any file transforms applied during sync. A project can declare multiple formats.
_Avoid_: platform, tool, AI

**Canonical format**:
The storage format used in the registry — always `claude`. All bundles are authored and stored as Claude bundles. Other formats are produced on-the-fly during sync via transforms.
_Avoid_: source format, native format

**Transform**:
A function that converts a bundle from canonical (claude) format to a target format. Maps known fields; unknown frontmatter fields are copied as-is. Platforms ignore unrecognised fields, so no passthrough mechanism needed.
_Avoid_: converter, adapter

**Sync**:
The operation that reads bundles from the registry, applies the appropriate transform per declared format, and writes files to the correct `(kind, format)` target path in the project. Idempotent.
_Avoid_: install, deploy, copy

**Lockfile**:
`.skillsync/sync.lock` — written by `sync`, tracks each synced file's source bundle, format, last-synced content hash, registry hash, local hash, and sync state (`clean` | `local_modified` | `registry_modified` | `conflict`). Used by `push` to find the originating bundle and by `sync` for conflict detection.
_Avoid_: manifest (use "lockfile"), state file

## Relationships

- A **Registry** contains many **Bundles**
- A **Bundle** belongs to exactly one **Kind**
- A **Bundle** carries zero or more **Tags**
- A **Project** references one or more **Bundles** (by `kind/name` string in config)
- A **Sync** reads the **Registry** and writes bundle files into the **Project**

## Example dialogue

> **Dev:** "I want every project to get our naming-convention rule and the tdd skill."
> **Domain expert:** "Tag them both `standard` in the registry. Then each project config just says `tags = ["standard"]` and sync handles the rest."

**bundle.toml**:
The metadata file at the root of every bundle folder. Contains `description` and `tags` only. Name and kind are derived from the folder path (`<kind>/<name>/`). All files under the bundle folder (excluding `bundle.toml` itself), including nested subdirectories, are synced — preserving their relative path (e.g. `scripts/helper.sh` lands at `<kind-target>/<name>/scripts/helper.sh`).
_Avoid_: manifest, config (use "config" only for the project-side `.skillsync/config.toml`)

**Registry shape**:
Every bundle is a folder (`<kind>/<name>/`), even single-file kinds like `rule`, `agent`, and `command`. Uniform structure — no special cases in the registry walker.

**Kind→path map**:
The hardcoded table mapping `(kind, format)` to the target path in a project:
- `(skill, claude)` → `.claude/skills/<name>/`
- `(skill, cursor)` → `.cursor/skills/<name>/`
- `(rule, claude)` → `.claude/rules/<name>.md`
- `(rule, cursor)` → `.cursor/rules/<name>.mdc`
- `(agent, claude)` → `.claude/agents/<name>.md`
- `(agent, cursor)` → `.cursor/agents/<name>.md`
- `(command, claude)` → `.claude/commands/<name>.md`
- `(command, cursor)` → `.cursor/commands/<name>.md`

## Flagged ambiguities

- "bundle" was initially unclear whether it meant the registry artifact or the config entry — resolved: **Bundle** is the registry artifact; the config stores a `kind/name` string reference to it (no separate type needed for v1).
