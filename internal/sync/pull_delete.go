package sync

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func currentTargetSet(spec pathSpec, bundleDir, name string, contentFiles []string) map[string]bool {
	set := make(map[string]bool, len(contentFiles))
	for _, src := range contentFiles {
		set[resolveTarget(spec, bundleDir, name, src)] = true
	}
	return set
}

func partitionOrphans(bundleRef, format string, current map[string]bool, entries []LockEntry) (orphans, kept []LockEntry) {
	for _, e := range entries {
		if e.Bundle != bundleRef || e.Format != format || current[e.Target] {
			kept = append(kept, e)
			continue
		}
		orphans = append(orphans, e)
	}
	return orphans, kept
}

func projectFileMatchesLock(projectRoot string, e LockEntry) (matches, exists bool, err error) {
	absTarget := filepath.Join(projectRoot, filepath.FromSlash(e.Target))
	diskHash, exists, err := hashOfFileIfExists(absTarget)
	if err != nil || !exists {
		return false, exists, err
	}
	expected := e.LocalHash
	if expected == "" {
		expected = e.Hash
	}
	return diskHash == expected, true, nil
}

func removeProjectFile(absTarget string) error {
	if err := os.Remove(absTarget); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing orphaned file %s: %w", absTarget, err)
	}
	return nil
}

func pullDeleteOrphans(projectRoot, format, bundleRef string, current map[string]bool, lf *lockfile, warnings io.Writer) error {
	orphans, kept := partitionOrphans(bundleRef, format, current, lf.Entries)
	for _, e := range orphans {
		matches, exists, err := projectFileMatchesLock(projectRoot, e)
		if err != nil {
			return err
		}
		absTarget := filepath.Join(projectRoot, filepath.FromSlash(e.Target))
		switch {
		case !exists:
		case !matches:
			if err := stashLocalCopy(projectRoot, absTarget, warnings); err != nil {
				return err
			}
		default:
			if err := removeProjectFile(absTarget); err != nil {
				return err
			}
		}
	}
	lf.Entries = kept
	return nil
}

func syncDeleteOrphans(projectRoot, format, bundleRef string, current map[string]bool, lf *lockfile) ([]string, error) {
	orphans, kept := partitionOrphans(bundleRef, format, current, lf.Entries)
	var conflicts []string
	for _, e := range orphans {
		matches, exists, err := projectFileMatchesLock(projectRoot, e)
		if err != nil {
			return nil, err
		}
		absTarget := filepath.Join(projectRoot, filepath.FromSlash(e.Target))
		switch {
		case !exists:
		case !matches:
			conflicts = append(conflicts, e.Target)
			e.State = StateConflict
			kept = append(kept, e)
		default:
			if err := removeProjectFile(absTarget); err != nil {
				return nil, err
			}
		}
	}
	lf.Entries = kept
	return conflicts, nil
}
