# dockyard

A Go CLI that builds and manages Dockerised development containers. Templates
(Dockerfile + docker-compose.yml) are embedded in the binary — one command to
create, one to deploy.

## Quick Start

```bash
# Build the binary
go build -o bin/dockyard ./cmd/dockyard

# One-time host setup — choose where deployment volumes are stored
dockyard init

# Create a new deployment (interactive prompts for name, UID, SSH port, etc.)
dockyard create openclaw <your-container-name>

# Build image, create host volumes, start container
dockyard deploy <your-container-name>

# SSH in
ssh dy-<your-container-name>
```

## Templates

| Template | Purpose | Status |
|---|---|---|
| [dev-env](templates/dev-env/) | SSH dev box: Neovim, tmux, Python, Node.js | Ready |
| [openclaw](templates/openclaw/) | Autonomous coding agent with LLM tooling | Ready |
| [hermes-agent](templates/hermes-agent/) | Nous Research Hermes Agent — Python-based autonomous coding agent | Ready |

```
$ dockyard templates
  dev-env       Plain SSH-accessible development environment (no agent)
  hermes-agent  Autonomous coding agent powered by Nous Research Hermes Agent framework
  openclaw      Autonomous coding agent with LLM access (systemd + openclaw CLI)
```

## CLI Reference

| Command | Description |
|---|---|
| `dockyard init` | Set the volumes root path (`~/.config/dockyard/config.yaml`) |
| `dockyard templates` | List embedded templates |
| `dockyard create <template> [name]` | Interactive setup — writes `config.yaml`, `.env`, SSH key + stanza |
| `dockyard deploy <name>` | Stage build context, create host volumes, `docker compose up -d --build` |
| `dockyard up <name>` | Start without rebuilding |
| `dockyard down <name>` | Stop the container |
| `dockyard restart <name>` | Down + up (no rebuild) |
| `dockyard status [name]` | Show state of one or all deployments |
| `dockyard list` | Alias for `status` (no args) |
| `dockyard shell <name>` | Interactive bash inside the container |
| `dockyard logs [-f] <name>` | Tail container logs |
| `dockyard rm [-f] <name>` | Stop container, delete images and volume data |

## Configuration

All instance-specific values live outside the repo under `$VolumesRoot/<name>/`.

```
~/.config/dockyard/
├── config.yaml              # global: volumes_root
└── known_hosts              # SSH known hosts

$VolumesRoot/<name>/
├── config.yaml              # deployment: template, uid, gid, port, build_args
├── .env                     # generated for docker compose --env-file
├── build/                   # staged build context (Dockerfile, etc.)
├── ssh/authorized_keys      # injected into container
├── secrets/                 # mounted read-only at /secrets
├── workspace/               # mounted at /workspace
├── nvim-data/               # ~/.local/share/nvim
├── nvim-state/              # ~/.local/state/nvim
└── logs/                    # /logs
```

| Field | Example | Purpose |
|---|---|---|
| `name` | `<your-container-name>` | Deployment name; container becomes `dy-<your-container-name>` |
| `template` | `openclaw` | Which embedded template to use |
| `agent_uid` | `1100` | UID of the agent user inside the container |
| `agent_gid` | `1100` | GID (usually matches UID) |
| `ssh_port` | `2200` | Host port mapped to container port 22 |
| `build_args` | `NODE_MAJOR: "22"` | Template-specific build arguments |

## Philosophy

- **Single binary.** Templates, entrypoints, and manifests are embedded via
  `go:embed`. No files to locate at runtime.
- **Self-contained templates.** Each template has its own Dockerfile,
  docker-compose.yml, and manifest.yaml. No shared base image.
- **Typed configuration.** YAML config with struct validation replaces raw
  `.env` files. Legacy `.env` deployments are read transparently.
- **Host volume permissions done right.** A shared `agents` group with setgid
  (2770) lets the deploying user and the container agent user both read/write
  persistent volumes.
- **SSH-first access.** Containers are accessed via SSH (key-only), not
  `docker exec`. Works identically whether the container is local or remote.
- **Security defaults.** No root login, no password auth, secrets mounted
  read-only, resource guardrails to prevent runaway containers.

## Repo Structure

```
dockyard/
├── cmd/dockyard/main.go       # CLI entrypoint
├── internal/
│   ├── cli/                   # cobra subcommands
│   ├── config/                # Global + Deployment config, fsutil
│   ├── template/              # manifest loader, build-context stager
│   ├── prompt/                # interactive stdin helpers
│   ├── sshcfg/                # ~/.ssh/config management
│   ├── dockercmd/             # docker / docker compose wrappers
│   └── volumes/               # host volume creation + permissions
├── assets.go                  # go:embed for templates/ and shared/
├── templates/
│   ├── dev-env/               # Plain dev environment
│   ├── hermes-agent/          # Nous Research Hermes Agent
│   └── openclaw/              # Autonomous coding agent
└── shared/
    └── entrypoint.sh          # Common startup script
```

## Host Prerequisites

- Go 1.22+ (build only — the compiled binary has no runtime dependency)
- Docker Engine with Compose v2
- `sudo` access (for volume directory creation and group management)

## Adding a New Template

1. Create `templates/<name>/` with a `Dockerfile`, `docker-compose.yml`, and
   `manifest.yaml`.
2. The manifest declares `agent_dirs`, `root_dirs`, `build_args`, and
   `shared_files`. See [templates/dev-env/manifest.yaml](templates/dev-env/manifest.yaml)
   for a minimal example.
3. Rebuild the binary (`go build ./cmd/dockyard`) — the new template is
   automatically embedded and will appear in `dockyard templates`.
