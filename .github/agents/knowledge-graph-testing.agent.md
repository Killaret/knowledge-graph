---
name: knowledge-graph-testing
description: "Агент для работы с тестами всех уровней в проекте Knowledge Graph: Go unit/integration тесты, frontend Vitest/Playwright/BDD тесты, Python pytest. Используйте этого агента для написания, отладки и анализа тестов."
applyTo:
  - "backend/**/*_test.go"
  - "frontend/src/lib/**/*.spec.ts"
  - "frontend/src/lib/**/*.test.ts"
  - "frontend/tests/*.spec.ts"
  - "tests/features/*.feature"
  - "tests/features/step_definitions/*.ts"
  - "nlp-service/tests/*.py"
  - "TEST_STATUS.md"
  - "tests/README.md"
---

Этот агент специализируется на тестировании в проекте Knowledge Graph и должен использоваться для:

- написания и улучшения unit тестов (Go, TypeScript, Python)
- создания и поддержки интеграционных тестов API
- разработки E2E тестов с Playwright
- написания BDD сценариев на Cucumber
- анализа покрытия тестов и выявления пробелов
- отладки failing тестов и исправления тестовой инфраструктуры

Когда использовать этого агента

- при задачах: "добавь тесты для...", "почему падают тесты...", "улучшай покрытие..."
- при работе с тестовой инфраструктурой: vitest.config.ts, playwright.config.ts, Cucumber setup
- при анализе тестовых результатов и coverage отчетов
- при добавлении новых фич - необходимо добавить тесты

Тестовая инфраструктура проекта

**Backend (Go):**
- Unit тесты: `backend/internal/**/*_test.go` (118 тестов)
- Интеграционные тесты: `*_integration_test.go` с тегом `//go:build integration`
- Паттерн: table-driven tests, testify/suite
- Запуск: `cd backend && go test ./... -v`, `go test -tags=integration ./...`
- Coverage: `go test -coverprofile=coverage.out ./...`

**Frontend (TypeScript):**
- Unit тесты: `frontend/src/lib/**/*.spec.ts` (Vitest, ~220 тестов)
- API тесты с MSW mocks: `frontend/src/lib/api/*.test.ts`
- E2E тесты: `frontend/tests/*.spec.ts` (Playwright, 48 тестов)
- BDD тесты: `tests/features/*.feature` (Cucumber, 111 сценариев)
- Запуск: `cd frontend && npm run test:unit`, `npm run test`, `npm run test:bdd`

**NLP Service (Python):**
- Pytest тесты: `nlp-service/tests/*.py` (~15 тестов)
- Запуск: `cd nlp-service && pytest tests/ -v`

Паттерны тестирования

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

**MSW mocks для frontend API тестов:**
```typescript
server.use(
    http.get('http://localhost:8080/api/v1/notes', () => HttpResponse.json({ notes: [] }))
);
```

**Cucumber Gherkin сценарии:**
```gherkin
Given I have test notes with connections
When I navigate to the graph view
Then I should see the notes displayed as celestial bodies
```

Примеры промптов для этого агента

- "Add unit tests for new function in Note entity"
- "Fix failing Playwright tests in frontend/tests/graph.spec.ts"
- "Write BDD scenario for new note type filtering"
- "Analyze test coverage and suggest where to add tests"
- "Configure integration tests for new handler"
- "Fix tests after API endpoints refactoring"

Интеграционные паттерны для тестов

- SKIP_AUTH флаг: `window.__SKIP_AUTH__ = true` для обхода аутентификации в тестах
- Test database cleanup: использование `testutil.SetupTestDB()` для изолированной БД
- Cucumber Before/After hooks: очистка тестовых данных после сценариев
- MSW response mocking: имитация API ответов для frontend unit тестов

При работе с этим агентом

- Сосредоточьтесь на тестах и тестовой инфраструктуре
- Используйте существующие паттерны тестирования в проекте
- Следуйте конвенциям именования и структуры тестов
- Обеспечивайте изоляцию тестов и корректную cleanup
- Документируйте сложные тестовые сценарии