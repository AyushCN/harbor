# Oracle Cloud Deployment Notes

## Overview

Harbor is designed to run well on **Oracle Cloud Infrastructure (OCI)**, starting with the Always Free tier and scaling later as needed.

This document explains how we plan to host Harbor on Oracle Cloud.

---

## Why Oracle Cloud?

- Generous **Always Free** tier (good for early development)
- Competitive pricing when moving to paid resources
- Good free egress allowance
- Suitable for running Docker-based applications

---

## Recommended Setup (Phase 1)

### Target Architecture on Oracle Cloud

```
Oracle Cloud VM (Ubuntu)
├── Docker + Docker Compose
├── Harbor Control Plane
│   ├── API
│   ├── Worker
│   ├── PostgreSQL
│   ├── Redis
│   └── NATS
├── Traefik (reverse proxy)
└── Environment Containers (created dynamically)
```

We will run everything on **a single VM** in the beginning.

---

## Free Tier Strategy

### Recommended Free Resources

Use the **Always Free Ampere A1** flex shape:

- Up to **2–4 OCPUs**
- Up to **12–24 GB RAM** (depending on current free tier limits)
- Enough to run:
  - Control plane (API + Worker)
  - PostgreSQL, Redis, NATS
  - A few user environments

> Note: Oracle has changed free tier limits over time. Always check the latest Always Free documentation.

### Suggested VM Shape

- **Shape**: VM.Standard.A1.Flex (ARM)
- **OCPUs**: 2
- **Memory**: 12 GB
- **Boot Volume**: 100–200 GB
- **OS**: Ubuntu 22.04 or 24.04

---

## Networking

### Required Ports

| Port | Service              | Notes                     |
|------|----------------------|---------------------------|
| 22   | SSH                  | Restricted to your IP     |
| 80   | HTTP (Traefik)       | Public                    |
| 443  | HTTPS (Traefik)      | Public (later)            |
| 8080 | API (optional)       | Prefer reverse proxy only |

### Recommendations

- Use a **Public Subnet**
- Assign a **Public IP**
- Restrict SSH (port 22) to your own IP address
- Later add HTTPS with Let's Encrypt via Traefik

---

## Storage

- Use the boot volume for the operating system and Docker data
- Store environment workspaces on the local disk:
  ```bash
  /var/lib/harbor/workspaces
  ```
- Later we can move workspaces to block volume if needed

---

## Installation Steps (High Level)

1. Create an Oracle Cloud account
2. Create a Compartment (optional but recommended)
3. Create a Virtual Cloud Network (VCN)
4. Create a Compute Instance (Ampere A1 Flex)
5. SSH into the instance
6. Install Docker + Docker Compose
7. Clone the Harbor repository
8. Configure `.env`
9. Start services with Docker Compose
10. Point a domain (optional) to the public IP

---

## Docker on Oracle Cloud (ARM Notes)

Because we are using **Ampere ARM** instances:

- Prefer multi-arch Docker images when possible
- Test that base images support `linux/arm64`
- Common images that work well:
  - `postgres:16-alpine`
  - `redis:7-alpine`
  - `nats:2.10-alpine`
  - `traefik:v3`
  - Official Golang and Node images (ARM variants)

If some images only support `amd64`, we may need to switch those services to AMD shapes later.

---

## Deployment Modes

### Development / Early Stage
- Single VM
- Docker Compose
- HTTP only
- Free tier resources

### Later Stage
- Still single VM (bigger shape) **or**
- Separate VM for databases
- HTTPS enabled
- Domain connected
- Basic monitoring

---

## Security Recommendations on OCI

- Disable password authentication for SSH (use keys only)
- Restrict port 22 to your IP
- Keep the Docker socket protected
- Do not expose PostgreSQL, Redis, or NATS publicly
- Run Traefik as the only public entry point
- Enable Oracle Cloud Firewall rules carefully

---

## Cost Awareness

| Stage              | Expected Cost      | Notes |
|--------------------|--------------------|-------|
| Early development  | $0                 | Always Free tier |
| Small usage        | Very low           | Still can stay free or cheap |
| Growing usage      | Paid VM            | Upgrade shape when needed |

Always set up **Budget Alerts** in Oracle Cloud to avoid surprise bills.

---

## Future Improvements

- Move to Oracle Kubernetes Engine (OKE) later if needed
- Use Oracle Object Storage for backups
- Add block volumes for persistent workspace storage
- Explore ARM vs AMD performance trade-offs

---

## Summary

For Harbor:

1. Start on **Oracle Cloud Always Free (Ampere A1)**
2. Run everything with **Docker Compose** on one VM
3. Keep architecture simple
4. Scale vertically first, then consider multi-node later