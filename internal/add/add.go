package add

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/rolfwessels/skillsync/internal/bundle"
	"github.com/rolfwessels/skillsync/internal/config"
	skillssync "github.com/rolfwessels/skillsync/internal/sync"
)

var ErrAlreadyAdded = errors.New("bundle already in config")
var ErrNotFound = errors.New("bundle not found in registry")

func Run(projectRoot, registryRoot, bundleRef string, warnings io.Writer) error {
	if err := validateBundleExists(registryRoot, bundleRef); err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if slices.Contains(cfg.Bundles, bundleRef) {
		return fmt.Errorf("%w: %s", ErrAlreadyAdded, bundleRef)
	}
	cfg.Bundles = append(cfg.Bundles, bundleRef)
	if err := config.Update(projectRoot, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	return skillssync.Run(projectRoot, registryRoot, cfg, warnings)
}

func validateBundleExists(registryRoot, bundleRef string) error {
	kind, name, err := bundle.ParseRef(bundleRef)
	if err != nil {
		return err
	}
	bundleDir := filepath.Join(registryRoot, kind, name)
	if _, err := os.Stat(bundleDir); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrNotFound, bundleRef)
	}
	return nil
}
