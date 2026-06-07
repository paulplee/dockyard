// Package dockyard provides the public Engine API for building and
// managing Dockerised development containers. It is imported by both
// the dockyard CLI and downstream products (e.g., operon) that extend
// dockyard with additional templates and behaviour.
package dockyard

import (
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

// Engine is a configured instance of the dockyard orchestrator. Defaults
// correspond to the open-source dockyard binary. Downstream consumers
// customise the engine via With* methods and pass it to the CLI builder.
type Engine struct {
	// Name is the binary / product name ("dockyard").
	Name string

	// ContainerPrefix prepended to container names ("dy-").
	ContainerPrefix string

	// ConfigDir is the host config directory (~/.config/<name>).
	ConfigDir string

	// DefaultAgentUser is the default OS username inside the container.
	DefaultAgentUser string

	// AdditionalFS are extra embed.FS instances that contain
	// templates/<name>/ subdirectories. Operon uses this to ship
	// its private templates/operon/ without forking dockyard.
	AdditionalFS []fs.FS

	// Hooks allow downstream products to inject custom behaviour at
	// key points in the create/deploy lifecycle.
	Hooks ExtensionHooks
}

// ExtensionHooks provides injection points for downstream products.
// Each hook receives the Engine instance so it can access naming config.
type ExtensionHooks struct {
	// OnBeforeCreate is called after the interactive prompts have
	// populated buildArgs but before config.yaml is written. The hook
	// may add to buildArgs (e.g., auto-generate API keys).
	OnBeforeCreate func(engine *Engine, buildArgs map[string]string) error

	// OnAfterCreate is called after config.yaml, .env, and SSH config
	// have been written. Typical use: write secrets/env or agent
	// profiles into the deployment base directory.
	OnAfterCreate func(engine *Engine, base string) error
}

// DefaultEngine returns an Engine configured for the open-source dockyard CLI.
func DefaultEngine() *Engine {
	return &Engine{
		Name:             "dockyard",
		ContainerPrefix:  "dy-",
		ConfigDir:        configDir("dockyard"),
		DefaultAgentUser: "dy-user",
	}
}

// WithName returns a copy of e with Name set to v.
func (e *Engine) WithName(v string) *Engine {
	e2 := *e
	e2.Name = v
	return &e2
}

// WithContainerPrefix returns a copy with ContainerPrefix set to v.
func (e *Engine) WithContainerPrefix(v string) *Engine {
	e2 := *e
	e2.ContainerPrefix = v
	return &e2
}

// WithDefaultAgentUser returns a copy with DefaultAgentUser set to v.
func (e *Engine) WithDefaultAgentUser(v string) *Engine {
	e2 := *e
	e2.DefaultAgentUser = v
	return &e2
}

// WithConfigDir returns a copy with ConfigDir set to v.
func (e *Engine) WithConfigDir(v string) *Engine {
	e2 := *e
	e2.ConfigDir = v
	return &e2
}

// WithHooks returns a copy with Hooks set to v.
func (e *Engine) WithHooks(v ExtensionHooks) *Engine {
	e2 := *e
	e2.Hooks = v
	return &e2
}

// ContainerName returns the Docker container name for a deployment.
func (e *Engine) ContainerName(depName string) string {
	return e.ContainerPrefix + depName
}

// SSHHostAlias returns the SSH Host alias for a deployment.
func (e *Engine) SSHHostAlias(depName string) string {
	return e.ContainerPrefix + depName
}

// SSHMaker returns the SSH config marker line for a deployment.
func (e *Engine) SSHMaker(depName string) string {
	return "# " + e.Name + ": " + e.ContainerPrefix + depName
}

// configDir returns ~/.config/<name>.
func configDir(name string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", name)
}

// NextFreeUID finds an unused UID starting from start.
func NextFreeUID(start int) int {
	for uid := start; uid < start+500; uid++ {
		if _, err := user.LookupId(strconv.Itoa(uid)); err != nil {
			return uid
		}
	}
	return start
}