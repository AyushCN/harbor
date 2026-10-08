# Architecture

## Overview

Harbor is a collaborative development platform that provisions isolated environments from GitHub repositories.  

It is designed as a **single-host system first**, with a clear separation between the control plane and the runtime. Collaboration features (ownership, roles, presence, and sharing) are first-class citizens in the architecture.

---

## High-Level System Diagram

```
┌─────────────────┐         ┌─────────────────┐
│     User A      │         │     User B      │
│   (Host)        │         │ (Collaborator)  │
└────────┬────────┘         └────────┬────────┘
         │                           │
         └───────────┬───────────────┘
                     ▼
         ┌───────────────────────┐
         │   Next.js Frontend    │
         │  (Dashboard + IDE)    │
         └───────────┬───────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │      Go API           │
         │  (Control Plane)      │
         │  - Auth & Roles       │
         │  - REST + WebSocket   │
         │  - Presence           │
         └───────────┬───────────┘
                     │
       ┌─────────────┼─────────────┐
       ▼             ▼             ▼
┌────────────┐ ┌────────────┐ ┌────────────┐
│ PostgreSQL │ │ Message Bus│ │   Redis    │
│ (State)    │ │ (NATS)     │ │ (Presence) │
└────────────┘ └─────┬──────┘ └────────────┘
                     │
                     ▼
              ┌────────────┐
              │   Worker   │
              │ (Go)       │
              └─────┬──────┘
                    │
                    ▼
              ┌────────────┐
              │   Docker   │
              │  Runtime   │
              └─────┬──────┘
                    │
                    ▼
         ┌─────────────────────┐
         │   Environment       │
         │  (Container)        │
         │   + Traefik URL     │
         └─────────────────────┘
```

---

## Core Components

| Component          | Responsibility                                      | Technology          |
|--------------------|-----------------------------------------------------|---------------------|
| Frontend           | Dashboard, IDE, presence UI, sharing UI             | Next.js 15 + React  |
| API Server         | Authentication, authorization, REST, WebSocket      | Go + Gin            |
| Worker             | Clone repos, detect runtime, manage containers      | Go                  |
| State Store        | Users, environments, roles, memberships             | PostgreSQL 16       |
| Job Queue          | Async environment lifecycle jobs                    | NATS JetStream      |
| Realtime Layer     | Presence, terminal, live updates                    | Redis Pub/Sub       |
| Container Runtime  | Isolated environments                               | Docker              |
| Router             | Public / private preview URLs                       | Traefik             |

---

## Collaboration Architecture

### Roles

| Role            | Permissions                                                                 |
|-----------------|-----------------------------------------------------------------------------|
| Owner / Host    | Full control: edit, delete, invite, change roles, start/stop environment   |
| Collaborator    | Edit files, use terminal, create branches, fork environment                |
| Viewer          | Read-only access to files and preview                                      |

### Presence System

- Every user who opens an environment establishes a WebSocket connection.
- Presence events (join / leave) are published via Redis Pub/Sub.
- The frontend shows a live list of users currently inside the environment.
- The Host is always highlighted.

### Sharing

- Direct invite by GitHub username
- Shareable links with:
  - Role (Collaborator or Viewer)
  - Expiry time
  - Maximum number of uses

---

## Environment Lifecycle

```
CREATED
   │
   ▼
BUILDING ──────► BUILD_FAILED
   │
   ▼
RUNNING ⇄ STOPPED
   │         │
   │         ▼
   │      SUSPENDED (idle)
   │
   ▼
CRASHED
```

**Key behaviors:**
- Host can start, stop, resume, and delete the environment.
- Collaborators can work while the environment is RUNNING.
- Idle environments are automatically suspended after a configurable timeout.
- Presence is only active while the environment is RUNNING.

---

## Data Flow

### Creating an Environment

1. User (Host) submits a GitHub repository URL.
2. API creates records in PostgreSQL (`workspaces` + `environments`).
3. API publishes a job to NATS (`environment.create`).
4. Worker consumes the job:
   - Clones the repository
   - Detects language and framework
   - Creates and starts a Docker container
   - Performs readiness check
5. Worker updates environment status to `RUNNING`.
6. Frontend receives the public preview URL and opens the IDE.

### Joining as Collaborator

1. User accepts an invite or opens a share link.
2. API validates the role and adds the user to `environment_members`.
3. User opens the environment → WebSocket connection is established.
4. Presence system broadcasts that the user has joined.
5. All connected clients update the live user list.

---

## Technology Decisions

| Decision                  | Choice              | Reason                                      |
|---------------------------|---------------------|---------------------------------------------|
| Language (Backend)        | Go                  | Performance, strong concurrency, single binary |
| Frontend                  | Next.js 15          | Modern React, App Router, good DX           |
| Database                  | PostgreSQL          | Relational model fits users, roles, environments |
| Job Queue                 | NATS JetStream      | Lightweight, reliable, good for single-host |
| Realtime                  | Redis Pub/Sub       | Simple and effective for presence           |
| Container Runtime         | Docker              | Mature, widely understood                   |
| Reverse Proxy             | Traefik             | Automatic routing and TLS support           |

---

## Design Principles

1. **Host is always visible** — Ownership must never be ambiguous.
2. **Collaboration is first-class** — Roles and presence are core, not add-ons.
3. **Single-host first** — Make it solid on one machine before scaling.
4. **Clear separation** — API handles control plane, Worker handles runtime.
5. **Security by default** — Containers run with dropped capabilities and resource limits.
6. **Progressive complexity** — Start simple, add multi-node and advanced isolation later.

---

## Future Evolution

The architecture is intentionally designed so that the following can be added later without major rewrites:

- Multi-host / multi-node support
- Stronger isolation (gVisor, Firecracker, or Kubernetes)
- Sidecar databases
- Background workers and cron jobs
- Custom domains
- Organization / team management