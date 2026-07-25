# Minimal Manual Test Checklist

Quick smoke checklist for Knowledge Graph before merging `feature/ai-agents-integration` to `main`.

## Pre-requisites

```bash
# Start test stack
.\scripts\testing\start-test.ps1
.\scripts\testing\seed-test-data.ps1
```

## Backend

- [ ] `GET http://127.0.0.1:8083/health` returns `{"status":"ok"}`
- [ ] `GET http://127.0.0.1:8083/api/v1/notes?limit=1` returns notes
- [ ] `POST http://127.0.0.1:8083/api/v1/auth/register` creates a user
- [ ] `POST http://127.0.0.1:8083/api/v1/auth/login` returns JWT
- [ ] Authenticated `GET http://127.0.0.1:8083/api/v1/me/graph/fresh` returns graph data

## Graph Service

- [ ] `GET http://127.0.0.1:19091/health` returns `{"status":"ok","service":"graph-service"}`
- [ ] `GET http://127.0.0.1:19091/api/v1/graph/public` returns public-only nodes
- [ ] `GET http://127.0.0.1:19091/api/v1/graph/full` without token returns `401`
- [ ] `GET http://127.0.0.1:19091/api/v1/graph/full` with valid JWT returns user-scoped nodes

## Frontend

- [ ] `http://127.0.0.1:3002` loads without console errors
- [ ] Home page lists notes
- [ ] Graph page renders 2D graph
- [ ] Switch to 3D view works
- [ ] Creating a note from the graph works
- [ ] Public graph toggle shows only public notes

## E2E

```bash
cd frontend
npm run test:smoke
npm run test:bdd:skipauth
```

- [ ] Smoke tests pass
- [ ] BDD smoke tests pass

## Cleanup

```bash
.\scripts\testing\stop-test.ps1
```
