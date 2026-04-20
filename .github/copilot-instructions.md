# Dockyard — Copilot Instructions

## Project overview

dockyard is a Go CLI (`github.com/paulplee/dockyard`) that builds and manages
Dockerised development containers. Templates (Dockerfile + docker-compose.yml +
manifest.yaml) are embedded in the binary via `go:embed`. Configuration is YAML.

## Architecture at a glance

```
cmd/dockyard/          → main entrypoint
internal/cli/          → cobra subcommands (one file per command group)
internal/config/       → Global + Deployment config, YAML ↔ struct, privileged fs ops
internal/template/     → manifest loader, build-context stager, embedded FS consumer
internal/prompt/       → interactive stdin helpers
internal/sshcfg/       → ~/.ssh/config stanza management, authorized_keys install
internal/dockercmd/    → thin wrappers around `docker` and `docker compose`
internal/volumes/      → host volume creation + chown/chmod
assets.go              → root-level go:embed of templates/ and shared/
templates/<name>/      → per-template build context (Dockerfile, docker-compose.yml, manifest.yaml, ...)
shared/                → files copied into build contexts at deploy time (entrypoint.sh)
```

## Conventions

- Go code follows standard `internal/` layout. No exported packages outside `assets.go`.
- Config files are YAML (`config.yaml`, `manifest.yaml`). Legacy `.env` files are
  read for backward compatibility but never produced as the canonical format.
- Container names are always prefixed `dy-` (e.g. `dy-<your-container-name>`).
- Privileged file operations (chown, mkdir in root-owned dirs) fall back to `sudo`
  via helpers in `internal/config/fsutil.go` — never use `os/exec` sudo calls inline.
- Subcommands live in `internal/cli/cmd_*.go`. Group related commands in one file
  (e.g. deploy/up/down/restart share `cmd_deploy.go`; status/list/shell/logs/rm
  share `cmd_inspect.go`).
- Template manifests (`manifest.yaml`) declare `agent_dirs`, `root_dirs`,
  `build_args`, and `shared_files`. The Go code stages a build directory from
  the embedded FS before running `docker compose`.

## Documentation rule

After any successful code change, **always update the relevant documentation**:

- **Root README** (`README.md`): update if the change affects CLI commands,
  repo structure, configuration layout, quick-start instructions, or deployment
  parameters.
- **Template READMEs** (`templates/<name>/README.md`): update if the change
  modifies a template's Dockerfile, docker-compose.yml, manifest.yaml, included
  tools, volumes, or environment variables.
- Keep docs concise and command-focused — show exact `dockyard` invocations,
  not prose descriptions of what to do.

## Style

- Error messages: lowercase, no trailing period, include the failing entity
  (e.g. `unknown deployment "foo"`).
- CLI output: minimal, one line per action (e.g. `>>> Creating host volumes...`).
- No third-party logging library — use `fmt.Printf` / `fmt.Fprintf(os.Stderr, ...)`.
- Tests: table-driven where possible, in `*_test.go` next to the code.
