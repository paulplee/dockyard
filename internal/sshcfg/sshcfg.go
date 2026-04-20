// Package sshcfg manages the ~/.ssh/config entry for a dockyard deployment
// and installs an authorized_keys file into $VolumesBase/ssh/.
package sshcfg

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/paulplee/dockyard/internal/config"
)

// Entry describes one deployment's SSH stanza.
type Entry struct {
	Name         string // deployment name (Host alias will be "dy-<Name>")
	Port         int
	IdentityFile string
	KnownHosts   string // absolute path
}

// Marker returns the sentinel comment line written above each stanza.
func (e Entry) Marker() string { return "# dockyard: dy-" + e.Name }

// Render returns the full ~/.ssh/config stanza (with a leading blank line).
func (e Entry) Render() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(e.Marker())
	b.WriteString("\n")
	fmt.Fprintf(&b, "Host dy-%s\n", e.Name)
	b.WriteString("    HostName 127.0.0.1\n")
	fmt.Fprintf(&b, "    Port %d\n", e.Port)
	b.WriteString("    User agent\n")
	fmt.Fprintf(&b, "    IdentityFile %s\n", e.IdentityFile)
	b.WriteString("    IdentitiesOnly yes\n")
	b.WriteString("    StrictHostKeyChecking accept-new\n")
	fmt.Fprintf(&b, "    UserKnownHostsFile %s\n", e.KnownHosts)
	return b.String()
}

// Install appends the entry to ~/.ssh/config if no stanza with its marker is
// already present. Returns (added, error) — added=false means a stanza already
// existed and nothing was written.
func Install(e Entry) (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}
	cfgPath := filepath.Join(home, ".ssh", "config")

	existing, err := os.ReadFile(cfgPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	if strings.Contains(string(existing), e.Marker()) {
		return false, nil
	}

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o700); err != nil {
		return false, err
	}
	f, err := os.OpenFile(cfgPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return false, err
	}
	defer f.Close()
	if _, err := f.WriteString(e.Render()); err != nil {
		return false, err
	}
	if err := os.Chmod(cfgPath, 0o600); err != nil {
		return false, err
	}
	return true, nil
}

// DefaultIdentityFile picks the best local SSH key (preferring ed25519).
func DefaultIdentityFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
		p := filepath.Join(home, ".ssh", name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	// Fall back to the first *.pub we find.
	entries, err := os.ReadDir(filepath.Join(home, ".ssh"))
	if err == nil {
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".pub") {
				return filepath.Join(home, ".ssh", strings.TrimSuffix(e.Name(), ".pub")), nil
			}
		}
	}
	return filepath.Join(home, ".ssh", "id_ed25519"), nil
}

// ListPublicKeys returns absolute paths to *.pub files under ~/.ssh/.
func ListPublicKeys() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(home, ".ssh"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".pub") {
			out = append(out, filepath.Join(home, ".ssh", e.Name()))
		}
	}
	return out, nil
}

// InstallAuthorizedKeys copies pubKeyPath to $volumesBase/ssh/authorized_keys
// (overwriting if present) and fixes ownership/permissions.
func InstallAuthorizedKeys(volumesBase string, pubKeyPath string) error {
	sshDir := filepath.Join(volumesBase, "ssh")
	if err := config.MkdirAllPrivileged(sshDir, 0o755); err != nil {
		return err
	}
	data, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", pubKeyPath, err)
	}
	return config.WriteFilePrivileged(filepath.Join(sshDir, "authorized_keys"), data, 0o600)
}

// KnownHostsPath returns ~/.config/dockyard/known_hosts.
func KnownHostsPath() (string, error) {
	dir, err := config.GlobalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "known_hosts"), nil
}

// HasStanza reports whether ~/.ssh/config already contains the marker for name.
func HasStanza(name string) (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}
	f, err := os.Open(filepath.Join(home, ".ssh", "config"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	marker := "# dockyard: dy-" + name
	for s.Scan() {
		if strings.TrimSpace(s.Text()) == marker {
			return true, s.Err()
		}
	}
	return false, s.Err()
}
