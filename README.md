# 🌐 skillsync

[![GitHub release](https://img.shields.io/github/v/release/rolfwessels/skillsync)](https://github.com/rolfwessels/skillsync/releases)
[![Go CI](https://github.com/rolfwessels/skillsync/actions/workflows/github-action.yml/badge.svg)](https://github.com/rolfwessels/skillsync/actions)

Sync AI assistant bundles — skills, rules, agents, slash commands — across projects, machines, and tools from a single registry.

## 😩 Why skillsync?

- You tweak a skill in `~/.claude/` on WSL — your Windows host still has the old copy.
- You build a slash command for a Claude project, then open the next one in Cursor and want the same thing, in Cursor's format.
- Copy-pasting bundles between machines and tools gets old fast.

## ✨ How it works

Point every project at one **registry** (a local folder or git repo). Bundles are stored once in canonical Claude format; `pull` transforms them on the way out into whatever the project's tool expects (`.claude/` or `.cursor/`). Edit locally, `push` back, and every other project picks it up on the next sync.

```bash
skillsync init    # set up this project (and optionally wire a git hook)
skillsync pull    # fetch from registry → ./.claude/ or ./.cursor/
skillsync push    # send local edits back to the registry
skillsync sync    # both directions, with conflict detection
```

Opt into the git pre-commit hook during `init` and every commit keeps your registry in lockstep.

## ⚡ Install

**Linux / macOS** — installs to `~/.local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/rolfwessels/skillsync/main/install.sh | sh
```

**Windows (PowerShell)** — installs to `%LOCALAPPDATA%\Programs\skillsync\` and adds it to your user PATH:

```powershell
irm https://raw.githubusercontent.com/rolfwessels/skillsync/main/install.ps1 | iex
```

Want a different location? Set `INSTALL_DIR` first:

```bash
INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/rolfwessels/skillsync/main/install.sh | sh
```

```powershell
$env:INSTALL_DIR = 'C:\tools\skillsync'; irm https://raw.githubusercontent.com/rolfwessels/skillsync/main/install.ps1 | iex
```

To upgrade, just re-run the same command. Both scripts pull the binary from the [latest GitHub release](https://github.com/rolfwessels/skillsync/releases/latest), which is published on every push to `main` under a new version tag.

Prefer to download by hand? Grab the right archive for your platform from the [releases page](https://github.com/rolfwessels/skillsync/releases), extract, and put the binary somewhere on your PATH.

## 📦 Technology

- [Cobra](https://github.com/spf13/cobra) for the CLI
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Huh](https://github.com/charmbracelet/huh) for the TUI
- [BurntSushi/toml](https://github.com/BurntSushi/toml) for config parsing
- Docker for the dev environment
- MakeFile because it just works!

## 🚀 Getting started with development

This project ships with a development container that has all the tooling required to build, test, and publish.

```bash
# bring up dev environment
make build up

# test the project
make test

# run the CLI
make start

# build release binaries for all platforms
make publish
```

To build and push a Docker image:

```bash
make docker-build docker-login docker-push
# or just
make docker-publish
```

## 🛠 Prerequisites

- [Docker](https://docs.docker.com/get-docker/) — for the dev container
- [Git](https://git-scm.com/) — for version control
- `make` — available via WSL on Windows, or natively on macOS/Linux

## 📋 Available make commands

### 💻 Commands outside the container

| Command      | Description                                          |
|--------------|------------------------------------------------------|
| `make up`    | Bring up the container & attach to the dev shell     |
| `make down`  | Stop the container                                   |
| `make build` | Rebuild the container                                |

### 🐳 Commands to run inside the container

| Command                   | Description                                    |
|---------------------------|------------------------------------------------|
| `make version`            | Show the current version                       |
| `make start`              | Run skillsync                                  |
| `make test`               | Run tests                                      |
| `make publish`            | Build release archives for 5 platforms         |
| `make docker-login`       | Login to Docker registry                       |
| `make docker-build`       | Build the production Docker image              |
| `make docker-push`        | Push the Docker image                          |
| `make docker-pull-short-tag` | Pull image by git short hash               |
| `make docker-tag-env`     | Tag image for an environment                   |
| `make docker-publish`     | Full build + push workflow                     |
| `make deploy`             | Deploy skillsync                               |
| `make update-packages`    | Update Go dependencies to latest               |

## 💻 Development

### Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: Backward-compatible new functionality
- **PATCH**: Backward-compatible bug fixes

`MAJOR` and `MINOR` are set manually via `versionPrefix` in the `Makefile`. `PATCH` is automatically derived from commit count.

```bash
make version
```

### Development Workflow

Feature branches are created off `main` with the prefix `feature/` or `bug/`.
PR builds attach archives as workflow artifacts. Merging to `main` publishes them as a new versioned [GitHub release](https://github.com/rolfwessels/skillsync/releases).

## FAQ

**Can I use this on Windows/macOS/Linux?**  
Yes — binaries are published for all three platforms.

**How do I update to the latest version?**  
Re-run the install command (`install.sh` or `install.ps1`). It overwrites the binary in place from the latest release.

## Research

- [Cobra CLI framework](https://cobra.dev/)
- [What is a Makefile?](https://opensource.com/article/18/8/what-how-makefile)
