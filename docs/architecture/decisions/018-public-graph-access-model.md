# ADR 018: Public Graph Access Model

## Status

Accepted — 2026-09-08

## Context

Knowledge Graph was built as a personal knowledge base, but the product intent
has always been wider: published notes should form a **community graph** that an
unauthenticated visitor can browse, read and search, with the option of later
splitting into sub-graphs for distinct communities. Signing in moves the visitor
from the shared graph to their own.

By September 2026 that intent was half implemented, and the half that existed
was misleading rather than merely incomplete.

### What the code actually did

A graph node carries only `ID`, `Title`, `Type`, `Creator`, `Public`
(`services/graph-service/internal/db/postgres_client.go`). It carries no note
content. Reading a single note goes through `GET /api/v1/notes/:id`, which sat
behind the global `JWTAuth` middleware, as did search:

```
GET /api/v1/notes/search?q=star   ->  401 Unauthorized
```

So an anonymous visitor saw a canvas of labelled points and could open none of
them. The showcase existed; the thing it was showcasing did not.

Separately, the choice between the two graphs was welded to the session state
(`frontend/src/shared/api/graph.ts`):

```ts
const endpoint = isAuthenticated() ? "v1/graph/full" : "v1/graph/public";
```

An authenticated user therefore could not return to the community graph at all —
the people most able to contribute to it were the only ones who could not see it.

A third, smaller problem: three endpoint names for two concepts, where the most
prominent name was the most misleading. `/api/v1/graph/all` is anonymous and
returns the public subset; `/api/v1/graph/public` and `/api/v1/graph/full` on the
graph service mean what they say. `all` reads as the widest scope and is the
narrowest — the author of this repository's external audit misread it on exactly
that basis.

## Problem Statement

How should the public perimeter be shaped so that the community graph is a usable
product surface, without weakening the isolation that keeps private notes private?

## Decision Drivers

- **The stated product model must be achievable.** A visitor who can see titles
  but open nothing is not browsing a community.
- **Private data isolation is non-negotiable.** Opening the perimeter must not
  become a path to someone else's notes.
- **Future community sub-graphs.** Whatever is decided now must not foreclose
  them.
- **Attribution.** A community graph without authorship is an anonymous dump.
- **Reversibility.** Naming and view choices should be cheap to revisit; access
  rules should not need revisiting at all.

## Considered Options

### 1. Leave the perimeter closed; the graph stays a teaser

Anonymous visitors see structure only, which nudges them to register.

Rejected. It contradicts the product model rather than implementing it, and it
makes the anonymous surface untestable in any meaningful way — a visual baseline
of that surface can only ever capture a graph and error banners.

### 2. Open reading only, keep search behind the login

Cheaper, and reduces query load from unauthenticated traffic.

Rejected. Navigation would be possible only by clicking through the graph. A
community whose contents cannot be searched is a curiosity, not a resource.

### 3. Open reading and search over the public subset — **chosen**

`GET /api/v1/notes/:id` and `GET /api/v1/notes/search` leave the JWT perimeter,
scoped to `is_public = true`. A private note answers `404` to a caller who does
not own it, so its existence is not confirmable.

### 4. View selection: welded to auth, or an explicit mode

Rejected the welded form. The graph choice became a **view mode** — *My graph* /
*Community* — with the session state supplying only the default. A third
combined mode was deliberately deferred: it needs a decision on how to
distinguish other people's nodes visually, which is a design question, not an
access one.

### 5. Endpoint naming: alias or hard rename

Rejected the alias. `/api/v1/graph/all` becomes `/api/v1/graph/public` with no
compatibility path. There are no external consumers, and a `deprecated` route
that nobody ever removes is the same drift, only politely dressed.

## Decision

1. **Anonymous callers may read a public note and search the public subset.**
   Refusals answer `404`, never `403`.
2. **The community graph is a view mode**, not a consequence of being logged out.
   Authentication sets the default; the user sets the mode.
3. **`/api/v1/graph/all` is renamed to `/api/v1/graph/public`**, without an alias.
4. **Author names stay in the public graph.** This is attribution and it is
   deliberate: usernames are public data on published notes, and the publish
   flow should say so.

## Consequences

### Positive

- The product model becomes achievable end to end: browse, open, search, and
  switch back to your own graph without signing out.
- Community sub-graphs now have somewhere to attach — they are additional modes,
  not a rewrite of the access rules.
- The anonymous surface becomes testable, which unblocks a genuinely anonymous
  visual baseline.

### Negative

- The public perimeter grows, and with it the surface that must be verified. The
  mitigation is structural rather than procedural: object-level authorisation on
  every `/notes/:id` route (ADR-adjacent work, `middleware.RequireNoteAccess`),
  with a router-level test that enumerates routes from the router itself, so a
  future route added without a guard fails the test rather than shipping.
- Unauthenticated search adds query load with no per-user rate context.
- The rename is breaking. It is recorded in `CHANGELOG.md` and applied in one
  change rather than left to decay.

### Neutral

- `Cache-Control` stays `private` with `Vary` on the shared routes. The same
  route now serves both audiences, so a `public` cache directive would leak an
  owner's response into a shared cache — a regression of an earlier audit finding.

## When to Reconsider

- If unauthenticated search becomes a load or abuse problem, the first move is
  rate limiting on the anonymous path, not closing the endpoint.
- If community sub-graphs arrive, revisit whether *Community* should become a
  selector over several graphs rather than a single mode.
- If author attribution becomes unwanted, the change is a display name rather
  than removal — removing authorship would undo the point of the surface.

## References

- Task specifications: `docs/tasks/PUB-1-anonymous-read-and-search.md`,
  `PUB-2-graph-view-mode.md`, `PUB-3-rename-graph-endpoints.md`
- Related work: `docs/tasks/SEC-1-note-idor.md` — object-level authorisation,
  which had to land first because it removes the barrier this ADR then opens
- Perimeter separation: `docs/GRAPH_SERVICE_AUTH.md`, ADR 013
