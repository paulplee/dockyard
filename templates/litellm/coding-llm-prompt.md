# LiteLLM Proxy — Coding LLM Configuration Prompt

Paste the block below into your coding assistant's system prompt, config file,
or environment variables. Replace `<MASTER_KEY>` with your `LITELLM_MASTER_KEY`.

---

## Prompt block

```
You have access to a private LLM API proxy. Use the following settings for all
LLM API calls:

  Base URL : http://b2ain.com:4000/v1
  API key  : <MASTER_KEY>
  Model    : gemini-flash          (routes to Google Gemini 2.5 Flash)

The proxy is OpenAI-compatible. Use the standard OpenAI SDK or any tool that
accepts a custom base_url. Do not call api.openai.com or any provider endpoint
directly — always route through the proxy.

For vision or OCR tasks, always encode the image as base64 and pass it as
a data URI — do not pass public image URLs, as CDNs block VPS fetch requests.
Gemini 2.5 Flash handles vision and OCR natively.
```

---

## Environment variables (for tools that read from env)

```bash
export OPENAI_BASE_URL="http://b2ain.com:4000/v1"
export OPENAI_API_KEY="<MASTER_KEY>"
export OPENAI_MODEL="gemini-flash"
```

## SDK snippet (Python)

```python
import base64, pathlib
from openai import OpenAI

client = OpenAI(
    base_url="http://b2ain.com:4000/v1",
    api_key="<MASTER_KEY>",
)

# Text
response = client.chat.completions.create(
    model="gemini-flash",
    messages=[{"role": "user", "content": "..."}],
)

# Vision / OCR — base64 only; URL image fetching is blocked by CDNs on VPS IPs
img_b64 = base64.b64encode(pathlib.Path("image.png").read_bytes()).decode()
response = client.chat.completions.create(
    model="gemini-flash",
    messages=[{"role": "user", "content": [
        {"type": "text", "text": "Extract all text from this image."},
        {"type": "image_url", "image_url": {"url": f"data:image/png;base64,{img_b64}"}},
    ]}],
)
```

## curl — Vision / OCR

```bash
# macOS: base64 -b0   Linux: base64 -w0
IMG=$(base64 -w0 /path/to/image.png)

curl -s http://b2ain.com:4000/v1/chat/completions \
  -H "Authorization: Bearer <MASTER_KEY>" \
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

## Available model aliases

| Alias | Upstream model |
|---|---|
| `gemini-flash` | `gemini/gemini-2.5-flash` |
| `gemini-pro` | `gemini/gemini-2.5-flash` (alias for compatibility) |
