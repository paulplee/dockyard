// Package cli wires up cobra subcommands for the dockyard binary.
package cli

import (
	"github.com/paulplee/dockyard/pkg/dockyard"
	"github.com/spf13/cobra"
)

// NewRootCmd builds the root command with all subcommands attached.
// Pass DefaultEngine() for the standard dockyard CLI; downstream
// products provide a customised engine.
func NewRootCmd(engine *dockyard.Engine, version string) *cobra.Command {
	root := &cobra.Command{
		Use:           engine.Name,
		Short:         "Deploy and manage Dockerised development containers",
		Long:          engine.Name + " builds and runs per-template Docker containers with persistent volumes, SSH access, and host-user-mapped UIDs.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		newInitCmd(engine),
		newTemplatesCmd(),
		newCreateCmd(engine),
		newDeployCmd(engine),
		newUpCmd(engine),
		newDownCmd(engine),
		newRestartCmd(engine),
		newStatusCmd(engine),
		newListCmd(engine),
		newShellCmd(engine),
		newLogsCmd(engine),
		newRmCmd(engine),
	)
	return root
}