# Minimal Manual Test Checklist

Quick smoke checklist for Knowledge Graph before merging feature branches to `main`.

> **Note:** The default test stack uses `SKIP_AUTH=true`. To test real JWT authentication, set `SKIP_AUTH=false` before starting the test stack and run the [real-auth checklist](#real-auth-mode) below.

## Pre-requisites

```powershell
# Default skip-auth test stack
.\scripts\testing\start-test.ps1
.\scripts\testing\seed-test-data.ps1
```

For **real-auth** testing on Windows:

```powershell
$env:SKIP_AUTH = "false"
# If port 3002 is in a Hyper-V excluded range, also set FRONTEND_PORT:
# $env:FRONTEND_PORT = "50070"
.\scripts\testing\start-test.ps1
.\scripts\testing\seed-test-data.ps1
```

On Linux/Mac:

```bash
export SKIP_AUTH=false
./scripts/testing/start-test.sh
./scripts/testing/seed-test-data.sh
```

**URLs**

- Frontend: `http://127.0.0.1:3002`
- Backend: `http://127.0.0.1:18083`
- Graph service (HTTP): `http://127.0.0.1:19091`
- Test user: `testuser` / `TestPassword123!`

## Backend

- [ ] `GET http://127.0.0.1:18083/health` returns `{"status":"ok"}`
- [ ] `GET http://127.0.0.1:18083/api/v1/notes?limit=1` returns notes (skip-auth) or `401` (real-auth)
- [ ] `POST http://127.0.0.1:18083/api/v1/auth/register` creates a user (real-auth) or `201`/`409` (skip-auth)
- [ ] `POST http://127.0.0.1:18083/api/v1/auth/login` returns JWT (real-auth) or `200`/`skip` (skip-auth)
- [ ] Authenticated `GET http://127.0.0.1:18083/api/v1/me/graph/fresh` with `Authorization: Bearer <token>` returns graph data

## Graph Service

- [ ] `GET http://127.0.0.1:19091/health` returns `{"status":"ok","service":"graph-service"}`
- [ ] `GET http://127.0.0.1:19091/api/v1/graph/public` returns public-only nodes (no auth required)
- [ ] `GET http://127.0.0.1:19091/api/v1/graph/full` without token returns `401` (real-auth) or user-scoped data (skip-auth)
- [ ] `GET http://127.0.0.1:19091/api/v1/graph/full` with valid `Authorization: Bearer <token>` returns user-scoped nodes

## Frontend

- [ ] `http://127.0.0.1:3002` loads without console errors
- [ ] Home page lists notes
- [ ] Graph page (`/graph`) renders 2D graph
- [ ] Switch to 3D view works
- [ ] Creating a note from the graph works
- [ ] Public graph toggle shows only public notes (when public notes exist)

## Real-auth mode

Run after `SKIP_AUTH=false` start:

1. [ ] Open `http://127.0.0.1:3002` in an incognito window.
2. [ ] Public home graph loads without `401` loops.
3. [ ] Click **Login** and sign in with `testuser` / `TestPassword123!`.
4. [ ] After login, home page (`/`) loads notes for `testuser`.
5. [ ] Open DevTools Network: graph-service requests include `Authorization: Bearer <token>`.
6. [ ] Switch to list view and filter by type — only matching notes remain.
7. [ ] Open `/graph` — 2D graph renders user-scoped nodes.
8. [ ] Create a new note from the graph (`N` or side panel) — it appears in list and graph.
9. [ ] Logout works and returns to public view.
10. [ ] `GET http://127.0.0.1:19091/api/v1/graph/full` without token returns `401`.
11. [ ] `GET http://127.0.0.1:19091/api/v1/graph/full` with token returns user data.

## E2E

```bash
cd frontend

# Skip-auth smoke (default test stack)
npm run test:smoke
npm run test:bdd:skipauth

# Real-auth smoke / BDD (SKIP_AUTH=false test stack)
npm run test:realauth
npm run test:bdd:realauth
```

- [ ] Smoke tests pass
- [ ] BDD smoke tests pass

## Known Non-Blocking Issues

- `smoke-real-auth` may log a `400` on `/api/v1/auth/refresh` before login and a `404` on `/api/v1/notes/00000000-0000-0000-0000-000000000001`; both are expected.
- During the login → home-page transition, `data-testid="graph-stats"` can briefly resolve to two elements; the smoke test uses `.first()`.

## Cleanup

```bash
.\scripts\testing\stop-test.ps1
```
