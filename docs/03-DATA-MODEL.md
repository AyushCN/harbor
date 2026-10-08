# Data Model

## Overview

The data model is designed around **collaboration**.  
Every environment has a clear **Host**, supports multiple members with roles, and tracks real-time presence.

The core entities are:

- Users
- Workspaces
- Environments
- Environment Members (roles)
- Share Links
- Presence (ephemeral)

---

## Entity Relationship Diagram (Simplified)

```
User
  │
  ├── owns ──────────────► Workspace
  │                           │
  │                           └── has ──► Environment
  │                                          │
  │                                          ├── has many ──► EnvironmentMember
  │                                          │                    │
  │                                          │                    └── role (OWNER, COLLABORATOR, VIEWER)
  │                                          │
  │                                          └── has many ──► ShareLink
  │
  └── can be member of ──► EnvironmentMember
```

---

## Core Tables

### 1. `users`

Stores authenticated users (via GitHub OAuth).

| Column              | Type         | Description                          |
|---------------------|--------------|--------------------------------------|
| id                  | UUID         | Primary key                          |
| github_id           | BIGINT       | GitHub user ID (unique)              |
| username            | TEXT         | GitHub username                      |
| name                | TEXT         | Display name                         |
| avatar_url          | TEXT         | Profile picture                      |
| encrypted_token     | TEXT         | Encrypted GitHub access token        |
| created_at          | TIMESTAMPTZ  |                                      |
| updated_at          | TIMESTAMPTZ  |                                      |

---

### 2. `workspaces`

Represents a cloned GitHub repository.

| Column              | Type         | Description                          |
|---------------------|--------------|--------------------------------------|
| id                  | UUID         | Primary key                          |
| owner_id            | UUID         | FK → users.id (the Host)             |
| git_url             | TEXT         | Full GitHub repository URL           |
| git_branch          | TEXT         | Default branch                       |
| name                | TEXT         | Display name                         |
| created_at          | TIMESTAMPTZ  |                                      |
| updated_at          | TIMESTAMPTZ  |                                      |

---

### 3. `environments`

A running (or stopped) instance of a workspace.

| Column              | Type         | Description                                      |
|---------------------|--------------|--------------------------------------------------|
| id                  | UUID         | Primary key                                      |
| workspace_id        | UUID         | FK → workspaces.id                               |
| host_id             | UUID         | FK → users.id (always the Owner/Host)            |
| status              | TEXT         | CREATED, BUILDING, RUNNING, STOPPED, SUSPENDED, CRASHED, BUILD_FAILED |
| container_id        | TEXT         | Docker container ID                              |
| port                | INTEGER      | Internal port the app is listening on            |
| public_url          | TEXT         | Preview URL (via Traefik)                        |
| last_activity_at    | TIMESTAMPTZ  | Used for idle suspend                            |
| created_at          | TIMESTAMPTZ  |                                                  |
| updated_at          | TIMESTAMPTZ  |                                                  |
| deleted_at          | TIMESTAMPTZ  | Soft delete                                      |

---

### 4. `environment_members`  (Collaboration Core)

Tracks who has access to an environment and with what role.

| Column              | Type         | Description                          |
|---------------------|--------------|--------------------------------------|
| id                  | UUID         | Primary key                          |
| environment_id      | UUID         | FK → environments.id                 |
| user_id             | UUID         | FK → users.id                        |
| role                | TEXT         | OWNER, COLLABORATOR, VIEWER          |
| invited_by          | UUID         | FK → users.id                        |
| created_at          | TIMESTAMPTZ  |                                      |

**Constraints:**
- Unique (`environment_id`, `user_id`)
- One user can only have one role per environment

---

### 5. `share_links`

Allows sharing an environment via a link.

| Column              | Type         | Description                          |
|---------------------|--------------|--------------------------------------|
| id                  | UUID         | Primary key                          |
| environment_id      | UUID         | FK → environments.id                 |
| created_by          | UUID         | FK → users.id                        |
| token               | TEXT         | Unique share token                   |
| role                | TEXT         | COLLABORATOR or VIEWER               |
| expires_at          | TIMESTAMPTZ  | Optional expiry                      |
| max_uses            | INTEGER      | Optional usage limit                 |
| use_count           | INTEGER      | Current number of uses               |
| created_at          | TIMESTAMPTZ  |                                      |

---

### 6. `environment_events` (Optional but recommended)

Audit log of important actions.

| Column              | Type         | Description                          |
|---------------------|--------------|--------------------------------------|
| id                  | UUID         | Primary key                          |
| environment_id      | UUID         | FK → environments.id                 |
| user_id             | UUID         | FK → users.id (nullable)             |
| event_type          | TEXT         | created, started, stopped, joined, left, role_changed, etc. |
| metadata            | JSONB        | Extra information                    |
| created_at          | TIMESTAMPTZ  |                                      |

---

## Presence (Ephemeral – Not stored in Postgres)

Presence is handled in **Redis**, not in the database.

**Redis Keys pattern:**

```
presence:environment:{environment_id} → Set of user_ids
```

- When a user opens an environment → `SADD`
- When a user leaves or disconnects → `SREM`
- Frontend receives real-time updates via WebSocket + Redis Pub/Sub

---

## Important Design Rules

1. **Host is immutable**  
   The `host_id` on an environment never changes. Ownership transfer is a future feature.

2. **Owner is also a member**  
   When an environment is created, the Host is automatically inserted into `environment_members` with role `OWNER`.

3. **Soft deletes**  
   Environments are soft-deleted (`deleted_at`) so history and audit logs remain intact.

4. **Role hierarchy**
   ```
   OWNER > COLLABORATOR > VIEWER
   ```

---

## Future Tables (Not required for v1)

- `organizations`
- `organization_members`
- `environment_databases` (sidecars)
- `environment_secrets`
- `audit_logs` (more detailed)