# Collaboration Model

## Roles

| Role          | Permissions                                      |
|---------------|--------------------------------------------------|
| **Owner / Host** | Full control: edit, delete, invite, change roles, stop/start |
| **Collaborator** | Edit files, use terminal, push to their own branch, fork |
| **Viewer**       | Read-only access to files and preview            |

## Visibility Rules

- Host is always visible on:
  - Environment list (dashboard)
  - Environment detail page
  - Inside the IDE header
- Active users (presence) are shown in real time inside the IDE
- Share links can be restricted by role and expiry

## Sharing Methods

1. **Invite by GitHub username**
2. **Generate shareable link** (with role + expiry + max uses)
3. **Fork** — Collaborator creates their own independent copy

## Presence System

- WebSocket connection per environment
- Redis Pub/Sub for broadcasting join/leave events
- Frontend shows live list of users currently inside the environment