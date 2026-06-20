package cli

import (
	"fmt"
	"os/user"

	"github.com/paulplee/dockyard/config"
	"github.com/paulplee/dockyard/pkg/dockyard"
	"github.com/paulplee/dockyard/prompt"
	"github.com/spf13/cobra"
)

func newInitCmd(engine *dockyard.Engine) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Configure this machine (choose a volumes root)",
		Long:  "Writes ~/.config/" + engine.Name + "/config.yaml with the path that will hold all deployment volumes.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			existing, err := config.LoadGlobal(engine)
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
				defVol = u.HomeDir + "/.config/" + engine.Name + "/volumes"
			}
			root, err := p.String("Volumes root path", defVol)
			if err != nil {
				return err
			}
			root = config.ExpandPath(root)
			g := &config.Global{VolumesRoot: root}
			if err := g.Save(engine); err != nil {
				return err
			}
			// Ensure the volumes root exists and is owned by the calling user.
			if err := config.MkdirAllPrivileged(root, 0o755); err != nil {
				return fmt.Errorf("create volumes root: %w", err)
			}
			if err := config.ChownToSelf(root); err != nil {
				return fmt.Errorf("chown volumes root: %w", err)
			}
			gp, _ := config.GlobalPath(engine)
			fmt.Printf("Wrote %s (volumes_root=%s)\n", gp, root)
			return nil
		},
	}
}

// mustLoadGlobal returns the global config or errors if the machine has not
// been initialised yet.
func mustLoadGlobal(engine *dockyard.Engine) (*config.Global, error) {
	g, err := config.LoadGlobal(engine)
	if err != nil {
		return nil, err
	}
	if g == nil || g.VolumesRoot == "" {
		return nil, fmt.Errorf("%s is not configured — run '%s init' first", engine.Name, engine.Name)
	}
	return g, nil
}