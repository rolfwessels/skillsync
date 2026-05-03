package integration_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/rolfwessels/skillsync/internal/config"
	"github.com/rolfwessels/skillsync/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Scenario S1: sync on a fresh project pulls all bundles (registry_modified → pull).
func TestSync_FreshProject_PullsAllBundles(t *testing.T) {
	// arrange
	regDir := multiFormatRegistry(t)
	projectRoot := t.TempDir()
	cfg := multiFormatConfig(regDir)

	// act
	err := sync.Sync(projectRoot, regDir, cfg, io.Discard, io.Discard)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(projectRoot, ".claude", "rules", "naming-convention.md"))
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "rules", "naming-convention.mdc"))
	assert.FileExists(t, filepath.Join(projectRoot, ".claude", "skills", "tdd", "SKILL.md"))
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "skills", "tdd", "SKILL.md"))
}

// Scenario S2: sync when registry changed and local unchanged → fast-forward pull.
func TestSync_RegistryChanged_LocalUnchanged_Pulls(t *testing.T) {
	// arrange
	regDir := multiFormatRegistry(t)
	projectRoot := t.TempDir()
	cfg := multiFormatConfig(regDir)
	require.NoError(t, sync.Sync(projectRoot, regDir, cfg, io.Discard, io.Discard))

	regRuleFile := filepath.Join(regDir, "rules", "naming-convention", "rule.md")
	orig := mustReadFile(t, regRuleFile)
	require.NoError(t, os.WriteFile(regRuleFile, append(orig, []byte("\n## Registry Sync Update\n")...), 0644))

	// act
	err := sync.Sync(projectRoot, regDir, cfg, io.Discard, io.Discard)

	// assert
	require.NoError(t, err)
	claudeRule := string(mustReadFile(t, filepath.Join(projectRoot, ".claude", "rules", "naming-convention.md")))
	assert.Contains(t, claudeRule, "Registry Sync Update")
}

// Scenario S3: sync when local changed and registry unchanged → push to registry.
func TestSync_LocalChanged_RegistryUnchanged_Pushes(t *testing.T) {
	// arrange
	regDir := multiFormatRegistry(t)
	projectRoot := t.TempDir()
	cfg := multiFormatConfig(regDir)
	require.NoError(t, config.Write(projectRoot, cfg))
	require.NoError(t, sync.Sync(projectRoot, regDir, cfg, io.Discard, io.Discard))

	cursorRulePath := filepath.Join(projectRoot, ".cursor", "rules", "naming-convention.mdc")
	orig := mustReadFile(t, cursorRulePath)
	require.NoError(t, os.WriteFile(cursorRulePath, append(orig, []byte("\n## Local Sync Edit\n")...), 0644))

	// act
	err := sync.Sync(projectRoot, regDir, cfg, io.Discard, io.Discard)

	// assert
	require.NoError(t, err)
	regRule := string(mustReadFile(t, filepath.Join(regDir, "rules", "naming-convention", "rule.md")))
	assert.Contains(t, regRule, "Local Sync Edit")
}

// Scenario S4: sync when both registry and local changed → conflict; non-zero exit; neither side modified.
func TestSync_BothChanged_Conflict_NeitherSideModified(t *testing.T) {
	// arrange
	regDir := multiFormatRegistry(t)
	projectRoot := t.TempDir()
	cfg := multiFormatConfig(regDir)
	require.NoError(t, config.Write(projectRoot, cfg))
	require.NoError(t, sync.Sync(projectRoot, regDir, cfg, io.Discard, io.Discard))

	regRuleFile := filepath.Join(regDir, "rules", "naming-convention", "rule.md")
	origReg := mustReadFile(t, regRuleFile)
	require.NoError(t, os.WriteFile(regRuleFile, append(origReg, []byte("\n## Registry Side\n")...), 0644))

	cursorRulePath := filepath.Join(projectRoot, ".cursor", "rules", "naming-convention.mdc")
	origLocal := mustReadFile(t, cursorRulePath)
	require.NoError(t, os.WriteFile(cursorRulePath, append(origLocal, []byte("\n## Local Side\n")...), 0644))

	snapshotReg := mustReadFile(t, regRuleFile)
	snapshotLocal := mustReadFile(t, cursorRulePath)

	// act
	err := sync.Sync(projectRoot, regDir, cfg, io.Discard, io.Discard)

	// assert — error returned, neither side changed
	require.Error(t, err)
	assert.Equal(t, snapshotReg, mustReadFile(t, regRuleFile), "registry should not be modified on conflict")
	assert.Equal(t, snapshotLocal, mustReadFile(t, cursorRulePath), "local file should not be modified on conflict")
}

// Scenario S5: sync on project with old-style lockfile (missing registry_hash/local_hash) migrates without data loss.
func TestSync_OldLockfile_MigratesWithoutDataLoss(t *testing.T) {
	// arrange
	regDir := multiFormatRegistry(t)
	projectRoot := t.TempDir()
	cfg := multiFormatConfig(regDir)

	// Simulate old-style pull (no registry_hash/local_hash in lockfile)
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, io.Discard))
	stripNewLockFields(t, projectRoot)

	localContent := mustReadFile(t, filepath.Join(projectRoot, ".cursor", "rules", "naming-convention.mdc"))

	// act — sync should not stash or overwrite unchanged files
	err := sync.Sync(projectRoot, regDir, cfg, io.Discard, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, localContent, mustReadFile(t, filepath.Join(projectRoot, ".cursor", "rules", "naming-convention.mdc")))
}

// stripNewLockFields rewrites the lockfile, removing registry_hash/local_hash/state to simulate an old-format lockfile.
func stripNewLockFields(t *testing.T, projectRoot string) {
	t.Helper()
	type oldEntry struct {
		Bundle string `toml:"bundle"`
		Format string `toml:"format"`
		Target string `toml:"target"`
		Hash   string `toml:"hash"`
	}
	type oldLock struct {
		Entries []oldEntry `toml:"entries"`
	}

	entries := readLockEntries(t, projectRoot)
	old := oldLock{}
	for _, e := range entries {
		old.Entries = append(old.Entries, oldEntry(e))
	}

	lockPath := filepath.Join(projectRoot, ".skillsync", "sync.lock")
	f, err := os.Create(lockPath)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, toml.NewEncoder(f).Encode(old))
}
