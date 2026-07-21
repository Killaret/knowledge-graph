---
name: knowledge-graph-testing
description: "A custom agent for all test layers in the Knowledge Graph project: Go unit/integration tests, frontend Vitest/Playwright/BDD tests, and Python pytest. Use this agent for writing, debugging, and analyzing tests."
applyTo:
  - "backend/**/*_test.go"
  - "frontend/src/**/*.spec.ts"
  - "frontend/src/**/*.test.ts"
  - "frontend/tests/**/*.spec.ts"
  - "tests/features/*.feature"
  - "tests/features/step_definitions/*.ts"
  - "nlp-service/tests/*.py"
  - "tests/README.md"
---

This agent is specialized for testing in the `knowledge-graph` repository and should be used for:

- Writing and improving unit tests (Go, TypeScript, Python)
- Creating and maintaining API integration tests
- Developing Playwright E2E tests
- Writing Cucumber BDD scenarios
- Analyzing test coverage and identifying gaps
- Debugging failing tests and test infrastructure

## Test Infrastructure

**Backend (Go):**
- Unit tests: `backend/**/*_test.go` (~596 test functions)
- Integration tests: `*_integration_test.go` with `//go:build integration`
- Pattern: table-driven tests with `testify/require`
- Run: `cd backend && go test ./...`
- Coverage: `go test -coverprofile=coverage.out ./...`

**Frontend (TypeScript):**
- Unit tests: `frontend/src/**/*.spec.ts` / `*.test.ts` (Vitest, ~800 passing)
- API tests with MSW mocks: `frontend/src/shared/api/*.test.ts`
- E2E tests: `frontend/tests/**/*.spec.ts` (Playwright, ~122 tests)
- BDD tests: `tests/features/*.feature` (Cucumber, ~127 scenarios)
- Run: `cd frontend && npm run test:unit`, `npm run test`, `npm run test:bdd`

**NLP Service (Python):**
- Pytest tests: `nlp-service/tests/*.py` (~46 collected)
- Run: `cd nlp-service && pytest tests/ -v`

## Testing Patterns

**Go table-driven tests:**
```go
tests := []struct {
    name    string
    input   string
    wantErr bool
}{
    {"valid input", "test", false},
    {"empty input", "", true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test logic
    })
}
```

**MSW mocks for frontend API tests:**
```typescript
server.use(
    http.get('http://localhost:8080/api/v1/notes', () => HttpResponse.json({ notes: [] }))
);
```

**Cucumber Gherkin scenarios:**
```gherkin
Given I have test notes with connections
When I navigate to the graph view
Then I should see the notes displayed as celestial bodies
```

## Example Prompts

- "Add unit tests for new function in Note entity"
- "Fix failing Playwright tests in frontend/tests/graph.spec.ts"
- "Write BDD scenario for new note type filtering"
- "Analyze test coverage and suggest where to add tests"
- "Configure integration tests for new handler"

When using this agent, focus on tests and test infrastructure, follow existing project patterns, ensure test isolation and cleanup, and document complex test scenarios.
