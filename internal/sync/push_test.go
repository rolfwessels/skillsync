package sync_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/rolfwessels/skillsync/internal/config"
	"github.com/rolfwessels/skillsync/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func copyRegistryToTemp(t *testing.T) string {
	t.Helper()
	src := filepath.Join("..", "..", "testdata", "registry")
	dst := t.TempDir()
	require.NoError(t, copyTree(src, dst))
	return dst
}

func copyTree(srcRoot, dstRoot string) error {
	return filepath.WalkDir(srcRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return fmt.Errorf("rel path: %w", err)
		}
		if rel == "." {
			return os.MkdirAll(dstRoot, 0755)
		}
		dst := filepath.Join(dstRoot, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		out, err := os.Create(dst)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return err
		}
		return out.Close()
	})
}

func sha256File(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

func TestPush_NoOpLeavesRegistryAndLockUnchanged(t *testing.T) {
	regDir := copyRegistryToTemp(t)
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"rules/naming-convention"})
	require.NoError(t, config.Write(projectRoot, cfg))
	var warn bytes.Buffer
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	rulePath := filepath.Join(regDir, "rules", "naming-convention", "rule.md")
	wantSum := sha256File(t, rulePath)
	lockBefore := mustReadFile(t, filepath.Join(projectRoot, ".skillsync", "sync.lock"))

	require.NoError(t, sync.Push(projectRoot, regDir, cfg))

	assert.Equal(t, wantSum, sha256File(t, rulePath))
	lockAfter := mustReadFile(t, filepath.Join(projectRoot, ".skillsync", "sync.lock"))
	assert.Equal(t, string(lockBefore), string(lockAfter))
}

func TestPush_CursorRuleEditWritesClaudeCanonicalToRegistry(t *testing.T) {
	regDir := copyRegistryToTemp(t)
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"rules/naming-convention"})
	require.NoError(t, config.Write(projectRoot, cfg))
	var warn bytes.Buffer
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	cursorRule := filepath.Join(projectRoot, ".cursor", "rules", "naming-convention.mdc")
	orig, err := os.ReadFile(cursorRule)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cursorRule, append(orig, []byte("\nPushed line\n")...), 0644))

	require.NoError(t, sync.Push(projectRoot, regDir, cfg))

	rulePath := filepath.Join(regDir, "rules", "naming-convention", "rule.md")
	got, err := os.ReadFile(rulePath)
	require.NoError(t, err)
	s := string(got)
	assert.Contains(t, s, "Pushed line")
	assert.Contains(t, s, "alwaysApply:")
	assert.Contains(t, s, "paths:")
	assert.NotContains(t, s, "always_apply:")
}

func TestPush_UpdatesLockEntryHashToMatchDisk(t *testing.T) {
	regDir := copyRegistryToTemp(t)
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"skills/tdd"})
	require.NoError(t, config.Write(projectRoot, cfg))
	var warn bytes.Buffer
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	target := filepath.Join(projectRoot, ".cursor", "skills", "tdd", "SKILL.md")
	body, err := os.ReadFile(target)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(target, append(body, []byte("edit")...), 0644))
	wantHash := sha256File(t, target)

	require.NoError(t, sync.Push(projectRoot, regDir, cfg))

	var lf struct {
		Entries []struct {
			Target string `toml:"target"`
			Hash   string `toml:"hash"`
		} `toml:"entries"`
	}
	_, err = toml.DecodeFile(filepath.Join(projectRoot, ".skillsync", "sync.lock"), &lf)
	require.NoError(t, err)
	var gotHash string
	for _, e := range lf.Entries {
		if e.Target == ".cursor/skills/tdd/SKILL.md" {
			gotHash = e.Hash
			break
		}
	}
	assert.Equal(t, wantHash, gotHash)
}

func TestPush_OnlyChangedBundleWritten(t *testing.T) {
	regDir := copyRegistryToTemp(t)
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"skills/tdd", "rules/naming-convention"})
	require.NoError(t, config.Write(projectRoot, cfg))
	var warn bytes.Buffer
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	rulePath := filepath.Join(regDir, "rules", "naming-convention", "rule.md")
	ruleBefore, err := os.ReadFile(rulePath)
	require.NoError(t, err)

	skillPath := filepath.Join(projectRoot, ".cursor", "skills", "tdd", "SKILL.md")
	skillBody, err := os.ReadFile(skillPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(skillPath, append(skillBody, []byte("x")...), 0644))

	require.NoError(t, sync.Push(projectRoot, regDir, cfg))

	ruleAfter, err := os.ReadFile(rulePath)
	require.NoError(t, err)
	assert.Equal(t, string(ruleBefore), string(ruleAfter))
}

func TestPush_EmptyRegistryPathErrors(t *testing.T) {
	cfg := sampleConfig([]string{"cursor"}, []string{"skills/tdd"})
	err := sync.Push(t.TempDir(), "", cfg)
	require.Error(t, err)
}

func TestPush_AgentReverseTransformUpdatesRegistry(t *testing.T) {
	regDir := copyRegistryToTemp(t)
	projectRoot := t.TempDir()
	cfg := sampleConfig([]string{"cursor"}, []string{"agents/code-reviewer"})
	require.NoError(t, config.Write(projectRoot, cfg))
	var warn bytes.Buffer
	require.NoError(t, sync.Run(projectRoot, regDir, cfg, &warn))

	agentPath := filepath.Join(projectRoot, ".cursor", "agents", "code-reviewer.md")
	b, err := os.ReadFile(agentPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(agentPath, append(b, []byte("\nLocal\n")...), 0644))

	require.NoError(t, sync.Push(projectRoot, regDir, cfg))

	regAgent := filepath.Join(regDir, "agents", "code-reviewer", "agent.md")
	got, err := os.ReadFile(regAgent)
	require.NoError(t, err)
	assert.Contains(t, string(got), "read_only:")
	assert.Contains(t, string(got), "Local")
}
