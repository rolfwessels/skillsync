package importcmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"

	"github.com/rolfwessels/skillsync/internal/bundle"
	"github.com/rolfwessels/skillsync/internal/config"
	skillssync "github.com/rolfwessels/skillsync/internal/sync"
)

type canonicalFile struct {
	content []byte
	regRel  string
}

var ErrRegistryBundleNotFound = errors.New("bundle not found in registry")

type Decision struct {
	LinkExisting bool
	BundleRef    string
}

type Prompter interface {
	PickGroups(all []Group) ([]Group, error)
	Decide(group Group, registryBundles []bundle.Bundle) (Decision, error)
	Confirm(summary string) error
}

type importPlan struct {
	group Group
	dec   Decision
}

func Run(projectRoot string, cfg config.ProjectConfig, registryRoot string, p Prompter, warnings io.Writer) error {
	groups, err := UntrackedGroups(projectRoot, cfg)
	if err != nil {
		return err
	}
	if len(groups) == 0 {
		fmt.Fprintln(warnings, "no untracked bundle files found")
		return nil
	}
	picked, err := p.PickGroups(groups)
	if err != nil {
		return fmt.Errorf("selection aborted: %w", err)
	}
	if len(picked) == 0 {
		fmt.Fprintln(warnings, "nothing selected")
		return nil
	}
	regBundles, err := bundle.Walk(registryRoot)
	if err != nil {
		return fmt.Errorf("walking registry: %w", err)
	}
	var plans []importPlan
	for _, g := range picked {
		dec, err := p.Decide(g, regBundles)
		if err != nil {
			return fmt.Errorf("decision aborted: %w", err)
		}
		if _, _, err := bundle.ParseRef(dec.BundleRef); err != nil {
			return fmt.Errorf("invalid bundle ref %q: %w", dec.BundleRef, err)
		}
		if err := bundleStemMatches(g, dec); err != nil {
			return err
		}
		plans = append(plans, importPlan{group: g, dec: dec})
	}
	if err := p.Confirm(fmt.Sprintf("Import %d item(s)", len(plans))); err != nil {
		return fmt.Errorf("aborted: %w", err)
	}
	return applyPlans(projectRoot, registryRoot, cfg, plans, warnings)
}

func applyPlans(projectRoot, registryRoot string, baseCfg config.ProjectConfig, plans []importPlan, warnings io.Writer) error {
	cfgPath := config.ConfigPath(projectRoot)
	cfgSnap, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("reading config for rollback: %w", err)
	}
	lockPath := filepath.Join(projectRoot, ".skillsync", "sync.lock")
	lockSnap, lockErr := os.ReadFile(lockPath)
	lockHad := lockErr == nil

	rollback := func() {
		_ = os.WriteFile(cfgPath, cfgSnap, 0644)
		if lockHad {
			_ = os.WriteFile(lockPath, lockSnap, 0644)
			return
		}
		_ = os.Remove(lockPath)
	}

	cfg := baseCfg
	cfg.Bundles = append([]string(nil), cfg.Bundles...)
	for _, pl := range plans {
		if err := materializeRegistry(registryRoot, pl.group, pl.dec); err != nil {
			rollback()
			return err
		}
		ref := pl.dec.BundleRef
		if !slices.Contains(cfg.Bundles, ref) {
			cfg.Bundles = append(cfg.Bundles, ref)
		}
	}
	if err := config.Update(projectRoot, cfg); err != nil {
		rollback()
		return fmt.Errorf("updating config: %w", err)
	}
	if err := skillssync.Run(projectRoot, registryRoot, cfg, warnings); err != nil {
		rollback()
		return fmt.Errorf("sync after import: %w", err)
	}
	return nil
}

func materializeRegistry(registryRoot string, g Group, dec Decision) error {
	files, err := canonicalsForGroup(g)
	if err != nil {
		return err
	}
	kind, name, err := bundle.ParseRef(dec.BundleRef)
	if err != nil {
		return err
	}
	if kind != g.Kind {
		return fmt.Errorf("bundle kind %q does not match detected kind %q", kind, g.Kind)
	}
	bundleDir := filepath.Join(registryRoot, kind, name)
	if dec.LinkExisting {
		fi, err := os.Stat(bundleDir)
		if err != nil || !fi.IsDir() {
			return fmt.Errorf("%w: %s", ErrRegistryBundleNotFound, dec.BundleRef)
		}
	} else {
		if fi, err := os.Stat(bundleDir); err == nil && fi.IsDir() {
			return fmt.Errorf("bundle %q already exists in registry", dec.BundleRef)
		}
		if err := os.MkdirAll(bundleDir, 0755); err != nil {
			return fmt.Errorf("creating bundle dir: %w", err)
		}
		desc := descriptionFromFiles(files)
		if err := writeBundleTOML(bundleDir, desc); err != nil {
			return err
		}
	}
	for _, f := range files {
		dest := filepath.Join(bundleDir, f.regRel)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("creating parent dirs: %w", err)
		}
		if err := os.WriteFile(dest, f.content, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", dest, err)
		}
	}
	return nil
}

func writeBundleTOML(bundleDir, description string) error {
	path := filepath.Join(bundleDir, "bundle.toml")
	type meta struct {
		Description string   `toml:"description"`
		Tags        []string `toml:"tags"`
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating bundle.toml: %w", err)
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(meta{Description: description, Tags: []string{}})
}

func canonicalsForGroup(g Group) ([]canonicalFile, error) {
	order := []string{"claude", "cursor"}
	for _, format := range order {
		for _, v := range g.Variants {
			if v.Format != format {
				continue
			}
			var out []canonicalFile
			for _, f := range v.Files {
				data, err := os.ReadFile(f.AbsPath)
				if err != nil {
					return nil, fmt.Errorf("reading %s: %w", f.AbsPath, err)
				}
				canon, err := skillssync.ContentToCanonical(g.Kind, format, data)
				if err != nil {
					return nil, fmt.Errorf("canonicalizing from %s: %w", format, err)
				}
				out = append(out, canonicalFile{
					content: canon,
					regRel:  registryFilename(g.Kind, f.InnerPath),
				})
			}
			return out, nil
		}
	}
	return nil, fmt.Errorf("no variants for group")
}

func descriptionFromFiles(files []canonicalFile) string {
	for _, f := range files {
		if d := extractDescription(f.content); d != "" {
			return truncate(d, 50)
		}
	}
	return "Imported by skillsync"
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

func extractDescription(content []byte) string {
	s := strings.TrimSpace(string(content))
	if !strings.HasPrefix(s, "---") {
		return ""
	}
	rest := s[3:]
	end := strings.Index(rest, "---")
	if end < 0 {
		return ""
	}
	var fm struct {
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(rest[:end]), &fm); err != nil {
		return ""
	}
	return strings.TrimSpace(fm.Description)
}

func bundleStemMatches(g Group, dec Decision) error {
	_, name, err := bundle.ParseRef(dec.BundleRef)
	if err != nil {
		return err
	}
	if name != g.BundleName {
		return fmt.Errorf("bundle ref must use path stem %q (got ref name %q)", g.BundleName, name)
	}
	return nil
}

func registryFilename(kind, innerPath string) string {
	switch kind {
	case "skills":
		return innerPath
	case "rules":
		return "rule.md"
	case "agents":
		return "agent.md"
	case "commands":
		return "command.md"
	default:
		return innerPath
	}
}
