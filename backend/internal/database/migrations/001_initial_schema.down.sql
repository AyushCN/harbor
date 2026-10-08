-- 001_initial_schema.down.sql
-- Rollback initial schema

DROP TRIGGER IF EXISTS update_environments_updated_at ON environments;
DROP TRIGGER IF EXISTS update_workspaces_updated_at ON workspaces;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS environment_events;
DROP TABLE IF EXISTS share_links;
DROP TABLE IF EXISTS environment_members;
DROP TABLE IF EXISTS environments;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS environment_status;
DROP TYPE IF EXISTS member_role;

DROP EXTENSION IF EXISTS "pgcrypto";