# Security

## Overview

Harbor is designed as a **collaborative development platform** that runs user code inside isolated containers.  

Because the platform executes arbitrary code from GitHub repositories and allows multiple users to collaborate, security is a critical concern.

This document describes the current security model, trust boundaries, and hardening measures for v1 (single-host).

---

## Trust Boundaries

| Boundary                    | What it protects against                  | Current Protection Level      |
|----------------------------|-------------------------------------------|-------------------------------|
| Authentication             | Unauthorized access                       | Strong (GitHub OAuth + JWT)  |
| Authorization (Roles)      | Privilege escalation between users        | Strong                       |
| Container Isolation        | Malicious code affecting other environments | Best-effort                  |
| Host Isolation             | Compromised container escaping to host    | Limited                      |
| Network Isolation          | Lateral movement between environments     | Basic                        |

> **Important**: In v1, the control plane mounts the Docker socket.  
> This means a compromise of the API or Worker process is equivalent to root access on the host.

---

## Authentication

- Users authenticate exclusively via **GitHub OAuth 2.0** (with PKCE recommended).
- After successful login, the platform issues a **JWT**.
- The JWT is stored in an **HttpOnly**, **Secure**, **SameSite=Lax** cookie.
- GitHub access tokens are encrypted at rest using AES-256-GCM before being stored in the database.

### Token Handling Rules

- Never log access tokens
- Never expose tokens to the frontend
- Encrypt tokens before writing to the database
- Support token revocation on logout

---

## Authorization & Roles

Harbor uses a clear role-based access control (RBAC) model:

| Role            | Privileges                                                                 |
|-----------------|----------------------------------------------------------------------------|
| **Owner / Host**    | Full control over the environment (start, stop, delete, invite, change roles) |
| **Collaborator**    | Can edit files, use terminal, and work inside the environment                 |
| **Viewer**          | Read-only access to files and preview                                         |

### Enforcement Rules

- Every API endpoint that mutates state must check the caller's role.
- The Host of an environment is immutable in v1.
- Collaborators cannot promote themselves or invite new members.
- Viewers cannot perform any write operations.

---

## Container Security

Each environment runs in its own Docker container with the following hardening:

```bash
--read-only
--cap-drop ALL
--security-opt no-new-privileges
--user 1000:1000
--memory   # limit
--cpus     # limit
--pids-limit
```

### Additional Measures

- Temporary filesystems (`tmpfs`) for `/tmp`, `/var/tmp`, and `/run`
- No privileged containers
- Resource limits enforced on every environment
- Containers are connected to isolated bridge networks when possible

---

## Network Security

- Each environment should ideally run on its own Docker bridge network.
- Direct communication between environments is blocked by default.
- Only the Traefik reverse proxy can route traffic to environment containers.
- The control plane (API + Worker) should not be exposed publicly in production without additional protection (VPN, IP allowlist, or reverse proxy authentication).

---

## Secrets Management

| Secret                    | Storage Method                     |
|---------------------------|------------------------------------|
| GitHub OAuth tokens       | AES-256-GCM encrypted in database  |
| JWT signing key           | Environment variable               |
| Encryption key            | Environment variable               |
| Database credentials      | Environment variable / secrets manager |

**Rules:**
- Never commit secrets to Git
- Use different keys for development and production
- Rotate encryption keys carefully (requires re-encryption of existing tokens)

---

## WebSocket Security

- WebSocket connections require a valid JWT.
- Presence and terminal sessions are scoped to a specific environment.
- Users can only join WebSocket rooms for environments they are members of.
- Origin checks should be enforced.

---

## Input Validation & Common Attacks

| Attack Type               | Mitigation                                      |
|---------------------------|-------------------------------------------------|
| Path Traversal            | Strict path validation against workspace root   |
| Command Injection         | Avoid shell execution; use direct exec          |
| SSRF                      | Validate Git URLs; block internal IPs           |
| XSS                       | Proper output encoding in the frontend          |
| CSRF                      | SameSite cookies + Origin checks                |

---

## Logging & Monitoring

- Log authentication events (login, logout, failed attempts)
- Log role changes and member invitations
- Log environment lifecycle events (create, start, stop, delete)
- Never log secrets or full access tokens
- Monitor for unusual activity (rapid environment creation, excessive invites)

---

## Known Limitations (v1)

1. **Docker socket access**  
   The Worker and API have access to the Docker socket. A compromise of either component gives full control over the host.

2. **Shared kernel**  
   Containers share the host kernel. Kernel exploits can lead to container escape.

3. **Single-host**  
   There is no strong multi-tenant isolation across hosts yet.

4. **No gVisor / Firecracker**  
   Stronger sandboxing technologies are planned for later phases.

---

## Security Roadmap

### Short term
- Enforce strict role checks on every endpoint
- Add rate limiting
- Improve network isolation between environments
- Add audit logging

### Medium term
- Move to rootless Docker or user namespaces
- Introduce gVisor or Firecracker for stronger isolation
- Support secrets injection into environments

### Long term
- Multi-node architecture with proper tenant isolation
- Network policies (Cilium / Calico)
- Formal security audit

---

## Reporting Security Issues

If you discover a security vulnerability, please report it privately instead of opening a public issue.

Recommended contact method: Create a private security advisory on GitHub or contact the maintainers directly.