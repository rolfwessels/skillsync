package add_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/rolfwessels/skillsync/internal/add"
	"github.com/rolfwessels/skillsync/internal/config"
	skillssync "github.com/rolfwessels/skillsync/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const registryPath = "../../testdata/registry"

func setupProject(t *testing.T, bundles []string) string {
	t.Helper()
	projectRoot := t.TempDir()
	cfg := config.ProjectConfig{
		Registry: registryPath,
		Formats:  []string{"cursor"},
		Bundles:  bundles,
	}
	require.NoError(t, config.Write(projectRoot, cfg))
	return projectRoot
}

func TestRun_AppendsToConfigAndSyncs(t *testing.T) {
	// arrange
	projectRoot := setupProject(t, []string{})
	var warn bytes.Buffer

	// act
	err := add.Run(projectRoot, registryPath, "skills/tdd", &warn)

	// assert
	require.NoError(t, err)
	cfg, err := config.Load(projectRoot)
	require.NoError(t, err)
	assert.Contains(t, cfg.Bundles, "skills/tdd")
	assert.FileExists(t, filepath.Join(projectRoot, ".cursor", "skills", "tdd", "SKILL.md"))
}

func TestRun_BundleNotInRegistry_ConfigUnchanged(t *testing.T) {
	// arrange
	projectRoot := setupProject(t, []string{})
	var warn bytes.Buffer

	// act
	err := add.Run(projectRoot, registryPath, "skills/nonexistent", &warn)

	// assert
	require.Error(t, err)
	cfg, loadErr := config.Load(projectRoot)
	require.NoError(t, loadErr)
	assert.NotContains(t, cfg.Bundles, "skills/nonexistent")
}

func TestRun_BundleAlreadyInConfig_Error(t *testing.T) {
	// arrange
	projectRoot := setupProject(t, []string{"skills/tdd"})
	var warn bytes.Buffer

	// act
	err := add.Run(projectRoot, registryPath, "skills/tdd", &warn)

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, add.ErrAlreadyAdded)
}

func TestRun_ResultIdenticalToDeclaringInConfig(t *testing.T) {
	// arrange — project with bundle declared from the start
	projectRoot1 := setupProject(t, []string{"skills/tdd"})
	var warn bytes.Buffer
	cfg1, err := config.Load(projectRoot1)
	require.NoError(t, err)
	require.NoError(t, skillssync.Run(projectRoot1, registryPath, cfg1, &warn))

	// arrange — project with bundle added via add command
	projectRoot2 := setupProject(t, []string{})
	require.NoError(t, add.Run(projectRoot2, registryPath, "skills/tdd", &warn))

	// assert — synced file contents are identical
	want, err := os.ReadFile(filepath.Join(projectRoot1, ".cursor", "skills", "tdd", "SKILL.md"))
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(projectRoot2, ".cursor", "skills", "tdd", "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
