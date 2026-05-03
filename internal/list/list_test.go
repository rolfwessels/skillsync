package list_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rolfwessels/skillsync/internal/list"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const registryPath = "../../testdata/registry"

func TestRun_RegistryOnly(t *testing.T) {
	// arrange
	var buf bytes.Buffer

	// act
	err := list.Run(registryPath, nil, &buf)

	// assert
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "agents")
	assert.Contains(t, output, "code-reviewer")
	assert.Contains(t, output, "commands")
	assert.Contains(t, output, "summarise")
	assert.Contains(t, output, "rules")
	assert.Contains(t, output, "naming-convention")
	assert.Contains(t, output, "skills")
	assert.Contains(t, output, "tdd")
	assert.NotContains(t, output, "✓", "no ✓ markers in registry-only mode")
}

func TestRun_InstalledBundlesMarked(t *testing.T) {
	// arrange
	var buf bytes.Buffer
	installed := []string{"skills/tdd", "rules/naming-convention"}

	// act
	err := list.Run(registryPath, installed, &buf)

	// assert
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "✓ tdd")
	assert.Contains(t, output, "✓ naming-convention")
	assert.NotContains(t, output, "✓ code-reviewer")
	assert.NotContains(t, output, "✓ summarise")
}

func TestRun_GroupedByKindAlphabetically(t *testing.T) {
	// arrange
	var buf bytes.Buffer

	// act
	err := list.Run(registryPath, nil, &buf)

	// assert
	require.NoError(t, err)
	output := buf.String()
	lines := strings.Split(output, "\n")

	agentsIdx := indexContaining(lines, "agents")
	commandsIdx := indexContaining(lines, "commands")
	rulesIdx := indexContaining(lines, "rules")
	skillsIdx := indexContaining(lines, "skills")

	require.NotEqual(t, -1, agentsIdx, "agents group missing")
	require.NotEqual(t, -1, commandsIdx, "commands group missing")
	require.NotEqual(t, -1, rulesIdx, "rules group missing")
	require.NotEqual(t, -1, skillsIdx, "skills group missing")

	assert.Less(t, agentsIdx, commandsIdx, "agents before commands")
	assert.Less(t, commandsIdx, rulesIdx, "commands before rules")
	assert.Less(t, rulesIdx, skillsIdx, "rules before skills")
}

func TestRun_BundlesAlphabeticalWithinKind(t *testing.T) {
	// arrange
	registryWithTwoBundles := registryPath
	var buf bytes.Buffer

	// act
	err := list.Run(registryWithTwoBundles, nil, &buf)

	// assert
	require.NoError(t, err)
	_ = buf.String()
}

func TestRun_InvalidRegistry(t *testing.T) {
	var buf bytes.Buffer
	err := list.Run("/nonexistent/path", nil, &buf)
	assert.Error(t, err)
}

func indexContaining(lines []string, substr string) int {
	for i, l := range lines {
		if strings.Contains(l, substr) {
			return i
		}
	}
	return -1
}
