# API Contract — Reading and Trying the Knowledge Graph API

The API contract is a single hand-written OpenAPI file: [`backend/openAPI.yaml`](../backend/openAPI.yaml).
It is the source of truth — it is *not* generated from code annotations, and a
router-level test (`backend/cmd/server/router_contract_test.go`) fails the build
when a route and the spec drift apart in either direction.

This page is the path from a clean machine to a rendered contract and a Postman
collection. Every URL below was verified against a running stack.

## 1. Bring the stack up

Use the isolated test stack — it is self-contained, disposable, and seeds its own
database:

```powershell
.\scripts\testing\start-test.ps1
.\scripts\testing\seed-test-data.ps1   # optional: 100 seeded notes, links, embeddings
```

This starts PostgreSQL, Redis, MongoDB, the NLP service, the graph service, the
backend and a frontend under the `kg-test-*` names on their own ports. It does
not start your dev or personal data — those stacks stay untouched.

The development stack (`docker compose up -d`) works too, but it needs a filled
`.env` and its own database; for contract exploration the test stack is simpler.

## 2. Where to look

Open Swagger UI directly on the backend:

- Test stack: <http://127.0.0.1:18083/swagger/index.html>
- Dev stack: <http://127.0.0.1:9000/swagger/index.html> (direct) or
  <http://127.0.0.1:18080/swagger/index.html> (through the nginx gateway)

The port always comes from the compose file you started — look at the backend's
published port (`"127.0.0.1:18083:8080"` in `docker-compose.test.yml`), not at
this document, if the numbers ever drift.

## 3. The spec file itself

Two ways to the same document:

- From a running server: `GET http://127.0.0.1:18083/openapi.yaml` (200, no auth).
- From the repository before anything is running: `backend/openAPI.yaml`.

The specification is OpenAPI 3.0.3 — deliberately the 3.0 line, because the
embedded Swagger UI and most client generators do not read 3.1 yet.

## 4. Postman

`Import` → `Link` → `http://127.0.0.1:18083/openapi.yaml` (or import the file
from `backend/openAPI.yaml`). Postman generates a collection straight from the
OpenAPI document. **There is no committed Postman collection** — a second copy
of the same contract would drift; the spec is the collection.

## 5. Authentication for probing

Most routes require a token. To get one:

```bash
curl -X POST http://127.0.0.1:18083/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"login":"me","email":"me@example.com","password":"Secret123!"}'
```

The response carries `access_token`. Send it as `Authorization: Bearer <token>`
on every protected call. In Swagger UI, press **Authorize** and paste the token.

Two routes are open anonymously since the public-read release: `GET
/api/v1/notes/{id}` for public notes, and `GET /api/v1/notes/search` over public
notes only. A `200` without a token on these is intended, not a broken guard —
private notes answer `404` to callers who do not own them, anonymous or not.

## 6. What `POST /api/v1/notes` returns

A created note answers `201` with the standard envelope, and the contract spells
the shape out, not just the code:

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Note Title",
    "content": "...",
    "type": "star",
    "metadata": {},
    "created_at": "2026-09-08T12:00:00Z",
    "updated_at": "2026-09-08T12:00:00Z"
  },
  "message": "Resource created successfully"
}
```

The full field list is `components.schemas.Note` in the spec; errors follow
`components.schemas.ErrorResponse` (`code`, `message`, `details[]`) — see
[`API_ERRORS_EN.md`](API_ERRORS_EN.md) for the error catalogue.

## Notes for maintainers

- The Swagger UI bundle is embedded via `swaggo/gin-swagger` and reads
  `/openapi.yaml`; `/swagger/doc.json` is intentionally empty — the old generated
  `backend/docs` package was removed so the stale spec could not be mistaken for
  the real one.
- Adding or renaming a route without updating `openAPI.yaml` fails
  `go test ./cmd/server` — update the spec in the same commit.
