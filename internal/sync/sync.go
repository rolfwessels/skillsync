package sync

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/rolfwessels/skillsync/internal/bundle"
	"github.com/rolfwessels/skillsync/internal/config"
)

func Run(projectRoot, registryRoot string, cfg config.ProjectConfig, warnings io.Writer) error {
	lf, err := loadLock(projectRoot)
	if err != nil {
		return err
	}

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
			if err := syncBundle(projectRoot, bundleDir, kind, name, format, bundleRef, &lf, warnings); err != nil {
				return fmt.Errorf("syncing %s for format %s: %w", bundleRef, format, err)
			}
		}
	}

	return saveLock(projectRoot, lf)
}

func syncBundle(projectRoot, bundleDir, kind, name, format, bundleRef string, lf *lockfile, warnings io.Writer) error {
	spec, err := lookupSpec(kind, format)
	if err != nil {
		return err
	}

	contentFiles, err := bundleContentFiles(bundleDir)
	if err != nil {
		return err
	}

	for _, srcPath := range contentFiles {
		targetPath := resolveTarget(spec, bundleDir, name, srcPath)
		absTarget := filepath.Join(projectRoot, filepath.FromSlash(targetPath))

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

		existing, hasLock := lf.lookup(bundleRef, format, targetPath)
		if hasLock && existing.LocalHash == "" {
			existing.LocalHash = existing.Hash
		}
		diskHash, diskExists, err := hashOfFileIfExists(absTarget)
		if err != nil {
			return fmt.Errorf("hashing %s: %w", absTarget, err)
		}

		if hasLock && existing.RegistryHash == registryHash && diskExists && diskHash == existing.LocalHash {
			continue
		}

		if diskExists && hasLock && diskHash != existing.LocalHash {
			if err := stashLocalCopy(projectRoot, absTarget, warnings); err != nil {
				return err
			}
		}

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
	}
	return nil
}

func hashOfFileIfExists(path string) (hash string, exists bool, err error) {
	data, readErr := os.ReadFile(path)
	if os.IsNotExist(readErr) {
		return "", false, nil
	}
	if readErr != nil {
		return "", false, fmt.Errorf("reading file: %w", readErr)
	}
	return hashBytes(data), true, nil
}

func stashLocalCopy(projectRoot, absTarget string, warnings io.Writer) error {
	rel, err := filepath.Rel(projectRoot, absTarget)
	if err != nil {
		return fmt.Errorf("stash relative path: %w", err)
	}
	stamp := fmt.Sprintf("%d", time.Now().UnixNano())
	destAbs := filepath.Join(projectRoot, ".skillsync", "stash", stamp, rel)
	if err := os.MkdirAll(filepath.Dir(destAbs), 0755); err != nil {
		return fmt.Errorf("creating stash dir: %w", err)
	}
	if err := os.Rename(absTarget, destAbs); err != nil {
		return fmt.Errorf("moving file to stash: %w", err)
	}
	stashRel := filepath.Join(".skillsync", "stash", stamp, rel)
	fmt.Fprintf(warnings, "warning: local edits to synced file %s were stashed at %s\n",
		filepath.ToSlash(rel), filepath.ToSlash(stashRel))
	return nil
}

func bundleContentFiles(bundleDir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(bundleDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() == "bundle.toml" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking bundle dir %s: %w", bundleDir, err)
	}
	return files, nil
}

func resolveTarget(spec pathSpec, bundleDir, name, srcPath string) string {
	if spec.kind == targetDir {
		rel, _ := filepath.Rel(bundleDir, srcPath)
		return spec.dir + "/" + name + "/" + filepath.ToSlash(rel)
	}
	return spec.dir + "/" + name + spec.ext
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating dir for %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}
