# Development

## Repository Layout

- `cmd/twilight`: Go backend entrypoint.
- `internal/api`: HTTP routing, auth, handlers, rate limits, sessions, API response helpers, and split business adapters.
- `internal/api/*_client.go`: remote service clients such as Emby, TMDB, Bangumi, and generic JSON HTTP helpers.
- `internal/api/*_handlers.go`: route handlers grouped by feature, including media requests, invites, regcodes, scheduler, system update, and database admin.
- `internal/store`: JSON and PostgreSQL-backed state store used by the Go backend.
- `internal/config`: TOML/env configuration loading.
- `internal/security`: password hashing and secure random helpers.
- `webui`: Next.js frontend.

## Backend Commands

```bash
go test ./...
go run ./cmd/twilight api --host 0.0.0.0 --port 5000 --config config.toml --debug
go build -o bin/twilight ./cmd/twilight
```

## Frontend Commands

```bash
cd webui
pnpm install --frozen-lockfile
pnpm lint
pnpm build
```

## API Development Rules

- Add routes in `internal/api/routes.go`.
- Keep handlers small enough to test, and move reusable behavior into feature files instead of adding unrelated logic to `handlers.go`.
- Put external service logic behind client/service helpers; handlers should validate input, authorize, call the helper, and shape the response.
- For public or auth-sensitive endpoints, add rate limits and tests.
- For admin destructive actions, require an explicit confirmation token and return structured `skipped`/`failed` details.
- Keep response envelopes compatible with `webui/src/lib/api.ts`.

## Data Model

The Go backend stores state in `db/twilight_go_state.json` by default and can also use PostgreSQL. Use `/api/v1/system/admin/database/migrate` with `dry_run=true` before changing storage backends; the preview reports entity counts, snapshot size, and target readiness. Production migrations from older deployments should be explicit one-time import tooling, not implicit startup behavior.

## Update Workflow

The admin Git update endpoint supports `dry_run` preflight and refuses dirty worktrees by default. It must remain argument-based (`exec.Command` with argv) and must not introduce shell command strings.
