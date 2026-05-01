# hermes-agent — Nous Research Hermes Agent

Extends the [dev-env](../dev-env/) base with Python 3, the [Hermes Agent](https://hermes-agent.nousresearch.com/) framework, and systemd support for autonomous operation.

## Quick Start

```bash
dockyard create hermes-agent <your-container-name>
dockyard deploy <your-container-name>
ssh dy-<your-container-name>
```

## What's Added (vs dev-env)

| Component | Purpose |
|---|---|
| Python 3 (11/12/13) | Runtime for the Hermes Agent framework |
| hermes-agent | Autonomous agent orchestration |
| systemd | PID 1 — manages hermes-agent as a user service |
| container-init.sh | Oneshot: SSH host-key generation, boot-seed, permission fixes, user systemd instance start |

### Python Packages

| Package | Purpose |
|---|---|
| hermes-agent | Core Hermes Agent framework |
| openai, anthropic, litellm | LLM client libraries |
| httpx | Async HTTP |
| pydantic | Structured data validation |
| structlog, rich | Logging and terminal output |
| tenacity | Retry logic |

## Persistent Volumes

| Directory | Container Mount | Owner |
|---|---|---|
| `config/` | `~/.config` | agent (2770) |
| `hermes-data/` | `~/.hermes-agent` | agent (2770) |
| `workspace/` | `/workspace` | agent (2770) |
| `nvim-data/` | `~/.local/share/nvim` | agent (2770) |
| `nvim-state/` | `~/.local/state/nvim` | agent (2770) |
| `logs/` | `/logs` | agent (2770) |
| `secrets/` | `/secrets` (read-only) | root (750) |
| `ssh/authorized_keys` | `~/.ssh/authorized_keys` | agent |

## First-Boot Seeding

On first start, `container-init.sh` sources `boot-seed.sh` which:

- Seeds **LazyVim** starter config into `~/.config/nvim/` (skipped if `init.lua` already exists)
- Seeds **TPM** (Tmux Plugin Manager) into `~/.config/tmux/plugins/tpm/` (skipped if already present)
- Installs the **hermes-agent** systemd user service into `~/.config/systemd/user/` (skipped if already present)

All seeding is idempotent — existing config is never overwritten.

## Secrets

An `env.example` file is seeded into `secrets/` on first deploy (never overwrites). Copy it and fill in your API keys:

```bash
cp /secrets/env.example /secrets/env
```

Source it on login by adding to `~/.bashrc`:

```bash
[ -f /secrets/env ] && set -a && source /secrets/env && set +a
```

## LiteLLM Access

The container is configured with `LITELLM_BASE_URL=http://host.gateway.internal:4000` via `extra_hosts` in docker-compose.yml. This routes to the LiteLLM proxy running on the host machine, providing unified access to multiple LLM providers.

## Resource Limits

Default limits set in docker-compose.yml:

| Resource | Limit | Reservation |
|---|---|---|
| CPU | 4 cores | 0.5 cores |
| Memory | 16 GB | 512 MB |

Override in docker-compose.yml or via `dockyard deploy` config.

## Configuration

Hermes Agent configuration is stored in `~/.hermes-agent/` (persistent volume). Key configuration options:

| Environment Variable | Purpose | Default |
|---|---|---|
| `HERMES_MODEL` | LLM model to use | `openai/gpt-4o` |
| `HERMES_MODE` | Agent operation mode | `autonomous` |
| `LITELLM_BASE_URL` | LiteLLM proxy endpoint | `http://host.gateway.internal:4000` |

## Agent Modes

| Mode | Description |
|---|---|
| `autonomous` | Full autonomous operation, agent makes all decisions |
| `assistant` | Semi-autonomous, seeks confirmation for major actions |
| `interactive` | Human-in-the-loop, requires approval for each step |

## Comparison: hermes-agent vs openclaw

| Feature | hermes-agent | openclaw |
|---|---|---|
| Language | Python | Node.js |
| Framework | Hermes Agent (Nous Research) | OpenClaw |
| Package Manager | pip (venv) | npm |
| Browser Automation | Via tool plugins | Chromium (built-in) |
| Task Queue | Via litellm/external | Celery + Redis |
