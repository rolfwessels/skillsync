# skillsync — Build Plan

Milestones sized so each one ends with something you can actually run.

---

## Milestone 0 — Scaffold (½ day)

**Goal:** `skillsync --help` runs.

- `go mod init github.com/<you>/skillsync`
- Repo layout:
  ```
  cmd/skillsync/main.go
  internal/cli/        # cobra commands
  internal/config/     # load/save .skillsync/config.toml
  internal/bundle/     # bundle.toml, registry walking
  internal/sync/       # the copy engine
  internal/git/        # exec wrapper around `git`
  internal/transform/  # transformer registry (stubs for v1)
  internal/hooks/      # install/uninstall git hooks
  testdata/            # fake registry + fake project for tests
  ```
- `main.go` wires cobra root command with version flag
- GitHub Actions: lint (`golangci-lint`) + test on push
- `goreleaser.yaml` skeleton (don\'t publish yet, just `goreleaser release --snapshot` works)

**Done when:** `go build && ./skillsync --help` prints subcommand stubs.

---

## Milestone 1 — Read-only registry (1 day)

**Goal:** point at a local folder, list what\'s there.

- Define `Bundle` struct + `bundle.toml` parser
- Walk `registry/<kind>/<name>/` discovering bundles
- New command: `skillsync list` — prints discovered bundles grouped by kind
- Test against `testdata/registry/`

**Done when:** `skillsync list --registry ./testdata/registry` shows your fake skills/commands/agents with descriptions and tags.

Pure read-and-print. No config, no project, no git. Builds your mental model of the data.

---

## Milestone 2 — Sync engine, no transforms (1 day)

**Goal:** copy bundles from registry into a project.

- Define project `config.toml` schema + loader
- Resolve config location: `.skillsync/config.toml` or `.git/.skillsync/config.toml`, error if both present
- New command: `skillsync sync` — for each bundle in config, copy `registry/<kind>/<name>/*` to the hardcoded target path for that kind
- Hardcoded kind→path map (start dumb, e.g. `skill → .claude/skills/<name>/`)
- No-op cleanly if `[bundles]` is empty

**Done when:** you have a fake project with a config, you run `skillsync sync`, and the right files appear in the right places. Re-running is idempotent.

---

## Milestone 3 — `init` with TUI (1–2 days)

**Goal:** greenfield project setup feels nice.

- New command: `skillsync init`
- Prompt for registry path/URL (default `~/.skillsync/registry`)
- Prompt for format(s) — multi-select via `huh`
- Multi-select bundle picker grouped by kind, populated from registry walk
- Write `.skillsync/config.toml`
- Auto-create empty registry folder if path doesn\'t exist
- Then: call `sync`

**Done when:** running `skillsync init` in an empty folder lands you with a config and synced files, all from a TUI.

This is the one that sells the tool. Spend time on it.

---

## Milestone 4 — Git registry + caching (1 day)

**Goal:** registry can be a git URL.

- `internal/git/` wrapper: `Clone`, `Pull`, all via `os/exec` to system `git`
- On `sync`/`list`: if registry is a URL, clone to `~/.skillsync/cache/<hash>/` on first use, `git pull` on subsequent runs
- If registry is a local path, use directly

**Done when:** pointing config at a GitHub repo Just Works.

---

## Milestone 5 — Git hooks (½ day)

**Goal:** sync happens automatically.

- `internal/hooks/`: write `post-checkout` and `post-merge` shell scripts that call `skillsync sync`
- Installed automatically by `init`
- `skillsync hooks install` / `hooks uninstall` for manual control
- Detect and don\'t clobber existing hooks (append or warn)

**Done when:** `git pull` on a project with skillsync configured re-syncs without you thinking about it.

---

## Milestone 6 — `add` (½ day)

**Goal:** modify config without hand-editing TOML.

- `skillsync add <kind>/<name>` — append to config, run sync

**Done when:** `skillsync add skills/tdd` updates config and synced files appear.

---

## Milestone 7 — Conflict handling on down-sync (½ day)

**Goal:** local edits don\'t silently die.

- Before overwriting a file during `sync`, hash-compare. If changed locally, copy to `.skillsync/stash/<timestamp>/<path>` then overwrite
- Print warning summarising what was stashed

**Done when:** edit a synced file, run sync, see the warning, find your edit in the stash folder.

---

## Milestone 8 — Transforms, scaffolded (1–2 days)

**Goal:** the architecture exists, even if only one transformer is implemented.

- `internal/transform/` registry: map `(kind, sourceFormat, targetFormat)` → transformer function
- Identity transformer (claude → claude) as the default
- One real transformer: claude skill → cursor skill (probably near-identity given how similar they are)
- `x-passthrough.<format>` frontmatter preserved opaquely — split on `---`, keep frontmatter as a string, hand-roll a flat `key: value` reader for the few fields we actually need
- Transformer can return "skip + warn" for kinds with no target equivalent

**Done when:** a project with `format = ["cursor"]` syncs claude-canonical skills into cursor-shaped output.

---

## Milestone 9 — `push` (1 day)

**Goal:** edits made in a project\'s format flow back up.

- `skillsync push` — for each synced bundle, run reverse transformer, write into the registry working copy, `git add`/`commit`/`push`
- Off by default; `auto_push = true` in config enables auto on `sync`

**Done when:** edit a cursor-format file in a project, run push, see a commit land in the registry repo with the canonical (claude) form updated.

---

## Milestone 10 — Release (½ day)

**Goal:** other people can install it.

- Finish `goreleaser.yaml`: matrix for darwin/linux/windows × amd64/arm64
- Homebrew tap, Scoop bucket, GitHub Releases
- `README.md` with the install one-liner
- Tag `v0.1.0`, push, watch CI publish

**Done when:** `brew install <you>/tap/skillsync` works on a fresh Mac.

---

## Total: ~8–10 working days for v1

### Stack reminder

- Language: **Go**
- CLI: `cobra`
- TUI: `bubbletea` + `huh`
- Git: shell out to system `git` via `os/exec`
- TOML: `BurntSushi/toml`
- Frontmatter: opaque string + tiny hand-rolled flat parser
- Release: `goreleaser`

### Out of scope for v1

- Versioning (always HEAD)
- Free-text search (tags only)
- External transformer plugins
- Configurable kind→path mapping
- `promote` command
- Per-project overrides
