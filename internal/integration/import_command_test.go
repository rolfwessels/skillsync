package integration_test

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/rolfwessels/skillsync/internal/config"
	"github.com/rolfwessels/skillsync/internal/importcmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bundleTOML struct {
	Description string `toml:"description"`
}

func TestImport_preExistingSkill_createsBundleAndLock(t *testing.T) {
	regDir := t.TempDir()
	projectRoot := t.TempDir()
	cfg := config.ProjectConfig{
		Registry: regDir,
		Formats:  []string{"cursor", "claude"},
		Bundles:  []string{},
	}
	require.NoError(t, config.Write(projectRoot, cfg))

	skillDir := filepath.Join(projectRoot, ".cursor", "skills", "myskill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Imported skill\n"), 0644))

	stub := &importcmd.StubPrompter{
		Indices: []int{0},
		Decisions: []importcmd.Decision{
			{LinkExisting: false, BundleRef: "skills/myskill"},
		},
	}
	err := importcmd.Run(projectRoot, cfg, regDir, stub, io.Discard)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(regDir, "skills", "myskill", "SKILL.md"))
	assert.FileExists(t, filepath.Join(regDir, "skills", "myskill", "bundle.toml"))

	loaded, err := config.Load(projectRoot)
	require.NoError(t, err)
	assert.Contains(t, loaded.Bundles, "skills/myskill")

	var lf struct {
		Entries []struct {
			Bundle string `toml:"bundle"`
			Format string `toml:"format"`
			Target string `toml:"target"`
		} `toml:"entries"`
	}
	_, err = toml.DecodeFile(filepath.Join(projectRoot, ".skillsync", "sync.lock"), &lf)
	require.NoError(t, err)

	var sawClaude, sawCursor bool
	for _, e := range lf.Entries {
		if e.Bundle != "skills/myskill" {
			continue
		}
		switch e.Format {
		case "claude":
			sawClaude = true
			assert.Equal(t, ".claude/skills/myskill/SKILL.md", e.Target)
		case "cursor":
			sawCursor = true
			assert.Equal(t, ".cursor/skills/myskill/SKILL.md", e.Target)
		}
	}
	assert.True(t, sawClaude, "expected claude lock entry")
	assert.True(t, sawCursor, "expected cursor lock entry")
}

func TestImport_skillWithFrontmatter_usesFrontmatterDescription(t *testing.T) {
	// arrange
	regDir := t.TempDir()
	projectRoot := t.TempDir()
	cfg := config.ProjectConfig{Registry: regDir, Formats: []string{"cursor"}, Bundles: []string{}}
	require.NoError(t, config.Write(projectRoot, cfg))

	skillContent := "---\nname: myskill\ndescription: My fancy skill description\n---\n\nDo the thing.\n"
	skillDir := filepath.Join(projectRoot, ".cursor", "skills", "myskill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644))

	stub := &importcmd.StubPrompter{
		Indices:   []int{0},
		Decisions: []importcmd.Decision{{LinkExisting: false, BundleRef: "skills/myskill"}},
	}

	// act
	err := importcmd.Run(projectRoot, cfg, regDir, stub, io.Discard)

	// assert
	require.NoError(t, err)
	var meta bundleTOML
	_, err = toml.DecodeFile(filepath.Join(regDir, "skills", "myskill", "bundle.toml"), &meta)
	require.NoError(t, err)
	assert.Equal(t, "My fancy skill description", meta.Description)
}

func TestImport_skillWithoutFrontmatter_usesDefaultDescription(t *testing.T) {
	// arrange
	regDir := t.TempDir()
	projectRoot := t.TempDir()
	cfg := config.ProjectConfig{Registry: regDir, Formats: []string{"cursor"}, Bundles: []string{}}
	require.NoError(t, config.Write(projectRoot, cfg))

	skillDir := filepath.Join(projectRoot, ".cursor", "skills", "plain")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("Do the thing.\n"), 0644))

	stub := &importcmd.StubPrompter{
		Indices:   []int{0},
		Decisions: []importcmd.Decision{{LinkExisting: false, BundleRef: "skills/plain"}},
	}

	// act
	err := importcmd.Run(projectRoot, cfg, regDir, stub, io.Discard)

	// assert
	require.NoError(t, err)
	var meta bundleTOML
	_, err = toml.DecodeFile(filepath.Join(regDir, "skills", "plain", "bundle.toml"), &meta)
	require.NoError(t, err)
	assert.Equal(t, "Imported by skillsync", meta.Description)
}

func TestImport_skillWithLongDescription_truncatesTo50(t *testing.T) {
	// arrange
	regDir := t.TempDir()
	projectRoot := t.TempDir()
	cfg := config.ProjectConfig{Registry: regDir, Formats: []string{"cursor"}, Bundles: []string{}}
	require.NoError(t, config.Write(projectRoot, cfg))

	long := "This is a very long description that exceeds fifty characters by quite some margin"
	skillContent := "---\nname: myskill\ndescription: " + long + "\n---\n\nDo the thing.\n"
	skillDir := filepath.Join(projectRoot, ".cursor", "skills", "longdesc")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644))

	stub := &importcmd.StubPrompter{
		Indices:   []int{0},
		Decisions: []importcmd.Decision{{LinkExisting: false, BundleRef: "skills/longdesc"}},
	}

	// act
	err := importcmd.Run(projectRoot, cfg, regDir, stub, io.Discard)

	// assert
	require.NoError(t, err)
	var meta bundleTOML
	_, err = toml.DecodeFile(filepath.Join(regDir, "skills", "longdesc", "bundle.toml"), &meta)
	require.NoError(t, err)
	assert.LessOrEqual(t, len([]rune(meta.Description)), 53, "description must be at most 50 chars + ellipsis")
	assert.True(t, strings.HasSuffix(meta.Description, "…"), "long description must end with ellipsis")
}

func TestImport_multiFileSkill_allFilesInRegistry(t *testing.T) {
	// arrange
	regDir := t.TempDir()
	projectRoot := t.TempDir()
	cfg := config.ProjectConfig{Registry: regDir, Formats: []string{"cursor"}, Bundles: []string{}}
	require.NoError(t, config.Write(projectRoot, cfg))

	skillDir := filepath.Join(projectRoot, ".cursor", "skills", "multi")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("main\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "EXTRA.md"), []byte("extra\n"), 0644))

	stub := &importcmd.StubPrompter{
		Indices:   []int{0},
		Decisions: []importcmd.Decision{{LinkExisting: false, BundleRef: "skills/multi"}},
	}

	// act
	err := importcmd.Run(projectRoot, cfg, regDir, stub, io.Discard)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(regDir, "skills", "multi", "SKILL.md"))
	assert.FileExists(t, filepath.Join(regDir, "skills", "multi", "EXTRA.md"))
}

func TestImport_multiFileSkillFromFixture_allFilesInRegistryAndLock(t *testing.T) {
	const projectFixture = "../../testdata/projects/scripted"

	// arrange – copy fixture to a writable temp dir so we don't pollute testdata
	projectRoot := t.TempDir()
	copyDir(t, projectFixture, projectRoot)

	regDir := t.TempDir()
	cfg := config.ProjectConfig{Registry: regDir, Formats: []string{"cursor"}, Bundles: []string{}}
	require.NoError(t, config.Write(projectRoot, cfg))

	stub := &importcmd.StubPrompter{
		Indices:   []int{0},
		Decisions: []importcmd.Decision{{LinkExisting: false, BundleRef: "skills/scripted"}},
	}

	// act
	err := importcmd.Run(projectRoot, cfg, regDir, stub, io.Discard)

	// assert – all files land in registry
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(regDir, "skills", "scripted", "SKILL.md"), "SKILL.md must be in registry")
	assert.FileExists(t, filepath.Join(regDir, "skills", "scripted", "scripts", "helper.sh"), "nested script must be in registry")

	// assert – lock has entries for every file
	var lf struct {
		Entries []struct {
			Bundle string `toml:"bundle"`
			Target string `toml:"target"`
		} `toml:"entries"`
	}
	_, err = toml.DecodeFile(filepath.Join(projectRoot, ".skillsync", "sync.lock"), &lf)
	require.NoError(t, err)
	targets := make(map[string]bool)
	for _, e := range lf.Entries {
		if e.Bundle == "skills/scripted" {
			targets[e.Target] = true
		}
	}
	assert.True(t, targets[".cursor/skills/scripted/SKILL.md"], "lock must track SKILL.md")
	assert.True(t, targets[".cursor/skills/scripted/scripts/helper.sh"], "lock must track nested script")

	// assert – files synced back to project
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "skills", "scripted", "SKILL.md"), "SKILL.md must exist in project after sync")
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "skills", "scripted", "scripts", "helper.sh"), "nested script must exist in project after sync")
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
	require.NoError(t, err)
}

func TestImport_skillWithScripts_nestedFilesInRegistry(t *testing.T) {
	// arrange
	regDir := t.TempDir()
	projectRoot := t.TempDir()
	cfg := config.ProjectConfig{Registry: regDir, Formats: []string{"cursor"}, Bundles: []string{}}
	require.NoError(t, config.Write(projectRoot, cfg))

	skillDir := filepath.Join(projectRoot, ".cursor", "skills", "scripted")
	scriptsDir := filepath.Join(skillDir, "scripts")
	require.NoError(t, os.MkdirAll(scriptsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("main\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(scriptsDir, "run.sh"), []byte("#!/bin/sh\n"), 0644))

	stub := &importcmd.StubPrompter{
		Indices:   []int{0},
		Decisions: []importcmd.Decision{{LinkExisting: false, BundleRef: "skills/scripted"}},
	}

	// act
	err := importcmd.Run(projectRoot, cfg, regDir, stub, io.Discard)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(regDir, "skills", "scripted", "SKILL.md"))
	assert.FileExists(t, filepath.Join(regDir, "skills", "scripted", "scripts", "run.sh"))
}
