package sync

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type transformFunc func([]byte) ([]byte, error)

func applyContentTransform(kind, format string, data []byte) ([]byte, error) {
	d, err := lookupDescriptor(kind, format)
	if err != nil {
		return nil, err
	}
	if d.forward == nil {
		return data, nil
	}
	return d.forward(data)
}

func ContentToCanonical(kind, format string, data []byte) ([]byte, error) {
	return applyReverseContentTransform(kind, format, data)
}

func applyReverseContentTransform(kind, format string, data []byte) ([]byte, error) {
	d, err := lookupDescriptor(kind, format)
	if err != nil {
		return nil, err
	}
	if d.reverse == nil {
		return data, nil
	}
	return d.reverse(data)
}

func transformRuleToClaudeFormat(data []byte) ([]byte, error) {
	meta, body, ok := splitYAMLFrontmatter(data)
	if !ok {
		return data, nil
	}
	out := mapRuleMetaToClaudeFormat(meta)
	return assembleFrontmatter(out, body)
}

func mapRuleMetaToClaudeFormat(meta map[string]any) map[string]any {
	out := make(map[string]any)
	for _, k := range sortedMetaKeys(meta) {
		v := meta[k]
		if ck, cv, mapped := claudeRuleField(k, v); mapped {
			out[ck] = cv
			continue
		}
		out[k] = v
	}
	return out
}

func claudeRuleField(k string, v any) (string, any, bool) {
	n := strings.ToLower(strings.ReplaceAll(k, "_", ""))
	switch n {
	case "globs", "glob", "paths":
		return "paths", toStringSlice(v), true
	case "alwaysapply":
		return "alwaysApply", v, true
	default:
		return "", nil, false
	}
}

func toStringSlice(v any) []string {
	switch val := v.(type) {
	case string:
		parts := strings.Split(val, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				result = append(result, t)
			}
		}
		return result
	case []any:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, strings.TrimSpace(s))
			}
		}
		return result
	default:
		return nil
	}
}

func transformRuleToClaude(data []byte) ([]byte, error) {
	return transformRuleToClaudeFormat(data)
}

func transformAgentToClaude(data []byte) ([]byte, error) {
	meta, body, ok := splitYAMLFrontmatter(data)
	if !ok {
		return data, nil
	}
	out := mapAgentMetaToClaude(meta)
	return assembleFrontmatter(out, body)
}

func mapAgentMetaToClaude(meta map[string]any) map[string]any {
	out := make(map[string]any)
	for _, k := range sortedMetaKeys(meta) {
		v := meta[k]
		if ck, mapped := claudeAgentKeyFromCursor(k); mapped {
			out[ck] = v
			continue
		}
		out[k] = v
	}
	return out
}

func claudeAgentKeyFromCursor(k string) (string, bool) {
	n := strings.ToLower(strings.ReplaceAll(k, "_", ""))
	if n == "readonly" {
		return "read_only", true
	}
	return "", false
}

func transformRuleToCursor(data []byte) ([]byte, error) {
	meta, body, ok := splitYAMLFrontmatter(data)
	if !ok {
		return data, nil
	}
	out := mapRuleMetaToCursor(meta)
	return assembleFrontmatter(out, body)
}

func transformAgentToCursor(data []byte) ([]byte, error) {
	meta, body, ok := splitYAMLFrontmatter(data)
	if !ok {
		return data, nil
	}
	out := mapAgentMetaToCursor(meta)
	return assembleFrontmatter(out, body)
}

func splitYAMLFrontmatter(data []byte) (map[string]any, []byte, bool) {
	if !bytes.HasPrefix(data, []byte("---")) {
		return nil, data, false
	}
	rest := data[len("---"):]
	switch {
	case bytes.HasPrefix(rest, []byte("\r\n")):
		rest = rest[2:]
	case bytes.HasPrefix(rest, []byte("\n")):
		rest = rest[1:]
	default:
		return nil, data, false
	}
	var fm, body []byte
	if i := bytes.Index(rest, []byte("\n---\n")); i >= 0 {
		fm = rest[:i]
		body = rest[i+len("\n---\n"):]
	} else if i := bytes.Index(rest, []byte("\r\n---\r\n")); i >= 0 {
		fm = rest[:i]
		body = rest[i+len("\r\n---\r\n"):]
	} else {
		return nil, data, false
	}
	meta, err := parseYAMLMap(fm)
	if err != nil {
		return nil, data, false
	}
	return meta, body, true
}

func parseYAMLMap(fm []byte) (map[string]any, error) {
	var m map[string]any
	if err := yaml.Unmarshal(fm, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

func assembleFrontmatter(meta map[string]any, body []byte) ([]byte, error) {
	fmOut, err := yaml.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmOut)
	buf.WriteString("---\n")
	buf.Write(body)
	return buf.Bytes(), nil
}

func sortedMetaKeys(meta map[string]any) []string {
	keys := make([]string, 0, len(meta))
	for k := range meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func mapRuleMetaToCursor(meta map[string]any) map[string]any {
	out := make(map[string]any)
	for _, k := range sortedMetaKeys(meta) {
		v := meta[k]
		if ck, mapped := cursorRuleKey(k); mapped {
			out[ck] = v
			continue
		}
		out[k] = v
	}
	return out
}

func cursorRuleKey(k string) (string, bool) {
	n := strings.ToLower(strings.ReplaceAll(k, "_", ""))
	switch n {
	case "alwaysapply":
		return "alwaysApply", true
	case "globs", "glob", "paths":
		return "globs", true
	case "description":
		return "description", true
	default:
		return k, false
	}
}

func mapAgentMetaToCursor(meta map[string]any) map[string]any {
	out := make(map[string]any)
	for _, k := range sortedMetaKeys(meta) {
		v := meta[k]
		if ck, mapped := cursorAgentKey(k); mapped {
			out[ck] = v
			continue
		}
		out[k] = v
	}
	return out
}

func cursorAgentKey(k string) (string, bool) {
	n := strings.ToLower(strings.ReplaceAll(k, "_", ""))
	switch n {
	case "name":
		return "name", true
	case "description":
		return "description", true
	case "model":
		return "model", true
	case "readonly":
		return "readonly", true
	default:
		return k, false
	}
}
