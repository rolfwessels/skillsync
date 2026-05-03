package init

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rolfwessels/skillsync/internal/config"
	"github.com/rolfwessels/skillsync/internal/hooks"
	skillssync "github.com/rolfwessels/skillsync/internal/sync"
)

type PromptResult struct {
	Config          config.ProjectConfig
	InstallGitHooks bool
}

type Prompter interface {
	Ask(projectRoot string, defaults config.ProjectConfig) (PromptResult, error)
}

func Run(projectRoot string, prompter Prompter, warnings io.Writer) error {
	if _, err := os.Stat(config.ConfigPath(projectRoot)); err == nil {
		return fmt.Errorf("already initialised: %s exists", config.ConfigPath(projectRoot))
	}
	result, err := prompter.Ask(projectRoot, globalDefaults())
	if err != nil {
		return fmt.Errorf("prompting: %w", err)
	}
	if err := config.Write(projectRoot, result.Config); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := skillssync.Run(projectRoot, result.Config.Registry, result.Config, warnings); err != nil {
		return fmt.Errorf("syncing: %w", err)
	}
	return applyGitHooks(projectRoot, result.InstallGitHooks, warnings)
}

func Reconfigure(projectRoot string, prompter Prompter, warnings io.Writer) error {
	current, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	result, err := prompter.Ask(projectRoot, current)
	if err != nil {
		return fmt.Errorf("prompting: %w", err)
	}
	if err := config.Update(projectRoot, result.Config); err != nil {
		return err
	}
	return applyGitHooks(projectRoot, result.InstallGitHooks, warnings)
}

func globalDefaults() config.ProjectConfig {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config.ProjectConfig{}
	}
	return config.ProjectConfig{Registry: config.GlobalRegistry(homeDir)}
}

func applyGitHooks(projectRoot string, install bool, warnings io.Writer) error {
	if _, err := os.Stat(filepath.Join(projectRoot, ".git")); err != nil {
		fmt.Fprintf(warnings, "skillsync: warning: not a git repository — skipping git hook installation\n")
		return nil
	}
	if !install {
		return nil
	}
	warns, err := hooks.Install(projectRoot)
	if err != nil {
		return fmt.Errorf("installing git hooks: %w", err)
	}
	for _, w := range warns {
		fmt.Fprintf(warnings, "skillsync: warning: %s\n", w)
	}
	return nil
}
