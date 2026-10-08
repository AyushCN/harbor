-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_build_logs_environment_id_sequence;
DROP INDEX IF EXISTS idx_build_logs_environment_id;
DROP INDEX IF EXISTS idx_build_logs_created_at;
DROP TABLE IF EXISTS build_logs;

-- +goose StatementEnd