#!/usr/bin/env bash
# ============================================================================
# Twilight Docker Entrypoint
# ============================================================================
# Initializes the container at startup:
#   - Waits for PostgreSQL to be ready
#   - Ensures required directories exist
#   - Generates a bot_internal_secret if not configured
#   - Launches the Twilight binary
# ============================================================================

set -euo pipefail

log() { echo "[entrypoint] $(date -u '+%Y-%m-%dT%H:%M:%SZ') $*" >&2; }

# ---- Wait for PostgreSQL ----
pg_host="${TWILIGHT_POSTGRES_HOST:-postgres}"
pg_port="${TWILIGHT_POSTGRES_PORT:-5432}"
pg_user="${TWILIGHT_POSTGRES_USER:-twilight}"
pg_db="${TWILIGHT_POSTGRES_DATABASE:-twilight}"

if command -v pg_isready &>/dev/null; then
    log "Waiting for PostgreSQL at ${pg_host}:${pg_port} ..."
    until pg_isready -h "$pg_host" -p "$pg_port" -U "$pg_user" -d "$pg_db" -t 3 >/dev/null 2>&1; do
        sleep 2
    done
    log "PostgreSQL is ready"
else
    log "pg_isready not available; sleeping 10s to let PostgreSQL start ..."
    sleep 10
fi

# ---- Ensure directories ----
for dir in uploads db/backups; do
    if [ ! -d "$dir" ]; then
        mkdir -p "$dir"
        log "Created directory: $dir"
    fi
done

# ---- Generate bot_internal_secret if empty ----
if [ -z "${TWILIGHT_BOT_INTERNAL_SECRET:-}" ]; then
    new_secret=$(openssl rand -base64 48 2>/dev/null || head -c 48 /dev/urandom | base64)
    export TWILIGHT_BOT_INTERNAL_SECRET="$new_secret"
    log "Generated bot_internal_secret (will not persist across restarts; set BOT_INTERNAL_SECRET env var for persistence)"
fi

# ---- Launch Twilight ----
log "Starting Twilight: $*"
exec ./twilight "$@"
