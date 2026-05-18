package init

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/charmbracelet/huh"

	"github.com/rolfwessels/skillsync/internal/bundle"
	"github.com/rolfwessels/skillsync/internal/config"
)

type TUIPrompter struct {
	PromptForGitHooks bool
}

func (t TUIPrompter) Ask(projectRoot string, defaults config.ProjectConfig) (PromptResult, error) {
	registry := defaults.Registry
	if registry == "" {
		registry = filepath.Join(os.Getenv("HOME"), ".skillsync", "registry")
	}
	formats := defaults.Formats
	bundles := defaults.Bundles
	installHooks := true

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Registry path").
				Value(&registry),
			huh.NewMultiSelect[string]().
				Title("Formats").
				Options(
					huh.NewOption("claude", "claude").Selected(slices.Contains(formats, "claude")),
					huh.NewOption("cursor", "cursor").Selected(slices.Contains(formats, "cursor")),
				).
				Validate(func(v []string) error {
					if len(v) == 0 {
						return fmt.Errorf("select at least one format")
					}
					return nil
				}).
				Value(&formats),
		),
	).Run(); err != nil {
		return PromptResult{}, fmt.Errorf("aborted: %w", err)
	}

	bundleOptions, err := buildBundleOptions(registry, bundles)
	if err != nil {
		return PromptResult{}, err
	}

	var groups []*huh.Group
	if len(bundleOptions) > 0 {
		groups = append(groups, huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Bundles to include").
				Options(bundleOptions...).
				Value(&bundles),
		))
	}
	if t.PromptForGitHooks && isGitWorkTree(projectRoot) {
		groups = append(groups, huh.NewGroup(
			huh.NewConfirm().
				Title("Install auto-sync git hooks? (Y/n)").
				Affirmative("Yes").
				Negative("No").
				Value(&installHooks),
		))
	}
	if len(groups) > 0 {
		if err := huh.NewForm(groups...).Run(); err != nil {
			return PromptResult{}, fmt.Errorf("aborted: %w", err)
		}
	}

	return PromptResult{
		Config: config.ProjectConfig{
			Registry: registry,
			Formats:  formats,
			Bundles:  bundles,
		},
		InstallGitHooks: installHooks,
	}, nil
}

func isGitWorkTree(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil
}

func buildBundleOptions(registry string, selected []string) ([]huh.Option[string], error) {
	bs, err := bundle.Walk(registry)
	if err != nil {
		return nil, nil
	}
	opts := make([]huh.Option[string], 0, len(bs))
	for _, b := range bs {
		label := fmt.Sprintf("%-12s %s  %v", b.Name, b.Description, b.Tags)
		value := b.Kind + "/" + b.Name
		opts = append(opts, huh.NewOption(label, value).Selected(slices.Contains(selected, value)))
	}
	return opts, nil
}
