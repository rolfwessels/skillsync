package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const configDir = ".skillsync"
const configFile = "config.toml"
const gitConfigDir = ".git/.skillsync"

var ErrNotFound = errors.New("config not found: no .skillsync/config.toml or .git/.skillsync/config.toml")

type ProjectConfig struct {
	Registry string   `toml:"registry"`
	Formats  []string `toml:"formats"`
	Bundles  []string `toml:"bundles"`
}

func ConfigPath(projectRoot string) string {
	return filepath.Join(projectRoot, configDir, configFile)
}

func gitConfigPath(projectRoot string) string {
	return filepath.Join(projectRoot, gitConfigDir, configFile)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func Load(projectRoot string) (ProjectConfig, error) {
	primary := ConfigPath(projectRoot)
	secondary := gitConfigPath(projectRoot)
	hasPrimary := exists(primary)
	hasSecondary := exists(secondary)

	switch {
	case hasPrimary && hasSecondary:
		return ProjectConfig{}, fmt.Errorf("ambiguous config: both %s and %s exist", primary, secondary)
	case hasPrimary:
		return decode(primary)
	case hasSecondary:
		return decode(secondary)
	default:
		return ProjectConfig{}, ErrNotFound
	}
}

func decode(path string) (ProjectConfig, error) {
	var cfg ProjectConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return ProjectConfig{}, fmt.Errorf("decoding config %s: %w", path, err)
	}
	return cfg, nil
}

func Write(projectRoot string, cfg ProjectConfig) error {
	path := ConfigPath(projectRoot)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}
	return save(path, cfg)
}

func Update(projectRoot string, cfg ProjectConfig) error {
	return save(ConfigPath(projectRoot), cfg)
}

func GlobalRegistry(homeDir string) string {
	path := filepath.Join(homeDir, configDir, configFile)
	if !exists(path) {
		return ""
	}
	var cfg struct {
		Defaults struct {
			Registry string `toml:"registry"`
		} `toml:"defaults"`
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return ""
	}
	return cfg.Defaults.Registry
}

func normalize(cfg ProjectConfig) ProjectConfig {
	cfg.Registry = strings.TrimSpace(cfg.Registry)
	return cfg
}

func save(path string, cfg ProjectConfig) error {
	cfg = normalize(cfg)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	return nil
}
