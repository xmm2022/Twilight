# Version Notes

## Go Backend Branch

- Backend source of truth: `cmd/` and `internal/`.
- Frontend source of truth: `webui/`.
- Legacy backend source and dependency metadata have been removed from this branch.

## Current Unreleased

- Refactored the Go backend by feature boundary: Emby client/library/inventory, TMDB, Bangumi, media requests, invite/regcode/code use, scheduler handlers/runner, database admin, and system update now live in separate files.
- Added PostgreSQL-compatible database migration preview details: source/target driver, snapshot size, full entity counts, target readiness, and restart/config warnings.
- Hardened Git update and auto-update: HTTPS-only repo URLs, credential rejection, branch validation, dry-run preflight, dirty worktree refusal, redacted response metadata, and `git pull --ff-only`.
- Updated the admin config frontend for database migration preview details and Git update safety preflight.
- Expanded `.gitignore` for Go build/test artifacts, coverage/pprof outputs, SQLite WAL files, runtime data, frontend coverage, and deployment/test caches.
- Documented the Go backend split, database migration/restore workflow, hot reload behavior, and update security model.

## Release Checklist

- Update backend version in `cmd/twilight/main.go` when cutting a release.
- Update frontend version in `webui/package.json` and lockfiles when frontend dependencies change.
- Run `go test ./...`.
- Run frontend lint/build from `webui/` when UI or API client behavior changes.
