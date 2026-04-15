# Test Structure

This directory contains all integration and end-to-end tests for the Knowledge Graph project.

## Structure

```
tests/
├── backend/
│   ├── integration/     # API integration tests
│   └── e2e/            # Backend end-to-end tests
├── frontend/
│   └── e2e/            # Playwright E2E tests
└── nlp/
    └── integration/    # NLP service integration tests
```

## Unit Tests

Unit tests are located next to the source code they test (Go best practice):

- `backend/internal/domain/*/*_test.go` - Domain layer tests
- `backend/internal/application/*/*_test.go` - Application layer tests
- `backend/internal/infrastructure/*/*_test.go` - Infrastructure tests
- `backend/internal/interfaces/*/*_test.go` - Interface layer tests
- `nlp-service/tests/` - NLP service unit tests

## Running Tests

### Backend Integration Tests
```bash
cd tests/backend/integration
go test -v ./...
```

### Frontend E2E Tests
```bash
cd tests/frontend/e2e
npx playwright test
```

### NLP Integration Tests
```bash
cd tests/nlp/integration
pytest ./...
```
