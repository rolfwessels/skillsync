package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

func ParseRef(ref string) (kind, name string, err error) {
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid bundle ref %q: expected kind/name", ref)
	}
	return parts[0], parts[1], nil
}

type Bundle struct {
	Name        string
	Kind        string
	Description string
	Tags        []string
	Path        string
}

type bundleTOML struct {
	Description string   `toml:"description"`
	Tags        []string `toml:"tags"`
}

func ParseBundleFile(dir, kind, name string) (Bundle, error) {
	tomlPath := filepath.Join(dir, "bundle.toml")
	data, err := os.ReadFile(tomlPath)
	if err != nil {
		return Bundle{}, fmt.Errorf("reading bundle.toml in %s: %w", dir, err)
	}
	var meta bundleTOML
	if err := toml.Unmarshal(data, &meta); err != nil {
		return Bundle{}, fmt.Errorf("parsing bundle.toml in %s: %w", dir, err)
	}
	return Bundle{
		Name:        name,
		Kind:        kind,
		Description: meta.Description,
		Tags:        meta.Tags,
		Path:        dir,
	}, nil
}
