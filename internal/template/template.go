// Package template manages dockyard's embedded template tree. Each template
// is a subdirectory under ./templates/<name>/ containing a manifest.yaml, a
// Dockerfile, a docker-compose.yml, and any supporting files. Shared files
// referenced via manifest.shared_files are embedded from ./shared/.
package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	dockyard "github.com/paulplee/dockyard"
	"gopkg.in/yaml.v3"
)

// embedded is the project-wide embedded FS containing templates/ and shared/.
var embedded = dockyard.Assets

// BuildArg describes a single prompt shown during `dockyard create`.
type BuildArg struct {
	Name    string   `yaml:"name"`
	Prompt  string   `yaml:"prompt"`
	Default string   `yaml:"default"`
	Options []string `yaml:"options,omitempty"`
	Help    string   `yaml:"help,omitempty"`
}

// SharedFile is a file copied from the repo's shared/ tree into the staged
// build context prior to `docker compose build`.
type SharedFile struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
}

// Manifest is parsed from templates/<name>/manifest.yaml.
type Manifest struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	AgentDirs   []string     `yaml:"agent_dirs"`
	RootDirs    []string     `yaml:"root_dirs"`
	BuildArgs   []BuildArg   `yaml:"build_args"`
	SharedFiles []SharedFile `yaml:"shared_files"`
	HostFiles   []SharedFile `yaml:"host_files"`
}

// List returns the names of all embedded templates (alphabetical).
func List() ([]string, error) {
	entries, err := fs.ReadDir(embedded, "templates")
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := fs.Stat(embedded, filepath.Join("templates", e.Name(), "manifest.yaml")); err != nil {
			continue
		}
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out, nil
}

// LoadManifest loads templates/<name>/manifest.yaml from the embedded FS.
func LoadManifest(name string) (*Manifest, error) {
	data, err := fs.ReadFile(embedded, filepath.Join("templates", name, "manifest.yaml"))
	if err != nil {
		return nil, fmt.Errorf("load manifest for %q: %w", name, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest for %q: %w", name, err)
	}
	if m.Name == "" {
		m.Name = name
	}
	return &m, nil
}

// StageBuildContext copies all files from templates/<name>/ (except
// manifest.yaml) plus every shared_files entry into dstDir. It overwrites
// existing files so that `dockyard deploy` always produces a fresh context.
func StageBuildContext(name, dstDir string) error {
	m, err := LoadManifest(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}

	root := filepath.Join("templates", name)
	err = fs.WalkDir(embedded, root, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if rel == "manifest.yaml" {
			return nil
		}
		dst := filepath.Join(dstDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		return copyEmbedded(p, dst)
	})
	if err != nil {
		return err
	}

	for _, sf := range m.SharedFiles {
		srcPath := strings.TrimPrefix(sf.Src, "./")
		dst := filepath.Join(dstDir, sf.Dst)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := copyEmbedded(srcPath, dst); err != nil {
			return fmt.Errorf("copy shared %s: %w", sf.Src, err)
		}
	}
	return nil
}

// WriteHostFiles copies manifest host_files entries from the embedded template
// into baseDir on the host. Unlike shared_files (which go into the Docker build
// context), host_files are seeded directly into the host volume tree so the
// user can edit them in place. Existing files are never overwritten.
func WriteHostFiles(name, baseDir string) error {
	m, err := LoadManifest(name)
	if err != nil {
		return err
	}
	root := filepath.Join("templates", name)
	for _, hf := range m.HostFiles {
		src := filepath.Join(root, hf.Src)
		dst := filepath.Join(baseDir, hf.Dst)
		if _, err := os.Stat(dst); err == nil {
			continue // never overwrite user edits
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := copyEmbedded(src, dst); err != nil {
			return fmt.Errorf("seed host file %s: %w", hf.Src, err)
		}
	}
	return nil
}

func copyEmbedded(src, dst string) error {
	data, err := fs.ReadFile(embedded, src)
	if err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if strings.HasSuffix(src, ".sh") {
		mode = 0o755
	}
	return os.WriteFile(dst, data, mode)
}
