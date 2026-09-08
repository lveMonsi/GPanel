# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository overview

GPanel is a Vue 3 server-management panel with a Go backend split into a control-plane service and a host agent. The repository contains three independent Go modules (`core`, `agent`, and `tools`) plus the Vite frontend; it is not a Go workspace.

- `frontend/` — Vue 3 + TypeScript UI, built with Vite. API wrappers live under `src/api`, reusable UI under `src/components`, and routed screens under `src/views`.
- `core/` — the main Gin HTTP service and control plane. It owns panel authentication, host/settings/session/quick-command data, and the embedded frontend. Its route handlers are in `controllers`, business logic in `service`, persistence in `repo`/`models`, and cross-cutting behavior in `middleware`.
- `agent/` — the host-side Gin service for operating on the machine: files, processes, firewall, SSH, terminal, monitoring, cron, and system logs. It has its own models/repositories/services and protects `/api` with an API key.
- `tools/` — the `gpctl` administrative CLI used for service/configuration management and upgrades.
- `gpanel.sh` — root-only installation/update/uninstallation/status script for released Linux binaries and systemd services.
- `Makefile` — local build, packaging, installation, and deployment entry points.

### Request flow and runtime boundaries

The browser talks to `core` under `/api/v1`. `core/routes/router.go` registers panel APIs, embeds `core/routes/web/dist` with `go:embed`, serves the SPA fallback, and exposes WebSocket/terminal endpoints. Operations that require host-level functionality are handled by core controllers and forwarded through `core/utils/agent_client.go` to the agent, preferably over the Unix socket `/var/run/gpanel/agent.sock` and otherwise over the configured agent HTTP address (default `localhost:9998`). The client sends `X-API-Key` when configured.

`agent` initializes its own SQLite database, migrations, cron scheduler, and monitoring scheduler before registering its API routes. Keep changes to core-facing DTOs/routes and agent-facing DTOs/routes compatible: many UI features cross both services. Both services use Gin and GORM with SQLite, but their databases and module paths are separate.

The frontend production output must exist at `core/routes/web/dist` before compiling `core`, because `core/routes/router.go` embeds that directory. `core` defaults to port `8080`; agent defaults to `9998`. Core reads database-backed settings through its configuration cache and environment overrides such as `GPANEL_JWT_SECRET` and `GPANEL_AGENT_API_KEY`.

## Development commands

Prerequisites used by CI: Go 1.23 and Node.js 20. Run Go commands from the relevant module directory.

### Install frontend dependencies and develop the UI

```bash
cd frontend
npm ci                         # reproducible install
npm run dev                    # Vite dev server on http://localhost:5173
npm run build                  # default Vite build
npm run build:pro              # production-mode build used for releases
npm run preview                # preview a built frontend
```

The Vite dev server proxies `/api` to `http://localhost:8080`; start core separately when exercising API calls. There is no configured npm lint or test script. For a type check, use `npx vue-tsc --noEmit` from `frontend` (the package already includes `vue-tsc`).

### Run services locally

Do not use `go run` or any other local Go compilation command in the development environment. Use GitHub Actions-generated release artifacts when a compiled Core, Agent, or `gpctl` binary is needed. The frontend development server may still be used for UI work:

```bash
cd frontend
npm ci
npm run dev                    # Vite dev server on http://localhost:5173
```

The Vite dev server proxies `/api` to `http://localhost:8080`; API-backed UI testing therefore requires a Core/Agent pair obtained from the remote release workflow and configured for the local environment. Core and Agent default to ports `8080` and `9998` respectively.

For local development, core and agent may use their default ports, or settings/environment configuration can be supplied as described in `core/global/config_cache.go` and `agent/global/config.go`. Core can use the agent over the Unix socket when it exists; otherwise ensure the configured agent address is reachable.

### Build and release binaries

**Do not compile binaries or run a full production build in the development environment.** Builds must be performed remotely by GitHub Actions and published to GitHub Releases. After making buildable changes, commit and push them to the appropriate remote branch/tag, then use the repository workflows:

- `.github/workflows/prerelease.yml` runs for pushes to `dev` (and can be started manually) and publishes a prerelease.
- `.github/workflows/release.yml` runs for `v*` tags (and can be started manually) and publishes a formal release.

Use the GitHub Actions run and its generated release assets to verify compilation and obtain binaries. Do not run `make build`, `make build_linux`, `go build`, `npm run build`, or equivalent local production compilation commands merely to validate a change. Local source-level checks and tests remain allowed where the required toolchains are available.

The repository Makefile contains these build targets for CI/reference purposes only:

```bash
make build
make build_linux
make build_frontend
make build_core
make build_core_linux
make build_agent
make build_agent_linux
make build_gpctl
make build_gpctl_linux
make clean
```

The release workflows install Go 1.23 and Node.js 20, build the frontend, copy its output to `core/routes/web/dist`, then produce `gpanel`, `gpanel-agent`, and `gpctl` for Linux amd64, Linux arm64, and Windows amd64. They create archives and SHA-256 checksums and attach them to GitHub Releases. Push only intentional commits/tags because pushing can trigger a release build.

The current Makefile uses `core/web/dist` for its copy target, while the Go embed directive uses `core/routes/web/dist`; CI workflows use the correct embed path. Do not work around this by performing a local production build; fix the source/configuration and let GitHub Actions validate it remotely.

### Test, format, and vet

Only the agent currently contains Go tests (`agent/service/ssh_test.go`). Run the full test suite per module (this also works for packages with no tests):

```bash
(cd agent && go test ./...)
(cd core && go test ./...)
(cd tools && go test ./...)
```

Run one test or one package with:

```bash
cd agent
 go test ./service -run '^TestParseSSHLogDate$' -count=1
 go test ./service -run 'TestParseSSHLogDateYearBoundary' -count=1
```

Before submitting Go changes, use the repository's standard toolchain commands:

```bash
for module in core agent tools; do
  (cd "$module" && find . -name '*.go' -type f -print0 | xargs -0 gofmt -w && go vet ./...)
done
```

For a focused change, prefer `gofmt -w path/to/changed.go`. There is no repository-wide configured linter beyond `go vet` and the frontend type checker.

### Deployment-related commands

`make install` and `make deploy` modify `/opt/gpanel`, `/var/lib/gpanel`, `/var/log/gpanel`, `/usr/local/bin`, and systemd unit files, and require root/systemd. Do not use them for ordinary development. The equivalent installed services are `gpanel-agent` and `gpanel`; logs are commonly inspected with `journalctl -u gpanel -u gpanel-agent -f`.

## Change-location guidance

- Add or change panel HTTP endpoints in `core/routes/router.go` and the corresponding `core/controllers`/`service`/`repo` layers. Apply `middleware.Auth()` to protected panel endpoints.
- Add or change host operations in the matching `agent/routes`, `agent/controllers`, `agent/service`, and `agent/repo` layers, then update the core proxy/controller and frontend API module when the operation is user-facing.
- Keep request/response shapes synchronized across `core/dto`, `agent/dto`, and `frontend/src/api/interface`.
- Database tables are auto-migrated during service startup in each service's `main.go`; model changes can affect existing SQLite data in `data/` or deployed `/var/lib/gpanel` data.
- Frontend routes and global UI setup are in `frontend/src/router` and `frontend/src/main.ts`; axios and WebSocket behavior is centralized in `frontend/src/utils`.
- After frontend changes, rebuild/copy `frontend/dist` into `core/routes/web/dist` before testing the embedded production UI or compiling core.

No `README.md`, Cursor rules, or GitHub Copilot instructions were present in the repository when this file was created. CI behavior described above is defined in `.github/workflows/release.yml` and `.github/workflows/prerelease.yml`.
