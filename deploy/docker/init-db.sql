-- ============================================================================
-- Twilight PostgreSQL Initialization
-- ============================================================================
-- This script runs automatically on first PostgreSQL container start.
-- It creates the required tables and indexes.
--
-- The Go backend also auto-creates these tables on startup, but having
-- this script ensures they exist before the backend connects.
-- ============================================================================

-- Extension for UUID generation (optional)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Main state table (single JSONB document)
CREATE TABLE IF NOT EXISTS twilight_state (
    id          BIGINT PRIMARY KEY DEFAULT 1,
    state       JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT  singleton CHECK (id = 1)
);

-- Sessions table (shared across processes)
CREATE TABLE IF NOT EXISTS twilight_sessions (
    token       TEXT PRIMARY KEY,
    uid         BIGINT NOT NULL,
    expires_at  BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS twilight_sessions_uid_idx ON twilight_sessions (uid);
CREATE INDEX IF NOT EXISTS twilight_sessions_expires_at_idx ON twilight_sessions (expires_at);

-- Runtime logs table (in-memory buffer persisted to DB)
CREATE TABLE IF NOT EXISTS twilight_runtime_logs (
    id          BIGSERIAL PRIMARY KEY,
    time        BIGINT NOT NULL,
    level       TEXT NOT NULL,
    message     TEXT NOT NULL,
    attrs       JSONB DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS twilight_runtime_logs_time_idx ON twilight_runtime_logs (time DESC);
CREATE INDEX IF NOT EXISTS twilight_runtime_logs_id_desc_idx ON twilight_runtime_logs (id DESC);

-- Insert initial empty state if not exists
INSERT INTO twilight_state (id, state) VALUES (1, '{}'::jsonb)
ON CONFLICT (id) DO NOTHING;
