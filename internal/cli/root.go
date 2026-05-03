package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)


var Version = "dev"

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "skillsync",
		Short:   "Sync AI skills, commands, and agents across your projects",
		Version: Version,
		Long: fmt.Sprintf(`skillsync v%s

Keeps your AI assistant bundles (skills, rules, agents) in sync
across all your projects from a central registry.`, Version),
	}

	root.AddCommand(newListCmd())
	root.AddCommand(newPullCmd())
	root.AddCommand(newInitCmd())
	root.AddCommand(newSetCmd())
	root.AddCommand(newAddCmd())
	root.AddCommand(newPushCmd())
	root.AddCommand(newSyncCmd())
	root.AddCommand(newImportCmd())

	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
