package cli

import (
	"fmt"
	"os/user"
	"strings"

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

			// --- Volumes root ---
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

			// --- Domain name (for HTTPS / Caddy) ---
			defDomain := ""
			if existing != nil && existing.DomainName != "" {
				defDomain = existing.DomainName
			}
			if defDomain == "" {
				defDomain = "localhost"
			}
			domainName, err := p.String("Domain name for HTTPS (enter for localhost)", defDomain)
			if err != nil {
				return err
			}
			domainName = strings.TrimSpace(domainName)
			if domainName == "" {
				domainName = "localhost"
			}

			// --- DeepSeek API key ---
			defKey := ""
			if existing != nil && existing.DeepSeekAPIKey != "" {
				defKey = existing.DeepSeekAPIKey[:4] + "…" + existing.DeepSeekAPIKey[len(existing.DeepSeekAPIKey)-4:]
			}
			promptStr := "DeepSeek API key"
			if defKey != "" {
				promptStr += " [" + defKey + "]"
			}
			deepSeekKey, err := p.String(promptStr, defKey)
			if err != nil {
				return err
			}
			deepSeekKey = strings.TrimSpace(deepSeekKey)
			if deepSeekKey == "" && defKey != "" {
				// User pressed Enter — restore the full key.
				deepSeekKey = existing.DeepSeekAPIKey
			}

			// Build and persist the global config.
			g := &config.Global{
				VolumesRoot:    root,
				DomainName:     domainName,
				DeepSeekAPIKey: deepSeekKey,
			}
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
			fmt.Printf("Wrote %s (volumes_root=%s, domain=%s)\n", gp, root, domainName)
			if deepSeekKey != "" {
				fmt.Printf("  DeepSeek API key: set (%s…%s)\n", deepSeekKey[:4], deepSeekKey[len(deepSeekKey)-4:])
			} else {
				fmt.Println("  DeepSeek API key: not set (you can add it later)")
			}
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