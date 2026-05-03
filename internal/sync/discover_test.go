package sync_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rolfwessels/skillsync/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnumeratePotentialTargets_skillWithNestedFile_included(t *testing.T) {
	// arrange
	root := t.TempDir()
	skillDir := filepath.Join(root, ".cursor", "skills", "myscript")
	scriptsDir := filepath.Join(skillDir, "scripts")
	require.NoError(t, os.MkdirAll(scriptsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("skill"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(scriptsDir, "run.sh"), []byte("#!/bin/sh"), 0644))

	// act
	got, err := sync.EnumeratePotentialTargets(root, []string{"cursor"})

	// assert
	require.NoError(t, err)
	var foundScript bool
	for _, p := range got {
		if p.TargetRel == ".cursor/skills/myscript/scripts/run.sh" {
			foundScript = true
			assert.Equal(t, "scripts/run.sh", p.InnerPath)
		}
	}
	assert.True(t, foundScript, "nested script file must appear as a potential target")
}

func TestEnumeratePotentialTargets_cursorSkillsAndRules(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, ".cursor", "skills", "foo")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("skill"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".cursor", "rules"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".cursor", "rules", "bar.mdc"), []byte("rule"), 0644))

	got, err := sync.EnumeratePotentialTargets(root, []string{"cursor"})
	require.NoError(t, err)

	var skillHit, ruleHit bool
	for _, p := range got {
		switch p.TargetRel {
		case ".cursor/skills/foo/SKILL.md":
			skillHit = true
			assert.Equal(t, "skills", p.Kind)
			assert.Equal(t, "foo", p.BundleName)
		case ".cursor/rules/bar.mdc":
			ruleHit = true
			assert.Equal(t, "rules", p.Kind)
			assert.Equal(t, "bar", p.BundleName)
		}
	}
	assert.True(t, skillHit, "expected skill target")
	assert.True(t, ruleHit, "expected rule target")
}
