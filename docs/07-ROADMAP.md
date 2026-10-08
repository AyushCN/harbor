# Roadmap

## Overview

This roadmap outlines the planned evolution of Harbor from a solid single-host collaborative platform to a more powerful and secure multi-user development platform.

The focus is on delivering a reliable collaboration experience first, then progressively improving isolation, scalability, and features.

---

## Phase 1 – Foundation (Current Focus)

**Goal:** Make the core collaboration platform work reliably on a single host.

### Features
- [x] Project vision, architecture, and documentation
- [ ] GitHub OAuth authentication
- [ ] Create environments from GitHub repositories
- [ ] Basic runtime detection (Node.js, Python, Go)
- [ ] Start / Stop / Resume environments
- [ ] Role-based access (Owner, Collaborator, Viewer)
- [ ] Invite members and share links
- [ ] Real-time presence (who is online)
- [ ] Browser IDE (Monaco editor + Terminal)
- [ ] Live preview URL via Traefik
- [ ] Idle suspend

### Success Criteria
- A Host can create an environment and invite collaborators
- Collaborators can open the environment and see the Host + other online users
- Basic editing and terminal access works
- Environments can be stopped and resumed cleanly

---

## Phase 2 – Collaboration Experience

**Goal:** Make collaboration feel polished and useful for real teams.

### Features
- [ ] Better presence UI (avatars, status indicators)
- [ ] Activity feed / event log per environment
- [ ] Fork environment (Collaborator creates their own copy)
- [ ] Branch management from the IDE
- [ ] Commit & push to GitHub from the browser
- [ ] File locking or awareness (who is editing what)
- [ ] Improved share link controls (expiry, max uses, revoke)

### Success Criteria
- Two or more people can comfortably work in the same environment
- Host has clear visibility and control over members
- Sharing feels simple and secure

---

## Phase 3 – Developer Experience & Reliability

**Goal:** Improve day-to-day usability and system stability.

### Features
- [ ] Better runtime detection (more languages + frameworks)
- [ ] Dependency caching
- [ ] Streaming logs
- [ ] Resource usage metrics (CPU / Memory)
- [ ] Environment templates
- [ ] Sidecar databases (PostgreSQL, Redis)
- [ ] Environment variables & secrets management
- [ ] Improved error messages and recovery

### Success Criteria
- Cold start times are acceptable
- Logs and metrics are easy to access
- Users can run common full-stack apps without friction

---

## Phase 4 – Hardening & Isolation

**Goal:** Significantly improve security and isolation.

### Features
- [ ] Stronger container isolation (gVisor or Firecracker exploration)
- [ ] Per-environment Docker networks
- [ ] Stricter resource limits and quotas
- [ ] Rate limiting
- [ ] Audit logging
- [ ] Rootless Docker support
- [ ] Security review of Docker socket usage

### Success Criteria
- A compromised environment has limited ability to affect the host or other environments
- Clear security documentation and threat model

---

## Phase 5 – Multi-Host & Scale

**Goal:** Move beyond single-host limitations.

### Features
- [ ] Multi-node support
- [ ] Scheduler / orchestrator for placing environments
- [ ] Shared storage strategy
- [ ] Centralized control plane
- [ ] Horizontal scaling of Workers
- [ ] Basic observability (metrics, tracing, alerting)

### Success Criteria
- Environments can run across multiple machines
- Control plane remains stable under higher load

---

## Phase 6 – Platform Features (Future)

**Goal:** Evolve toward a more complete PaaS experience.

### Possible Features
- Organizations and teams
- Custom domains
- Background workers & cron jobs
- Persistent volumes
- Usage quotas and simple billing (optional)
- Public environment gallery / templates
- API tokens for automation

---

## Guiding Principles

1. **Collaboration first** — Ownership, roles, and presence must always remain clear.
2. **Single-host excellence before multi-host complexity**
3. **Security improves progressively** — never pretend strong isolation exists before it does
4. **Developer experience matters** — prefer simple and reliable over feature-rich but fragile
5. **Documentation stays up to date** with every major phase

---

## Current Priority Order

1. Finish Phase 1 (core collaboration loop)
2. Polish Phase 2 (real teamwork experience)
3. Improve reliability and DX (Phase 3)
4. Then invest heavily in isolation and security (Phase 4)