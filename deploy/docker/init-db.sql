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
    data        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_expires ON twilight_sessions (expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_uid ON twilight_sessions ((data->>'uid'));

-- Runtime logs table (in-memory buffer persisted to DB)
CREATE TABLE IF NOT EXISTS twilight_runtime_logs (
    id          BIGSERIAL PRIMARY KEY,
    level       TEXT NOT NULL,
    message     TEXT NOT NULL,
    fields      JSONB DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_runtime_logs_created ON twilight_runtime_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_runtime_logs_level ON twilight_runtime_logs (level);

-- Insert initial empty state if not exists
INSERT INTO twilight_state (id, state) VALUES (1, '{}'::jsonb)
ON CONFLICT (id) DO NOTHING;
