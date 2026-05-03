package importcmd

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/rolfwessels/skillsync/internal/bundle"
)

type LivePrompter struct{}

func (LivePrompter) PickGroups(all []Group) ([]Group, error) {
	opts := make([]huh.Option[int], len(all))
	for i, g := range all {
		opts[i] = huh.NewOption(g.Title(), i)
	}
	var picked []int
	form := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[int]().
			Title("Untracked bundle files").
			Description("Select items to link or copy into the registry").
			Options(opts...).
			Value(&picked),
	))
	if err := form.Run(); err != nil {
		return nil, err
	}
	var out []Group
	for _, i := range picked {
		if i < 0 || i >= len(all) {
			continue
		}
		out = append(out, all[i])
	}
	return out, nil
}

func (LivePrompter) Decide(g Group, registryBundles []bundle.Bundle) (Decision, error) {
	mode := "new"
	f := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title(g.Title()).
			Options(
				huh.NewOption("Create new bundle", "new"),
				huh.NewOption("Link to existing bundle", "link"),
			).
			Value(&mode),
	))
	if err := f.Run(); err != nil {
		return Decision{}, err
	}
	if mode == "link" {
		return decideLinkExisting(g, registryBundles)
	}
	return decideCreateNew(g)
}

func decideLinkExisting(g Group, registryBundles []bundle.Bundle) (Decision, error) {
	var opts []huh.Option[string]
	for _, b := range registryBundles {
		if b.Kind != g.Kind {
			continue
		}
		ref := b.Kind + "/" + b.Name
		label := fmt.Sprintf("%s — %s", ref, b.Description)
		opts = append(opts, huh.NewOption(label, ref))
	}
	if len(opts) == 0 {
		return Decision{}, fmt.Errorf("no %s bundles in registry to link", g.Kind)
	}
	var ref string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Existing bundle").
			Options(opts...).
			Value(&ref),
	))
	if err := form.Run(); err != nil {
		return Decision{}, err
	}
	return Decision{LinkExisting: true, BundleRef: ref}, nil
}

func decideCreateNew(g Group) (Decision, error) {
	defaultRef := fmt.Sprintf("%s/%s", g.Kind, g.BundleName)
	var ref string
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("New bundle reference (kind/name)").
			Value(&ref).
			Placeholder(defaultRef).
			Validate(func(s string) error {
				if s == "" {
					s = defaultRef
				}
				_, _, err := bundle.ParseRef(s)
				return err
			}),
	))
	if err := form.Run(); err != nil {
		return Decision{}, err
	}
	if ref == "" {
		ref = defaultRef
	}
	if _, _, err := bundle.ParseRef(ref); err != nil {
		return Decision{}, err
	}
	return Decision{LinkExisting: false, BundleRef: ref}, nil
}

func (LivePrompter) Confirm(summary string) error {
	ok := true
	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(summary).
			Affirmative("Apply").
			Negative("Cancel").
			Value(&ok),
	))
	if err := form.Run(); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("cancelled")
	}
	return nil
}
