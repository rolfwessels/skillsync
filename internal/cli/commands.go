package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rolfwessels/skillsync/internal/add"
	"github.com/rolfwessels/skillsync/internal/config"
	"github.com/rolfwessels/skillsync/internal/importcmd"
	skillsinit "github.com/rolfwessels/skillsync/internal/init"
	"github.com/rolfwessels/skillsync/internal/list"
	skillssync "github.com/rolfwessels/skillsync/internal/sync"
)

func newListCmd() *cobra.Command {
	var registryFlag string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available bundles in the registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			cfg, err := config.Load(cwd)
			if err != nil && !errors.Is(err, config.ErrNotFound) {
				return fmt.Errorf("loading config: %w", err)
			}
			registryRoot := registryFlag
			if registryRoot == "" {
				registryRoot = cfg.Registry
			}
			if registryRoot == "" {
				return fmt.Errorf("no registry configured: use --registry or set registry in config")
			}
			var installed []string
			if !errors.Is(err, config.ErrNotFound) {
				installed = cfg.Bundles
			}
			return list.Run(registryRoot, installed, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&registryFlag, "registry", "", "Path to registry (overrides config)")
	return cmd
}

func newPullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Pull bundles from the registry into the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			cfg, err := config.Load(cwd)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			return skillssync.Run(cwd, cfg.Registry, cfg, cmd.ErrOrStderr())
		},
	}
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialise skillsync in the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			return skillsinit.Run(cwd, skillsinit.TUIPrompter{PromptForGitHooks: true}, cmd.ErrOrStderr())
		},
	}
}

func newSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Reconfigure the current project interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			return skillsinit.Reconfigure(cwd, skillsinit.TUIPrompter{PromptForGitHooks: true}, cmd.ErrOrStderr())
		},
	}
}

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <kind>/<name>",
		Short: "Add a bundle to the project config and sync it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			cfg, err := config.Load(cwd)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			return add.Run(cwd, cfg.Registry, args[0], cmd.ErrOrStderr())
		},
	}
}

func newPushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Push local edits back to the registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			cfg, err := config.Load(cwd)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			return skillssync.Push(cwd, cfg.Registry, cfg)
		},
	}
}

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync bundles between the registry and the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			cfg, err := config.Load(cwd)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			return skillssync.Sync(cwd, cfg.Registry, cfg, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}

func newImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Link untracked skill files into the registry and project config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			cfg, err := config.Load(cwd)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			return importcmd.Run(cwd, cfg, cfg.Registry, importcmd.LivePrompter{}, cmd.ErrOrStderr())
		},
	}
}
