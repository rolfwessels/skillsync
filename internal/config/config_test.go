package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rolfwessels/skillsync/internal/config"
)

func sampleConfig() config.ProjectConfig {
	return config.ProjectConfig{
		Registry: "~/.skillsync/registry",
		Formats:  []string{"claude", "cursor"},
		Bundles:  []string{"skills/tdd", "rules/naming"},
	}
}

func writeAt(t *testing.T, dir, rel string, cfg config.ProjectConfig) {
	t.Helper()
	path := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	_, err = f.WriteString("registry = \"~/.skillsync/registry\"\nformats = [\"claude\", \"cursor\"]\nbundles = [\"skills/tdd\", \"rules/naming\"]\n")
	require.NoError(t, err)
}

func TestGlobalRegistry(t *testing.T) {
	t.Run("returns registry from global config when present", func(t *testing.T) {
		// arrange
		homeDir := t.TempDir()
		path := filepath.Join(homeDir, ".skillsync", "config.toml")
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
		require.NoError(t, os.WriteFile(path, []byte("[defaults]\nregistry = \"/global/registry\"\n"), 0644))

		// act & assert
		assert.Equal(t, "/global/registry", config.GlobalRegistry(homeDir))
	})

	t.Run("returns empty string when global config absent", func(t *testing.T) {
		assert.Empty(t, config.GlobalRegistry(t.TempDir()))
	})

	t.Run("returns empty string when file is malformed", func(t *testing.T) {
		// arrange
		homeDir := t.TempDir()
		path := filepath.Join(homeDir, ".skillsync", "config.toml")
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
		require.NoError(t, os.WriteFile(path, []byte("not valid toml :::"), 0644))

		// act & assert
		assert.Empty(t, config.GlobalRegistry(homeDir))
	})
}

func TestLoad(t *testing.T) {
	t.Run("loads config from .skillsync/config.toml", func(t *testing.T) {
		// arrange
		dir := t.TempDir()
		writeAt(t, dir, ".skillsync/config.toml", sampleConfig())

		// act
		cfg, err := config.Load(dir)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "~/.skillsync/registry", cfg.Registry)
		assert.Equal(t, []string{"claude", "cursor"}, cfg.Formats)
		assert.Equal(t, []string{"skills/tdd", "rules/naming"}, cfg.Bundles)
	})

	t.Run("loads config from .git/.skillsync/config.toml", func(t *testing.T) {
		// arrange
		dir := t.TempDir()
		writeAt(t, dir, ".git/.skillsync/config.toml", sampleConfig())

		// act
		cfg, err := config.Load(dir)

		// assert
		require.NoError(t, err)
		assert.Equal(t, "~/.skillsync/registry", cfg.Registry)
	})

	t.Run("errors if both locations exist", func(t *testing.T) {
		// arrange
		dir := t.TempDir()
		writeAt(t, dir, ".skillsync/config.toml", sampleConfig())
		writeAt(t, dir, ".git/.skillsync/config.toml", sampleConfig())

		// act
		_, err := config.Load(dir)

		// assert
		assert.ErrorContains(t, err, "ambiguous")
	})

	t.Run("errors if neither location exists", func(t *testing.T) {
		// arrange
		dir := t.TempDir()

		// act
		_, err := config.Load(dir)

		// assert
		assert.ErrorContains(t, err, "not found")
	})
}

func TestWrite(t *testing.T) {
	t.Run("writes config to .skillsync/config.toml", func(t *testing.T) {
		// arrange
		dir := t.TempDir()
		cfg := config.ProjectConfig{
			Registry: "~/.skillsync/registry",
			Formats:  []string{"claude", "cursor"},
			Bundles:  []string{"skills/tdd", "rules/naming-convention"},
		}

		// act
		err := config.Write(dir, cfg)

		// assert
		require.NoError(t, err)
		data, err := os.ReadFile(filepath.Join(dir, ".skillsync", "config.toml"))
		require.NoError(t, err, "config file not created")
		content := string(data)
		assert.Contains(t, content, `registry = "~/.skillsync/registry"`)
		assert.Contains(t, content, `"claude"`)
		assert.Contains(t, content, `"cursor"`)
		assert.Contains(t, content, `"skills/tdd"`)
	})

	t.Run("trims whitespace from registry path", func(t *testing.T) {
		// arrange
		dir := t.TempDir()
		cfg := config.ProjectConfig{Registry: "  /some/path  ", Formats: []string{"claude"}}

		// act
		err := config.Write(dir, cfg)

		// assert
		require.NoError(t, err)
		loaded, err := config.Load(dir)
		require.NoError(t, err)
		assert.Equal(t, "/some/path", loaded.Registry)
	})

	t.Run("errors if config already exists", func(t *testing.T) {
		// arrange
		dir := t.TempDir()
		cfg := config.ProjectConfig{Registry: "~/.skillsync/registry"}
		require.NoError(t, config.Write(dir, cfg))

		// act
		err := config.Write(dir, cfg)

		// assert
		assert.ErrorContains(t, err, "already exists")
	})
}
