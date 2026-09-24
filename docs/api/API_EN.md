# API Contract — Reading and Trying the Knowledge Graph API

The API contract is a single hand-written OpenAPI file: [`backend/openAPI.yaml`](../../backend/openAPI.yaml).
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

## 2. Pre-built images and the running contract

If you use the deploy compose (`docker-compose.deploy.yml`) with pre-built Docker
Hub images, the image tag selects the contract version you run:

- `main` — the latest green build from the `main` branch; moves on every push.
- `YYYY-MM-DD-<short-sha>` — a frozen release from that commit.

`GET /openapi.yaml` from a running container must match `backend/openAPI.yaml`
from the same commit. If it does not, the image on Docker Hub is behind the
source-of-truth spec. Pin to a dated tag when you want the contract frozen.

## 3. Where to look

Open Swagger UI directly on the backend:

- Test stack: <http://127.0.0.1:18083/swagger/index.html>
- Dev stack: <http://127.0.0.1:9000/swagger/index.html> (direct) or
  <http://127.0.0.1:18080/swagger/index.html> (through the nginx gateway)

The port always comes from the compose file you started — look at the backend's
published port (`"127.0.0.1:18083:8080"` in `docker-compose.test.yml`), not at
this document, if the numbers ever drift.

## 4. The spec file itself

Two ways to the same document:

- From a running server: `GET http://127.0.0.1:18083/openapi.yaml` (200, no auth).
- From the repository before anything is running: `backend/openAPI.yaml`.

The specification is OpenAPI 3.0.3 — deliberately the 3.0 line, because the
embedded Swagger UI and most client generators do not read 3.1 yet.

## 5. Postman

`Import` → `Link` → `http://127.0.0.1:18083/openapi.yaml` (or import the file
from `backend/openAPI.yaml`). Postman generates a collection straight from the
OpenAPI document. **There is no committed Postman collection** — a second copy
of the same contract would drift; the spec is the collection.

## 6. Authentication for probing

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

## 7. What `POST /api/v1/notes` returns

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

## 8. Publishing and unpublishing a note

A note is created **private** by default. The only supported way to change
public visibility is through the dedicated publish endpoints:

```bash
# Make a note public (owner or admin only)
POST /api/v1/notes/{id}/publish

# Revert a note to private
POST /api/v1/notes/{id}/unpublish
```

Both answer with the standard `Note` envelope and set `is_public` accordingly.

**These endpoints are not simple field setters, and that is why they exist.**
Publishing and unpublishing also invalidate the cached public graph and notify the
graph-service subscriber, so a freshly published note appears in the community graph
without waiting for a cache expiry. Setting the column directly — in the database, or
through any future write path that bypasses these routes — would leave the public
graph serving a stale answer. If a new code path ever needs to change visibility, it
must go through the same handlers rather than the column.

Do **not** set `is_public` in `POST /api/v1/notes` or `PUT /api/v1/notes/{id}` —
`createNoteRequest` and `UpdateNoteRequest` do not include the field, and any
`is_public` or `source_url` sent on `PUT` is silently ignored. The public
read endpoints described in section 5 only return notes whose `is_public` is
`true`.

## 9. Graph view: personal and community

The graph is served by **graph-service**, not the main backend, and is reached through
the `/graph-service/api` proxy — `http://127.0.0.1:29091/api/v1/graph/...` when talking
to the test stack directly.

```bash
GET /api/v1/graph/full     # the caller's own graph; requires a token
GET /api/v1/graph/public   # the community graph; open to anonymous callers
```

Which one the frontend asks for is a **view mode**, not a consequence of being logged
in. An anonymous visitor always gets `public` and is shown no switcher; an authenticated
user defaults to `personal`, can switch to *Community*, and the choice survives a reload
(`localStorage`). Anonymous callers receive `401` from `full`.

Verified on the test stack: a user with five notes, two of them published, sees five
nodes in `full` and two in `public`, and an anonymous caller sees the same two. The
community graph is what is shared, not a trimmed copy of someone's own.

The main backend exposes `GET /api/v1/graph/public`, an anonymous route that returns
the public subset. It was renamed from the old `all` path (PUB-3) and has no alias.

## 10. Creating notes in batches

For clients that produce many notes at once — the Java source-text handler, bulk
import — three synchronous routes exist on the main backend:

```bash
POST /api/v1/notes/batch/create   # array of notes, max 50
POST /api/v1/notes/batch/delete   # array of ids
POST /api/v1/import/batch         # notes *and* links between them, max 50 each
```

Two asymmetries are deliberate. `import/batch` accepts links but has **no delete** —
an external source should not be able to erase. The user-facing `batch/create` accepts
no links.

Linking notes that do not exist yet is the open question: the client currently supplies
its own UUIDs in `import/batch`, and a foreign id that already exists is rejected rather
than overwritten. The contract for that is still being decided (see
`docs/tasks/BATCH-1-api-design.md`), so treat it as provisional.

Batch creation runs the same post-processing as single creation — keywords, embeddings,
link weights and recommendations. Mass **import** does not yet enqueue recommendations;
that gap is tracked as IMP-4.

## Notes for maintainers

- The Swagger UI bundle is embedded via `swaggo/gin-swagger` and reads
  `/openapi.yaml`; `/swagger/doc.json` is intentionally empty — the old generated
  `backend/docs` package was removed so the stale spec could not be mistaken for
  the real one.
- Adding or renaming a route without updating `openAPI.yaml` fails
  `go test ./cmd/server` — update the spec in the same commit.
