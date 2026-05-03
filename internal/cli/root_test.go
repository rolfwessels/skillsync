package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rolfwessels/skillsync/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := cli.NewRootCmd()
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected help output, got empty string")
	}
}

func TestListCommand_RegistryFlag(t *testing.T) {
	// arrange
	buf := new(bytes.Buffer)
	cmd := cli.NewRootCmd()
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"list", "--registry", "../../testdata/registry"})

	// act
	err := cmd.Execute()

	// assert
	require.NoError(t, err)
	output := buf.String()
	assert.True(t, strings.Contains(output, "skills"), "output should contain skills group")
	assert.True(t, strings.Contains(output, "tdd"), "output should contain tdd bundle")
	assert.True(t, strings.Contains(output, "rules"), "output should contain rules group")
}

func TestRootCommand_Version(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := cli.NewRootCmd()
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--version"})

	_ = cmd.Execute()

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected version output, got empty string")
	}
}
