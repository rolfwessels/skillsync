package init

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildBundleOptions(t *testing.T) {
	t.Run("loads bundles from the given registry path", func(t *testing.T) {
		// arrange
		regDir := t.TempDir()
		bundleDir := filepath.Join(regDir, "skills", "mybundle")
		require.NoError(t, os.MkdirAll(bundleDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(bundleDir, "bundle.toml"), []byte("description = \"a bundle\"\n"), 0o644))

		// act
		opts, err := buildBundleOptions(regDir, nil)

		// assert
		require.NoError(t, err)
		require.Len(t, opts, 1)
		assert.Equal(t, "skills/mybundle", opts[0].Value)
	})

	t.Run("returns nil when registry path does not exist", func(t *testing.T) {
		// act
		opts, err := buildBundleOptions("/no/such/registry/path", nil)

		// assert
		require.NoError(t, err)
		assert.Nil(t, opts)
	})

	t.Run("returns empty slice when registry exists but has no bundles", func(t *testing.T) {
		// arrange
		regDir := t.TempDir()

		// act
		opts, err := buildBundleOptions(regDir, nil)

		// assert
		require.NoError(t, err)
		assert.Empty(t, opts)
	})

	t.Run("marks previously selected bundle", func(t *testing.T) {
		// arrange
		regDir := t.TempDir()
		bundleDir := filepath.Join(regDir, "skills", "mybundle")
		require.NoError(t, os.MkdirAll(bundleDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(bundleDir, "bundle.toml"), []byte("description = \"a bundle\"\n"), 0o644))

		// act
		opts, err := buildBundleOptions(regDir, []string{"skills/mybundle"})

		// assert
		require.NoError(t, err)
		require.Len(t, opts, 1)
		assert.Equal(t, "skills/mybundle", opts[0].Value)
	})
}
