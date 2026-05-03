package sync

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rolfwessels/skillsync/internal/bundle"
	"github.com/rolfwessels/skillsync/internal/config"
)

var ErrConflict = errors.New("sync conflict")

// Sync resolves divergence between the registry and the project in a single step.
// It evaluates each (bundle, format) entry and applies pull, push, or reports conflict.
func Sync(projectRoot, registryRoot string, cfg config.ProjectConfig, out, warnings io.Writer) error {
	lf, err := loadLock(projectRoot)
	if err != nil {
		return err
	}

	var conflicts []string
	for _, bundleRef := range cfg.Bundles {
		kind, name, err := bundle.ParseRef(bundleRef)
		if err != nil {
			fmt.Fprintf(warnings, "warning: skipping %q: %v\n", bundleRef, err)
			continue
		}
		bundleDir := filepath.Join(registryRoot, kind, name)
		if _, err := os.Stat(bundleDir); os.IsNotExist(err) {
			fmt.Fprintf(warnings, "warning: bundle %q not found in registry, skipping\n", bundleRef)
			continue
		}
		for _, format := range cfg.Formats {
			cs, err := syncEvaluate(projectRoot, bundleDir, kind, name, format, bundleRef, &lf)
			if err != nil {
				return fmt.Errorf("evaluating %s for format %s: %w", bundleRef, format, err)
			}
			conflicts = append(conflicts, cs...)
		}
	}

	if len(conflicts) > 0 {
		for _, c := range conflicts {
			fmt.Fprintf(out, "conflict: %s\n", c)
		}
		return fmt.Errorf("%w: %d file(s) have changes on both sides", ErrConflict, len(conflicts))
	}

	return saveLock(projectRoot, lf)
}

func syncEvaluate(projectRoot, bundleDir, kind, name, format, bundleRef string, lf *lockfile) ([]string, error) {
	spec, err := lookupSpec(kind, format)
	if err != nil {
		return nil, err
	}
	contentFiles, err := bundleContentFiles(bundleDir)
	if err != nil {
		return nil, err
	}

	var conflicts []string
	for _, srcPath := range contentFiles {
		targetPath := resolveTarget(spec, bundleDir, name, srcPath)
		absTarget := filepath.Join(projectRoot, filepath.FromSlash(targetPath))

		data, err := os.ReadFile(srcPath)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", srcPath, err)
		}
		currentRegistryHash := hashBytes(data)

		diskHash, _, err := hashOfFileIfExists(absTarget)
		if err != nil {
			return nil, fmt.Errorf("hashing %s: %w", absTarget, err)
		}

		existing, hasLock := lf.lookup(bundleRef, format, targetPath)
		if hasLock && existing.LocalHash == "" {
			existing.LocalHash = existing.Hash
		}
		lockRegistryHash := existing.RegistryHash
		lockLocalHash := existing.LocalHash
		if !hasLock {
			lockRegistryHash = ""
			lockLocalHash = ""
		}

		state := evaluateState(currentRegistryHash, diskHash, lockRegistryHash, lockLocalHash)

		switch state {
		case StateClean:
			// no-op
		case StateRegistryModified:
			if err := applyPull(projectRoot, srcPath, absTarget, kind, format, bundleRef, targetPath, lf); err != nil {
				return nil, err
			}
		case StateLocalModified:
			if err := applyPush(projectRoot, bundleDir, kind, name, format, bundleRef, targetPath, absTarget, lf); err != nil {
				return nil, err
			}
		case StateConflict:
			conflicts = append(conflicts, targetPath)
			lf.upsertState(bundleRef, format, targetPath, StateConflict)
		}
	}
	return conflicts, nil
}

func applyPull(projectRoot, srcPath, absTarget, kind, format, bundleRef, targetPath string, lf *lockfile) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", srcPath, err)
	}
	registryHash := hashBytes(data)
	out, err := applyContentTransform(kind, format, data)
	if err != nil {
		return fmt.Errorf("transform %s for format %s: %w", srcPath, format, err)
	}
	hash := hashBytes(out)
	if err := writeFile(absTarget, out); err != nil {
		return err
	}
	lf.upsert(LockEntry{
		Bundle:       bundleRef,
		Format:       format,
		Target:       targetPath,
		Hash:         hash,
		RegistryHash: registryHash,
		LocalHash:    hash,
		State:        StateClean,
	})
	return nil
}

func applyPush(projectRoot, bundleDir, kind, name, format, bundleRef, targetPath, absTarget string, lf *lockfile) error {
	data, err := os.ReadFile(absTarget)
	if err != nil {
		return fmt.Errorf("reading project file %s: %w", targetPath, err)
	}
	diskHash := hashBytes(data)

	spec, err := lookupSpec(kind, format)
	if err != nil {
		return err
	}
	contentFiles, err := bundleContentFiles(bundleDir)
	if err != nil {
		return err
	}
	for _, p := range contentFiles {
		if resolveTarget(spec, bundleDir, name, p) == targetPath {
			canonical, err := applyReverseContentTransform(kind, format, data)
			if err != nil {
				return fmt.Errorf("reverse transform %s: %w", targetPath, err)
			}
			if err := writeFile(p, canonical); err != nil {
				return fmt.Errorf("writing registry file: %w", err)
			}
			registryHash := hashBytes(canonical)
			lf.upsert(LockEntry{
				Bundle:       bundleRef,
				Format:       format,
				Target:       targetPath,
				Hash:         diskHash,
				RegistryHash: registryHash,
				LocalHash:    diskHash,
				State:        StateClean,
			})
			return nil
		}
	}
	return fmt.Errorf("no bundle file maps to target %q", targetPath)
}
