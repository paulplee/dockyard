# openclaw — Autonomous Coding Agent

Extends the [dev-env](../dev-env/) base with Node.js, the OpenClaw CLI, systemd
(for the openclaw-gateway user service), and LiteLLM proxy access. Designed for
autonomous coding agents running inside persistent SSH-accessible containers.

## Quick Start

```bash
dockyard create openclaw <your-container-name>
dockyard deploy <your-container-name>
ssh dy-<your-container-name>
openclaw onboard    # first-time: configure provider, model, and gateway
openclaw gateway    # start the gateway
```

## What's Added (vs dev-env)

| Component | Purpose |
|---|---|
| Node.js (22 or 24) | Runtime for the openclaw CLI |
| openclaw CLI | Agent orchestration and gateway |
| Chromium | Browser automation for tools |
| Homebrew (Linuxbrew) | Skill installation via `brew install` (ephemeral — see below) |
| systemd | PID 1 — manages openclaw-gateway as a user service |
| container-init.sh | Oneshot: SSH key gen, boot-seed, service install, workspace path fix |
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
| `OPENCLAW_VERSION` | `latest` | any npm tag or version | Changing this busts the Docker layer cache and forces a fresh npm install on next deploy |

Override at deploy time by editing `$VOLUMES_BASE/<name>/config.yaml` and
re-running `dockyard deploy <name>`.

## Persistent Volumes

Everything OpenClaw creates — config, credentials, sessions, memories, cron jobs — lives inside `~/.openclaw`. The workspace (skills, AGENTS.md, code) lives at `/workspace`. Both are bind-mounted to the host so **nothing is lost across container restarts or image rebuilds**.

| Host path (under `VOLUMES_BASE`) | Container mount | What lives here |
|---|---|---|
| `openclaw-data/` | `~/.openclaw` | **All OpenClaw data** (see breakdown below) |
| `workspace/` | `/workspace` | Agent workspace: code, skills, AGENTS.md, SOUL.md |
| `config/` | `~/.config` | systemd user services, nvim, tmux config |
| `nvim-data/` | `~/.local/share/nvim` | Neovim plugin data |
| `nvim-state/` | `~/.local/state/nvim` | Neovim state |
| `logs/` | `/logs` | Container log output |
| `secrets/` | `/secrets` (read-only) | API keys and tokens |
| `ssh/authorized_keys` | `~/.ssh/authorized_keys` | SSH public keys for login |

### What's inside `~/.openclaw`

| Path | Contents | Risk if lost |
|---|---|---|
| `openclaw.json` | Provider, model, gateway, and tool config | Must re-run `openclaw onboard` |
| `credentials/` | Auth tokens (Telegram, Discord, etc.) | Must re-authenticate all channels |
| `workspace/` | Not used — redirected to `/workspace` by container-init.sh | — |

### What's inside `/workspace` (the agent workspace)

OpenClaw's workspace root is redirected from `~/.openclaw/workspace` to `/workspace` by `container-init.sh` on every boot. This keeps agent files on the dedicated `workspace` volume, separate from openclaw config.

| Path | Contents | Risk if lost |
|---|---|---|
| `skills/` | File-drop skills (`.md` files, SKILL.md) | All custom skills lost |
| `AGENTS.md` | Agent system prompt injection | Context lost |
| `SOUL.md` | Agent persona | Persona reset |
| `TOOLS.md` | Tool configuration prompt | Tool config lost |
| (code files) | Files the agent creates and edits | Work lost |

### Homebrew — ephemeral by design

Homebrew (`/home/linuxbrew/.linuxbrew`) is baked into the image and is **not** bind-mounted. Packages installed with `brew install <skill>` write to the container's writable layer and are **lost when the container is recreated**.

**For persistent skills, always use file-drop instead:**
```bash
mkdir -p /workspace/skills/my-skill
# create /workspace/skills/my-skill/SKILL.md
```
This lands in the `workspace` volume and survives all restarts and upgrades.

## Updating OpenClaw

OpenClaw is installed from npm at image build time. Docker layer caching means `dockyard deploy` alone will **not** pull a newer version unless you bust the cache by changing `OPENCLAW_VERSION`.

### To update to the latest release

1. Edit `OPENCLAW_VERSION` in the deployment config to any new value — e.g. the exact version tag:
   ```
   OPENCLAW_VERSION=2026.4.29
   ```
   Or append today's date to `latest` to force a cache miss:
   ```
   OPENCLAW_VERSION=latest-20260503
   ```
   > Note: only real npm versions/tags work for the actual install. Using a fake suffix won't change what npm installs — to get truly "latest", use a real version number. Get the current latest version from: `npm view openclaw version`

2. Delete the staged build directory and redeploy:
   ```bash
   sudo rm -rf /opt/dckyard/<name>/build
   dockyard deploy <name>
   ```

### To pin to a specific release

Set `OPENCLAW_VERSION=2026.4.29` (or any npm version string) and redeploy as above.

### What the update replaces vs preserves

| Item | After update |
|---|---|
| openclaw binary + node_modules | ✅ Updated (baked into new image) |
| `~/.openclaw/` (config, credentials, sessions) | ✅ Untouched — bind mount |
| `/workspace` (skills, code, AGENTS.md) | ✅ Untouched — bind mount |
| Homebrew-installed packages | ❌ Lost — must reinstall after each image rebuild |

## First-Boot Seeding

On every container start, `container-init.sh` runs as a systemd oneshot:

1. Regenerates SSH host keys if missing.
2. Seeds LazyVim and TPM configs (idempotent — never overwrites existing config).
3. Installs `openclaw-gateway.service` into `~/.config/systemd/user/` if not present.
4. Fixes the openclaw workspace path in `~/.openclaw/openclaw.json` to point to `/workspace` if it references a non-existent path.
5. Ensures the user systemd instance is running and starts the gateway service.

## Services

One systemd user service runs inside the container:

| Service | Command | Purpose |
|---|---|---|
| `openclaw-gateway.service` | `openclaw gateway` | Messaging gateway (Telegram, Discord, Slack, etc.) |

Check status:
```bash
systemctl --user status openclaw-gateway.service
journalctl --user -u openclaw-gateway.service -f
```

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
