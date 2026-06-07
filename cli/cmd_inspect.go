package cli

import (
	"fmt"
	"os"

	"github.com/paulplee/dockyard/config"
	"github.com/paulplee/dockyard/dockercmd"
	"github.com/paulplee/dockyard/pkg/dockyard"
	"github.com/paulplee/dockyard/prompt"
	"github.com/spf13/cobra"
)

func newStatusCmd(engine *dockyard.Engine) *cobra.Command {
	return &cobra.Command{
		Use:   "status [name]",
		Short: "Show status of deployments",
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := mustLoadGlobal(engine)
			if err != nil {
				return err
			}
			var names []string
			if len(args) > 0 {
				names = args
				for _, n := range names {
					if d, _ := config.LoadDeployment(g.VolumesRoot, n); d == nil {
						fmt.Printf("unknown deployment %q\n", n)
						continue
					}
				}
			} else {
				names, _ = config.ListDeployments(g.VolumesRoot)
			}
			if len(names) == 0 {
				fmt.Printf("(no deployments — run '%s create' first)\n", engine.Name)
				return nil
			}
			fmt.Printf("%-20s  %-10s  %-6s  %-10s  %s\n", "NAME", "TEMPLATE", "PORT", "STATE", "CONTAINER")
			for _, n := range names {
				d, err := config.LoadDeployment(g.VolumesRoot, n)
				if err != nil {
					return err
				}
				if d == nil {
					continue
				}
				state, _ := dockercmd.ContainerState(engine.ContainerName(d.ContainerName))
				fmt.Printf("%-20s  %-10s  %-6d  %-10s  %s\n",
					d.Name, d.Template, d.SSHPort, state, engine.ContainerName(d.ContainerName))
			}
			return nil
		},
	}
}

func newListCmd(engine *dockyard.Engine) *cobra.Command {
	cmd := newStatusCmd(engine)
	cmd.Use = "list"
	cmd.Short = "List all deployments (alias for status)"
	return cmd
}

func newShellCmd(engine *dockyard.Engine) *cobra.Command {
	return &cobra.Command{
		Use:   "shell <name>",
		Short: "Start an interactive shell inside the container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := mustLoadGlobal(engine)
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
			u := d.GetAgentUser(engine)
			uid := d.AgentUID
			envVars := []string{fmt.Sprintf("XDG_RUNTIME_DIR=/run/user/%d", uid)}
			return dockercmd.Exec(engine.ContainerName(d.ContainerName), u, "/home/"+u, envVars, "bash")
		},
	}
}

func newLogsCmd(engine *dockyard.Engine) *cobra.Command {
	var follow bool
	cmd := &cobra.Command{
		Use:   "logs <name>",
		Short: "Tail container logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := mustLoadGlobal(engine)
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
			return dockercmd.Logs(engine.ContainerName(d.ContainerName), follow)
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow log output")
	return cmd
}

func newRmCmd(engine *dockyard.Engine) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a deployment (stop container, delete images and volume data)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			g, err := mustLoadGlobal(engine)
			if err != nil {
				return err
			}
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
				ok, err := p.Confirm(fmt.Sprintf("Remove container %s and delete %s?",
					engine.ContainerName(d.ContainerName), base), false)
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}
			buildDir := base + "/build"
			if _, err := os.Stat(buildDir); err == nil {
				// --rmi local removes only locally-built images (agent, litellm).
				// --rmi all would also delete pulled images (open-webui, pgvector,
				// debian) forcing a full re-download on next deploy.
				_ = dockercmd.Compose(buildDir, config.DeploymentEnvPath(g.VolumesRoot, name),
					nil, "down", "--rmi", "local", "--remove-orphans")
			}
			if err := config.RemoveAllPrivileged(base); err != nil {
				return fmt.Errorf("remove %s: %w", base, err)
			}
			fmt.Printf("Removed %s (deployment %q)\n", base, name)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")
	return cmd
}