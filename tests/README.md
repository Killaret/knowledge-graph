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
├── src/lib/
│   ├── components/*.spec.ts      # Component unit tests (Vitest)
│   ├── api/*.test.ts             # API client tests
│   ├── services/*.test.ts        # Service tests
│   ├── stores/*.test.ts          # Svelte store tests
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
```

### Frontend E2E Tests (Playwright)
```bash
cd frontend
npx playwright test                    # All E2E tests
npm run test:smoke                     # Smoke tests only
npm run test:headed                    # With browser visible
```

### BDD Tests (Cucumber)
```bash
cd frontend
npm run test:cucumber                  # or from project root:
# node --import ./frontend/node_modules/tsx/dist/loader.mjs ./frontend/node_modules/@cucumber/cucumber/bin/cucumber.js --config ./cucumber.mjs
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
| **Go Unit** | 31 | ~118 test functions | ✅ pass |
| **Go Integration** | 5+ | repository/handler suites | ⚠️ requires clean Docker testcontainers (reaper conflicts on current host) |
| **Frontend Unit** | 52 | 526 tests | ✅ pass (43 skipped) |
| **Playwright E2E** | 18 | 89 tests | ⚠️ 74 passed, 14 failed, 1 skipped |
| **BDD Features** | 1 | 5 scenarios | ⚠️ 2 passed, 3 failed |
| **NLP Python** | 2 | 33 collected | ✅ 28 passed, 5 skipped |

> **Note:** Go integration tests are resource-intensive and failed on this run due to Docker testcontainers reaper/container-name conflicts and Docker Engine instability. They should be run on a clean Docker environment for accurate results.

## Known Issues

1. **BDD runner:** `npm run test:cucumber` in `frontend/` exits with 0 scenarios because the `cucumber.mjs` config uses paths relative to the project root. Run from the project root with the loader path shown above.
2. **E2E failures:** Several failures are in `preload-full-cycle.spec.ts` (login form selectors) and `auth-skip-auth.spec.ts` (`page` undefined / profile content visibility), indicating outdated selectors or test environment drift.
3. **Go integration tests:** require Docker Desktop to be stable and free of stale testcontainers reaper containers.
