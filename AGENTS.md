# dockyard — Coding Agent Instructions

## Project Overview

dockyard is an open-source Go CLI for building and managing Dockerised
development containers. It is the independent foundation that downstream
products build on by importing `pkg/dockyard` as a Go module and registering
extra templates via `SetAdditionalFS`.

Templates (Dockerfile + docker-compose.yml) are embedded in the binary — one
command to create, one to deploy.

## Commit Hygiene (important)

dockyard is a **standalone open-source project**. Its commit history must read
as a self-contained foundation with **no references to any downstream product**
(no product names, no product-specific rationale).

- When a change is motivated by a downstream need, frame it generically in
  dockyard's own terms — e.g. "export StageAndEnv so downstream consumers can
  reuse staging logic without duplicating it", never "operon needs StageAndEnv
  for its pack system".
- Refer to "downstream products" or "consumers of `pkg/dockyard`", never to a
  specific product by name.
- If a change is genuinely dockyard-internal, just describe it directly.

This rule is enforced by a `commit-msg` git hook (see `githooks/commit-msg`)
that rejects any dockyard commit message mentioning `operon`. Install it with:

```bash
git config core.hooksPath githooks
```

## Architecture

- **Public API surface is `pkg/dockyard`** — the `Engine`, `ExtensionHooks`,
  `SetAdditionalFS`, and selected exported helpers in `cli/` (e.g.
  `StageAndEnv`). Changes here are consumed by downstream modules; keep them
  stable and documented.
- **Single binary**: templates, entrypoints, and manifests are embedded via
  `go:embed`. No files to locate at runtime.
- **Self-contained templates**: each `templates/<name>/` has its own
  `Dockerfile`, `docker-compose.yml`, and `manifest.yaml`. No shared base image.
- **Typed configuration**: YAML config with struct validation replaces raw
  `.env` files. Legacy `.env` deployments are read transparently.
- **No downstream code duplication**: if a downstream product needs dockyard
  behaviour, expose it here (e.g. as an exported helper or an extension hook)
  rather than letting the product fork the logic.

## Repo Structure

```
cmd/dockyard/main.go       # CLI entrypoint
cli/                       # cobra subcommands (create, deploy, up, down, ...)
config/                    # Global + Deployment config, fsutil
template/                  # manifest loader, build-context stager
prompt/                    # interactive stdin helpers
sshcfg/                    # ~/.ssh/config management
dockercmd/                 # docker / docker compose wrappers
volumes/                   # host volume creation + permissions
pkg/dockyard/engine.go     # public Engine API + ExtensionHooks
assets.go                  # go:embed for templates/ and shared/
templates/                 # dev-env, hermes-agent, litellm, openclaw, openwebui
shared/                    # entrypoint.sh and other shared files
githooks/                  # project git hooks (commit-msg)
```

## Build

```bash
go build -o bin/dockyard ./cmd/dockyard
```

Requires Go 1.22+, Docker Engine with Compose v2, and `sudo` access.

## Key Rules

- **dockyard is the foundation.** Features that only benefit a specific
  downstream product belong in that product, not here.
- **Public API stays stable.** `pkg/dockyard` and exported `cli/` helpers are a
  contract with downstream consumers — change them deliberately and document
  the change in the commit message.
- **No product names in commits.** See "Commit Hygiene" above.
- **Host volume permissions**: a shared `agents` group with setgid (2770) lets
  the deploying user and the container agent user both read/write persistent
  volumes.
- **Security defaults**: no root login, no password auth, secrets mounted
  read-only.

## Adding a New Template

1. Create `templates/<name>/` with a `Dockerfile`, `docker-compose.yml`, and
   `manifest.yaml`.
2. The manifest declares `agent_dirs`, `root_dirs`, `build_args`, and
   `shared_files`. See `templates/dev-env/manifest.yaml` for a minimal example.
3. Rebuild the binary (`go build ./cmd/dockyard`) — the new template is
   automatically embedded and will appear in `dockyard templates`.
