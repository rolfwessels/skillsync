package importcmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rolfwessels/skillsync/internal/config"
	"github.com/rolfwessels/skillsync/internal/sync"
)

type VariantFile struct {
	InnerPath string
	TargetRel string
	AbsPath   string
}

type Variant struct {
	Format string
	Files  []VariantFile
}

type Group struct {
	Kind           string
	BundleName     string
	Variants       []Variant
	ProjectFormats []string
}

func (g Group) Title() string {
	var vf []string
	for _, v := range g.Variants {
		vf = append(vf, v.Format)
	}
	sort.Strings(vf)
	base := fmt.Sprintf("%s/%s [%s]", g.Kind, g.BundleName, strings.Join(vf, "+"))
	miss := missingFormats(g.Variants, g.ProjectFormats)
	if len(miss) > 0 {
		base += fmt.Sprintf(" — pull will materialise: %s", strings.Join(miss, ", "))
	}
	return base
}

func missingFormats(variants []Variant, cfgFormats []string) []string {
	have := map[string]struct{}{}
	for _, v := range variants {
		have[v.Format] = struct{}{}
	}
	var out []string
	for _, f := range cfgFormats {
		if _, ok := have[f]; !ok {
			out = append(out, f)
		}
	}
	return out
}

func UntrackedGroups(projectRoot string, cfg config.ProjectConfig) ([]Group, error) {
	entries, err := sync.LoadLockEntries(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("loading lockfile: %w", err)
	}
	locked := map[string]struct{}{}
	for _, e := range entries {
		locked[e.Target] = struct{}{}
	}
	potentials, err := sync.EnumeratePotentialTargets(projectRoot, cfg.Formats)
	if err != nil {
		return nil, fmt.Errorf("scanning project paths: %w", err)
	}
	var untracked []sync.PotentialTarget
	for _, p := range potentials {
		if _, ok := locked[p.TargetRel]; ok {
			continue
		}
		untracked = append(untracked, p)
	}
	return mergePotentialTargets(untracked, cfg.Formats), nil
}

func mergePotentialTargets(targets []sync.PotentialTarget, projectFormats []string) []Group {
	type groupKey struct {
		kind, bundle string
	}
	type variantKey struct {
		gk     groupKey
		format string
	}
	groups := map[groupKey]*Group{}
	variantIdx := map[variantKey]int{}

	for _, pt := range targets {
		gk := groupKey{kind: pt.Kind, bundle: pt.BundleName}
		g := groups[gk]
		if g == nil {
			g = &Group{
				Kind:           pt.Kind,
				BundleName:     pt.BundleName,
				ProjectFormats: append([]string(nil), projectFormats...),
			}
			groups[gk] = g
		}
		vk := variantKey{gk: gk, format: pt.Format}
		idx, ok := variantIdx[vk]
		if !ok {
			idx = len(g.Variants)
			g.Variants = append(g.Variants, Variant{Format: pt.Format})
			variantIdx[vk] = idx
		}
		g.Variants[idx].Files = append(g.Variants[idx].Files, VariantFile{
			InnerPath: pt.InnerPath,
			TargetRel: pt.TargetRel,
			AbsPath:   pt.AbsPath,
		})
	}

	keys := make([]groupKey, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		aHasClaude := hasFormat(groups[a], "claude")
		bHasClaude := hasFormat(groups[b], "claude")
		if aHasClaude != bHasClaude {
			return aHasClaude
		}
		if a.kind != b.kind {
			return a.kind < b.kind
		}
		return a.bundle < b.bundle
	})

	out := make([]Group, 0, len(groups))
	for _, k := range keys {
		grp := groups[k]
		sort.Slice(grp.Variants, func(i, j int) bool {
			return grp.Variants[i].Format < grp.Variants[j].Format
		})
		out = append(out, *grp)
	}
	return out
}

func hasFormat(g *Group, format string) bool {
	for _, v := range g.Variants {
		if v.Format == format {
			return true
		}
	}
	return false
}
