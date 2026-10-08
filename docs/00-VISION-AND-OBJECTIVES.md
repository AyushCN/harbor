# Vision & Objectives

## Vision

Harbor is a **collaborative development platform**.

It allows developers and teams to create isolated, on-demand environments from any GitHub repository with zero configuration. Every environment has a clear **Host**, supports role-based collaboration, and provides real-time presence so everyone can see who owns the environment and who is currently working inside it.

The platform combines the simplicity of tools like Render with the collaborative power of Codespaces and Gitpod — while remaining self-hostable and starting as a reliable single-host system.

**Core Promise:**  
Paste a GitHub repository → get a live environment with a public preview URL, browser-based IDE, and the ability to invite others to collaborate — all with clear ownership and presence.

---

## Objectives

### Primary Objectives (v1)

1. **Zero-config environments**  
   Users can create a fully working environment from a GitHub repository URL without writing Dockerfiles or configuration files.

2. **Clear ownership**  
   Every environment has a visible **Host** (owner). The Host is always displayed on the dashboard, environment page, and inside the IDE.

3. **Role-based collaboration**  
   Support three core roles:
   - **Owner / Host** — full control
   - **Collaborator** — can edit and work in the environment
   - **Viewer** — read-only access

4. **Easy sharing**  
   Environments can be shared via:
   - Direct invite (by GitHub username)
   - Shareable links (with role, expiry, and usage limits)

5. **Real-time presence**  
   Users can see who is currently online inside an environment in real time.

6. **Browser-based IDE**  
   Provide a collaborative IDE experience with:
   - File editor (Monaco)
   - Integrated terminal
   - Git panel
   - Live preview

7. **Live preview URL**  
   Every running environment gets a public (or private) preview URL.

8. **GitHub-native experience**  
   Full support for GitHub OAuth, cloning, committing, pushing, and branch management directly from the platform.

9. **Single-host reliability**  
   The platform must work solidly on a single Linux host before expanding to multi-node deployments.

---

### Secondary Objectives (Later Phases)

- Sidecar databases (PostgreSQL, Redis, MySQL)
- Background workers and cron jobs
- Custom domains
- Environment forking
- Resource quotas and limits
- Multi-host / multi-node support
- Team and organization management

---

### Non-Goals (for v1)

- Public multi-tenant cloud with billing
- Guaranteed sub-second cold starts
- Full Kubernetes complexity
- Production-grade multi-tenant isolation on day one