# private — Private Application Scaffold

A generic, brand-neutral scaffold for deploying a **closed-source private application** inside a dockyard-managed container.

This template ships **no proprietary code**. It provides the standard dockyard container infrastructure (SSH access, Neovim, tmux, persistent volumes, systemd, Git SSH identity) plus a clearly-marked placeholder section where you wire in your own private application — whether built from source, downloaded as a prebuilt binary, or copied in at build time.

## Quick Start

```bash
dockyard create private <your-container-name>
dockyard deploy <your-container-name>
ssh dy-<your-container-name>
```

## Data Directory Convention

Your private app's data is stored in a directory whose name is **derived from the deployment name** (`CONTAINER_NAME`), keeping the template generic and each deployment fully isolated.

| Location | Path |
|---|---|
| Host | `${VOLUMES_BASE}/${CONTAINER_NAME}-data` |
| Container | `/home/${AGENT_USER}/.${CONTAINER_NAME}` |

**Example** — deployment named `acme`:

| Location | Path |
|---|---|
| Host | `/home/user/.config/dockyard/volumes/acme/acme-data` |
| Container | `/home/agent/.acme` |

The directory is created automatically by `make init` (via the shared `volumes.mk`) and its permissions are fixed on every boot by `container-init.sh`.

## Persistent Volumes

All state lives on the host under `${VOLUMES_BASE}/` (typically `~/.config/dockyard/volumes/<name>/`):

| Host path | Container mount | Purpose |
|---|---|---|
| `${VOLUMES_BASE}/${CONTAINER_NAME}-data/` | `/home/${AGENT_USER}/.${CONTAINER_NAME}` | **Private app data** — persistent app state |
| `${VOLUMES_BASE}/config/` | `~/.config` | App config (Neovim, tmux, git, etc.) |
| `${VOLUMES_BASE}/workspace/` | `/workspace` | Code, projects, and working files |
| `${VOLUMES_BASE}/nvim-data/` | `~/.local/share/nvim` | Neovim plugins (persist across rebuilds) |
| `${VOLUMES_BASE}/nvim-state/` | `~/.local/state/nvim` | Neovim session state |
| `${VOLUMES_BASE}/logs/` | `/logs` | Structured log output |
| `${VOLUMES_BASE}/secrets/` | `/secrets` (read-only) | API keys and tokens |
| `${VOLUMES_BASE}/ssh/` | `~/.ssh` | SSH authorized keys + persistent git identity keypair |
| `/sys/fs/cgroup` | `/sys/fs/cgroup:rw` | cgroup v2 filesystem (required by systemd) |

## Wiring In Your Private App

The [`Dockerfile`](Dockerfile) contains a commented section titled **"Private app installation"** between the SSH hardening block and the workspace layout block. Uncomment and adapt **one** of the patterns below.

### Pattern A — Build from a private git repository

Use this when you have source code in a private GitHub/GitLab repository and want to compile it inside the container image.

#### Step 1: Create a personal access token

- **GitHub**: Settings → Developer settings → Personal access tokens → Fine-grained tokens
  - Repository access: Only select repositories
  - Permissions: Contents (read-only)
- **GitLab**: Preferences → Access Tokens
  - Scopes: `read_repository`

#### Step 2: Edit the Dockerfile

Uncomment the `RUN --mount=type=secret,id=private_repo_token ...` block and replace the placeholders:

```dockerfile
RUN --mount=type=secret,id=private_repo_token \
    TOKEN=$(cat /run/secrets/private_repo_token) \
    git clone https://oauth2:${TOKEN}@github.com/<org>/<repo>.git \
          --branch ${PRIVATE_APP_VERSION} --depth 1 /opt/private-app \
    && cd /opt/private-app \
    && <build steps> \
    && ln -sf /opt/private-app/<binary> /usr/local/bin/private-app
```

#### Step 3: Pass the secret at build time

The token is **never written into the image**. It is injected via BuildKit's `--secret` flag at build time. You must patch the Makefile's `up` target to pass it through.

Add the `--secret` flag to the `docker compose build` invocation. Edit the `up` target in [`Makefile`](Makefile):

```makefile
up: prepare
	docker compose --env-file $(VOLUMES_BASE)/.env build \
	  --secret id=private_repo_token,src=$(HOME)/.config/dockyard/tokens/private_repo_token
	docker compose --env-file $(VOLUMES_BASE)/.env up -d
```

Then store your token on the host:

```bash
mkdir -p ~/.config/dockyard/tokens
# Write your token — ensure no trailing newline
printf '%s' 'ghp_your_token_here' > ~/.config/dockyard/tokens/private_repo_token
chmod 600 ~/.config/dockyard/tokens/private_repo_token
```

#### Common build-step examples by language

| Language | Build commands |
|---|---|
| **Python** | `python3 -m venv /opt/private-app/.venv && /opt/private-app/.venv/bin/pip install .` |
| **Node.js** | `npm ci && npm run build` |
| **Go** | `go build -o /usr/local/bin/private-app .` |
| **Rust** | `cargo build --release && cp target/release/app /usr/local/bin/private-app` |
| **Makefile / generic** | `make && cp ./build/app /usr/local/bin/private-app` |

### Pattern B — Install a prebuilt binary

Choose between downloading from a private URL or copying a local binary.

#### Pattern B1 — Download from a private release URL

Use when your binary is hosted behind authentication (e.g., GitHub Releases, a private S3 bucket, or an artifact server).

```dockerfile
RUN --mount=type=secret,id=private_repo_token \
    TOKEN=$(cat /run/secrets/private_repo_token) \
    && curl -fsSL -H "Authorization: Bearer ${TOKEN}" \
       "https://releases.example.com/private-app-${PRIVATE_APP_VERSION}-linux-amd64" \
       -o /usr/local/bin/private-app \
    && chmod +x /usr/local/bin/private-app
```

#### Pattern B2 — COPY a binary placed next to the Dockerfile

Place your binary in `templates/private/` alongside the Dockerfile, then uncomment:

```dockerfile
COPY private-app /usr/local/bin/private-app
RUN chmod +x /usr/local/bin/private-app
```

> ⚠️ This embeds the binary directly in the image. For development iteration, prefer Pattern B1 or a bind mount.

### Adding a systemd service

If your app runs as a long-lived process (a web server, background worker, gateway, etc.), add a systemd user service so dockyard manages its lifecycle.

#### 1. Create a `.service` file

Place it in the image during the "Private app installation" section of the Dockerfile. For example, after the binary installation:

```dockerfile
COPY private-app.service /usr/local/share/dockyard/private-app.service
```

#### 2. Enable linger for the agent user

The Dockerfile already contains:

```dockerfile
RUN mkdir -p /var/lib/systemd/linger \
    && touch /var/lib/systemd/linger/${AGENT_USER}
```

If this line is not present (check the Dockerfile), add it after the user creation block.

#### 3. Wire it up in container-init.sh

Uncomment the **"Private app service startup"** block at the end of [`container-init.sh`](container-init.sh). It performs these steps on every boot:

```bash
# ── Private app service startup ─────────────────────────────────────────────
SVC_DIR="/home/${AGENT_USER}/.config/systemd/user"
if [ ! -f "${SVC_DIR}/private-app.service" ]; then
    echo ">>> Installing private-app systemd user service..."
    su -s /bin/bash "${AGENT_USER}" -c \
        "mkdir -p ${SVC_DIR}/default.target.wants \
         && cp /usr/local/share/dockyard/private-app.service ${SVC_DIR}/private-app.service \
         && ln -sf ../private-app.service ${SVC_DIR}/default.target.wants/private-app.service"
fi

AGENT_UID="$(id -u "${AGENT_USER}")"
if ! systemctl is-active --quiet "user@${AGENT_UID}.service" 2>/dev/null; then
    echo ">>> Starting user@${AGENT_UID} systemd instance..."
    systemctl start "user@${AGENT_UID}.service" || true
fi

DBUS="unix:path=/run/user/${AGENT_UID}/bus"
su -s /bin/bash "${AGENT_USER}" -c \
    "XDG_RUNTIME_DIR=/run/user/${AGENT_UID} DBUS_SESSION_BUS_ADDRESS=${DBUS} \
     systemctl --user start private-app.service 2>/dev/null || true"
```

The block is idempotent — the service file is only copied if absent, and `systemctl start` is safe to run repeatedly.

## Build Arguments

These arguments are prompted during `dockyard create private <name>`:

| Argument | Default | Description |
|---|---|---|
| `PRIVATE_APP_VERSION` | `main` | Git ref, tag, or release version for your private app. Used as a Docker layer cache-bust key — change this value to force a rebuild. |
| `PRIVATE_APP_PORT` | `7700` | Port on the host that maps to the app container port. Set to `0` if your app does not expose a network service. |
| `DNS_SERVER` | `10.10.1.2` | Primary DNS resolver injected into the container. The secondary is always `1.1.1.1`. |

## Git SSH Access

On first boot, `container-init.sh` generates a persistent `ed25519` keypair at `~/.ssh/id_ed25519` (inside the `$VOLUMES_BASE/ssh/` bind mount) — so the key **survives reboots and container rebuilds**.

If you pre-place your own `id_ed25519` + `id_ed25519.pub` in `$VOLUMES_BASE/ssh/` before starting, auto-generation is skipped.

### Authorise the key on GitHub / GitLab

1. Read the generated public key:
   ```bash
   cat /logs/git-ssh-pubkey.txt
   ```
   (Also available inside the container at `~/.ssh/id_ed25519.pub`.)

2. Add it to your account:
   - **GitHub**: Settings → SSH and GPG keys → New SSH key
   - **GitLab**: Preferences → SSH Keys → Add key

3. Verify from inside the container:
   ```bash
   ssh -T git@github.com
   ssh -T git@gitlab.com
   ```

### Set commit identity

Add to `/secrets/env` (see [Persistent Volumes](#persistent-volumes)):

```bash
GIT_AUTHOR_NAME=Private App
GIT_AUTHOR_EMAIL=app@example.com
```

`container-init.sh` reads these on every boot and writes them to `~/.config/git/config`, which persists via the `config/` volume.

## Resource Limits

Default limits set in [`docker-compose.yml`](docker-compose.yml):

| Resource | Limit | Reservation |
|---|---|---|
| CPU | 4 cores | 0.5 cores |
| Memory | 16 GB | 512 MB |
