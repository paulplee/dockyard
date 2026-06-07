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

// additionalFS holds extra template filesystems registered by downstream
// products (e.g., operon adds its private templates/operon/ template).
var additionalFS []fs.FS

// SetAdditionalFS registers additional embedded filesystems that contain
// templates/<name>/ subdirectories. Call before List/LoadManifest/StageBaseContext.
func SetAdditionalFS(filesystems []fs.FS) {
	additionalFS = filesystems
}

// openTemplateFile opens a file from the primary or additional FS, returning
// the FS and the file data. Checks additional FS first so downstream products
// can override template files.
func openTemplateFile(path string) ([]byte, error) {
	// Check additional FS in reverse order (last registered wins for overrides).
	for i := len(additionalFS) - 1; i >= 0; i-- {
		data, err := fs.ReadFile(additionalFS[i], path)
		if err == nil {
			return data, nil
		}
	}
	return fs.ReadFile(embedded, path)
}

// templateFSForPath returns the fs.FS that contains path, preferring
// additionalFS (last registered wins), then falling back to embedded.
func templateFSForPath(path string) (fs.FS, error) {
	for i := len(additionalFS) - 1; i >= 0; i-- {
		if _, err := fs.Stat(additionalFS[i], path); err == nil {
			return additionalFS[i], nil
		}
	}
	if _, err := fs.Stat(embedded, path); err == nil {
		return embedded, nil
	}
	return nil, fmt.Errorf("template path %q not found in any FS", path)
}

// readTemplateDir reads directory entries from the primary or additional FS.
func readTemplateDir(path string) ([]os.DirEntry, error) {
	seen := map[string]bool{}
	var out []os.DirEntry
	for i := len(additionalFS) - 1; i >= 0; i-- {
		entries, err := fs.ReadDir(additionalFS[i], path)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !seen[e.Name()] {
				out = append(out, e)
				seen[e.Name()] = true
			}
		}
	}
	entries, err := fs.ReadDir(embedded, path)
	if err != nil {
		if len(out) > 0 {
			sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
			return out, nil
		}
		return nil, err
	}
	for _, e := range entries {
		if !seen[e.Name()] {
			out = append(out, e)
			seen[e.Name()] = true
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out, nil
}

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

// List returns the names of all templates (primary + additional FS).
func List() ([]string, error) {
	entries, err := readTemplateDir("templates")
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := openTemplateFile(filepath.Join("templates", e.Name(), "manifest.yaml")); err != nil {
			continue
		}
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out, nil
}

// LoadManifest loads templates/<name>/manifest.yaml from the primary or additional FS.
func LoadManifest(name string) (*Manifest, error) {
	data, err := openTemplateFile(filepath.Join("templates", name, "manifest.yaml"))
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
	tfs, err := templateFSForPath(root)
	if err != nil {
		return fmt.Errorf("stage build context for %q: %w", name, err)
	}

	err = fs.WalkDir(tfs, root, func(p string, d fs.DirEntry, walkErr error) error {
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
		return copyTemplateFile(tfs, p, dst)
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
		if err := copyTemplateFile(nil, srcPath, dst); err != nil {
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
	tfs, err := templateFSForPath(root)
	if err != nil {
		return err
	}
	for _, hf := range m.HostFiles {
		src := filepath.Join(root, hf.Src)
		dst := filepath.Join(baseDir, hf.Dst)
		if _, err := os.Stat(dst); err == nil {
			continue // never overwrite user edits
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := copyTemplateFile(tfs, src, dst); err != nil {
			return fmt.Errorf("seed host file %s: %w", hf.Src, err)
		}
	}
	return nil
}

// copyTemplateFile copies a file from tfs (if provided) or the first FS that
// contains srcPath. If tfs is nil, it searches additionalFS then embedded.
func copyTemplateFile(tfs fs.FS, src, dst string) error {
	if tfs == nil {
		var err error
		tfs, err = templateFSForPath(src)
		if err != nil {
			// Fall back to searching all additional FS then embedded.
			return copyViaOpenTemplate(src, dst)
		}
	}
	data, err := fs.ReadFile(tfs, src)
	if err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if strings.HasSuffix(src, ".sh") {
		mode = 0o755
	}
	return os.WriteFile(dst, data, mode)
}

// copyViaOpenTemplate uses openTemplateFile to read src and write dst.
func copyViaOpenTemplate(src, dst string) error {
	data, err := openTemplateFile(src)
	if err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if strings.HasSuffix(src, ".sh") {
		mode = 0o755
	}
	return os.WriteFile(dst, data, mode)
}