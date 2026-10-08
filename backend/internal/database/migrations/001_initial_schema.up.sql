-- +goose Up
-- +goose StatementBegin

-- Initial schema for Harbor collaborative development platform

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- USERS
-- ============================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    github_id       BIGINT UNIQUE NOT NULL,
    username        TEXT UNIQUE NOT NULL,
    name            TEXT,
    avatar_url      TEXT,
    encrypted_token TEXT, -- Encrypted GitHub access token
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_github_id ON users(github_id);
CREATE INDEX idx_users_username ON users(username);

-- ============================================
-- WORKSPACES
-- ============================================
CREATE TABLE workspaces (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    git_url         TEXT NOT NULL,
    git_branch      TEXT NOT NULL DEFAULT 'main',
    name            TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workspaces_owner_id ON workspaces(owner_id);

-- ============================================
-- ENVIRONMENTS
-- ============================================
CREATE TYPE environment_status AS ENUM (
    'CREATED',
    'BUILDING',
    'RUNNING',
    'STOPPED',
    'SUSPENDED',
    'CRASHED',
    'BUILD_FAILED'
);

CREATE TABLE environments (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id     UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    host_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status           environment_status NOT NULL DEFAULT 'CREATED',
    container_id     TEXT,
    port             INTEGER,
    public_url       TEXT,
    last_activity_at TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX idx_environments_workspace_id ON environments(workspace_id);
CREATE INDEX idx_environments_host_id ON environments(host_id);
CREATE INDEX idx_environments_status ON environments(status);

-- ============================================
-- ENVIRONMENT MEMBERS (Collaboration Core)
-- ============================================
CREATE TYPE member_role AS ENUM ('OWNER', 'COLLABORATOR', 'VIEWER');

CREATE TABLE environment_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    environment_id  UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            member_role NOT NULL DEFAULT 'VIEWER',
    invited_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (environment_id, user_id)
);

CREATE INDEX idx_environment_members_environment_id ON environment_members(environment_id);
CREATE INDEX idx_environment_members_user_id ON environment_members(user_id);

-- ============================================
-- SHARE LINKS
-- ============================================
CREATE TABLE share_links (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    environment_id  UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    created_by      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token           TEXT UNIQUE NOT NULL,
    role            member_role NOT NULL DEFAULT 'VIEWER',
    expires_at      TIMESTAMPTZ,
    max_uses        INTEGER,
    use_count       INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_share_links_environment_id ON share_links(environment_id);
CREATE INDEX idx_share_links_token ON share_links(token);

-- ============================================
-- ENVIRONMENT EVENTS (Audit Log)
-- ============================================
CREATE TABLE environment_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    environment_id  UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    event_type      TEXT NOT NULL,
    metadata        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_environment_events_environment_id ON environment_events(environment_id);
CREATE INDEX idx_environment_events_created_at ON environment_events(created_at);

-- ============================================
-- UPDATED_AT TRIGGER
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_workspaces_updated_at BEFORE UPDATE ON workspaces
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_environments_updated_at BEFORE UPDATE ON environments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd