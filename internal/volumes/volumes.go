// Package volumes creates and owns the per-deployment directory tree under
// $VolumesBase (equivalent to the `init` target in shared/makefiles/volumes.mk).
package volumes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paulplee/dockyard/internal/config"
	"github.com/paulplee/dockyard/internal/template"
)

// Prepare creates the template's agent_dirs and root_dirs under base, then
// chowns agent_dirs to (uid, gid) with mode 2770. secrets/ is tightened to
// mode 750, and an existing ssh/authorized_keys is set to uid:gid mode 600.
func Prepare(base string, m *template.Manifest, uid, gid int) error {
	fmt.Printf(">>> Creating host volumes under %s\n", base)
	for _, name := range m.AgentDirs {
		d := filepath.Join(base, name)
		if err := config.MkdirAllPrivileged(d, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", d, err)
		}
		fmt.Printf("  OK  %s\n", d)
	}
	for _, name := range m.RootDirs {
		d := filepath.Join(base, name)
		if err := config.MkdirAllPrivileged(d, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", d, err)
		}
		fmt.Printf("  OK  %s\n", d)
	}
	fmt.Printf(">>> Setting ownership (UID=%d, GID=%d) on agent_dirs\n", uid, gid)
	for _, name := range m.AgentDirs {
		d := filepath.Join(base, name)
		if err := config.ChownPrivileged(d, uid, gid, true); err != nil {
			return err
		}
		if err := config.ChmodPrivileged(d, 0o2770, true); err != nil {
			return err
		}
	}
	secrets := filepath.Join(base, "secrets")
	if _, err := os.Stat(secrets); err == nil {
		if err := config.ChmodPrivileged(secrets, 0o750, false); err != nil {
			return err
		}
	}
	ak := filepath.Join(base, "ssh", "authorized_keys")
	if _, err := os.Stat(ak); err == nil {
		if err := config.ChownPrivileged(ak, uid, gid, false); err != nil {
			return err
		}
		if err := config.ChmodPrivileged(ak, 0o600, false); err != nil {
			return err
		}
	}
	fmt.Println(">>> Volume preparation complete.")
	if err := template.WriteHostFiles(m.Name, base); err != nil {
		return fmt.Errorf("seed host files: %w", err)
	}
	return nil
}
