# openclaw — Autonomous Coding Agent

Extends the [dev-env](../dev-env/) base with Node.js, the OpenClaw CLI, systemd
(for the openclaw-gateway user service), and LiteLLM proxy access. Designed for
autonomous coding agents running inside persistent SSH-accessible containers.

## Quick Start

```bash
dockyard create openclaw <your-container-name>
dockyard deploy <your-container-name>
ssh dy-<your-container-name>
```

## What's Added (vs dev-env)

| Component | Purpose |
|---|---|
| Node.js (22 or 24) | Runtime for the openclaw CLI |
| openclaw CLI | Agent orchestration and gateway |
| systemd | PID 1 — manages openclaw-gateway as a user service |
| container-init.sh | Oneshot: SSH host-key generation, boot-seed, permission fixes |
| openclaw-gateway.service | Auto-starts via systemd user linger |

### Python packages (via `requirements.txt`)

| Package | Purpose |
|---|---|
| openai, anthropic, litellm | LLM client libraries |
| httpx, aiohttp | Async HTTP |
| pydantic | Structured data validation |
| structlog, rich | Logging and terminal output |
| celery, redis | Task queue |
| tenacity | Retry logic |

## Build Args

Declared in `manifest.yaml` and prompted during `dockyard create`:

| Arg | Default | Options | Notes |
|---|---|---|---|
| `NODE_MAJOR` | `24` | `22`, `24` | Node 24 has a jiti/V8 regression on Westmere/Sandy Bridge era x86 CPUs — use 22 on such hardware |

Override at deploy time by editing `$VolumesRoot/<name>/config.yaml` and
re-running `dockyard deploy <name>`.

## Persistent Volumes

| Directory | Container Mount | Owner |
|---|---|---|
| `config/` | `~/.config` | agent (2770) |
| `openclaw-data/` | `~/.openclaw` | agent (2770) |
| `workspace/` | `/workspace` | agent (2770) |
| `nvim-data/` | `~/.local/share/nvim` | agent (2770) |
| `nvim-state/` | `~/.local/state/nvim` | agent (2770) |
| `logs/` | `/logs` | agent (2770) |
| `secrets/` | `/secrets` (read-only) | root (750) |
| `ssh/authorized_keys` | `~/.ssh/authorized_keys` | agent |

## First-Boot Seeding

On first start, `container-init.sh` sources `boot-seed.sh` which:

- Seeds **LazyVim** starter config into `~/.config/nvim/` (skipped if `init.lua`
  already exists)
- Seeds **TPM** (Tmux Plugin Manager) into `~/.config/tmux/plugins/tpm/` (skipped
  if already present)
- Installs the **openclaw-gateway** systemd user service into
  `~/.config/systemd/user/` (skipped if already present)

All seeding is idempotent — existing config is never overwritten.

## Secrets

An `env.example` file is seeded into `secrets/` on first deploy (never
overwrites). Copy it and fill in your API keys:

```bash
cp /secrets/env.example /secrets/env
```

Source it on login by adding to `~/.bashrc`:

```bash
[ -f /secrets/env ] && set -a && source /secrets/env && set +a
```

## LiteLLM Access

The container is configured with `LITELLM_BASE_URL=http://host.gateway.internal:4000`
via `extra_hosts` in docker-compose.yml. This routes to the LiteLLM proxy running on
the host machine, providing unified access to multiple LLM providers.

## Resource Limits

Default limits set in docker-compose.yml:

| Resource | Limit | Reservation |
|---|---|---|
| CPU | 4 cores | 0.5 cores |
| Memory | 16 GB | 512 MB |
