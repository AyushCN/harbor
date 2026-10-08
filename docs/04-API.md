# API Reference

## Overview

The Harbor API is a RESTful API with WebSocket support for real-time features (presence, terminal, logs).

**Base URL (development):** `http://localhost:8080/api`

**Authentication:**  
All endpoints (except auth) require a valid JWT token sent via:

- `Authorization: Bearer <token>` header, **or**
- HttpOnly cookie (`harbor_token`)

---

## Authentication

### `GET /auth/github`
Redirects the user to GitHub OAuth.

### `GET /auth/github/callback`
GitHub OAuth callback. Sets the session cookie and redirects to the frontend.

### `POST /auth/logout`
Clears the session.

### `GET /auth/me`
Returns the currently authenticated user.

**Response**
```json
{
  "id": "uuid",
  "username": "ayushcn",
  "name": "Ayush C N",
  "avatar_url": "https://..."
}
```

---

## Environments

### `GET /environments`
List environments the current user has access to (as Host, Collaborator, or Viewer).

**Query parameters:**
- `role` (optional) — filter by role (`OWNER`, `COLLABORATOR`, `VIEWER`)

**Response**
```json
[
  {
    "id": "uuid",
    "name": "my-app",
    "status": "RUNNING",
    "public_url": "https://abc123.harbor.localhost",
    "host": {
      "id": "uuid",
      "username": "ayushcn",
      "avatar_url": "https://..."
    },
    "role": "OWNER",
    "created_at": "2026-10-07T10:00:00Z"
  }
]
```

### `POST /environments`
Create a new environment (current user becomes the Host).

**Request**
```json
{
  "git_url": "https://github.com/user/repo",
  "git_branch": "main",
  "name": "optional-name"
}
```

**Response**
```json
{
  "id": "uuid",
  "status": "CREATED",
  "host_id": "uuid"
}
```

### `GET /environments/:id`
Get detailed information about an environment.

**Response includes:**
- Host information
- Current user role
- List of members
- Current presence (online users)
- Public URL
- Status

### `POST /environments/:id/start`
Start a stopped or suspended environment.  
**Only Host or Collaborator can call this.**

### `POST /environments/:id/stop`
Stop a running environment.  
**Only Host can call this.**

### `POST /environments/:id/resume`
Resume a suspended environment.

### `DELETE /environments/:id`
Delete an environment (soft delete).  
**Only Host can call this.**

---

## Collaboration & Members

### `GET /environments/:id/members`
List all members of an environment.

**Response**
```json
[
  {
    "user": {
      "id": "uuid",
      "username": "ayushcn",
      "avatar_url": "https://..."
    },
    "role": "OWNER",
    "joined_at": "2026-10-07T10:00:00Z"
  }
]
```

### `POST /environments/:id/members`
Invite a user to the environment.

**Request**
```json
{
  "username": "collaborator_github_username",
  "role": "COLLABORATOR"
}
```

**Only Host can invite.**

### `PATCH /environments/:id/members/:user_id`
Change a member's role.

**Request**
```json
{
  "role": "VIEWER"
}
```

**Only Host can change roles.**

### `DELETE /environments/:id/members/:user_id`
Remove a member from the environment.  
**Only Host can remove members.**

---

## Share Links

### `POST /environments/:id/share-links`
Create a shareable link.

**Request**
```json
{
  "role": "COLLABORATOR",
  "expires_in_hours": 48,
  "max_uses": 10
}
```

**Response**
```json
{
  "token": "abc123xyz",
  "url": "https://harbor.app/share/abc123xyz",
  "role": "COLLABORATOR",
  "expires_at": "2026-10-09T10:00:00Z"
}
```

### `POST /share/:token/join`
Join an environment using a share link.

---

## Presence (Real-time)

Presence is handled via **WebSocket**.

### WebSocket Endpoint
```
WS /ws/environments/:id
```

**Events sent by server:**

```json
{
  "type": "presence.join",
  "user": {
    "id": "uuid",
    "username": "ayushcn",
    "avatar_url": "https://..."
  }
}
```

```json
{
  "type": "presence.leave",
  "user_id": "uuid"
}
```

```json
{
  "type": "presence.sync",
  "users": [ ... ]
}
```

---

## IDE Related Endpoints

### `GET /environments/:id/files`
List files in the environment workspace.

### `GET /environments/:id/files/*path`
Read a file.

### `PUT /environments/:id/files/*path`
Write / update a file.  
**Requires Collaborator or Owner role.**

### `GET /environments/:id/terminal`
WebSocket endpoint for interactive terminal.

### `GET /environments/:id/logs`
Stream container logs (WebSocket or SSE).

---

## Error Format

All errors follow this structure:

```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Only the Host can perform this action"
  }
}
```

Common error codes:
- `UNAUTHORIZED`
- `FORBIDDEN`
- `NOT_FOUND`
- `VALIDATION_ERROR`
- `CONFLICT`

---

## Permission Summary

| Action                    | Owner | Collaborator | Viewer |
|---------------------------|-------|--------------|--------|
| View environment          | ✅    | ✅           | ✅     |
| Edit files                | ✅    | ✅           | ❌     |
| Use terminal              | ✅    | ✅           | ❌     |
| Start / Resume            | ✅    | ✅           | ❌     |
| Stop / Delete             | ✅    | ❌           | ❌     |
| Invite members            | ✅    | ❌           | ❌     |
| Change roles              | ✅    | ❌           | ❌     |
| Create share links        | ✅    | ❌           | ❌     |