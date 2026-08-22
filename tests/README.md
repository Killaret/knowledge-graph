# Test Structure

This directory contains BDD/Cucumber tests and test support files for the Knowledge Graph project.

## Actual Test Structure

```
backend/                          # Go backend
├── internal/
│   ├── domain/*/*_test.go        # Domain entities & value objects
│   ├── application/*/*_test.go   # Use cases
│   ├── infrastructure/*/*_test.go  # Repositories, queue, nlp, db
│   └── interfaces/api/*/*_test.go # HTTP handlers & middleware
└── cmd/checkconfig/main.go       # Config validation CLI

frontend/                         # SvelteKit frontend
├── src/components/**/*.spec.ts  # Component unit tests (Vitest)
├── src/features/**/*.test.ts    # Feature logic tests
├── src/shared/
│   ├── api/*.test.ts             # API client tests
│   ├── services/*.test.ts        # Service tests
│   ├── stores/*.svelte.test.ts   # Svelte runes store tests
│   └── utils/*.test.ts           # Utility tests
├── tests/*.spec.ts               # Playwright E2E tests
└── tests/features/*.feature      # BDD scenarios

nlp-service/                      # Python NLP service
└── tests/*.py                     # pytest tests
```

## Running Tests

### Backend Unit Tests
```bash
cd backend
go test ./...                         # All unit tests
go test -race ./...                  # Race detection (requires CGO_ENABLED=1)
go test -tags=integration ./...      # Integration tests with testcontainers
go test -coverprofile=coverage.out ./...
```

### Frontend Unit Tests
```bash
cd frontend
npm run test:unit                      # Vitest
npm run test:unit:watch               # Watch mode
npm run test:coverage                 # With coverage
npx svelte-check --tsconfig ./tsconfig.json  # Svelte/TypeScript check
```

### Frontend E2E Tests (Playwright)
```bash
cd frontend
npx playwright test                    # All E2E tests
npm run test:smoke                     # Smoke tests only
npm run test:headed                    # With browser visible
```

### BDD Tests (Cucumber)
Run from the project root; the `cucumber.mjs` config uses paths relative to the project root.
```bash
cd /d:/knowledge-graph
node --import ./frontend/node_modules/tsx/dist/loader.mjs ./frontend/node_modules/@cucumber/cucumber/bin/cucumber.js --config ./cucumber.mjs
```

### NLP Service Tests
```bash
cd nlp-service
pytest
pytest tests/ -v
```

## Test Counts Summary

| Category | Files | Tests/Scenarios | Latest Result |
|----------|-------|-------------------|---------------|
| **Go Unit** | 99 | 596 test functions | ✅ pass |
| **Go Race** | - | - | ⚠️ skipped (CGO_ENABLED=0 on Windows) |
| **Go Integration** | 5+ | repository/handler suites | ⚠️ requires Linux/WSL Docker + resources |
| **Frontend Unit** | 84 | 976+ tests | ✅ pass (37 skipped) |
| **Frontend svelte-check** | - | - | ✅ 0 errors, warnings acceptable |
| **Frontend Build** | - | - | ✅ success |
| **Playwright E2E** | 16 | 122 tests | ✅ most pass, some skipped |
| **BDD Features** | 14 | 127 scenarios | ✅ most pass |
| **Smoke Tests** | 1 | 9 tests | ✅ 9 passed |
| **NLP Python** | 2 | 46 collected | ✅ most pass |

> **Note:** Go integration tests are resource-intensive and failed on this run because PostgreSQL testcontainers could not reach the ready state within the startup timeout. They should be run on a clean Docker environment with adequate CPU/memory for accurate results.

## Known Issues

1. **Docker Desktop not available** - Cannot run dev and personal stacks for full integration testing
2. **Go integration tests:** require a stable Docker Desktop setup with sufficient resources and no stale testcontainers reaper containers
3. **Swagger UI:** `http://localhost:8080/swagger/` returns 404 because the `gin-swagger` handler requires a `swag init` generated `docs` package, which is not produced during the Docker build
4. **OpenAPI spec:** Missing 403 Forbidden response codes for all endpoints (403 should be documented for authorization errors)
