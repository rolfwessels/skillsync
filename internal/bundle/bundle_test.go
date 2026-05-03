package bundle_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rolfwessels/skillsync/internal/bundle"
)

func TestParseBundleFile(t *testing.T) {
	t.Run("parses description and tags", func(t *testing.T) {
		// arrange
		dir := t.TempDir()
		writeTOML(t, filepath.Join(dir, "bundle.toml"), `
description = "A test skill"
tags = ["go", "testing"]
`)
		// act
		b, err := bundle.ParseBundleFile(dir, "skills", "my-skill")

		// assert
		require.NoError(t, err)
		assert.Equal(t, "my-skill", b.Name)
		assert.Equal(t, "skills", b.Kind)
		assert.Equal(t, "A test skill", b.Description)
		assert.Equal(t, []string{"go", "testing"}, b.Tags)
		assert.Equal(t, dir, b.Path)
	})

	t.Run("missing bundle.toml is an error", func(t *testing.T) {
		// arrange
		dir := t.TempDir()

		// act
		_, err := bundle.ParseBundleFile(dir, "skills", "missing")

		// assert
		assert.Error(t, err)
	})
}
