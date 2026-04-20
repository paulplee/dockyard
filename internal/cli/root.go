// Package cli wires up cobra subcommands for the dockyard binary.
package cli

import "github.com/spf13/cobra"

// NewRootCmd builds the root `dockyard` command with all subcommands attached.
func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "dockyard",
		Short:         "Deploy and manage Dockerised development containers",
		Long:          "dockyard builds and runs per-template Docker containers with persistent volumes, SSH access, and host-user-mapped UIDs.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		newInitCmd(),
		newTemplatesCmd(),
		newCreateCmd(),
		newDeployCmd(),
		newUpCmd(),
		newDownCmd(),
		newRestartCmd(),
		newStatusCmd(),
		newListCmd(),
		newShellCmd(),
		newLogsCmd(),
		newRmCmd(),
	)
	return root
}
