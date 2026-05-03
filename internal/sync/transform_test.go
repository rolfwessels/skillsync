package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForwardTransform(t *testing.T) {
	tests := []struct {
		name        string
		kind        string
		format      string
		give        string
		wantContain []string
		wantAbsent  []string
		wantEqual   string
	}{
		// skill — no frontmatter, identity for all formats
		{
			name:      "skill claude: identity",
			kind:      "skills", format: "claude",
			give:      "# My Skill\nbody content\n",
			wantEqual: "# My Skill\nbody content\n",
		},
		{
			name:      "skill cursor: identity",
			kind:      "skills", format: "cursor",
			give:      "# My Skill\nbody content\n",
			wantEqual: "# My Skill\nbody content\n",
		},
		// rule — claude forward transform
		{
			name:        "rule claude: globs → paths as YAML sequence",
			kind:        "rules", format: "claude",
			give:        "---\nglobs: '**/*.go'\n---\n# Rule\n",
			wantContain: []string{"paths:"},
			wantAbsent:  []string{"globs:", "glob:"},
		},
		{
			name:        "rule claude: alwaysApply passes through unchanged",
			kind:        "rules", format: "claude",
			give:        "---\nalwaysApply: false\n---\n# Rule\n",
			wantContain: []string{"alwaysApply:"},
			wantAbsent:  []string{"always_apply:"},
		},
		{
			name:        "rule claude: multi-pattern globs → paths sequence",
			kind:        "rules", format: "claude",
			give:        "---\nglobs: '**/*.go, **/go.mod'\n---\n# Rule\n",
			wantContain: []string{"paths:", "**/*.go", "**/go.mod"},
			wantAbsent:  []string{"globs:"},
		},
		{
			name:        "rule claude: glob singular → paths sequence",
			kind:        "rules", format: "claude",
			give:        "---\nglob: \"*.go\"\nalways_apply: true\n---\n# Rule\n",
			wantContain: []string{"paths:", "alwaysApply:"},
			wantAbsent:  []string{"glob:", "always_apply:"},
		},
		{
			name:      "rule claude: no frontmatter, identity",
			kind:      "rules", format: "claude",
			give:      "# Rule without frontmatter\n",
			wantEqual: "# Rule without frontmatter\n",
		},
		{
			name: "rule claude: empty file, no error",
			kind: "rules", format: "claude",
			give: "",
		},
		// rule — cursor: known field mapping
		{
			name:        "rule cursor: maps always_apply → alwaysApply",
			kind:        "rules", format: "cursor",
			give:        "---\nalways_apply: true\n---\n# Rule\n",
			wantContain: []string{"alwaysApply:"},
			wantAbsent:  []string{"always_apply:"},
		},
		{
			name:        "rule cursor: maps glob → globs",
			kind:        "rules", format: "cursor",
			give:        "---\nglob: \"*.go\"\n---\n# Rule\n",
			wantContain: []string{"globs:"},
			wantAbsent:  []string{"\nglob:"},
		},
		{
			name:        "rule cursor: description passes through unchanged",
			kind:        "rules", format: "cursor",
			give:        "---\ndescription: lint rule\n---\n# Rule\n",
			wantContain: []string{"description:"},
		},
		{
			name:        "rule cursor: unknown frontmatter fields pass through",
			kind:        "rules", format: "cursor",
			give:        "---\nalways_apply: true\ncustom_field: 42\n---\n# Rule\n",
			wantContain: []string{"alwaysApply:", "custom_field:"},
			wantAbsent:  []string{"always_apply:"},
		},
		{
			name:        "rule cursor: missing optional glob field omitted",
			kind:        "rules", format: "cursor",
			give:        "---\ndescription: just a desc\n---\n# Rule\n",
			wantContain: []string{"description:"},
			wantAbsent:  []string{"globs:", "alwaysApply:"},
		},
		{
			name:      "rule cursor: no frontmatter, identity",
			kind:      "rules", format: "cursor",
			give:      "# Rule without frontmatter\n",
			wantEqual: "# Rule without frontmatter\n",
		},
		{
			name:   "rule cursor: empty file, no error",
			kind:   "rules", format: "cursor",
			give:   "",
		},
		// agent — claude is canonical (identity)
		{
			name:      "agent claude: already canonical, identity",
			kind:      "agents", format: "claude",
			give:      "---\nread_only: true\n---\n# Agent\n",
			wantEqual: "---\nread_only: true\n---\n# Agent\n",
		},
		// agent — cursor: known field mapping
		{
			name:        "agent cursor: maps read_only → readonly",
			kind:        "agents", format: "cursor",
			give:        "---\nread_only: true\nname: myagent\n---\n# Agent\n",
			wantContain: []string{"readonly:", "name:"},
			wantAbsent:  []string{"read_only:"},
		},
		{
			name:        "agent cursor: model and description pass through unchanged",
			kind:        "agents", format: "cursor",
			give:        "---\nmodel: claude-3\ndescription: my agent\n---\n# Agent\n",
			wantContain: []string{"model:", "description:"},
		},
		{
			name:        "agent cursor: unknown frontmatter fields pass through",
			kind:        "agents", format: "cursor",
			give:        "---\nread_only: false\nextra_field: foo\n---\n# Agent\n",
			wantContain: []string{"readonly:", "extra_field:"},
			wantAbsent:  []string{"read_only:"},
		},
		{
			name:        "agent cursor: missing optional read_only field omitted",
			kind:        "agents", format: "cursor",
			give:        "---\nname: myagent\n---\n# Agent\n",
			wantContain: []string{"name:"},
			wantAbsent:  []string{"readonly:", "read_only:"},
		},
		{
			name:   "agent cursor: empty file, no error",
			kind:   "agents", format: "cursor",
			give:   "",
		},
		// command — identity for all formats
		{
			name:      "command claude: identity",
			kind:      "commands", format: "claude",
			give:      "# Command\ndo stuff\n",
			wantEqual: "# Command\ndo stuff\n",
		},
		{
			name:      "command cursor: identity",
			kind:      "commands", format: "cursor",
			give:      "# Command\ndo stuff\n",
			wantEqual: "# Command\ndo stuff\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyContentTransform(tt.kind, tt.format, []byte(tt.give))
			require.NoError(t, err)
			if tt.wantEqual != "" {
				assert.Equal(t, tt.wantEqual, string(got))
			}
			for _, s := range tt.wantContain {
				assert.Contains(t, string(got), s)
			}
			for _, s := range tt.wantAbsent {
				assert.NotContains(t, string(got), s)
			}
		})
	}
}

func TestReverseTransform(t *testing.T) {
	tests := []struct {
		name        string
		kind        string
		format      string
		give        string
		wantContain []string
		wantAbsent  []string
		wantEqual   string
	}{
		// rule cursor → claude (canonical)
		{
			name:        "rule cursor: alwaysApply passes through unchanged",
			kind:        "rules", format: "cursor",
			give:        "---\nalwaysApply: true\n---\n# Rule\n",
			wantContain: []string{"alwaysApply:"},
			wantAbsent:  []string{"always_apply:"},
		},
		{
			name:        "rule cursor: globs → paths as YAML sequence",
			kind:        "rules", format: "cursor",
			give:        "---\nglobs:\n  - \"*.go\"\n---\n# Rule\n",
			wantContain: []string{"paths:"},
			wantAbsent:  []string{"globs:", "glob:"},
		},
		{
			name:        "rule cursor: unknown frontmatter fields pass through",
			kind:        "rules", format: "cursor",
			give:        "---\nalwaysApply: false\ncustom_key: 99\n---\n# Rule\n",
			wantContain: []string{"alwaysApply:", "custom_key:"},
		},
		{
			name:      "rule cursor: no frontmatter, identity",
			kind:      "rules", format: "cursor",
			give:      "# Rule\n",
			wantEqual: "# Rule\n",
		},
		{
			name: "rule cursor: empty file, no error",
			kind: "rules", format: "cursor",
			give: "",
		},
		// rule claude → claude (reverse is identity — no cursor fields to map back)
		{
			name:      "rule claude: identity",
			kind:      "rules", format: "claude",
			give:      "---\nalwaysApply: true\npaths:\n    - \"*.go\"\n---\n# Rule\n",
			wantEqual: "---\nalwaysApply: true\npaths:\n    - \"*.go\"\n---\n# Rule\n",
		},
		// agent cursor → claude
		{
			name:        "agent cursor: maps readonly → read_only",
			kind:        "agents", format: "cursor",
			give:        "---\nreadonly: true\nname: n\n---\n# Agent\n",
			wantContain: []string{"read_only:", "name:"},
			wantAbsent:  []string{"readonly:"},
		},
		{
			name:        "agent cursor: unknown frontmatter fields pass through",
			kind:        "agents", format: "cursor",
			give:        "---\nreadonly: false\nextra: bar\n---\n# Agent\n",
			wantContain: []string{"read_only:", "extra:"},
			wantAbsent:  []string{"readonly:"},
		},
		// skill cursor → claude (no reverse transform, identity)
		{
			name:      "skill cursor: identity",
			kind:      "skills", format: "cursor",
			give:      "# Skill\n",
			wantEqual: "# Skill\n",
		},
		// command cursor → claude (no reverse transform, identity)
		{
			name:      "command cursor: identity",
			kind:      "commands", format: "cursor",
			give:      "# Cmd\n",
			wantEqual: "# Cmd\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyReverseContentTransform(tt.kind, tt.format, []byte(tt.give))
			require.NoError(t, err)
			if tt.wantEqual != "" {
				assert.Equal(t, tt.wantEqual, string(got))
			}
			for _, s := range tt.wantContain {
				assert.Contains(t, string(got), s)
			}
			for _, s := range tt.wantAbsent {
				assert.NotContains(t, string(got), s)
			}
		})
	}
}

func TestRegisterFormat_AppliesForwardTransform(t *testing.T) {
	registerFormat("skills", "testfmt", formatDescriptor{
		pathSpec: pathSpec{dir: ".test/skills", kind: targetDir},
		forward: func(b []byte) ([]byte, error) {
			return append([]byte("pfx:"), b...), nil
		},
	})
	got, err := applyContentTransform("skills", "testfmt", []byte("a"))
	require.NoError(t, err)
	assert.Equal(t, "pfx:a", string(got))
}
