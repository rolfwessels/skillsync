package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const lockFile = ".skillsync/sync.lock"

type LockEntry struct {
	Bundle       string `toml:"bundle"`
	Format       string `toml:"format"`
	Target       string `toml:"target"`
	Hash         string `toml:"hash"`
	RegistryHash string `toml:"registry_hash,omitempty"`
	LocalHash    string `toml:"local_hash,omitempty"`
	State        State  `toml:"state,omitempty"`
}

type lockfile struct {
	Entries []LockEntry `toml:"entries"`
}

func LoadLockEntries(projectRoot string) ([]LockEntry, error) {
	lf, err := loadLock(projectRoot)
	if err != nil {
		return nil, err
	}
	return lf.Entries, nil
}

func loadLock(projectRoot string) (lockfile, error) {
	path := filepath.Join(projectRoot, lockFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return lockfile{}, nil
	}
	if err != nil {
		return lockfile{}, fmt.Errorf("reading lockfile: %w", err)
	}
	var lf lockfile
	if err := toml.Unmarshal(data, &lf); err != nil {
		return lockfile{}, fmt.Errorf("parsing lockfile: %w", err)
	}
	return lf, nil
}

func saveLock(projectRoot string, lf lockfile) error {
	path := filepath.Join(projectRoot, lockFile)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating lockfile dir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating lockfile: %w", err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(lf); err != nil {
		return fmt.Errorf("encoding lockfile: %w", err)
	}
	return nil
}

func (lf lockfile) lookup(bundle, format, target string) (LockEntry, bool) {
	for _, e := range lf.Entries {
		if e.Bundle == bundle && e.Format == format && e.Target == target {
			return e, true
		}
	}
	return LockEntry{}, false
}

func (lf *lockfile) upsert(entry LockEntry) {
	for i, e := range lf.Entries {
		if e.Bundle == entry.Bundle && e.Format == entry.Format && e.Target == entry.Target {
			lf.Entries[i] = entry
			return
		}
	}
	lf.Entries = append(lf.Entries, entry)
}

func (lf *lockfile) upsertState(bundleRef, format, target string, state State) {
	for i, e := range lf.Entries {
		if e.Bundle == bundleRef && e.Format == format && e.Target == target {
			lf.Entries[i].State = state
			return
		}
	}
}
