# Graph-Service Authentication & Authorization

This document describes how authentication and authorization work in the `services/graph-service` component.

## Overview

`graph-service` is a separate Go service that owns graph layout computation, caching, and analytics. It receives HTTP and gRPC requests either directly from the frontend (via the SvelteKit proxy) or internally from the main backend. All non-public endpoints require authentication.

## Authentication Mechanisms

### 1. JWT Bearer Token (`Authorization: Bearer <token>`)

- Uses the same `JWT_SECRET` as the main backend.
- Token must have `token_type: "access"`.
- `user_id` claim is extracted and used to scope database queries.

### 2. Internal Service Token (`X-Internal-Auth`)

- Shared secret configured via `GRAPH_SERVICE_INTERNAL_TOKEN`.
- Used for server-to-server calls from the main backend.
- If `X-User-Id` header is also present, the request is scoped to that user.
- If `X-User-Id` is missing, the request is treated as anonymous/public-only.

### 3. `SKIP_AUTH` Mode

- When `SKIP_AUTH=true` (development/test only), requests are treated as **public**.
- This prevents leaking private notes when authentication is disabled.

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

Previously, an unauthenticated request or `SKIP_AUTH` mode could result in an empty `NotesFilter`, causing `noteVisibilitySQL` to return `TRUE` and expose private notes. The logic now defaults to `IsPublic = true` when no valid user ID is present.

## Files

- `services/graph-service/internal/api/auth.go` - middleware and token validation
- `services/graph-service/internal/api/auth_test.go` - unit tests
- `services/graph-service/internal/api/http_server.go` - HTTP handlers and filter construction
- `services/graph-service/internal/api/grpc_server.go` - gRPC handlers and filter construction
- `services/graph-service/internal/db/postgres_client.go` - visibility SQL helpers
