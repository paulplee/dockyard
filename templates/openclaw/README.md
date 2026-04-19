# openclaw — Autonomous Coding Agent

Extends the [dev-env](../dev-env/) template with LLM client libraries, async HTTP,
task queue support, and LiteLLM proxy access. Designed for autonomous coding agents
that run inside a persistent SSH-accessible container.

## Quick Start

```bash
cd templates/openclaw

# 1. Interactive setup — writes .env to the deployment volume path
make setup

# 2. Deploy
make deploy AGENT_NAME=<name>

# 3. SSH in
ssh -p <SSH_PORT> agent@localhost
```

## What's Added (vs dev-env)

| Package | Purpose |
|---|---|
| openai, anthropic, litellm | LLM client libraries |
| httpx, aiohttp | Async HTTP |
| pydantic | Structured data validation |
| structlog, rich | Logging and terminal output |
| celery, redis | Task queue (matches ae86 Celery stack) |
| tenacity | Retry logic |

## LiteLLM Access

The container is configured with `LITELLM_BASE_URL=http://host.gateway.internal:4000`
via `extra_hosts` in docker-compose.yml. This routes to the LiteLLM proxy running on
the host machine, providing unified access to multiple LLM providers.

## Agent Code

Place your agent code in `/workspace` (bind-mounted from the host). When ready to
auto-start the agent on container boot, update `docker-compose.yml`:

```yaml
command: ["/workspace/start.sh"]
```

The image never needs to change — only the code in `/workspace` evolves.
