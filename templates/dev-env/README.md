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
| tmux | system | Stock install — configure from `workspace/` |
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
| `config/` | `~/.config` | All app config (Neovim, tmux, etc.) |
| `nvim-data/` | `~/.local/share/nvim` | Neovim plugins (persist across rebuilds) |
| `nvim-state/` | `~/.local/state/nvim` | Neovim session state |
| `logs/` | `/logs` | Structured log output |
| `secrets/` | `/secrets` (read-only) | API keys / tokens |
| `ssh/authorized_keys` | `~/.ssh/authorized_keys` | SSH public keys |

## Persisting Your Config

`~/.config` is bind-mounted from the host (`config/` in the deployment directory),
so all config survives rebuilds and is directly editable from the host.

### Neovim

On first container start the entrypoint seeds the
[LazyVim starter](https://github.com/LazyVim/starter) into `config/nvim/` if it
doesn't exist yet. Open nvim and run `:Lazy sync` to download plugins (stored in
the persistent `nvim-data/` volume). Your config is immediately editable on the
host at:

```
~/.config/dockyard/volumes/<name>/config/nvim/
```

### tmux

tmux config lives at `~/.config/tmux/tmux.conf` — already persistent. To
bootstrap from scratch inside the container:

```bash
mkdir -p ~/.config/tmux
# write your tmux.conf, or clone a dotfiles repo into workspace/ and symlink it:
# ln -s /workspace/dotfiles/tmux.conf ~/.config/tmux/tmux.conf
```

To install TPM:

```bash
git clone https://github.com/tmux-plugins/tpm ~/.config/tmux/plugins/tpm
```

Plugins install to `~/.config/tmux/plugins/` and persist automatically.

### Shell dotfiles

Store your bashrc additions in the persistent `workspace/` and source them:

```bash
cat >> /workspace/.bashrc_local << 'EOF'
# your customisations here
EOF
echo '[ -f /workspace/.bashrc_local ] && source /workspace/.bashrc_local' >> ~/.bashrc
```

## Secrets

`secrets/env.example` is seeded into your deployment's `secrets/` directory on
first deploy. Copy it to `env` and fill in your credentials:

```bash
# on the host
cp ~/.config/dockyard/volumes/mybox/secrets/env.example \
   ~/.config/dockyard/volumes/mybox/secrets/env
# edit with real values
```

Inside the container, source it on login:

```bash
# add to ~/.bashrc inside the container
[ -f /secrets/env ] && set -a && source /secrets/env && set +a
```

The `secrets/` directory is never added to the Docker image and is mounted
read-only in the container.

## Build Args

This template has no template-specific build arguments. The standard deployment
fields (agent UID/GID, SSH port) are configured during `dockyard create`.
