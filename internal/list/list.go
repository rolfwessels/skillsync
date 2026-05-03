package list

import (
	"fmt"
	"io"
	"sort"

	"github.com/rolfwessels/skillsync/internal/bundle"
)

func Run(registryRoot string, installed []string, w io.Writer) error {
	bundles, err := bundle.Walk(registryRoot)
	if err != nil {
		return fmt.Errorf("walking registry: %w", err)
	}

	installedSet := toSet(installed)
	byKind := groupByKind(bundles)

	kinds := sortedKeys(byKind)
	for i, kind := range kinds {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, kind)
		group := byKind[kind]
		sort.Slice(group, func(a, b int) bool { return group[a].Name < group[b].Name })
		for _, b := range group {
			key := kind + "/" + b.Name
			marker := "  "
			if installedSet[key] {
				marker = "✓ "
			}
			fmt.Fprintf(w, "  %s%-20s %s\n", marker, b.Name, b.Description)
		}
	}
	return nil
}

func groupByKind(bundles []bundle.Bundle) map[string][]bundle.Bundle {
	m := make(map[string][]bundle.Bundle)
	for _, b := range bundles {
		m[b.Kind] = append(m[b.Kind], b)
	}
	return m
}

func sortedKeys(m map[string][]bundle.Bundle) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}
