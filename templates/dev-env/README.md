# dev-env — SSH Development Environment

A hardened, SSH-accessible Debian container with Neovim (LazyVim), tmux (TPM),
Python 3, Node.js, and modern CLI tools. This is the baseline template — other
templates (like `openclaw`) extend it with purpose-specific tooling.

## Quick Start

```bash
cd templates/dev-env

# 1. Interactive setup — writes .env to the deployment volume path
make setup

# 2. Deploy (creates host dirs, sets permissions, builds image, starts container)
make deploy CONTAINER_NAME=<name>

# 3. Activate group membership (once per login session)
newgrp agents

# 4. SSH in
ssh -p <SSH_PORT> agent@localhost
```

## What's Included

| Tool | Version | Notes |
|---|---|---|
| Neovim | 0.12.1 | LazyVim config, plugins pre-warmed |
| tmux | system | TPM + custom config from paulplee/mancave |
| Python 3 | system | pip, venv, dev headers |
| Node.js | system | npm included |
| bat, ripgrep, fd, fzf, jq | system | Modern CLI essentials |
| btop | system | Interactive process/resource monitor |
| uv | latest | Fast Python package manager (Astral) |
| SSH server | hardened | Key-only auth, no root login |

## Persistent Volumes

All state lives on the host at the directory path specified when `make setup` is
first run:

| Directory | Container mount | Owner |
|---|---|---|
| `workspace/` | `/workspace` | agent (2770) |
| `nvim-data/` | `~/.local/share/nvim` | agent (2770) |
| `nvim-state/` | `~/.local/state/nvim` | agent (2770) |
| `logs/` | `/logs` | agent (2770) |
| `secrets/` | `/secrets` (read-only) | root (750) |
| `ssh/authorized_keys` | `~/.ssh/authorized_keys` | agent |

## Makefile Targets

| Target | Description |
|---|---|
| `make setup` | Interactive prompt → writes `.env` to deployment volume path |
| `make deploy CONTAINER_NAME=<n>` | group → init → up |
| `make up CONTAINER_NAME=<n>` | Build and start the container |
| `make down CONTAINER_NAME=<n>` | Stop the container |
| `make shell CONTAINER_NAME=<n>` | Exec into the container |
| `make logs CONTAINER_NAME=<n>` | Tail container logs |
| `make clean CONTAINER_NAME=<n>` | Remove container, image, and volume data (preserves SSH keys) |
| `make reset CONTAINER_NAME=<n>` | Clean + deploy |
