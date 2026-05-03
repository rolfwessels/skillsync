package integration_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/rolfwessels/skillsync/internal/config"
	skillsinit "github.com/rolfwessels/skillsync/internal/init"
	"github.com/rolfwessels/skillsync/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRule  = "rules/naming-convention"
	testSkill = "skills/tdd"
)

func multiFormatRegistry(t *testing.T) string {
	t.Helper()
	src := filepath.Join("..", "..", "testdata", "registry")
	dst := t.TempDir()
	require.NoError(t, copyTree(src, dst))
	return dst
}

func multiFormatConfig(regDir string) config.ProjectConfig {
	return config.ProjectConfig{
		Registry: regDir,
		Formats:  []string{"cursor", "claude"},
		Bundles:  []string{testRule, testSkill},
	}
}

// Scenario A: init with formats=["cursor","claude"] writes config with both formats and no error.
func TestMultiFormat_Init_WritesConfigWithBothFormats(t *testing.T) {
	// arrange
	regDir := multiFormatRegistry(t)
	projectRoot := t.TempDir()
	stub := &stubPrompter{cfg: multiFormatConfig(regDir)}

	// act
	err := skillsinit.Run(projectRoot, stub, io.Discard)

	// assert
	require.NoError(t, err)
	loaded, err := config.Load(projectRoot)
	require.NoError(t, err)
	assert.Equal(t, []string{"cursor", "claude"}, loaded.Formats)
	assert.Contains(t, loaded.Bundles, testRule)
	assert.Contains(t, loaded.Bundles, testSkill)
}

// Scenario B: pull writes files for both formats; lockfile has one entry per (bundle, format) pair.
func TestMultiFormat_Pull_WritesBothFormatsAndPopulatesLockfile(t *testing.T) {
	// arrange
	regDir := multiFormatRegistry(t)
	projectRoot := t.TempDir()
	cfg := multiFormatConfig(regDir)

	// act
	err := sync.Run(projectRoot, regDir, cfg, io.Discard)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(projectRoot, ".claude", "rules", "naming-convention.md"))
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "rules", "naming-convention.mdc"))
	assert.FileExists(t, filepath.Join(projectRoot, ".claude", "skills", "tdd", "SKILL.md"))
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "skills", "tdd", "SKILL.md"))

	entries := readLockEntries(t, projectRoot)
	assertLockEntry(t, entries, testRule, "claude")
	assertLockEntry(t, entries, testRule, "cursor")
	assertLockEntry(t, entries, testSkill, "claude")
	assertLockEntry(t, entries, testSkill, "cursor")
}

// Scenario C: update a bundle in the registry, re-pull; both format outputs reflect new content;
// lockfile hashes are updated.
func TestMultiFormat_Pull_AfterRegistryUpdate_BothFormatsAndHashesUpdated(t *testing.T) {
	// arrange
	regDir := multiFormatRegistry(t)
	projectRoot := t.TempDir()
	cfg := multiFormatConfig(regDir)
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, io.Discard))

	entriesBefore := readLockEntries(t, projectRoot)
	hashBefore := lockEntryHash(entriesBefore, testRule, "claude")
	require.NotEmpty(t, hashBefore)

	regRuleFile := filepath.Join(regDir, "rules", "naming-convention", "rule.md")
	orig := mustReadFile(t, regRuleFile)
	require.NoError(t, os.WriteFile(regRuleFile, append(orig, []byte("\n## Registry Update\nNew content.\n")...), 0644))

	// act
	err := sync.Run(projectRoot, regDir, cfg, io.Discard)

	// assert
	require.NoError(t, err)

	claudeRule := string(mustReadFile(t, filepath.Join(projectRoot, ".claude", "rules", "naming-convention.md")))
	assert.Contains(t, claudeRule, "Registry Update")

	cursorRule := string(mustReadFile(t, filepath.Join(projectRoot, ".cursor", "rules", "naming-convention.mdc")))
	assert.Contains(t, cursorRule, "Registry Update")

	entriesAfter := readLockEntries(t, projectRoot)
	assert.NotEqual(t, hashBefore, lockEntryHash(entriesAfter, testRule, "claude"), "claude lock hash should update after registry change")
	assert.NotEqual(t, hashBefore, lockEntryHash(entriesAfter, testRule, "cursor"), "cursor lock hash should update after registry change")
}

// Scenario D: edit the cursor-format output of a bundle, push; change reflected in registry;
// next pull produces consistent claude-format output.
func TestMultiFormat_Push_CursorEdit_ReflectedInRegistryAndConsistentOnNextPull(t *testing.T) {
	// arrange
	regDir := multiFormatRegistry(t)
	projectRoot := t.TempDir()
	cfg := multiFormatConfig(regDir)
	require.NoError(t, config.Write(projectRoot, cfg))
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, io.Discard))

	cursorRulePath := filepath.Join(projectRoot, ".cursor", "rules", "naming-convention.mdc")
	orig := mustReadFile(t, cursorRulePath)
	require.NoError(t, os.WriteFile(cursorRulePath, append(orig, []byte("\n## Pushed Section\n")...), 0644))

	// act
	err := sync.Push(projectRoot, regDir, cfg)

	// assert — registry has Claude canonical format with the edit
	require.NoError(t, err)
	regRule := string(mustReadFile(t, filepath.Join(regDir, "rules", "naming-convention", "rule.md")))
	assert.Contains(t, regRule, "Pushed Section")
	assert.Contains(t, regRule, "alwaysApply:")
	assert.Contains(t, regRule, "paths:")
	assert.NotContains(t, regRule, "always_apply:")

	// fresh pull into new project — both formats contain the pushed change
	freshProject := t.TempDir()
	require.NoError(t, sync.Run(freshProject, regDir, cfg, io.Discard))

	claudeRule := string(mustReadFile(t, filepath.Join(freshProject, ".claude", "rules", "naming-convention.md")))
	assert.Contains(t, claudeRule, "Pushed Section")

	cursorRule := string(mustReadFile(t, filepath.Join(freshProject, ".cursor", "rules", "naming-convention.mdc")))
	assert.Contains(t, cursorRule, "Pushed Section")
}

// --- helpers ---

type stubPrompter struct {
	cfg config.ProjectConfig
}

func (s *stubPrompter) Ask(_ string, _ config.ProjectConfig) (skillsinit.PromptResult, error) {
	return skillsinit.PromptResult{Config: s.cfg}, nil
}

type lockEntry struct {
	Bundle string `toml:"bundle"`
	Format string `toml:"format"`
	Target string `toml:"target"`
	Hash   string `toml:"hash"`
}

func readLockEntries(t *testing.T, projectRoot string) []lockEntry {
	t.Helper()
	var lf struct {
		Entries []lockEntry `toml:"entries"`
	}
	_, err := toml.DecodeFile(filepath.Join(projectRoot, ".skillsync", "sync.lock"), &lf)
	require.NoError(t, err)
	return lf.Entries
}

func assertLockEntry(t *testing.T, entries []lockEntry, bundle, format string) {
	t.Helper()
	for _, e := range entries {
		if e.Bundle == bundle && e.Format == format {
			return
		}
	}
	t.Errorf("lockfile missing entry for bundle=%s format=%s", bundle, format)
}

func lockEntryHash(entries []lockEntry, bundle, format string) string {
	for _, e := range entries {
		if e.Bundle == bundle && e.Format == format {
			return e.Hash
		}
	}
	return ""
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return b
}

func copyTree(srcRoot, dstRoot string) error {
	return filepath.WalkDir(srcRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return fmt.Errorf("rel path: %w", err)
		}
		if rel == "." {
			return os.MkdirAll(dstRoot, 0755)
		}
		dst := filepath.Join(dstRoot, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		out, err := os.Create(dst)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return err
		}
		return out.Close()
	})
}
