package cli

import (
	"fmt"
	"os/user"

	"github.com/paulplee/dockyard/internal/config"
	"github.com/paulplee/dockyard/internal/prompt"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Configure this machine (choose a volumes root)",
		Long:  "Writes ~/.config/dockyard/config.yaml with the path that will hold all deployment volumes.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			existing, err := config.LoadGlobal()
			if err != nil {
				return err
			}
			p := prompt.New()
			defVol := ""
			if existing != nil {
				defVol = existing.VolumesRoot
			}
			if defVol == "" {
				u, _ := user.Current()
				defVol = u.HomeDir + "/.config/dockyard/volumes"
			}
			root, err := p.String("Volumes root path", defVol)
			if err != nil {
				return err
			}
			g := &config.Global{VolumesRoot: root}
			if err := g.Save(); err != nil {
				return err
			}
			// Ensure the volumes root exists and is owned by the calling user.
			if err := config.MkdirAllPrivileged(root, 0o755); err != nil {
				return fmt.Errorf("create volumes root: %w", err)
			}
			if err := config.ChownToSelf(root); err != nil {
				return fmt.Errorf("chown volumes root: %w", err)
			}
			gp, _ := config.GlobalPath()
			fmt.Printf("Wrote %s (volumes_root=%s)\n", gp, root)
			return nil
		},
	}
}

// mustLoadGlobal returns the global config or errors if the machine has not
// been initialised yet.
func mustLoadGlobal() (*config.Global, error) {
	g, err := config.LoadGlobal()
	if err != nil {
		return nil, err
	}
	if g == nil || g.VolumesRoot == "" {
		return nil, fmt.Errorf("dockyard is not configured — run 'dockyard init' first")
	}
	return g, nil
}
