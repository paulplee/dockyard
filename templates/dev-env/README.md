# dev-env — SSH Development Environment

A hardened, SSH-accessible Debian container with Neovim (LazyVim), tmux, Python 3,
Node.js, and modern CLI tools. This is the baseline template — other templates
(like `openclaw`) extend it with purpose-specific tooling.

## Quick Start

```bash
dockyard create dev-env mybox
dockyard deploy mybox
ssh dy-mybox
```

## What's Included

| Tool | Version | Notes |
|---|---|---|
| Neovim | 0.12.1 | LazyVim starter config, plugins pre-warmed at build time |
| tmux | system | TPM + [mancave](https://github.com/paulplee/mancave) config baked in |
| Python 3 | system | pip, venv, dev headers |
| Node.js | system | npm included |
| bat, ripgrep, fd, fzf, jq | system | Modern CLI essentials |
| btop | system | Interactive process/resource monitor |
| uv | latest | Fast Python package manager (Astral) |
| SSH server | hardened | Key-only auth, no root login |

## Persistent Volumes

All state lives on the host under `$VolumesRoot/<name>/`:

| Directory | Container Mount | Purpose |
|---|---|---|
| `workspace/` | `/workspace` | Your code and projects |
| `nvim-data/` | `~/.local/share/nvim` | Neovim plugins (persist across rebuilds) |
| `nvim-state/` | `~/.local/state/nvim` | Neovim session state |
| `logs/` | `/logs` | Structured log output |
| `secrets/` | `/secrets` (read-only) | API keys / tokens |
| `ssh/authorized_keys` | `~/.ssh/authorized_keys` | SSH public keys |

## Persisting Your Config

The LazyVim starter config (`~/.config/nvim`) and tmux config are baked into
the image and will be reset on `dockyard deploy` (rebuild). To keep your
customisations across rebuilds, store them in the persistent `workspace/` volume
and source/symlink them on login.

### Neovim

Neovim **plugin data** (`nvim-data/`) is already persistent — any plugins you
install with `:Lazy sync` survive rebuilds. To also persist your config:

```bash
# inside the container
mkdir -p /workspace/.config/nvim
cp -r ~/.config/nvim/* /workspace/.config/nvim/
echo 'export XDG_CONFIG_HOME=/workspace/.config' >> ~/.bashrc
```

On next login the `XDG_CONFIG_HOME` override points Neovim at your workspace
copy. Plugins still load from the persistent `nvim-data/` volume.

### tmux

Store your `tmux.conf` in the workspace and point tmux at it:

```bash
mkdir -p /workspace/.config/tmux
cp ~/.config/tmux/tmux.conf /workspace/.config/tmux/tmux.conf
# edit to taste, then:
echo 'source-file /workspace/.config/tmux/tmux.conf' >> ~/.config/tmux/tmux.conf
```

Or replace the config file entirely by overriding `XDG_CONFIG_HOME` as above.

### Shell dotfiles

```bash
# Store your bashrc additions in the workspace:
cat >> /workspace/.bashrc_local << 'EOF'
# your customisations here
EOF
echo '[ -f /workspace/.bashrc_local ] && source /workspace/.bashrc_local' >> ~/.bashrc
```

## Build Args

This template has no template-specific build arguments. The standard deployment
fields (agent UID/GID, SSH port) are configured during `dockyard create`.
