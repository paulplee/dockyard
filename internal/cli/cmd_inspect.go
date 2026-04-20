package cli

import (
	"fmt"
	"os"

	"github.com/paulplee/dockyard/internal/config"
	"github.com/paulplee/dockyard/internal/dockercmd"
	"github.com/paulplee/dockyard/internal/prompt"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [name]",
		Short: "Show container state for one or all deployments",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := mustLoadGlobal()
			if err != nil {
				return err
			}
			var names []string
			if len(args) == 1 {
				names = []string{args[0]}
			} else {
				names, err = config.ListDeployments(g.VolumesRoot)
				if err != nil {
					return err
				}
			}
			if len(names) == 0 {
				fmt.Println("(no deployments — run 'dockyard create' first)")
				return nil
			}
			fmt.Printf("%-20s  %-10s  %-6s  %-10s  %s\n", "NAME", "TEMPLATE", "PORT", "STATE", "CONTAINER")
			for _, n := range names {
				d, err := config.LoadDeployment(g.VolumesRoot, n)
				if err != nil {
					fmt.Printf("%-20s  (error: %v)\n", n, err)
					continue
				}
				if d == nil {
					continue
				}
				state, _ := dockercmd.ContainerState("dy-" + d.ContainerName)
				fmt.Printf("%-20s  %-10s  %-6d  %-10s  dy-%s\n",
					d.Name, d.Template, d.SSHPort, state, d.ContainerName)
			}
			return nil
		},
	}
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all deployments (alias of `status`)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return newStatusCmd().RunE(cmd, nil)
		},
	}
}

func newShellCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shell <name>",
		Short: "Open an interactive bash in the container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := mustLoadGlobal()
			if err != nil {
				return err
			}
			d, err := config.LoadDeployment(g.VolumesRoot, args[0])
			if err != nil {
				return err
			}
			if d == nil {
				return fmt.Errorf("unknown deployment %q", args[0])
			}
				u := d.GetAgentUser()
				return dockercmd.Exec("dy-"+d.ContainerName, u, "/home/"+u, "bash")
		},
	}
}

func newLogsCmd() *cobra.Command {
	var follow bool
	cmd := &cobra.Command{
		Use:   "logs <name>",
		Short: "Tail the container's docker logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := mustLoadGlobal()
			if err != nil {
				return err
			}
			d, err := config.LoadDeployment(g.VolumesRoot, args[0])
			if err != nil {
				return err
			}
			if d == nil {
				return fmt.Errorf("unknown deployment %q", args[0])
			}
			return dockercmd.Logs("dy-"+d.ContainerName, follow)
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow log output")
	return cmd
}

func newRmCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "rm <name>",
		Short: "Stop the container and delete the deployment's files (DANGEROUS)",
		Long:  "Runs `docker compose down --rmi all` then removes $VolumesBase entirely. Prompts unless --force is given.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := mustLoadGlobal()
			if err != nil {
				return err
			}
			name := args[0]
			d, err := config.LoadDeployment(g.VolumesRoot, name)
			if err != nil {
				return err
			}
			if d == nil {
				return fmt.Errorf("unknown deployment %q", name)
			}
			base := d.Base(g.VolumesRoot)
			if !force {
				p := prompt.New()
				ok, err := p.Confirm(fmt.Sprintf("Remove container dy-%s and delete %s?", d.ContainerName, base), false)
				if err != nil {
					return err
				}
				if !ok {
					fmt.Println("aborted")
					return nil
				}
			}
			buildDir := base + "/build"
			if _, err := os.Stat(buildDir); err == nil {
				_ = dockercmd.Compose(buildDir, config.DeploymentEnvPath(g.VolumesRoot, name),
					nil, "down", "--rmi", "all", "--remove-orphans")
			}
			if err := config.RemoveAllPrivileged(base); err != nil {
				return fmt.Errorf("remove %s: %w", base, err)
			}
			fmt.Printf("removed %s\n", base)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")
	return cmd
}
