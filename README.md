# Harbor

**Collaborative development environments from GitHub repositories.**

Harbor lets teams create isolated, on-demand environments from any GitHub repository.  
Every environment has a clear **Host**, supports role-based collaboration, and shows real-time presence so everyone knows who owns the environment and who is currently working inside it.

> Paste a GitHub repo → get a live preview URL + browser IDE + collaboration tools.

---

## Vision

Harbor is a self-hostable collaborative development platform.  
It combines the simplicity of tools like Render with the teamwork features of Codespaces — while remaining simple enough to run on a single Oracle Cloud instance.

**Core principles:**
- Clear ownership (every environment has a visible Host)
- Real collaboration (roles + presence)
- Zero-config deployments from GitHub
- Single-host first, designed to grow

---

## Features

| Feature | Status | Notes |
|---------|--------|-------|
| **Create environments from GitHub repositories** | ✅ Done | Auto-detects language, branch selector |
| **Automatic language & framework detection** | ✅ Done | Node.js, Python, Go, Rust, Java, etc. |
| **Browser-based IDE (editor + terminal)** | ✅ Done | Monaco Editor + xterm.js |
| **Live preview URLs** | ✅ Done | Via Traefik reverse proxy |
| **Role-based access (Owner / Collaborator / Viewer)** | ✅ Done | Enforced in API + UI |
| **Real-time presence** | ✅ Done | Redis-backed, WebSocket |
| **Share links and direct invites** | ✅ Done | Role-based, expiry, max uses |
| **Idle suspend & resume** | 🚧 In Progress | NATS worker handles suspend/resume |
| **GitHub OAuth + Git operations** | ✅ Done | OAuth flow + token encryption |
| **Branch auto-fetch in Create modal** | ✅ Done | Fetches branches from GitHub API |
| **Build terminal with live logs** | ✅ Done | WebSocket streaming, ANSI colors |
| **Environment detail page** | ✅ Done | Build stages, actions, members |
| **Share links** | ✅ Done | Role-based, expiry, max uses |

---

## Architecture Overview

```
Browser (Next.js 14)
    ↓
Go API (Control Plane) — port 8082
    ↓
Worker + Docker (Environments) — NATS JetStream
```

- **Frontend**: Next.js 14 (App Router) + TypeScript + Tailwind
- **Backend**: Go (API + Worker) — Gin + NATS JetStream
- **Database**: PostgreSQL 15 (GORM-less, raw pgx)
- **Queue**: NATS JetStream (durable consumers, retry)
- **Realtime**: Redis (presence) + NATS (build logs)
- **Runtime**: Docker (multi-stage builds, security hardening)
- **Reverse Proxy**: Traefik (dynamic preview URLs)
- **Target Hosting**: Oracle Cloud (Always Free tier → scale)

---

## Documentation

| Document | Description |
|---------|-------------|
| [Vision & Objectives](docs/00-VISION-AND-OBJECTIVES.md) | Product vision and goals |
| [Architecture](docs/02-ARCHITECTURE.md) | System design |
| [Data Model](docs/03-DATA-MODEL.md) | Database schema |
| [API Reference](docs/04-API.md) | REST + WebSocket endpoints |
| [Development Guide](docs/05-DEVELOPMENT.md) | Local setup |
| [Security](docs/06-SECURITY.md) | Security model |
| [Roadmap](docs/07-ROADMAP.md) | Development phases |
| [Oracle Cloud](docs/08-ORACLE-CLOUD.md) | Hosting target |

---

## Quick Start (Development)

```bash
# 1. Clone the repository
git clone https://github.com/AyushCN/harbor.git
cd harbor

# 2. Copy environment file
cp .env.example .env

# 3. Start infrastructure (PostgreSQL, Redis, NATS, Traefik)
docker compose -f docker-compose.dev.yml up -d

# 4. Run API (port 8082)
cd backend && go run ./cmd/api

# 5. Run Worker (new terminal)
cd backend && MODE=worker go run ./cmd/worker

# 6. Run Frontend (new terminal, port 3000)
cd frontend && npm install && npm run dev
```

Open [http://localhost:3000](http://localhost:3000)

---

## Hosting Target

Harbor is designed to run well on **Oracle Cloud**, starting with the Always Free tier and scaling as needed.

---

## Project Status

Harbor is in active development.  
Current focus: **building a solid single-host collaborative foundation**.

### Recently Completed (October 2024)
- ✅ GitHub OAuth + token encryption (AES-256-GCM)
- ✅ Environment CRUD + status lifecycle
- ✅ Role-based access (Owner/Collaborator/Viewer)
- ✅ Build terminal with WebSocket streaming
- ✅ GitHub branch auto-fetch in Create modal
- ✅ Share link generation + join flow
- ✅ Build terminal WebSocket streaming
- ✅ Build logs persistence (DB + NATS)
- ✅ Environment detail page with build stages
- ✅ Dashboard with stats, tabs, filters, actions
- ✅ Real-time presence (Redis + WebSocket)
- ✅ Fix: logout clears all state, redirects to login
- ✅ Fix: CORS via Next.js proxy + backend middleware
- ✅ Fix: GitHub OAuth redirect URI mismatch
- ✅ Fix: Build terminal CORS (proxy through Next.js)
- ✅ Fix: GitHub branch auto-fetch in Create modal
- ✅ Fix: Dockerfile auto-detect entry point from package.json

### In Progress
- 🚧 Idle suspend / resume logic
- 🚧 Activity feed / environment events
- 🚧 Environment forking

---

## Quick Start (Development)

```bash
# 1. Clone the repository
git clone https://github.com/AyushCN/harbor.git
cd harbor

# 2. Copy environment file
cp .env.example .env

# 3. Start infrastructure
docker compose -f docker-compose.dev.yml up -d

# 4. Run API (port 8082)
cd backend && go run ./cmd/api

# 6. Run Worker (new terminal)
cd backend && MODE=worker go run ./cmd/worker

# 7. Run Frontend (new terminal, port 3000)
cd frontend && npm install && npm run dev
```

Open [http://localhost:3000](http://localhost:3000)

---

## Environment Variables

Key variables in `.env`:

| Variable | Description |
|----------|-------------|
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | GitHub OAuth app credentials |
| `JWT_SECRET` | 32+ char secret for JWT signing |
| `ENCRYPTION_KEY` | 32-byte hex for token encryption (AES-256-GCM) |
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Redis connection string |
| `NATS_URL` | NATS JetStream URL |
| `FRONTEND_URL` | Frontend origin (for CORS) |
| `API_URL` | Backend API URL |
| `TRAEFIK_DOMAIN` | Domain for preview URLs |
| `DOCKER_HOST` | Docker socket path |

---

## License

MIT