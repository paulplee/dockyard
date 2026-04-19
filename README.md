# dockyard

Best-practice Docker container templates for purpose-built environments.
Each template is a self-contained, parameterized deployment — just `make setup`
and `make deploy`.

## Templates

| Template | Purpose | Status |
|---|---|---|
| [dev-env](templates/dev-env/) | SSH dev box: Neovim, tmux, Python, Node.js | Ready |
| [openclaw](templates/openclaw/) | Autonomous coding agent with LLM tooling | Ready |
| crawler | Web scraper / Scrapy worker | Planned |
| data-worker | Data pipeline worker | Planned |
| llm-inference | LLM serving (vLLM / Ollama) | Planned |

## Philosophy

- **Self-contained templates.** Each template has its own Dockerfile, docker-compose.yml,
  Makefile, and README. No shared base image — every template builds independently.
- **Parameterized deployments.** Instance-specific values (name, UID, SSH port) live in
  `.env` files outside the repo, in the deployment volume path. Nothing is hardcoded.
- **Shared operational logic.** Common Makefile targets (setup, group, init, volumes,
  docker operations) live in `shared/makefiles/` and are included by each template.
- **Host volume permissions done right.** A shared `agents` group with setgid (2770)
  lets the deploying user and the container agent user both read/write persistent volumes.
- **SSH-first access.** Containers are accessed via SSH (key-only), not `docker exec`.
  This works identically whether the container is local or remote.
- **Security defaults.** No root login, no password auth, secrets mounted read-only,
  resource guardrails to prevent runaway containers.

## Repo Structure

```
dockyard/
├── README.md
├── .gitignore
├── shared/
│   ├── entrypoint.sh              # Common startup: fix perms, start sshd, exec CMD
│   └── makefiles/
│       ├── setup.mk               # Interactive .env generator
│       ├── group.mk               # Host group creation
│       ├── volumes.mk             # Host volume init + chown
│       └── docker.mk              # up/down/logs/shell/clean/reset
└── templates/
    ├── dev-env/                   # Plain dev environment
    │   ├── Dockerfile
    │   ├── docker-compose.yml
    │   ├── Makefile
    │   └── README.md
    └── openclaw/                  # Autonomous coding agent
        ├── Dockerfile
        ├── docker-compose.yml
        ├── requirements.txt
        ├── Makefile
        └── README.md
```

## Quick Start

```bash
# Pick a template
cd templates/openclaw

# Interactive setup — prompts for name, UID, SSH port
# Writes .env to the deployment volume path (outside the repo)
make setup

# Deploy: creates host dirs, sets permissions, builds image, starts container
make deploy CONTAINER_NAME=<your-container-name>

# SSH in
ssh -p 2201 agent@localhost
```

## Deployment Parameters

All instance-specific values are stored in `.env` at the deployment volume path
(`<VOLUMES_ROOT>/<CONTAINER_NAME>/.env`), never in the repo. The host-level
`VOLUMES_ROOT` is stored in `~/.config/dockyard/.env` (written by `make setup`).

| Variable | Example | Purpose |
|---|---|---|
| `CONTAINER_NAME` | `(<your-container-name>)` | Container name, hostname, volume paths |
| `AGENT_UID` | `1101` | UID of the agent user inside the container |
| `AGENT_GID` | `1101` | GID (usually matches UID) |
| `SSH_PORT` | `2201` | Host port mapped to container SSH |

Multiple instances of the same template can run on one host — just use different
names, UIDs, and ports.

## Host Prerequisites

- Docker Engine with Compose v2
- `sudo` access (for volume directory creation and group management)

## Adding a New Template

1. Create a new directory under `templates/`.
2. Add a `Dockerfile`, `docker-compose.yml`, `Makefile`, and `README.md`.
3. Include the shared makefiles: `include ../../shared/makefiles/*.mk`
4. Copy `shared/entrypoint.sh` into the build context via a `prepare` target.

See [templates/dev-env/](templates/dev-env/) for the minimal example.
