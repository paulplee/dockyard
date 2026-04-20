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
| container-init.sh | Oneshot: SSH host-key generation, permission fixes |
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

## LiteLLM Access

The container is configured with `LITELLM_BASE_URL=http://host.gateway.internal:4000`
via `extra_hosts` in docker-compose.yml. This routes to the LiteLLM proxy running on
the host machine, providing unified access to multiple LLM providers.

## Persistent Volumes

| Directory | Container Mount | Owner |
|---|---|---|
| `openclaw-data/` | `~/.openclaw` | agent (2770) |
| `workspace/` | `/workspace` | agent (2770) |
| `nvim-data/` | `~/.local/share/nvim` | agent (2770) |
| `nvim-state/` | `~/.local/state/nvim` | agent (2770) |
| `logs/` | `/logs` | agent (2770) |
| `secrets/` | `/secrets` (read-only) | root (750) |
| `ssh/authorized_keys` | `~/.ssh/authorized_keys` | agent |

## Resource Limits

Default limits set in docker-compose.yml:

| Resource | Limit | Reservation |
|---|---|---|
| CPU | 4 cores | 0.5 cores |
| Memory | 16 GB | 512 MB |
