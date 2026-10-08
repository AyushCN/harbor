# Development Guide

## Prerequisites

Before you start, make sure you have the following installed:

| Tool              | Version     | Notes                          |
|-------------------|-------------|--------------------------------|
| Go                | 1.22+       | Backend language               |
| Node.js           | 20+         | Frontend                       |
| Docker            | 24+         | Container runtime              |
| Docker Compose    | v2          | Local infrastructure           |
| Git               | Latest      |                                |

Optional but recommended:
- `golangci-lint`
- `sqlc` (if using SQLC)
- `air` or `gow` for live reload

---

## Project Structure

```
harbor/
├── backend/
│   ├── cmd/
│   │   ├── api/          # API server entrypoint
│   │   └── worker/       # Background worker entrypoint
│   ├── internal/
│   │   ├── auth/
│   │   ├── environment/
│   │   ├── collaboration/
│   │   ├── presence/
│   │   ├── docker/
│   │   └── ...
│   ├── migrations/
│   └── go.mod
├── frontend/
│   ├── app/              # Next.js App Router
│   ├── components/
│   ├── lib/
│   └── package.json
├── docs/
├── docker-compose.dev.yml
├── .env.example
└── README.md
```

---

## 1. Clone & Setup

```bash
git clone https://github.com/yourusername/harbor.git
cd harbor
```

Copy the example environment file:

```bash
cp .env.example .env
```

---

## 2. Environment Variables

Edit `.env` and fill in the required values:

```env
# Database
DATABASE_URL=postgres://harbor:harbor@localhost:5432/harbor?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# NATS
NATS_URL=nats://localhost:4222

# Auth
JWT_SECRET=super-secret-change-me-in-production-at-least-32-chars
ENCRYPTION_KEY=your-32-byte-hex-key-here

# GitHub OAuth
GITHUB_CLIENT_ID=your_github_oauth_client_id
GITHUB_CLIENT_SECRET=your_github_oauth_client_secret

# Frontend
FRONTEND_URL=http://localhost:3000

# Docker
DOCKER_HOST=unix:///var/run/docker.sock
WORKSPACE_ROOT=/tmp/harbor-workspaces

# Environment
ENV=development
```

> Generate a secure `ENCRYPTION_KEY`:
> ```bash
> openssl rand -hex 16
> ```

---

## 3. Start Infrastructure

Start PostgreSQL, Redis, NATS, and Traefik:

```bash
docker compose -f docker-compose.dev.yml up -d
```

Verify services are healthy:

```bash
docker compose -f docker-compose.dev.yml ps
```

---

## 4. Run Backend

### Terminal 1 – API Server

```bash
cd backend
go run ./cmd/api
```

The API will start on `http://localhost:8080`.

### Terminal 2 – Worker

```bash
cd backend
MODE=worker go run ./cmd/worker
```

---

## 5. Run Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend will be available at:  
**http://localhost:3000**

---

## 6. GitHub OAuth Setup

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Create a new OAuth App
3. Set the following:
   - **Homepage URL**: `http://localhost:3000`
   - **Authorization callback URL**: `http://localhost:8080/api/auth/github/callback`
4. Copy the Client ID and Client Secret into your `.env` file

---

## Development Workflow

### Recommended Daily Flow

1. Start infrastructure:
   ```bash
   docker compose -f docker-compose.dev.yml up -d
   ```

2. Start API + Worker (two terminals)

3. Start Frontend

4. Open http://localhost:3000 and log in with GitHub

### Useful Commands

```bash
# Run backend tests
cd backend && go test ./...

# Format Go code
cd backend && go fmt ./...

# Frontend lint + build
cd frontend && npm run lint && npm run build

# View logs of infrastructure
docker compose -f docker-compose.dev.yml logs -f
```

---

## Database Migrations

Migrations are applied automatically when the API or Worker starts.

To run them manually (if needed):

```bash
cd backend
go run ./cmd/migrate up
```

---

## Common Issues

### Docker Permission Denied
```bash
sudo usermod -aG docker $USER
# Then log out and log back in
```

### Port already in use
Change the ports in `docker-compose.dev.yml` or stop the conflicting service.

### GitHub OAuth not working
- Double-check the callback URL
- Make sure `FRONTEND_URL` and `GITHUB_CLIENT_*` are set correctly

---

## Next Steps After Setup

Once everything is running:

1. Create your first environment from a public GitHub repository
2. Invite another GitHub user as Collaborator
3. Open the environment in two browsers and test presence
4. Try the browser IDE (editor + terminal)

---

## Development Guidelines

- Keep the **Host** concept visible in every UI related to environments
- Always check roles before allowing actions (Owner vs Collaborator vs Viewer)
- Prefer clear error messages over generic ones
- Write tests for collaboration logic (invites, role changes, permissions)