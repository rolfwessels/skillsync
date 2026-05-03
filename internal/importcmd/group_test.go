package importcmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rolfwessels/skillsync/internal/config"
	"github.com/rolfwessels/skillsync/internal/importcmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUntrackedGroups_excludesLockedPaths(t *testing.T) {
	root := t.TempDir()
	cfg := config.ProjectConfig{
		Formats: []string{"cursor"},
		Bundles: []string{},
	}
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".skillsync"), 0755))
	lock := `[[entries]]
bundle = "skills/existing"
format = "cursor"
target = ".cursor/skills/existing/SKILL.md"
hash = "abc"
`
	require.NoError(t, os.WriteFile(filepath.Join(root, ".skillsync", "sync.lock"), []byte(lock), 0644))

	skillDir := filepath.Join(root, ".cursor", "skills", "existing")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("x"), 0644))

	freeDir := filepath.Join(root, ".cursor", "skills", "orphan")
	require.NoError(t, os.MkdirAll(freeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(freeDir, "SKILL.md"), []byte("y"), 0644))

	groups, err := importcmd.UntrackedGroups(root, cfg)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, "orphan", groups[0].BundleName)
}

func TestUntrackedGroups_sameSkillTwoFormats_oneGroup(t *testing.T) {
	root := t.TempDir()
	cfg := config.ProjectConfig{
		Formats: []string{"cursor", "claude"},
		Bundles: []string{},
	}
	body := []byte("# Skill\n")
	cur := filepath.Join(root, ".cursor", "skills", "dup", "SKILL.md")
	cla := filepath.Join(root, ".claude", "skills", "dup", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(cur), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(cla), 0755))
	require.NoError(t, os.WriteFile(cur, body, 0644))
	require.NoError(t, os.WriteFile(cla, body, 0644))

	groups, err := importcmd.UntrackedGroups(root, cfg)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	require.Len(t, groups[0].Variants, 2)
	assert.Len(t, groups[0].Variants[0].Files, 1)
}

func TestUntrackedGroups_multipleFilesInSkillBundle_oneGroup(t *testing.T) {
	// arrange
	root := t.TempDir()
	cfg := config.ProjectConfig{Formats: []string{"cursor"}, Bundles: []string{}}
	skillDir := filepath.Join(root, ".cursor", "skills", "mymulti")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("x"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "EXTRA.md"), []byte("y"), 0644))

	// act
	groups, err := importcmd.UntrackedGroups(root, cfg)

	// assert
	require.NoError(t, err)
	require.Len(t, groups, 1, "multi-file skill bundle must produce one group")
	require.Len(t, groups[0].Variants, 1)
	assert.Len(t, groups[0].Variants[0].Files, 2, "both files must be in the single variant")
}

func TestUntrackedGroups_skillWithNestedFile_innerPathPreservesSubdir(t *testing.T) {
	// arrange
	root := t.TempDir()
	cfg := config.ProjectConfig{Formats: []string{"cursor"}, Bundles: []string{}}
	skillDir := filepath.Join(root, ".cursor", "skills", "nested")
	scriptsDir := filepath.Join(skillDir, "scripts")
	require.NoError(t, os.MkdirAll(scriptsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("x"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(scriptsDir, "run.sh"), []byte("y"), 0644))

	// act
	groups, err := importcmd.UntrackedGroups(root, cfg)

	// assert
	require.NoError(t, err)
	require.Len(t, groups, 1)
	require.Len(t, groups[0].Variants, 1)
	v := groups[0].Variants[0]
	require.Len(t, v.Files, 2)
	innerPaths := []string{v.Files[0].InnerPath, v.Files[1].InnerPath}
	assert.Contains(t, innerPaths, "SKILL.md")
	assert.Contains(t, innerPaths, "scripts/run.sh")
}

func TestUntrackedGroups_sortsCloudeBeforeCursor(t *testing.T) {
	// arrange
	root := t.TempDir()
	cfg := config.ProjectConfig{Formats: []string{"cursor", "claude"}, Bundles: []string{}}

	cursorOnly := filepath.Join(root, ".cursor", "skills", "cursor-skill")
	claudeOnly := filepath.Join(root, ".claude", "skills", "claude-skill")
	require.NoError(t, os.MkdirAll(cursorOnly, 0755))
	require.NoError(t, os.MkdirAll(claudeOnly, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(cursorOnly, "SKILL.md"), []byte("x"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(claudeOnly, "SKILL.md"), []byte("x"), 0644))

	// act
	groups, err := importcmd.UntrackedGroups(root, cfg)

	// assert
	require.NoError(t, err)
	require.Len(t, groups, 2)
	assert.Equal(t, "claude-skill", groups[0].BundleName, "claude-format group must sort first")
	assert.Equal(t, "cursor-skill", groups[1].BundleName)
}
