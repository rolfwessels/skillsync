package init_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rolfwessels/skillsync/internal/config"
	skillsinit "github.com/rolfwessels/skillsync/internal/init"
	"github.com/rolfwessels/skillsync/internal/hooks"
)

func TestRun(t *testing.T) {
	t.Run("writes config", func(t *testing.T) {
		projectDir := t.TempDir()
		stub := &stubPrompter{cfg: config.ProjectConfig{
			Registry: "/some/registry",
			Formats:  []string{"claude"},
			Bundles:  []string{"skills/tdd"},
		}}

		err := skillsinit.Run(projectDir, stub, io.Discard)

		require.NoError(t, err)
		_, statErr := os.Stat(filepath.Join(projectDir, ".skillsync", "config.toml"))
		assert.NoError(t, statErr, "config.toml should exist")
	})

	t.Run("errors immediately if already initialised without prompting", func(t *testing.T) {
		projectDir := t.TempDir()
		stub := &stubPrompter{cfg: config.ProjectConfig{Registry: "/some/registry", Formats: []string{"claude"}}}
		require.NoError(t, skillsinit.Run(projectDir, stub, io.Discard))
		stub.callCount = 0

		err := skillsinit.Run(projectDir, stub, io.Discard)

		assert.ErrorContains(t, err, "already initialised")
		assert.Equal(t, 0, stub.callCount, "prompter should not be called if already initialised")
	})

	t.Run("installs git hooks when in git repo and prompter accepts", func(t *testing.T) {
		projectDir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".git", "hooks"), 0o755))
		stub := &stubPrompter{
			cfg:          config.ProjectConfig{Registry: "/some/registry", Formats: []string{"claude"}},
			installHooks: true,
		}

		err := skillsinit.Run(projectDir, stub, io.Discard)

		require.NoError(t, err)
		b, err := os.ReadFile(filepath.Join(projectDir, ".git", "hooks", "post-merge"))
		require.NoError(t, err)
		assert.Contains(t, string(b), hooks.MarkerLine)
		assert.Contains(t, string(b), "skillsync sync")
	})

	t.Run("does not install hooks when prompter declines", func(t *testing.T) {
		projectDir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".git", "hooks"), 0o755))
		stub := &stubPrompter{
			cfg:          config.ProjectConfig{Registry: "/some/registry", Formats: []string{"claude"}},
			installHooks: false,
		}

		err := skillsinit.Run(projectDir, stub, io.Discard)

		require.NoError(t, err)
		_, statErr := os.Stat(filepath.Join(projectDir, ".git", "hooks", "post-merge"))
		assert.ErrorIs(t, statErr, os.ErrNotExist)
	})

	t.Run("does not fail when not a git repo", func(t *testing.T) {
		projectDir := t.TempDir()
		stub := &stubPrompter{
			cfg:          config.ProjectConfig{Registry: "/some/registry", Formats: []string{"claude"}},
			installHooks: true,
		}

		err := skillsinit.Run(projectDir, stub, io.Discard)

		require.NoError(t, err)
		_, statErr := os.Stat(filepath.Join(projectDir, ".git", "hooks", "post-merge"))
		assert.ErrorIs(t, statErr, os.ErrNotExist)
	})

	t.Run("syncs bundles after writing config", func(t *testing.T) {
		projectDir := t.TempDir()
		regDir := writeTestRegistryWithSkill(t)
		stub := &stubPrompter{
			cfg: config.ProjectConfig{
				Registry: regDir,
				Formats:  []string{"claude"},
				Bundles:  []string{"skills/syncdemo"},
			},
			installHooks: false,
		}

		err := skillsinit.Run(projectDir, stub, io.Discard)

		require.NoError(t, err)
		synced := filepath.Join(projectDir, ".claude", "skills", "syncdemo", "body.md")
		b, readErr := os.ReadFile(synced)
		require.NoError(t, readErr)
		assert.Equal(t, "synced-content", string(b))
		_, statErr := os.Stat(filepath.Join(projectDir, ".skillsync", "sync.lock"))
		assert.NoError(t, statErr)
	})

	t.Run("writes missing-bundle warnings to writer", func(t *testing.T) {
		projectDir := t.TempDir()
		regDir := t.TempDir()
		stub := &stubPrompter{
			cfg: config.ProjectConfig{
				Registry: regDir,
				Formats:  []string{"claude"},
				Bundles:  []string{"skills/absent"},
			},
			installHooks: false,
		}
		var warn bytes.Buffer

		runErr := skillsinit.Run(projectDir, stub, &warn)

		require.NoError(t, runErr)
		assert.Contains(t, warn.String(), "warning:")
		assert.Contains(t, warn.String(), "skills/absent")
	})

	t.Run("pre-fills registry default from global config", func(t *testing.T) {
		// arrange
		projectDir := t.TempDir()
		t.Setenv("HOME", t.TempDir())
		globalPath := filepath.Join(os.Getenv("HOME"), ".skillsync", "config.toml")
		require.NoError(t, os.MkdirAll(filepath.Dir(globalPath), 0755))
		require.NoError(t, os.WriteFile(globalPath, []byte("[defaults]\nregistry = \"/global/registry\"\n"), 0644))
		stub := &stubPrompter{cfg: config.ProjectConfig{Registry: "/global/registry", Formats: []string{"claude"}}}

		// act
		_ = skillsinit.Run(projectDir, stub, io.Discard)

		// assert
		assert.Equal(t, "/global/registry", stub.receivedDefaults.Registry)
	})

	t.Run("warns and skips hooks when hook exists without marker", func(t *testing.T) {
		projectDir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".git", "hooks"), 0o755))
		legacyPath := filepath.Join(projectDir, ".git", "hooks", "post-merge")
		require.NoError(t, os.WriteFile(legacyPath, []byte("#!/bin/sh\necho legacy\n"), 0o755))
		stub := &stubPrompter{
			cfg:          config.ProjectConfig{Registry: "/some/registry", Formats: []string{"claude"}},
			installHooks: true,
		}
		var warn bytes.Buffer

		runErr := skillsinit.Run(projectDir, stub, &warn)

		require.NoError(t, runErr)
		assert.Contains(t, warn.String(), "post-merge")
		b, err := os.ReadFile(legacyPath)
		require.NoError(t, err)
		assert.Contains(t, string(b), "legacy")
		assert.NotContains(t, string(b), hooks.MarkerLine)
	})
}

func TestReconfigure(t *testing.T) {
	t.Run("installs git hooks when prompter accepts", func(t *testing.T) {
		projectDir := initialisedProject(t)
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".git", "hooks"), 0o755))
		stub := &stubPrompter{
			cfg:          config.ProjectConfig{Registry: "/some/registry", Formats: []string{"claude"}},
			installHooks: true,
		}

		err := skillsinit.Reconfigure(projectDir, stub, io.Discard)

		require.NoError(t, err)
		b, err := os.ReadFile(filepath.Join(projectDir, ".git", "hooks", "post-merge"))
		require.NoError(t, err)
		assert.Contains(t, string(b), hooks.MarkerLine)
		assert.Contains(t, string(b), "skillsync sync")
	})

	t.Run("refreshes stale skillsync hook content", func(t *testing.T) {
		projectDir := initialisedProject(t)
		hooksDir := filepath.Join(projectDir, ".git", "hooks")
		require.NoError(t, os.MkdirAll(hooksDir, 0o755))
		stale := "#!/bin/sh\n" + hooks.MarkerLine + "\nskillsync pull\n"
		require.NoError(t, os.WriteFile(filepath.Join(hooksDir, "post-merge"), []byte(stale), 0o755))
		stub := &stubPrompter{
			cfg:          config.ProjectConfig{Registry: "/some/registry", Formats: []string{"claude"}},
			installHooks: true,
		}

		err := skillsinit.Reconfigure(projectDir, stub, io.Discard)

		require.NoError(t, err)
		b, err := os.ReadFile(filepath.Join(hooksDir, "post-merge"))
		require.NoError(t, err)
		assert.Contains(t, string(b), "skillsync sync")
		assert.NotContains(t, string(b), "skillsync pull")
	})

	t.Run("does not install hooks when prompter declines", func(t *testing.T) {
		projectDir := initialisedProject(t)
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".git", "hooks"), 0o755))
		stub := &stubPrompter{
			cfg:          config.ProjectConfig{Registry: "/some/registry", Formats: []string{"claude"}},
			installHooks: false,
		}

		err := skillsinit.Reconfigure(projectDir, stub, io.Discard)

		require.NoError(t, err)
		_, statErr := os.Stat(filepath.Join(projectDir, ".git", "hooks", "post-merge"))
		assert.ErrorIs(t, statErr, os.ErrNotExist)
	})
}

func initialisedProject(t *testing.T) string {
	t.Helper()
	projectDir := t.TempDir()
	stub := &stubPrompter{cfg: config.ProjectConfig{Registry: "/some/registry", Formats: []string{"claude"}}}
	require.NoError(t, skillsinit.Run(projectDir, stub, io.Discard))
	return projectDir
}

func writeTestRegistryWithSkill(t *testing.T) string {
	t.Helper()
	regDir := t.TempDir()
	bundleDir := filepath.Join(regDir, "skills", "syncdemo")
	require.NoError(t, os.MkdirAll(bundleDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(bundleDir, "bundle.toml"), []byte("description = \"d\"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(bundleDir, "body.md"), []byte("synced-content"), 0o644))
	return regDir
}

type stubPrompter struct {
	cfg              config.ProjectConfig
	installHooks     bool
	callCount        int
	receivedDefaults config.ProjectConfig
}

func (s *stubPrompter) Ask(_ string, defaults config.ProjectConfig) (skillsinit.PromptResult, error) {
	s.callCount++
	s.receivedDefaults = defaults
	return skillsinit.PromptResult{Config: s.cfg, InstallGitHooks: s.installHooks}, nil
}
