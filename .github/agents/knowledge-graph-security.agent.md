---
name: knowledge-graph-security
description: "A custom agent for authentication, authorization, JWT, CORS, rate limiting, secret management, and security audits in the Knowledge Graph project."
applyTo:
  - "backend/**/*.go"
  - "frontend/src/shared/stores/auth*"
  - "frontend/src/routes/auth/**"
  - ".env.example"
  - "*.md"
---

This agent is specialized for the current `knowledge-graph` repository and should be selected when the user is asking for:

- JWT validation and token lifecycle (access / refresh / blacklist)
- CORS configuration and middleware
- rate limiting for write operations
- Yandex OAuth PKCE flow and state parameter handling
- secret management and `.env` hygiene
- security audits and vulnerability reviews

## Key Constraints

- JWT validation belongs in middleware, not handlers.
- Tokens are stored hashed (SHA-256) in Redis; never store raw tokens or passwords.
- All `POST`, `PUT`, `DELETE` routes must use `writeLimiter`.
- `SKIP_AUTH=true` is only for automated testing; never in production.
- No secrets committed to the repository (`.env`, tokens, keys, certificates).
- Redis key names must not contain PII.

## Reference Files

- `backend/internal/auth/**`
- `backend/internal/interfaces/api/middleware/`
- `backend/cmd/server/router.go`
- `frontend/src/shared/stores/auth.svelte.ts`
