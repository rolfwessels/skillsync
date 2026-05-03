package sync

import "fmt"

type targetKind int

const (
	targetDir  targetKind = iota
	targetFile
)

type pathSpec struct {
	dir  string
	ext  string
	kind targetKind
}

type formatDescriptor struct {
	pathSpec
	forward transformFunc
	reverse transformFunc
}

var formatRegistry = map[string]map[string]formatDescriptor{}

func registerFormat(kind, format string, d formatDescriptor) {
	if formatRegistry[kind] == nil {
		formatRegistry[kind] = map[string]formatDescriptor{}
	}
	formatRegistry[kind][format] = d
}

func lookupDescriptor(kind, format string) (formatDescriptor, error) {
	formats, ok := formatRegistry[kind]
	if !ok {
		return formatDescriptor{}, fmt.Errorf("unknown kind %q", kind)
	}
	d, ok := formats[format]
	if !ok {
		return formatDescriptor{}, fmt.Errorf("unknown format %q for kind %q", format, kind)
	}
	return d, nil
}

func lookupSpec(kind, format string) (pathSpec, error) {
	d, err := lookupDescriptor(kind, format)
	if err != nil {
		return pathSpec{}, err
	}
	return d.pathSpec, nil
}

func init() {
	registerFormat("skills", "claude", formatDescriptor{pathSpec: pathSpec{dir: ".claude/skills", kind: targetDir}})
	registerFormat("skills", "cursor", formatDescriptor{pathSpec: pathSpec{dir: ".cursor/skills", kind: targetDir}})
	registerFormat("rules", "claude", formatDescriptor{
		pathSpec: pathSpec{dir: ".claude/rules", ext: ".md", kind: targetFile},
		forward:  transformRuleToClaudeFormat,
	})
	registerFormat("rules", "cursor", formatDescriptor{
		pathSpec: pathSpec{dir: ".cursor/rules", ext: ".mdc", kind: targetFile},
		forward:  transformRuleToCursor,
		reverse:  transformRuleToClaude,
	})
	registerFormat("agents", "claude", formatDescriptor{pathSpec: pathSpec{dir: ".claude/agents", ext: ".md", kind: targetFile}})
	registerFormat("agents", "cursor", formatDescriptor{
		pathSpec: pathSpec{dir: ".cursor/agents", ext: ".md", kind: targetFile},
		forward:  transformAgentToCursor,
		reverse:  transformAgentToClaude,
	})
	registerFormat("commands", "claude", formatDescriptor{pathSpec: pathSpec{dir: ".claude/commands", ext: ".md", kind: targetFile}})
	registerFormat("commands", "cursor", formatDescriptor{pathSpec: pathSpec{dir: ".cursor/commands", ext: ".md", kind: targetFile}})
}
