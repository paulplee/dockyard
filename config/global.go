// Package config manages on-disk configuration: a global config file at
// ~/.config/<name>/config.yaml and per-deployment config.yaml files stored
// under $VolumesRoot/<name>/.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/paulplee/dockyard/pkg/dockyard"
	"gopkg.in/yaml.v3"
)

// Global holds host-wide settings.
type Global struct {
	VolumesRoot string `yaml:"volumes_root"`
}

// GlobalDir returns ~/.config/<product>.
func GlobalDir(engine *dockyard.Engine) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", engine.Name), nil
}

// GlobalPath returns the path to the global YAML config file.
func GlobalPath(engine *dockyard.Engine) (string, error) {
	dir, err := GlobalDir(engine)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// legacyGlobalEnvPath is the pre-Go Make-era config file.
func legacyGlobalEnvPath(engine *dockyard.Engine) (string, error) {
	dir, err := GlobalDir(engine)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".env"), nil
}

// LoadGlobal reads the YAML config, falling back to the legacy .env file if
// the YAML form does not yet exist. Returns (nil, nil) if nothing is found.
func LoadGlobal(engine *dockyard.Engine) (*Global, error) {
	p, err := GlobalPath(engine)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err == nil {
		var g Global
		if err := yaml.Unmarshal(data, &g); err != nil {
			return nil, fmt.Errorf("parse %s: %w", p, err)
		}
		return &g, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	// Legacy fallback.
	legacy, err := legacyGlobalEnvPath(engine)
	if err != nil {
		return nil, err
	}
	kv, err := readEnvFile(legacy)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if vr := kv["VOLUMES_ROOT"]; vr != "" {
		return &Global{VolumesRoot: vr}, nil
	}
	return nil, nil
}

// Save writes the global config as YAML.
func (g *Global) Save(engine *dockyard.Engine) error {
	p, err := GlobalPath(engine)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(g)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// readEnvFile parses a KEY=VALUE .env file. Blank lines and # comments are
// ignored. Quoted values have their quotes stripped.
func readEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])
		v = strings.Trim(v, `"'`)
		out[k] = v
	}
	return out, s.Err()
}