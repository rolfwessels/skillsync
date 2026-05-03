package bundle_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rolfwessels/skillsync/internal/bundle"
)

func TestWalk(t *testing.T) {
	t.Run("returns bundles from testdata registry", func(t *testing.T) {
		// arrange
		root := filepath.Join("..", "..", "testdata", "registry")

		// act
		bundles, err := bundle.Walk(root)

		// assert
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(bundles), 4, "expected at least one bundle per kind")
		kinds := map[string]bool{}
		for _, b := range bundles {
			kinds[b.Kind] = true
		}
		for _, k := range []string{"skills", "rules", "agents", "commands"} {
			assert.True(t, kinds[k], "missing bundle for kind %q", k)
		}
	})

	t.Run("skips folders without bundle.toml", func(t *testing.T) {
		// arrange
		dir := t.TempDir()
		kindDir := filepath.Join(dir, "skills")
		noToml := filepath.Join(kindDir, "no-toml")
		withToml := filepath.Join(kindDir, "with-toml")
		os.MkdirAll(noToml, 0755)
		os.MkdirAll(withToml, 0755)
		writeTOML(t, filepath.Join(withToml, "bundle.toml"), `description = "has toml"`)

		// act
		bundles, err := bundle.Walk(dir)

		// assert
		require.NoError(t, err)
		require.Len(t, bundles, 1)
		assert.Equal(t, "with-toml", bundles[0].Name)
	})
}

func writeTOML(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}
