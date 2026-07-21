---
name: knowledge-graph-infrastructure
description: "A custom agent for Docker, docker-compose, nginx, volumes, health checks, and local infrastructure configuration in the Knowledge Graph project."
applyTo:
  - "docker-compose*.yml"
  - "Dockerfile*"
  - "nginx*.conf"
  - "backend/Dockerfile"
  - "frontend/Dockerfile"
  - "scripts/*"
  - "*.md"
---

This agent is specialized for the current `knowledge-graph` repository and should be selected when the user is asking for:

- Docker multi-stage builds and image optimization
- docker-compose stack configuration (dev, personal, test)
- nginx reverse proxy and routing
- service health checks and `depends_on` conditions
- volume naming and persistence
- local infrastructure scripts and environment setup

## Key Constraints

- All services must expose a `/health` endpoint.
- NLP service healthcheck needs `start_period: 600s` because of lazy model loading.
- Use Docker embedded DNS with variables (`set $backend_host backend:8080`) in nginx; never use static `upstream` blocks.
- `VITE_*` variables are baked at build time; pass them as `ARG`/`ENV` during `docker compose build`.
- Secrets live in `.env` (gitignored); never hard-code them in compose files.

## Reference Files

- `docker-compose.yml`, `docker-compose.personal.yml`, `docker-compose.test.yml`
- `nginx.conf`, `nginx.personal.conf`
- `backend/Dockerfile`, `frontend/Dockerfile`, `nlp-service/Dockerfile`
