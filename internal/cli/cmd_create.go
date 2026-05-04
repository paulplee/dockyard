package cli

import (
	"fmt"
	"net"
	"os/user"
	"strconv"
	"strings"

	"github.com/paulplee/dockyard/internal/config"
	"github.com/paulplee/dockyard/internal/prompt"
	"github.com/paulplee/dockyard/internal/sshcfg"
	"github.com/paulplee/dockyard/internal/template"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <template> [name]",
		Short: "Interactively configure a new deployment (writes config.yaml + SSH entry)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := mustLoadGlobal()
			if err != nil {
				return err
			}
			tmplName := args[0]
			m, err := template.LoadManifest(tmplName)
			if err != nil {
				return err
			}
			p := prompt.New()
			p.Info("=== Dockyard Create (%s) ===", tmplName)

			name := ""
			if len(args) == 2 {
				name = args[1]
			}
			if name == "" {
				name, err = p.String("Deployment name", "")
				if err != nil {
					return err
				}
			}
			if name == "" {
				return fmt.Errorf("deployment name is required")
			}

			agentUser, err := p.String("Agent username", "dy-user")
			if err != nil {
				return err
			}
			defUID := nextFreeUID(1100)
			uid, err := p.Int("Agent UID", defUID)
			if err != nil {
				return err
			}
			defPort := nextFreePort(2200)
			port, err := p.Int("SSH port on host", defPort)
			if err != nil {
				return err
			}

			buildArgs := map[string]string{}
			volumesBase := config.DeploymentDir(g.VolumesRoot, name)
			for _, ba := range m.BuildArgs {
				if ba.Help != "" {
					p.Info("%s", ba.Help)
				}
				// Expand {VOLUMES_BASE} placeholder in defaults so manifests
				// can reference the deployment's local volumes directory.
				dflt := strings.ReplaceAll(ba.Default, "{VOLUMES_BASE}", volumesBase)
				var val string
				if len(ba.Options) > 0 {
					val, err = p.Choice(ba.Prompt, ba.Options, dflt)
				} else {
					val, err = p.String(ba.Prompt, dflt)
				}
				if err != nil {
					return err
				}
				buildArgs[ba.Name] = val
			}

			d := &config.Deployment{
				Name:          name,
				Template:      tmplName,
				ContainerName: name,
				AgentUser:     agentUser,
				AgentUID:      uid,
				AgentGID:      uid,
				SSHPort:       port,
				BuildArgs:     buildArgs,
			}
			if err := d.Save(g.VolumesRoot); err != nil {
				return err
			}
			p.Info("Wrote %s", config.DeploymentConfigPath(g.VolumesRoot, name))

			// SSH key + config stanza.
			keys, _ := sshcfg.ListPublicKeys()
			var chosen string
			if len(keys) > 0 {
				p.Info("Available SSH public keys:")
				for i, k := range keys {
					p.Info("  %d) %s", i+1, k)
				}
				defKey := keys[0]
				ans, err := p.String("SSH public key to authorize (number or path)", defKey)
				if err != nil {
					return err
				}
				if n, err := strconv.Atoi(ans); err == nil && n >= 1 && n <= len(keys) {
					chosen = keys[n-1]
				} else {
					chosen = ans
				}
			} else {
				p.Info("WARNING: no SSH public keys found in ~/.ssh/ — skipping authorized_keys install")
			}
			base := d.Base(g.VolumesRoot)
			if chosen != "" {
				if err := sshcfg.InstallAuthorizedKeys(base, chosen); err != nil {
					return fmt.Errorf("install authorized_keys: %w", err)
				}
				p.Info("Installed authorized_keys at %s/ssh/authorized_keys", base)
			}

			identity, _ := sshcfg.DefaultIdentityFile()
			if chosen != "" {
				identity = strings.TrimSuffix(chosen, ".pub")
			}
			kh, err := sshcfg.KnownHostsPath()
			if err != nil {
				return err
			}
			added, err := sshcfg.Install(sshcfg.Entry{
				Name: name, Port: port, IdentityFile: identity, KnownHosts: kh,
			})
			if err != nil {
				return fmt.Errorf("ssh config: %w", err)
			}
			if added {
				p.Info("Added 'dy-%s' to ~/.ssh/config  →  ssh dy-%s", name, name)
			} else {
				p.Info("~/.ssh/config already has 'dy-%s' — skipping", name)
			}

			p.Info("")
			p.Info("Next: dockyard deploy %s   then   ssh dy-%s", name, name)
			return nil
		},
	}
}

func nextFreeUID(start int) int {
	for uid := start; uid < start+500; uid++ {
		if _, err := user.LookupId(strconv.Itoa(uid)); err != nil {
			return uid
		}
	}
	return start
}

func nextFreePort(start int) int {
	for port := start; port < start+500; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			continue
		}
		_ = ln.Close()
		return port
	}
	return start
}
