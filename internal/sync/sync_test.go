package sync_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rolfwessels/skillsync/internal/config"
	"github.com/rolfwessels/skillsync/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const registryPath = "../../testdata/registry"

func sampleConfig(formats, bundles []string) config.ProjectConfig {
	return config.ProjectConfig{
		Registry: registryPath,
		Formats:  formats,
		Bundles:  bundles,
	}
}

func TestRun_SkillLandsAtCorrectPath(t *testing.T) {
	// arrange
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"skills/tdd"})
	var warn bytes.Buffer

	// act
	err := sync.Run(projectRoot, registryPath, cfg, &warn)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "skills", "tdd", "SKILL.md"))
}

func TestRun_SkillWithNestedFile_preservesSubdirInTarget(t *testing.T) {
	// arrange
	regDir := t.TempDir()
	bundleDir := filepath.Join(regDir, "skills", "scripted")
	scriptsDir := filepath.Join(bundleDir, "scripts")
	require.NoError(t, os.MkdirAll(scriptsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(bundleDir, "bundle.toml"), []byte("description = \"d\"\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(bundleDir, "SKILL.md"), []byte("skill"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(scriptsDir, "run.sh"), []byte("#!/bin/sh"), 0644))

	projectRoot := t.TempDir()
	cfg := config.ProjectConfig{Registry: regDir, Formats: []string{"cursor"}, Bundles: []string{"skills/scripted"}}
	var warn bytes.Buffer

	// act
	err := sync.Run(projectRoot, regDir, cfg, &warn)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "skills", "scripted", "SKILL.md"))
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "skills", "scripted", "scripts", "run.sh"))
}

func TestRun_RuleLandsAtCorrectPath(t *testing.T) {
	// arrange
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"rules/naming-convention"})
	var warn bytes.Buffer

	// act
	err := sync.Run(projectRoot, registryPath, cfg, &warn)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "rules", "naming-convention.mdc"))
}

func TestRun_RuleDualFormat_ClaudeIdentityAndCursorTransform(t *testing.T) {
	// arrange
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"claude", "cursor"}, []string{"rules/naming-convention"})
	var warn bytes.Buffer

	// act
	err := sync.Run(projectRoot, registryPath, cfg, &warn)

	// assert
	require.NoError(t, err)
	claudePath := filepath.Join(projectRoot, ".claude", "rules", "naming-convention.md")
	claudeB, err := os.ReadFile(claudePath)
	require.NoError(t, err)
	assert.Contains(t, string(claudeB), "alwaysApply:")
	assert.Contains(t, string(claudeB), "paths:")
	assert.NotContains(t, string(claudeB), "always_apply:")
	cursorB, err := os.ReadFile(filepath.Join(projectRoot, ".cursor", "rules", "naming-convention.mdc"))
	require.NoError(t, err)
	cs := string(cursorB)
	assert.Contains(t, cs, "alwaysApply:")
	assert.Contains(t, cs, "custom_rule_field:")
	assert.NotContains(t, cs, "always_apply:")
}

func TestRun_AgentCursor_HasMappedFrontmatter(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"agents/code-reviewer"})
	var warn bytes.Buffer
	require.NoError(t, sync.Run(projectRoot, registryPath, cfg, &warn))
	b, err := os.ReadFile(filepath.Join(projectRoot, ".cursor", "agents", "code-reviewer.md"))
	require.NoError(t, err)
	s := string(b)
	assert.Contains(t, s, "readonly:")
	assert.Contains(t, s, "vendor_agent_field:")
	assert.NotContains(t, s, "read_only:")
}

func TestRun_MultipleFormats(t *testing.T) {
	// arrange
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"claude", "cursor"}, []string{"skills/tdd"})
	var warn bytes.Buffer

	// act
	err := sync.Run(projectRoot, registryPath, cfg, &warn)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "skills", "tdd", "SKILL.md"))
	assert.FileExists(t, filepath.Join(projectRoot, ".claude", "skills", "tdd", "SKILL.md"))
}

func TestRun_AllKindFormatPaths(t *testing.T) {
	tests := []struct {
		name       string
		bundle     string
		format     string
		wantTarget string
	}{
		{"skill cursor", "skills/tdd", "cursor", ".cursor/skills/tdd/SKILL.md"},
		{"skill claude", "skills/tdd", "claude", ".claude/skills/tdd/SKILL.md"},
		{"rule cursor", "rules/naming-convention", "cursor", ".cursor/rules/naming-convention.mdc"},
		{"rule claude", "rules/naming-convention", "claude", ".claude/rules/naming-convention.md"},
		{"agent cursor", "agents/code-reviewer", "cursor", ".cursor/agents/code-reviewer.md"},
		{"agent claude", "agents/code-reviewer", "claude", ".claude/agents/code-reviewer.md"},
		{"command cursor", "commands/summarise", "cursor", ".cursor/commands/summarise.md"},
		{"command claude", "commands/summarise", "claude", ".claude/commands/summarise.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			projectRoot := t.TempDir()
			cfg := sampleConfig([]string{tt.format}, []string{tt.bundle})
			var warn bytes.Buffer

			// act
			err := sync.Run(projectRoot, registryPath, cfg, &warn)

			// assert
			require.NoError(t, err)
			assert.FileExists(t, filepath.Join(projectRoot, filepath.FromSlash(tt.wantTarget)))
		})
	}
}

func TestRun_LockfileWritten(t *testing.T) {
	// arrange
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"skills/tdd", "rules/naming-convention"})
	var warn bytes.Buffer

	// act
	err := sync.Run(projectRoot, registryPath, cfg, &warn)

	// assert
	require.NoError(t, err)
	lockPath := filepath.Join(projectRoot, ".skillsync", "sync.lock")
	assert.FileExists(t, lockPath)
	data, err := os.ReadFile(lockPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "skills/tdd")
	assert.Contains(t, content, "rules/naming-convention")
	assert.Contains(t, content, "cursor")
}

func TestRun_Idempotent(t *testing.T) {
	// arrange
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"skills/tdd"})
	var warn bytes.Buffer

	require.NoError(t, sync.Run(projectRoot, registryPath, cfg, &warn))

	targetPath := filepath.Join(projectRoot, ".cursor", "skills", "tdd", "SKILL.md")
	info1, err := os.Stat(targetPath)
	require.NoError(t, err)

	// act — second run
	require.NoError(t, sync.Run(projectRoot, registryPath, cfg, &warn))

	// assert — file not rewritten
	info2, err := os.Stat(targetPath)
	require.NoError(t, err)
	assert.Equal(t, info1.ModTime(), info2.ModTime(), "file should not be rewritten on second sync")
}

func TestRun_MissingBundleWarnsAndContinues(t *testing.T) {
	// arrange
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"skills/nonexistent", "skills/tdd"})
	var warn bytes.Buffer

	// act
	err := sync.Run(projectRoot, registryPath, cfg, &warn)

	// assert
	require.NoError(t, err)
	assert.True(t, strings.Contains(warn.String(), "nonexistent"), "warn output should mention missing bundle")
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "skills", "tdd", "SKILL.md"), "valid bundle still synced")
}

func TestRun_LocalEditStashedBeforeOverwrite(t *testing.T) {
	// arrange
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"skills/tdd"})
	var warn bytes.Buffer
	require.NoError(t, sync.Run(projectRoot, registryPath, cfg, &warn))

	target := filepath.Join(projectRoot, ".cursor", "skills", "tdd", "SKILL.md")
	orig := mustReadFile(t, target)
	require.NoError(t, os.WriteFile(target, append(orig, []byte("\nLOCAL EDIT\n")...), 0644))

	wantBody, err := os.ReadFile(filepath.Join(registryPath, "skills", "tdd", "SKILL.md"))
	require.NoError(t, err)

	// act
	warn.Reset()
	err = sync.Run(projectRoot, registryPath, cfg, &warn)

	// assert
	require.NoError(t, err)
	gotBody, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, string(wantBody), string(gotBody), "synced file should match registry after stash+overwrite")

	ws := warn.String()
	assert.Contains(t, ws, ".cursor/skills/tdd/SKILL.md")
	assert.Contains(t, ws, ".skillsync/stash/")

	matches, err := filepath.Glob(filepath.Join(projectRoot, ".skillsync", "stash", "*", ".cursor", "skills", "tdd", "SKILL.md"))
	require.NoError(t, err)
	require.Len(t, matches, 1, "exactly one stashed copy with preserved path")
	stashed, err := os.ReadFile(matches[0])
	require.NoError(t, err)
	assert.Contains(t, string(stashed), "LOCAL EDIT")
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return b
}

func TestRun_RegistryFileRemoved_DeletesProjectFileAndDropsLockEntry(t *testing.T) {
	// arrange
	regDir := copyRegistryToTemp(t)
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"skills/scripted"})
	var warn bytes.Buffer
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	projectFile := filepath.Join(projectRoot, ".cursor", "skills", "scripted", "scripts", "helper.sh")
	require.FileExists(t, projectFile)
	require.NoError(t, os.Remove(filepath.Join(regDir, "skills", "scripted", "scripts", "helper.sh")))

	// act
	warn.Reset()
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	// assert
	_, err := os.Stat(projectFile)
	assert.True(t, os.IsNotExist(err), "project file should be deleted, got err=%v", err)

	entries, err := sync.LoadLockEntries(projectRoot)
	require.NoError(t, err)
	for _, e := range entries {
		assert.NotEqual(t, ".cursor/skills/scripted/scripts/helper.sh", e.Target,
			"lockfile entry for removed registry file should be dropped")
	}
}

func TestRun_RegistryFileRemoved_LocallyModified_StashesAndDeletes(t *testing.T) {
	// arrange
	regDir := copyRegistryToTemp(t)
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"skills/scripted"})
	var warn bytes.Buffer
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	projectFile := filepath.Join(projectRoot, ".cursor", "skills", "scripted", "scripts", "helper.sh")
	require.NoError(t, os.WriteFile(projectFile, []byte("LOCAL\n"), 0644))
	require.NoError(t, os.Remove(filepath.Join(regDir, "skills", "scripted", "scripts", "helper.sh")))

	// act
	warn.Reset()
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	// assert
	_, err := os.Stat(projectFile)
	assert.True(t, os.IsNotExist(err), "project file should be removed after stash, got err=%v", err)

	matches, err := filepath.Glob(filepath.Join(projectRoot, ".skillsync", "stash", "*", ".cursor", "skills", "scripted", "scripts", "helper.sh"))
	require.NoError(t, err)
	require.Len(t, matches, 1, "exactly one stashed copy of the locally modified file")
	stashed, err := os.ReadFile(matches[0])
	require.NoError(t, err)
	assert.Equal(t, "LOCAL\n", string(stashed))

	entries, err := sync.LoadLockEntries(projectRoot)
	require.NoError(t, err)
	for _, e := range entries {
		assert.NotEqual(t, ".cursor/skills/scripted/scripts/helper.sh", e.Target)
	}
}

func TestRun_RegistryFileRemoved_ProjectFileAlreadyAbsent_DropsLockEntrySilently(t *testing.T) {
	// arrange
	regDir := copyRegistryToTemp(t)
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"skills/scripted"})
	var warn bytes.Buffer
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	projectFile := filepath.Join(projectRoot, ".cursor", "skills", "scripted", "scripts", "helper.sh")
	require.NoError(t, os.Remove(projectFile))
	require.NoError(t, os.Remove(filepath.Join(regDir, "skills", "scripted", "scripts", "helper.sh")))

	// act
	warn.Reset()
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	// assert
	assert.Empty(t, warn.String(), "no warning expected for already-absent project file")
	entries, err := sync.LoadLockEntries(projectRoot)
	require.NoError(t, err)
	for _, e := range entries {
		assert.NotEqual(t, ".cursor/skills/scripted/scripts/helper.sh", e.Target)
	}
}

func TestRun_RegistryFileRemoved_DoesNotTouchOtherBundlesLockEntries(t *testing.T) {
	// arrange
	regDir := copyRegistryToTemp(t)
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"skills/scripted", "skills/tdd"})
	var warn bytes.Buffer
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	require.NoError(t, os.Remove(filepath.Join(regDir, "skills", "scripted", "scripts", "helper.sh")))

	// act
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	// assert — tdd entries untouched
	entries, err := sync.LoadLockEntries(projectRoot)
	require.NoError(t, err)
	var tddCount int
	for _, e := range entries {
		if e.Bundle == "skills/tdd" {
			tddCount++
		}
	}
	assert.Greater(t, tddCount, 0, "skills/tdd lock entries must remain")
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "skills", "tdd", "SKILL.md"))
}
