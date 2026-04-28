# Test Structure

This directory contains BDD/Cucumber tests and test support files for the Knowledge Graph project.

## Actual Test Structure

```
backend/                          # Go backend
├── internal/
│   ├── domain/*/*_test.go        # 25 tests (entities, value objects)
│   ├── application/*/*_test.go   # 17 tests (use cases)
│   ├── infrastructure/*/*_test.go  # 62 tests (repositories, queue, nlp)
│   └── interfaces/api/*/*_test.go # 14 tests (HTTP handlers)
└── cmd/checkconfig/main.go       # Config validation CLI

frontend/                         # SvelteKit frontend
├── src/lib/
│   ├── components/*.spec.ts      # 18 files, ~200 unit tests
│   ├── api/*.test.ts             # 3 files, API client tests
│   └── utils/*.test.ts           # 1 file, utility tests
├── tests/*.spec.ts               # 10 files, Playwright E2E tests
└── tests/features/*.feature      # 3 files, BDD scenarios

tests/                            # BDD tests (this directory)
├── features/*.feature            # 11 files, 98 BDD scenarios
├── features/step_definitions/    # 6 TypeScript files
└── support/                      # Test support files

nlp-service/                      # Python NLP service
└── tests/*.py                    # 2 files, ~15 pytest tests
```

## Unit Tests Location

Unit tests follow language-specific best practices:

### Go (backend/)
- `*_test.go` files next to source code
- **Total: 31 files, 118 test functions**

### TypeScript (frontend/)
- `*.spec.ts` — component tests (Vitest + Testing Library)
- `*.test.ts` — API client and utility tests (Vitest)
- **Total: 22 test files (~220 tests)**
  - 18 component tests (.spec.ts)
  - 4 API/utils tests (.test.ts)

### Python (nlp-service/)
- `tests/*.py` directory
- **Total: 2 files, ~15 tests**

## Running Tests

### Backend Unit Tests
```bash
cd backend
go test ./... -v                    # All tests
go test ./internal/domain/... -v     # Domain only
go test -race -coverprofile=coverage.out ./...
```

### Frontend Unit Tests
```bash
cd frontend
npm run test:unit                   # Vitest (204 tests)
npm run test:unit:watch            # Watch mode
npm run test:coverage              # With coverage
```

### Frontend E2E Tests (Playwright)
```bash
cd frontend
npm run test                       # All E2E tests
npm run test:smoke                 # Smoke tests only
npm run test:headed               # With browser visible
```

### BDD Tests (Cucumber)
```bash
cd frontend
npm run test:bdd                   # or npm run test:cucumber
```

### NLP Service Tests
```bash
cd nlp-service
pytest tests/ -v
pytest tests/test_api.py -v
pytest tests/test_nlp_utils.py -v
```

## Test Counts Summary

| Category | Files | Tests/Scenarios | Status |
|----------|-------|-----------------|--------|
| **Go Unit** | 31 | 118 test functions | ✅ Active |
| **Frontend Unit** | 22 | ~220 tests | ✅ Active |
| **Playwright E2E** | 10 | 48 tests | ✅ Active |
| **BDD Features** | 14 | 111 scenarios | ✅ Active |
| **NLP Python** | 2 | ~15 tests | ✅ Active |
| **Total** | **79** | **~512** | ✅ |
