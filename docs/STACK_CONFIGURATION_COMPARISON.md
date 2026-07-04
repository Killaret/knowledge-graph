# Dev Stack vs Personal Stack Configuration Comparison

**Document purpose:** compare the two Docker Compose stacks and their proxy/routing configurations.

## Stacks

| Stack | Compose file | Purpose |
|-------|--------------|---------|
| Dev (main) | `docker-compose.yml` | Full development environment with all services and a public API gateway |
| Personal | `docker-compose.personal.yml` | Isolated single-user environment with backup scheduler and relative URLs |

## Service Port Mapping

| Service | Dev stack host port | Personal stack host port |
|---------|---------------------|--------------------------|
| nginx | `8080` (API), `8081` (frontend) | `8082` (API), `8083` (frontend) |
| backend | `9000` → container `8080` | `8085` → container `8080` |
| graph-service (HTTP) | `9091` → `9091` | `9092` → `9091` |
| graph-service (gRPC) | `9090` → `9090` | not exposed on host |
| postgres | `15432` → `5432` | `5433` → `5432` |
| redis | `6379` → `6379` | `6380` → `6379` |
| mongo | `27017` → `27017` | `27018` → `27017` |
| nlp | `5000` → `5000` | `5001` → `5000` |
| frontend | `5173` → `3000` | `3001` → `3000` |

## Proxy Architecture

### Dev Stack (`docker-compose.yml` + `nginx.conf`)

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Browser   │────▶│ nginx:8080  │────▶│ backend:8080│
└─────────────┘     └─────────────┘     └─────────────┘
                            │
                            ▼
                    ┌─────────────────┐
                    │ graph-service:  │
                    │ 9091            │
                    └─────────────────┘
```

- Client-side API calls go directly to `nginx:8080`.
- SSR `hooks.server.ts` proxies `/api/v1` to `backend:8080` and `/graph-service` to `graph-service:9091`.
- Nginx is an explicit gateway for both browser and frontend container requests.

### Personal Stack (`docker-compose.personal.yml` + `nginx.personal.conf`)

```
┌─────────────┐     ┌──────────────────┐     ┌──────────────────┐
│   Browser   │────▶│ frontend_personal│────▶│ backend_personal │
└─────────────┘     │   :3000 (SSR)     │     │    :8080          │
                    └──────────────────┘     └──────────────────┘
                              │
                              ▼
                    ┌─────────────────────┐
                    │ graph-service-      │
                    │ personal:9091       │
                    └─────────────────────┘
```

- Client-side API calls use relative URLs (`/api`, `/graph-service`) and hit the SvelteKit frontend server.
- SSR `hooks.server.ts` proxies to `backend_personal:8080` and `graph-service-personal:9091`.
- Nginx is available on `8082/8083` but the frontend can also be used standalone on `3001`.

## Frontend Configuration Differences

| Setting | Dev stack | Personal stack |
|---------|-----------|----------------|
| `VITE_API_URL` (browser) | `http://localhost:8080` | `/api` (relative) |
| `VITE_GRAPH_SERVICE_URL` (browser) | `http://localhost:8080/graph-service` | `/graph-service` (relative) |
| `VITE_API_TARGET` (SSR) | `http://backend:8080` | `http://backend_personal:8080` |
| `GRAPH_SERVICE_URL` (SSR) | `http://graph-service:9091` | `http://graph-service-personal:9091` |

## Nginx Configuration Differences

| Aspect | `nginx.conf` (dev) | `nginx.personal.conf` |
|--------|---------------------|------------------------|
| Backend upstream | `backend:8080` | `backend_personal:8080` |
| Graph service upstream | `graph-service:9091` | `graph-service-personal:9091` |
| Frontend upstream | `frontend:3000` (port 8081) | `frontend_personal:3000` (port 8081) |
| Port 8081 API proxy | not present | `/api/` and `/graph-service/` also proxied on 8081 |

The personal nginx exposes API routes on both `8080` and `8081`; the dev nginx only exposes them on `8080`.

## Vite Dev Proxy Defaults

File: `frontend/vite.config.ts`

| Proxy path | Default target (dev stack) | Override env var |
|------------|----------------------------|------------------|
| `/api/v1` | `http://127.0.0.1:9000` | `VITE_API_TARGET` or `VITE_API_URL` |
| `/graph-service/api` | `http://127.0.0.1:9091` | `VITE_GRAPH_SERVICE_URL` |

These defaults match the dev stack host ports. Override the environment variables when running against the personal stack.

## Bugs Fixed During This Review

1. `frontend/src/hooks.server.ts` default `VITE_API_TARGET` was `http://backend_personal:8080`, which broke the dev stack SSR proxy. Changed to `http://backend:8080`.
2. `frontend/Dockerfile` was missing `ARG VITE_API_TARGET`, so the build argument was silently ignored.
3. `docker-compose.yml` frontend build arg `VITE_API_TARGET` had trailing `/api`, causing double `/api` paths in SSR (`/api/api/v1/...`). Removed from build args; runtime env now sets the correct value.
4. `docker-compose.yml` and `docker-compose.personal.yml` frontend runtime environments did not set `VITE_API_TARGET`, relying on a fallback that was wrong for dev. Added explicit values per stack.
5. `docker-compose.yml` frontend `VITE_API_URL` and `VITE_GRAPH_SERVICE_URL` defaults used the Docker network hostname `http://nginx:8080`, which browsers cannot resolve. Changed to `http://localhost:8080` (host port).
6. `docker-compose.yml` `frontend` service now depends on `nginx` so the API gateway is available when the frontend starts.
7. `frontend/vite.config.ts` default proxy targets were `8085` (backend) and `9092` (graph service), which match the personal stack rather than the dev stack. Changed to `9000` and `9091`.

## Remaining Architectural Differences

- **Dev stack** is designed for full-service integration with nginx as the central gateway.
- **Personal stack** is designed to be self-contained: the frontend can proxy API requests through its own SSR layer without nginx.
- The personal stack includes extra services: `cli_personal` (profile `cli`) and `backup_scheduler`.
- The dev stack exposes the gRPC port of the graph service (`9090`); the personal stack does not.

## Recommendation

- Keep the two stacks separate but document which one a developer is running.
- Always set `VITE_API_TARGET` and `GRAPH_SERVICE_URL` explicitly when extending either compose file.
- For `npm run dev` against the personal stack, set `VITE_API_TARGET=http://127.0.0.1:8085` and `VITE_GRAPH_SERVICE_URL=http://127.0.0.1:9092` in a local `.env` file.
