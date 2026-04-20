package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/paulplee/dockyard/internal/config"
	"github.com/paulplee/dockyard/internal/dockercmd"
	"github.com/paulplee/dockyard/internal/template"
	"github.com/paulplee/dockyard/internal/volumes"
	"github.com/spf13/cobra"
)

func newDeployCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy <name>",
		Short: "Create host volumes, stage build context, then `docker compose up -d --build`",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(args[0], true)
		},
	}
}

// stageAndEnv prepares the build directory and returns (buildDir, envPath, extraEnv).
// The shared build directory lives at $VolumesBase/build/.
func stageAndEnv(name string) (*config.Global, *config.Deployment, *template.Manifest, string, string, []string, error) {
	g, err := mustLoadGlobal()
	if err != nil {
		return nil, nil, nil, "", "", nil, err
	}
	d, err := config.LoadDeployment(g.VolumesRoot, name)
	if err != nil {
		return nil, nil, nil, "", "", nil, fmt.Errorf("load deployment %q: %w", name, err)
	}
	if d == nil {
		return nil, nil, nil, "", "", nil, fmt.Errorf("unknown deployment %q — run 'dockyard create' first", name)
	}
	if d.Template == "" {
		return nil, nil, nil, "", "", nil, fmt.Errorf("deployment %q has no template set — re-run 'dockyard create' or edit config.yaml", name)
	}
	m, err := template.LoadManifest(d.Template)
	if err != nil {
		return nil, nil, nil, "", "", nil, err
	}
	base := d.Base(g.VolumesRoot)
	buildDir := filepath.Join(base, "build")
	// Ensure build dir exists and is writable by the calling user (the
	// deployment base may be root-owned from a legacy Make setup).
	if err := config.MkdirAllPrivileged(buildDir, 0o755); err != nil {
		return nil, nil, nil, "", "", nil, fmt.Errorf("create build dir: %w", err)
	}
	if err := config.ChownToSelf(buildDir); err != nil {
		return nil, nil, nil, "", "", nil, fmt.Errorf("chown build dir: %w", err)
	}
	if err := template.StageBuildContext(d.Template, buildDir); err != nil {
		return nil, nil, nil, "", "", nil, fmt.Errorf("stage build context: %w", err)
	}
	// Re-write .env so it is always in sync with config.yaml.
	if err := d.WriteEnvFile(g.VolumesRoot); err != nil {
		return nil, nil, nil, "", "", nil, err
	}
	envPath := config.DeploymentEnvPath(g.VolumesRoot, name)
	extraEnv := []string{
		"VOLUMES_BASE=" + base,
		"CONTAINER_NAME=" + d.ContainerName,
		"AGENT_UID=" + strconv.Itoa(d.AgentUID),
		"AGENT_GID=" + strconv.Itoa(d.AgentGID),
		"SSH_PORT=" + strconv.Itoa(d.SSHPort),
	}
	for k, v := range d.BuildArgs {
		extraEnv = append(extraEnv, k+"="+v)
	}
	return g, d, m, buildDir, envPath, extraEnv, nil
}

// runDeploy implements `dockyard deploy` and is shared by the `up` alias.
func runDeploy(name string, withBuild bool) error {
	g, d, m, buildDir, envPath, extraEnv, err := stageAndEnv(name)
	if err != nil {
		return err
	}
	if err := config.EnsureGroup("agents", d.AgentGID); err != nil {
		return err
	}
	if err := volumes.Prepare(d.Base(g.VolumesRoot), m, d.AgentUID, d.AgentGID); err != nil {
		return err
	}
	args := []string{"up", "-d"}
	if withBuild {
		args = append(args, "--build")
	}
	return dockercmd.Compose(buildDir, envPath, extraEnv, args...)
}

// newUpCmd brings the stack up without forcing a rebuild.
func newUpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "up <name>",
		Short: "Start the deployment (no rebuild)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, _, buildDir, envPath, extraEnv, err := stageAndEnv(args[0])
			if err != nil {
				return err
			}
			return dockercmd.Compose(buildDir, envPath, extraEnv, "up", "-d")
		},
	}
}

func newDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down <name>",
		Short: "Stop the deployment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := mustLoadGlobal()
			if err != nil {
				return err
			}
			name := args[0]
			envPath := config.DeploymentEnvPath(g.VolumesRoot, name)
			buildDir := filepath.Join(config.DeploymentDir(g.VolumesRoot, name), "build")
			if _, err := os.Stat(buildDir); err != nil {
				return fmt.Errorf("deployment %q has not been deployed yet (no %s)", name, buildDir)
			}
			return dockercmd.Compose(buildDir, envPath, nil, "down")
		},
	}
}

func newRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart <name>",
		Short: "Restart the deployment (down + up, no rebuild)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			g, err := mustLoadGlobal()
			if err != nil {
				return err
			}
			envPath := config.DeploymentEnvPath(g.VolumesRoot, name)
			buildDir := filepath.Join(config.DeploymentDir(g.VolumesRoot, name), "build")
			if err := dockercmd.Compose(buildDir, envPath, nil, "down"); err != nil {
				return err
			}
			_, _, _, _, _, extraEnv, err := stageAndEnv(name)
			if err != nil {
				return err
			}
			return dockercmd.Compose(buildDir, envPath, extraEnv, "up", "-d")
		},
	}
}
