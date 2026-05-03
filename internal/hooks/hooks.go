package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const MarkerLine = "# skillsync-managed-hook"

func hookScript() string {
	return "#!/bin/sh\n" + MarkerLine + " — installed by skillsync; safe to delete this file to remove auto-sync\nskillsync sync\n"
}

func Install(repoRoot string) ([]string, error) {
	hooksDir := filepath.Join(repoRoot, ".git", "hooks")
	if _, err := os.Stat(hooksDir); err != nil {
		return nil, fmt.Errorf(".git/hooks: %w", err)
	}
	var warnings []string
	for _, name := range []string{"post-checkout", "post-merge"} {
		w, err := installOne(filepath.Join(hooksDir, name), name)
		if w != "" {
			warnings = append(warnings, w)
		}
		if err != nil {
			return warnings, fmt.Errorf("install %s: %w", name, err)
		}
	}
	return warnings, nil
}

func installOne(path, basename string) (warning string, err error) {
	content := hookScript()
	existing, readErr := os.ReadFile(path)
	if readErr == nil {
		s := strings.TrimSpace(string(existing))
		if s != "" && !strings.Contains(string(existing), MarkerLine) {
			return fmt.Sprintf("skip %s: existing hook without %s marker", basename, MarkerLine), nil
		}
	} else if !os.IsNotExist(readErr) {
		return "", fmt.Errorf("reading hook: %w", readErr)
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		return "", fmt.Errorf("writing hook: %w", err)
	}
	return "", nil
}

func Uninstall(repoRoot string) error {
	hooksDir := filepath.Join(repoRoot, ".git", "hooks")
	for _, name := range []string{"post-checkout", "post-merge"} {
		path := filepath.Join(hooksDir, name)
		b, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("reading %s: %w", name, err)
		}
		if !strings.Contains(string(b), MarkerLine) {
			continue
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove %s: %w", name, err)
		}
	}
	return nil
}
