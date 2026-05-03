# Sync state tracked in a lockfile, not inferred from paths

`sync` writes `.skillsync/sync.lock` mapping every synced project file to its source bundle and a content hash. We chose this over inferring the bundle from the target path (reversing the kind→path map) because inference breaks when files are moved or renamed, and because we need the last-synced hash anyway for conflict detection (M7). One file serves both `push` (reverse-transform source lookup) and `sync` (local-edit detection).
