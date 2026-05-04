# hermes-agent — Nous Research Hermes Agent

Extends the [dev-env](../dev-env/) base with Python 3, the [Hermes Agent](https://hermes-agent.nousresearch.com/) framework, and systemd support for autonomous operation.

## Quick Start

```bash
dockyard create hermes-agent <your-container-name>
dockyard deploy <your-container-name>
ssh dy-<your-container-name>
hermes setup          # first-time: choose provider and model
hermes                # start chatting
```

The web dashboard auto-starts at `http://<host-ip>:9119` (default port).

## What's Added (vs dev-env)

| Component | Purpose |
|---|---|
| Python 3 | Runtime for the Hermes Agent framework |
| hermes-agent `[web]` | Autonomous agent + web dashboard |
| nodejs / npm | Required to build the dashboard frontend |
| systemd | PID 1 — manages hermes-agent and dashboard as user services |
| container-init.sh | Oneshot: venv seeding, service install, service start |

### Python Packages

| Package | Purpose |
|---|---|
| hermes-agent | Core Hermes Agent framework |
| openai, anthropic, litellm | LLM client libraries |
| httpx | Async HTTP |
| pydantic | Structured data validation |
| structlog, rich | Logging and terminal output |
| tenacity | Retry logic |
| python-telegram-bot | Telegram gateway support |

## Persistent Volumes

Everything that Hermes creates — config, memory, skills, sessions, kanban, plugins, cron jobs, conversation history — lives inside `~/.hermes`. That single directory is bind-mounted to the host so **nothing is lost across container restarts or image rebuilds**.

| Host path | Container mount | What lives here |
|---|---|---|
| `$VOLUMES_BASE/hermes/` | `~/.hermes` | **All Hermes data** (see breakdown below) |
| `$VOLUMES_BASE/hermes-venv/` | `~/.hermes-venv` | Python virtualenv with hermes-agent installed |
| `$VOLUMES_BASE/config/` | `~/.config` | Neovim, tmux, and other tool config |
| `$WORKSPACE_PATH` | `/workspace` | Code, projects, files the agent works on (local dir or NAS mount) |
| `$VOLUMES_BASE/nvim-data/` | `~/.local/share/nvim` | Neovim plugin data |
| `$VOLUMES_BASE/nvim-state/` | `~/.local/state/nvim` | Neovim state |
| `$VOLUMES_BASE/logs/` | `/logs` | System/container log output |
| `$VOLUMES_BASE/secrets/` | `/secrets` (read-only) | API keys and tokens |
| `$VOLUMES_BASE/ssh/authorized_keys` | `~/.ssh/authorized_keys` | SSH public keys for login |

`WORKSPACE_PATH` is a build arg set during `dockyard create`. The default is `$VOLUMES_BASE/workspace` (a local directory, appropriate for laptop deploys). For server deployments, set it to an NFS/CIFS mount point (e.g. `/mnt/workspace`) to share a workspace across containers.

### What's inside `~/.hermes`

The `hermes/` volume is the single source of truth for Hermes state. Nothing in this directory is ever lost on restart or upgrade:

| Path | Contents | Risk if lost |
|---|---|---|
| `config.yaml` | Provider, model, tool, and feature config | Must re-run `hermes setup` |
| `SOUL.md` | Agent personality / persona | Personality reset to default |
| `memories/` | Persistent memory (`MEMORY.md`, `USER.md`) | Agent loses what it knows about you |
| `skills/` | Agent-created and user-created skills | All procedural memory lost |
| `state.db` | SQLite — kanban boards, session metadata | Kanban and session index lost |
| `sessions/` | Full conversation history (FTS5 searchable) | Chat history lost |
| `plugins/` | Installed plugins | Plugins must be reinstalled |
| `cron/` | Scheduled automation jobs | All scheduled tasks lost |
| `auth.json` | Provider auth tokens | Must re-authenticate |
| `sandboxes/` | Docker terminal backend metadata | Sandbox references lost |
| `logs/` | Hermes-internal logs | (safe to lose) |
| `profiles/` | Per-profile data (if using `hermes profile`) | All profiles lost |

> **Note**: `hermes-data/` maps `~/.hermes-agent` which Hermes does not currently use — it is reserved for future use and will remain empty.

## Updating Hermes Agent

Hermes is installed from git at image build time. Docker layer caching means `dockyard deploy` alone will **not** pull a newer version unless you bust the cache by changing `HERMES_VERSION`.

### To update to the latest `main`

1. Edit `HERMES_VERSION` in `/opt/dockyard/<name>/config.yaml` (`build_args.HERMES_VERSION`) to any new value — e.g. append today's date:
   ```
   HERMES_VERSION=main-20260503
   ```
2. Delete the staged build directory and redeploy:
   ```bash
   sudo rm -rf /opt/dockyard/<name>/build
   dockyard deploy <name>
   ```
   Docker sees a changed `ARG`, busts the venv layer cache, re-runs `pip install` and `npm run build`. The new venv seed is written to the image; on next boot, `container-init.sh` detects the changed build stamp and re-seeds the live venv volume.

### To pin to a specific release

Set `HERMES_VERSION` to a tag (e.g. `v2026.4.30`) and redeploy as above. Downgrading works the same way.

### What the update replaces vs preserves

| Item | After update |
|---|---|
| Python packages (hermes-agent, deps) | ✅ Updated to new version |
| Dashboard frontend (`web_dist/`) | ✅ Rebuilt from source — frontend matches backend |
| `~/.hermes/` (all user data) | ✅ Untouched — bind mount survives rebuild |
| Neovim config, tmux config | ✅ Untouched — bind mount |
| Workspace files | ✅ Untouched — bind mount |

## First-Boot Seeding

On every container start, `container-init.sh` runs as a systemd oneshot:

1. Compares the image build stamp against the live venv stamp — if different, re-seeds the venv from `/opt/hermes-venv-seed` into `~/.hermes-venv` (preserving user site-packages changes is not supported; the seed always wins when the stamp changes).
2. Installs `hermes-agent.service` and `hermes-dashboard.service` into `~/.config/systemd/user/` if not present.
3. Starts both user services.

All steps are idempotent — if the stamp matches and services are already installed, boot is instant.

## Services

Two systemd user services run inside the container:

| Service | Command | Purpose |
|---|---|---|
| `hermes-agent.service` | `hermes gateway run` | Telegram / Discord / Slack gateway |
| `hermes-dashboard.service` | `hermes dashboard --host 0.0.0.0 --no-open --insecure` | Web dashboard on port 9119 |

Check status:
```bash
systemctl --user status hermes-agent.service
systemctl --user status hermes-dashboard.service
journalctl --user -u hermes-dashboard.service -f
```

> **Security**: The dashboard is exposed on all interfaces (`0.0.0.0`) with `--insecure` — suitable for a trusted LAN only. For internet-facing deployments, put it behind a reverse proxy with authentication, or access it via SSH tunnel: `ssh -L 9119:localhost:9119 dy-<name>`.

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
