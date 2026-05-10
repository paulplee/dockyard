# openwebui — Open WebUI for Hermes Agent

Browser-based chat interface for the [hermes-agent](../hermes-agent/) template's built-in OpenAI-compatible API server.

## Quick Start

Deploy hermes-agent first (it provides the API backend), then:

```bash
dockyard create openwebui <your-container-name>
dockyard deploy <your-container-name>
```

Open WebUI will be available at `http://<host-ip>:8643` (default port).

## Connecting to hermes-agent

The compose file targets `http://host.gateway.internal:8642/v1` by default — the standard port for hermes-agent's API server on the same Docker host. If you used a different `API_PORT` when deploying hermes-agent, update `HERMES_API_URL` in the `.env` file:

```
# $VOLUMES_BASE/.env
HERMES_API_URL=http://host.gateway.internal:<your-api-port>/v1
```

The `OPENAI_API_KEY` in `/secrets/env` must match `API_SERVER_KEY` in your hermes-agent deployment.

## Configuration

Edit `/secrets/env.example`, copy to `/secrets/env`, and fill in:

| Variable | Purpose |
|---|---|
| `OPENAI_API_KEY` | Must match hermes-agent's `API_SERVER_KEY` |
| `WEBUI_SECRET_KEY` | Signs Open WebUI session tokens — set to a random 32-char string |
| `WEBUI_AUTH` | Set to `false` to disable login (LAN-only use) |
| `HERMES_API_URL` | Override API endpoint if hermes-agent is on a different host/port |

Generate secure random values:

```bash
openssl rand -hex 32
```

## Persistent Volumes

| Host path | Container mount | Contents |
|---|---|---|
| `$VOLUMES_BASE/data/` | `/app/backend/data` | Open WebUI database, user accounts, chat history, settings |

## Ports

| Host port | Container port | Service |
|---|---|---|
| `${WEBUI_PORT:-8643}` | `8080` | Open WebUI web interface |

## Architecture

```
browser  ──▶  Open WebUI :8643
                    │
                    │  OpenAI-compatible API
                    ▼
            host.gateway.internal:8642
                    │
                    ▼
            dy-<hermes-container> (hermes-agent API server)
```

Ollama integration is disabled (`ENABLE_OLLAMA_API=false`). To add other OpenAI-compatible backends, configure additional connections from within the Open WebUI admin panel.
