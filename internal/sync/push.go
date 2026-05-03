package sync

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
	modified := false
	for _, bundleRef := range cfg.Bundles {
		kind, name, err := bundle.ParseRef(bundleRef)
		if err != nil {
			return fmt.Errorf("bundle ref %q: %w", bundleRef, err)
		}
		bundleDir := filepath.Join(registryRoot, kind, name)
		if _, err := os.Stat(bundleDir); os.IsNotExist(err) {
			continue
		}
		for _, format := range cfg.Formats {
			changed, err := pushBundleFormat(projectRoot, bundleDir, kind, name, format, bundleRef, &lf)
			if err != nil {
				return fmt.Errorf("pushing %s for format %s: %w", bundleRef, format, err)
			}
			if changed {
				modified = true
			}
		}
	}
	if !modified {
		return nil
	}
	return saveLock(projectRoot, lf)
}

func pushBundleFormat(projectRoot, bundleDir, kind, name, format, bundleRef string, lf *lockfile) (bool, error) {
	spec, err := lookupSpec(kind, format)
	if err != nil {
		return false, err
	}
	targets, err := discoverProjectTargets(projectRoot, spec, name)
	if err != nil {
		return false, err
	}
	modified := false
	for _, target := range targets {
		changed, err := pushTarget(projectRoot, bundleDir, kind, name, format, bundleRef, target, spec, lf)
		if err != nil {
			return modified, err
		}
		if changed {
			modified = true
		}
	}
	dropped, err := deleteMissingTargets(bundleDir, name, format, bundleRef, spec, targets, lf)
	if err != nil {
		return modified, err
	}
	if dropped {
		modified = true
	}
	return modified, nil
}

func discoverProjectTargets(projectRoot string, spec pathSpec, name string) ([]string, error) {
	if spec.kind == targetDir {
		return discoverDirTargets(projectRoot, spec, name)
	}
	return discoverFileTarget(projectRoot, spec, name)
}

func discoverDirTargets(projectRoot string, spec pathSpec, name string) ([]string, error) {
	baseRel := spec.dir + "/" + name
	baseAbs := filepath.Join(projectRoot, filepath.FromSlash(baseRel))
	if _, err := os.Stat(baseAbs); os.IsNotExist(err) {
		return nil, nil
	}
	var out []string
	err := filepath.WalkDir(baseAbs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(baseAbs, path)
		if relErr != nil {
			return relErr
		}
		out = append(out, baseRel+"/"+filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking project bundle %s: %w", baseAbs, err)
	}
	return out, nil
}

func discoverFileTarget(projectRoot string, spec pathSpec, name string) ([]string, error) {
	target := spec.dir + "/" + name + spec.ext
	abs := filepath.Join(projectRoot, filepath.FromSlash(target))
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("statting %s: %w", abs, err)
	}
	return []string{target}, nil
}

func pushTarget(projectRoot, bundleDir, kind, name, format, bundleRef, target string, spec pathSpec, lf *lockfile) (bool, error) {
	absTarget := filepath.Join(projectRoot, filepath.FromSlash(target))
	data, err := os.ReadFile(absTarget)
	if err != nil {
		return false, fmt.Errorf("reading project file %s: %w", target, err)
	}
	diskHash := hashBytes(data)
	existing, hasLock := lf.lookup(bundleRef, format, target)
	if hasLock && existing.Hash == diskHash {
		return false, nil
	}
	canonical, err := applyReverseContentTransform(kind, format, data)
	if err != nil {
		return false, fmt.Errorf("reverse transform %s: %w", target, err)
	}
	src, err := resolveBundleSource(spec, bundleDir, name, target)
	if err != nil {
		return false, err
	}
	if err := writeFile(src, canonical); err != nil {
		return false, fmt.Errorf("writing registry file %s: %w", src, err)
	}
	registryHash := hashBytes(canonical)
	lf.upsert(LockEntry{
		Bundle:       bundleRef,
		Format:       format,
		Target:       target,
		Hash:         diskHash,
		RegistryHash: registryHash,
		LocalHash:    diskHash,
		State:        StateClean,
	})
	return true, nil
}

func resolveBundleSource(spec pathSpec, bundleDir, name, target string) (string, error) {
	if spec.kind == targetDir {
		prefix := spec.dir + "/" + name + "/"
		if !strings.HasPrefix(target, prefix) {
			return "", fmt.Errorf("target %q outside bundle %q", target, name)
		}
		inner := strings.TrimPrefix(target, prefix)
		return filepath.Join(bundleDir, filepath.FromSlash(inner)), nil
	}
	files, err := bundleContentFiles(bundleDir)
	if err != nil {
		return "", err
	}
	if len(files) > 0 {
		return files[0], nil
	}
	return "", fmt.Errorf("no content file in bundle %q to push to", name)
}

func deleteMissingTargets(bundleDir, name, format, bundleRef string, spec pathSpec, projectTargets []string, lf *lockfile) (bool, error) {
	keep := targetSet(projectTargets)
	var remaining []LockEntry
	modified := false
	for _, e := range lf.Entries {
		if e.Bundle != bundleRef || e.Format != format || keep[e.Target] {
			remaining = append(remaining, e)
			continue
		}
		src, err := resolveBundleSource(spec, bundleDir, name, e.Target)
		if err != nil {
			return modified, err
		}
		if err := os.Remove(src); err != nil && !os.IsNotExist(err) {
			return modified, fmt.Errorf("deleting registry file %s: %w", src, err)
		}
		modified = true
	}
	lf.Entries = remaining
	return modified, nil
}

func targetSet(targets []string) map[string]bool {
	set := make(map[string]bool, len(targets))
	for _, t := range targets {
		set[t] = true
	}
	return set
}
