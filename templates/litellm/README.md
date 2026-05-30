# litellm — LiteLLM Proxy Gateway

OpenAI-compatible API proxy that routes LLM calls through a VPS, masking your origin location from upstream providers (OpenAI, Anthropic, Gemini, Groq, …).

Clients connect to this proxy using a standard OpenAI SDK — no client-side changes required beyond pointing `base_url` at your VPS.

## Quick Start

```bash
dockyard create litellm <your-container-name>
dockyard deploy <your-container-name>
```

Before starting the container, configure the proxy:

```bash
# 1. Copy the proxy config example and add your model routing rules
cp $VOLUMES_BASE/config/proxy_config.yaml.example \
   $VOLUMES_BASE/config/proxy_config.yaml

# 2. Copy the secrets template and fill in your API keys + master key
cp $VOLUMES_BASE/secrets/env.example $VOLUMES_BASE/secrets/env
$EDITOR $VOLUMES_BASE/secrets/env

# 3. Start the container
dockyard up <your-container-name>
```

The proxy is available at `http://<vps-ip>:4000` (default port).

## Client Configuration

Point any OpenAI-compatible client at the proxy. The endpoint is plain HTTP — do not use `https://`.

```bash
# Text
curl -s http://<vps-ip>:4000/v1/chat/completions \
  -H "Authorization: Bearer $LITELLM_MASTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "gemini-flash", "messages": [{"role": "user", "content": "hi"}]}'
```

```bash
# Vision / OCR — always send images as base64; URL-based image fetching is
# blocked by most CDNs when the request originates from a VPS IP.
#
# macOS:  base64 -b0 image.png
# Linux:  base64 -w0 image.png
IMG=$(base64 -w0 /path/to/image.png)   # Linux; use -b0 on macOS

curl -s http://<vps-ip>:4000/v1/chat/completions \
  -H "Authorization: Bearer $LITELLM_MASTER_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"gemini-flash\",
    \"messages\": [{
      \"role\": \"user\",
      \"content\": [
        {\"type\": \"text\", \"text\": \"Extract all text from this image.\"},
        {\"type\": \"image_url\", \"image_url\": {\"url\": \"data:image/png;base64,${IMG}\"}}
      ]
    }]
  }" | python3 -c "import sys,json; print(json.load(sys.stdin)['choices'][0]['message']['content'])"
```

```python
# Python (openai SDK)
from openai import OpenAI
import base64, pathlib

client = OpenAI(
    base_url="http://<vps-ip>:4000/v1",
    api_key="<your-LITELLM_MASTER_KEY>",
)

# Text
response = client.chat.completions.create(
    model="gemini-flash",
    messages=[{"role": "user", "content": "hi"}],
)

# Vision / OCR
img_b64 = base64.b64encode(pathlib.Path("image.png").read_bytes()).decode()
response = client.chat.completions.create(
    model="gemini-flash",
    messages=[{"role": "user", "content": [
        {"type": "text", "text": "Extract all text from this image."},
        {"type": "image_url", "image_url": {"url": f"data:image/png;base64,{img_b64}"}},
    ]}],
)
```

```bash
# hermes-agent — set in $VOLUMES_BASE/secrets/env
LITELLM_BASE_URL=http://<vps-ip>:4000
OPENAI_API_KEY=<your-LITELLM_MASTER_KEY>
```

## Configuration

### Proxy config (`$VOLUMES_BASE/config/proxy_config.yaml`)

Defines model aliases, routing rules, and global settings. Edit this file to add or remove providers. Changes take effect on container restart.

See `proxy_config.yaml.example` for a full annotated example. By default the template ships with Gemini 2.5 Flash; add other providers as needed.

### Secrets (`$VOLUMES_BASE/secrets/env`)

| Variable | Purpose |
|---|---|
| `LITELLM_MASTER_KEY` | Bearer token required by all clients — generate with `openssl rand -hex 32` |
| `GEMINI_API_KEY` | Google Gemini provider key |
| `OPENAI_API_KEY` | OpenAI provider key (optional) |
| `ANTHROPIC_API_KEY` | Anthropic provider key (optional) |
| `GROQ_API_KEY` | Groq provider key (optional) |

Generate a secure master key:

```bash
openssl rand -hex 32
```

## Persistent Volumes

| Host path | Container mount | Contents |
|---|---|---|
| `$VOLUMES_BASE/config/proxy_config.yaml` | `/app/config.yaml` | Proxy routing rules |
| `$VOLUMES_BASE/data/` | `/app/data/` | SQLite usage + virtual key DB |
| `$VOLUMES_BASE/secrets/env` | injected via `env_file` | Provider API keys |

## Updating

```bash
docker compose -f $VOLUMES_BASE/build/docker-compose.yml \
  --env-file $VOLUMES_BASE/.env \
  pull

docker compose -f $VOLUMES_BASE/build/docker-compose.yml \
  --env-file $VOLUMES_BASE/.env \
  up -d --force-recreate
```

## Security Notes

- Keep `LITELLM_MASTER_KEY` secret — anyone with it can call upstream providers at your cost.
- The default compose exposes port 4000 on all interfaces. On a public VPS, consider placing Nginx or Caddy in front with TLS.
- Provider keys in `secrets/env` are never embedded in the image — they are injected at runtime only.
- `redact_user_api_key_info: true` is set in the proxy config to keep keys out of logs.
