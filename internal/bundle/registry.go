package bundle

import (
	"fmt"
	"os"
	"path/filepath"
)

func Walk(registryRoot string) ([]Bundle, error) {
	kindEntries, err := os.ReadDir(registryRoot)
	if err != nil {
		return nil, fmt.Errorf("reading registry root %s: %w", registryRoot, err)
	}
	var bundles []Bundle
	for _, kindEntry := range kindEntries {
		if !kindEntry.IsDir() {
			continue
		}
		kind := kindEntry.Name()
		kindPath := filepath.Join(registryRoot, kind)
		bs, err := walkKind(kindPath, kind)
		if err != nil {
			return nil, err
		}
		bundles = append(bundles, bs...)
	}
	return bundles, nil
}

func walkKind(kindPath, kind string) ([]Bundle, error) {
	nameEntries, err := os.ReadDir(kindPath)
	if err != nil {
		return nil, fmt.Errorf("reading kind dir %s: %w", kindPath, err)
	}
	var bundles []Bundle
	for _, nameEntry := range nameEntries {
		if !nameEntry.IsDir() {
			continue
		}
		name := nameEntry.Name()
		dir := filepath.Join(kindPath, name)
		tomlPath := filepath.Join(dir, "bundle.toml")
		if _, err := os.Stat(tomlPath); os.IsNotExist(err) {
			continue
		}
		b, err := ParseBundleFile(dir, kind, name)
		if err != nil {
			return nil, err
		}
		bundles = append(bundles, b)
	}
	return bundles, nil
}
