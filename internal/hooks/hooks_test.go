package hooks_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rolfwessels/skillsync/internal/hooks"
)

func TestInstall_writesHooksAndMakesExecutable(t *testing.T) {
	repo := fakeGitRepo(t)
	warn, err := hooks.Install(repo)
	require.NoError(t, err)
	assert.Empty(t, warn)
	for _, name := range []string{"post-checkout", "post-merge"} {
		p := filepath.Join(repo, ".git", "hooks", name)
		b, err := os.ReadFile(p)
		require.NoError(t, err)
		assert.Contains(t, string(b), hooks.MarkerLine)
		assert.Contains(t, string(b), "skillsync sync")
		st, err := os.Stat(p)
		require.NoError(t, err)
		assert.True(t, st.Mode().IsRegular())
		assert.Equal(t, os.FileMode(0o755), st.Mode().Perm())
	}
}

func TestInstall_skipsExistingWithoutMarker(t *testing.T) {
	repo := fakeGitRepo(t)
	p := filepath.Join(repo, ".git", "hooks", "post-checkout")
	require.NoError(t, os.WriteFile(p, []byte("#!/bin/sh\necho hi\n"), 0o755))
	warn, err := hooks.Install(repo)
	require.NoError(t, err)
	require.Len(t, warn, 1)
	assert.Contains(t, warn[0], "post-checkout")
	assert.Contains(t, warn[0], "without")
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, "#!/bin/sh\necho hi\n", string(b))
	merge := filepath.Join(repo, ".git", "hooks", "post-merge")
	mb, err := os.ReadFile(merge)
	require.NoError(t, err)
	assert.Contains(t, string(mb), hooks.MarkerLine)
}

func TestInstall_overwritesSkillsyncHook(t *testing.T) {
	repo := fakeGitRepo(t)
	p := filepath.Join(repo, ".git", "hooks", "post-checkout")
	old := "#!/bin/sh\n" + hooks.MarkerLine + " old\nskillsync sync\n"
	require.NoError(t, os.WriteFile(p, []byte(old), 0o644))
	_, err := hooks.Install(repo)
	require.NoError(t, err)
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Contains(t, string(b), hooks.MarkerLine)
	assert.Contains(t, string(b), "skillsync sync")
}

func TestUninstall_removesOnlyMarked(t *testing.T) {
	repo := fakeGitRepo(t)
	_, err := hooks.Install(repo)
	require.NoError(t, err)
	other := filepath.Join(repo, ".git", "hooks", "post-commit")
	require.NoError(t, os.WriteFile(other, []byte("#!/bin/sh\nother\n"), 0o755))
	require.NoError(t, hooks.Uninstall(repo))
	_, err = os.Stat(filepath.Join(repo, ".git", "hooks", "post-checkout"))
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(repo, ".git", "hooks", "post-merge"))
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(other)
	assert.NoError(t, err)
}

func fakeGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0o755))
	return dir
}
