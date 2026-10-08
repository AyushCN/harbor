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

## Features (Planned / In Progress)

- Create environments from GitHub repositories
- Automatic language & framework detection
- Browser-based IDE (editor + terminal)
- Live preview URLs
- Role-based access (Owner / Collaborator / Viewer)
- Real-time presence
- Share links and direct invites
- Idle suspend & resume
- GitHub OAuth + Git operations

---

## Architecture Overview

```
Browser (Next.js)
    ↓
Go API (Control Plane)
    ↓
Worker + Docker (Environments)
```

- **Frontend**: Next.js
- **Backend**: Go (API + Worker)
- **Database**: PostgreSQL
- **Queue**: NATS
- **Realtime**: Redis
- **Runtime**: Docker
- **Target Hosting**: Oracle Cloud

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
git clone https://github.com/yourusername/harbor.git
cd harbor

# 2. Copy environment file
cp .env.example .env

# 3. Start infrastructure
docker compose -f docker-compose.dev.yml up -d

# 4. Run API
cd backend && go run ./cmd/api

# 5. Run Worker (new terminal)
cd backend && MODE=worker go run ./cmd/worker

# 6. Run Frontend (new terminal)
cd frontend && npm install && npm run dev
```

Open [http://localhost:3000](http://localhost:3000)

---

## Hosting Target

Harbor is designed to run well on **Oracle Cloud**, starting with the Always Free tier and scaling as needed.

---

## Project Status

Harbor is in early development.  
The current focus is building a solid single-host collaborative foundation.

---

## License

MIT