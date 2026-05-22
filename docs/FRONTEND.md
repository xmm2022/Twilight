# Frontend Development

The frontend is a Next.js app in `webui/`.

## Local Setup

```bash
cd webui
pnpm install --frozen-lockfile
pnpm dev
```

Run the Go backend separately:

```bash
bash start_backend_dev.sh
```

## API Contract

- Frontend API calls live in `webui/src/lib/api.ts`.
- Backend routes are registered in `internal/api/routes.go`.
- Keep response shapes stable and use the shared `{ success, code, message, data, timestamp }` envelope.

## Verification

- Run frontend lint/build after UI or API client changes.
- Run `go test ./...` after backend changes.
