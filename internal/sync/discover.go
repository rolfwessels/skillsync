package sync

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type PotentialTarget struct {
	Kind       string
	Format     string
	BundleName string
	InnerPath  string
	TargetRel  string
	AbsPath    string
}

func bundleKindsOrdered() []string {
	ks := make([]string, 0, len(formatRegistry))
	for k := range formatRegistry {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func EnumeratePotentialTargets(projectRoot string, formats []string) ([]PotentialTarget, error) {
	var out []PotentialTarget
	for _, kind := range bundleKindsOrdered() {
		for _, format := range formats {
			spec, err := lookupSpec(kind, format)
			if err != nil {
				return nil, err
			}
			rootAbs := filepath.Join(projectRoot, filepath.FromSlash(spec.dir))
			fi, err := os.Stat(rootAbs)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, fmt.Errorf("stat %s: %w", rootAbs, err)
			}
			if !fi.IsDir() {
				continue
			}
			switch spec.kind {
			case targetDir:
				part, err := scanDirBundles(kind, format, spec, rootAbs)
				if err != nil {
					return nil, err
				}
				out = append(out, part...)
			case targetFile:
				part, err := scanFlatFiles(kind, format, spec, rootAbs)
				if err != nil {
					return nil, err
				}
				out = append(out, part...)
			}
		}
	}
	return out, nil
}

func scanDirBundles(kind, format string, spec pathSpec, rootAbs string) ([]PotentialTarget, error) {
	entries, err := os.ReadDir(rootAbs)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", rootAbs, err)
	}
	var out []PotentialTarget
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		bundleAbs := filepath.Join(rootAbs, name)
		err := filepath.WalkDir(bundleAbs, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || d.Name() == "bundle.toml" {
				return nil
			}
			inner, relErr := filepath.Rel(bundleAbs, path)
			if relErr != nil {
				return relErr
			}
			inner = filepath.ToSlash(inner)
			out = append(out, PotentialTarget{
				Kind:       kind,
				Format:     format,
				BundleName: name,
				InnerPath:  inner,
				TargetRel:  spec.dir + "/" + name + "/" + inner,
				AbsPath:    path,
			})
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walking %s: %w", bundleAbs, err)
		}
	}
	return out, nil
}

func scanFlatFiles(kind, format string, spec pathSpec, rootAbs string) ([]PotentialTarget, error) {
	entries, err := os.ReadDir(rootAbs)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", rootAbs, err)
	}
	var out []PotentialTarget
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if spec.ext != "" && !strings.HasSuffix(name, spec.ext) {
			continue
		}
		stem := strings.TrimSuffix(name, spec.ext)
		targetRel := spec.dir + "/" + stem + spec.ext
		out = append(out, PotentialTarget{
			Kind:       kind,
			Format:     format,
			BundleName: stem,
			TargetRel:  targetRel,
			AbsPath:    filepath.Join(rootAbs, name),
		})
	}
	return out, nil
}
