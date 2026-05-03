package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/rolfwessels/skillsync/internal/bundle"
	"github.com/rolfwessels/skillsync/internal/config"
)

func Push(projectRoot, registryRoot string, cfg config.ProjectConfig) error {
	if registryRoot == "" {
		return fmt.Errorf("no registry path configured")
	}
	lf, err := loadLock(projectRoot)
	if err != nil {
		return fmt.Errorf("loading lockfile: %w", err)
	}
	if len(lf.Entries) == 0 {
		return nil
	}
	modified := false
	for i := range lf.Entries {
		e := &lf.Entries[i]
		if !slices.Contains(cfg.Bundles, e.Bundle) {
			continue
		}
		if !slices.Contains(cfg.Formats, e.Format) {
			continue
		}
		kind, name, err := bundle.ParseRef(e.Bundle)
		if err != nil {
			return fmt.Errorf("bundle ref %q in lockfile: %w", e.Bundle, err)
		}
		srcAbs, err := resolveRegistrySource(registryRoot, kind, name, e.Format, e.Target)
		if err != nil {
			return fmt.Errorf("resolving registry file for %s: %w", e.Bundle, err)
		}
		absTarget := filepath.Join(projectRoot, filepath.FromSlash(e.Target))
		data, err := os.ReadFile(absTarget)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("reading project file %s: %w", e.Target, err)
		}
		diskHash := hashBytes(data)
		if diskHash == e.Hash {
			continue
		}
		canonical, err := applyReverseContentTransform(kind, e.Format, data)
		if err != nil {
			return fmt.Errorf("reverse transform %s: %w", e.Target, err)
		}
		if err := writeFile(srcAbs, canonical); err != nil {
			return fmt.Errorf("writing registry file: %w", err)
		}
		e.Hash = diskHash
		modified = true
	}
	if !modified {
		return nil
	}
	return saveLock(projectRoot, lf)
}

func resolveRegistrySource(registryRoot, kind, name, format, target string) (string, error) {
	spec, err := lookupSpec(kind, format)
	if err != nil {
		return "", err
	}
	bundleDir := filepath.Join(registryRoot, kind, name)
	contentFiles, err := bundleContentFiles(bundleDir)
	if err != nil {
		return "", err
	}
	for _, p := range contentFiles {
		if resolveTarget(spec, bundleDir, name, p) == target {
			return p, nil
		}
	}
	return "", fmt.Errorf("no bundle file maps to target %q", target)
}

