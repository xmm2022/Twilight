# Developer Workflow

This branch is the Go backend branch. Treat `cmd/` and `internal/` as the backend source of truth.

## Backend

- Run checks with `go test ./...`.
- Start locally with `bash start_backend_dev.sh`.
- Build production binary with `go build -o bin/twilight ./cmd/twilight`.
- Run production mode with `bash start_backend_prod.sh`.

## Frontend

- Install frontend dependencies in `webui/` with `pnpm install --frozen-lockfile`.
- Keep API calls centralized in `webui/src/lib/api.ts`.
- When adding backend endpoints, register them in `internal/api/routes.go` and add focused Go tests when behavior is security-sensitive or shared.

## Safety Baseline

- Prefer Redis-backed sessions and rate limits in production by setting `RedisURL`.
- Keep destructive admin actions behind explicit confirmation strings.
- Do not return secrets except one-time generated passwords or API keys, and only on creation/reset responses.
- Use `http.MaxBytesReader`, content-type checks, path cleaning, and strict response envelopes for upload and asset flows.

## Release Checks

- `go test ./...`
- Frontend lint/build from `webui/` when UI or API client behavior changes.
- Confirm `start_backend_prod.sh` and `deploy/*.service` point at `bin/twilight`, not any removed Python runtime.
