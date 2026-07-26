# Graph-Service Authentication & Authorization

This document describes how authentication and authorization work in the `services/graph-service` component.

## Overview

`graph-service` is a separate Go service that owns graph layout computation, caching, and analytics. It receives HTTP and gRPC requests either directly from the frontend (via the SvelteKit `/graph-service/api` proxy) or internally from the main backend. All non-public endpoints require authentication.

## Authentication Mechanisms

### 1. JWT Bearer Token (`Authorization: Bearer <token>`) and HttpOnly Cookie

- Uses the same `JWT_SECRET` as the main backend.
- Token must have `token_type: "access"`.
- `user_id` claim is extracted and used to scope database queries.
- The frontend `ky` client (`frontend/src/shared/api/client.ts`) adds the `Authorization` header to all API requests when an access token is available in memory.
- The backend also sets HttpOnly cookies (`access_token`, `refresh_token`) on login/refresh.
- The SvelteKit server proxy (`frontend/src/hooks.server.ts`) forwards both `Authorization` and `Cookie` headers to the backend and graph-service, so real-auth requests work through the same-origin `/api` and `/graph-service/api` proxies.

### 2. Internal Service Token (`X-Internal-Auth`)

- Shared secret configured via `GRAPH_SERVICE_INTERNAL_TOKEN`.
- Used for server-to-server calls from the main backend.
- If `X-User-Id` header is also present, the request is scoped to that user.
- If `X-User-Id` is missing, the request is treated as anonymous and only public notes are returned.

### 3. `SKIP_AUTH` Mode

- When `SKIP_AUTH=true` (development/test only), JWT validation is bypassed.
- The request context is marked as `skip-auth`, **not** as `public`.
- Visibility filtering is disabled: all notes are accessible, as this is a trusted local test mode.
- **Do not use `SKIP_AUTH=true` in production.**

## Authorization Rules

| Endpoint | Auth Required | Visibility Filter |
|----------|--------------|-------------------|
| `GET /api/v1/graph/public` | No | `is_public = true` |
| `GET /api/v1/graph/note/:id` | Yes | `creator_id = <user_id>` or public fallback |
| `GET /api/v1/graph/full` | Yes | `creator_id = <user_id>` or public fallback |
| `GET /api/v1/graph/delta` | Yes | `creator_id = <user_id>` or public fallback |
| `GET /api/v1/graph/path` | Yes | `creator_id = <user_id>` or public fallback |
| `GET /api/v1/graph/recommendations` | Yes | `creator_id = <user_id>` or public fallback |

## Security Fix: Anonymous/Default to Public

Previously, an unauthenticated request or `SKIP_AUTH` mode could result in an empty `NotesFilter`, causing `noteVisibilitySQL` to return `TRUE` and expose private notes. The logic now:

- Defaults anonymous requests to `IsPublic = true` when no valid user ID is present.
- Treats `SKIP_AUTH` as a separate trusted context that does **not** apply visibility restrictions, rather than treating it as a public-only request.

## HTTP Endpoints

- Direct: `http://127.0.0.1:19091/api/v1/graph/...`
- Via SvelteKit proxy (preferred in the browser): `http://127.0.0.1:3002/graph-service/api/v1/graph/...`
  - Proxy target is controlled by `VITE_API_TARGET` (backend) and `GRAPH_SERVICE_URL` (graph service) in `frontend/src/hooks.server.ts`.

## Known Cross-Service Security Notes

The following items were identified during a recent auth/token audit. They are **non-blocking** for real-auth manual testing but should be addressed before hardening the system.

1. **JWT token accepted from query parameter.**
   - `backend/internal/interfaces/api/middleware/jwt.go` falls back to `c.Query("token")`.
   - Tokens in URLs can leak to browser history and server access logs; prefer header/cookie only.

2. **Refresh token returned in JSON response.**
   - `backend` returns `refresh_token` in the login/refresh JSON body even though it is already set as an HttpOnly cookie.
   - This exposes the refresh token to JS memory; the frontend does not need it in the response body.

3. **`window.__ACCESS_TOKEN__` injection available in production builds.**
   - `frontend/src/shared/stores/auth.svelte.ts` uses `window.__ACCESS_TOKEN__` to support Playwright/Cucumber real-auth tests.
   - Any script running in the page context can set this value and force a login.

4. **Graph-service does not read the HttpOnly `access_token` cookie.**
   - `services/graph-service/internal/api/auth.go` only checks the `Authorization` header (and `X-Internal-Auth`).
   - After a full page reload, the frontend has a valid access cookie but graph-service returns `401`, triggering an unnecessary refresh.

5. **CORS defaults for the test stack are incomplete.**
   - `backend/cmd/server/middleware.go` does not include `http://127.0.0.1:3002` / `http://localhost:3002` in `defaultOrigins` and does not allow `X-API-Key` in `Access-Control-Allow-Headers`.
   - Currently mitigated by the same-origin `/api` proxy, but direct backend calls from the test frontend would be blocked.

## Files

- `services/graph-service/internal/api/auth.go` - middleware and token validation
- `services/graph-service/internal/api/auth_test.go` - unit tests
- `services/graph-service/internal/api/context.go` - request context helpers (`withUserID`, `withPublic`, `withSkipAuth`)
- `services/graph-service/internal/api/http_server.go` - HTTP handlers and filter construction
- `services/graph-service/internal/api/grpc_server.go` - gRPC handlers and filter construction
- `services/graph-service/internal/db/postgres_client.go` - visibility SQL helpers
