-- +goose Up
-- +goose StatementBegin

-- ============================================
-- BUILD LOGS (Real-time Terminal Streaming)
-- ============================================
CREATE TABLE build_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    environment_id  UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    sequence        INTEGER NOT NULL,          -- Log line sequence number
    message         TEXT NOT NULL,             -- Raw log message (with ANSI codes)
    level           TEXT NOT NULL DEFAULT 'info', -- info, warn, error, success
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_build_logs_environment_id ON build_logs(environment_id);
CREATE INDEX idx_build_logs_environment_id_sequence ON build_logs(environment_id, sequence);
CREATE INDEX idx_build_logs_created_at ON build_logs(created_at);

-- ============================================
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_build_logs_environment_id_sequence;
DROP INDEX IF EXISTS idx_build_logs_environment_id;
DROP INDEX IF EXISTS idx_build_logs_created_at;
DROP TABLE IF EXISTS build_logs;

-- +goose StatementEnd