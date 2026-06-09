# Agents in Knowledge Graph

Этот документ описывает агентские файлы, используемые в репозитории для поддержки работы с frontend, backend, тестирования, интеграции и документацией.

## Список агентов

### `knowledge-graph-frontend-svelte`
- **Фокус:** `frontend/`, Svelte 5, UI/UX, тестовая инфраструктура (Playwright/Vitest), паттерны хранения временных артефактов
- **Идеален для:** задач по визуальной части, компонентам, архитектуре frontend, тестированию UI и документации frontend

**Основные задачи:**
- Анализ и доработка UI/UX компонентов на Svelte 5
- Документирование архитектурных паттернов и frontend-конвенций
- Управление тестовой инфраструктурой (Playwright, Vitest, Lighthouse)
- Политика хранения временных тестовых артефактов

**Архитектурные паттерны:**
- Компонентная архитектура: атомы → молекулы → организмы
- Централизованное состояние через stores
- Изоляция логики: `src/lib`/`src/utils` для бизнес-логики, `src/components` для UI
- Vite + SvelteKit для сборки
- Snapshot-тесты и визуальные снимки

**Политика временных данных:**
- `frontend/test-results/temp/` — временные файлы (очищаются CI)
- `frontend/test-results/baseline/` — золотые эталоны (сохраняются)
- Скриншоты: `frontend/test-results/temp/screenshots/`
- Логи: `frontend/test-results/temp/logs/`

**Примеры промптов:**
- "Проанализируй `frontend/src/components/GraphCanvas` и предложи упрощения для Svelte 5"
- "Настрой playwright для сохранения скриншотов в `frontend/test-results/temp/screenshots`"
- "Добавь описание паттернов UI/UX в документацию"

---

### `knowledge-graph-backend-go`
- **Фокус:** `backend/`, Go 1.23+, архитектура DDD, Docker, PostgreSQL, Redis, API, миграции
- **Идеален для:** задач по backend, инфраструктуре, миграциям БД, сервисам и технической документации

**Основные задачи:**
- Backend Go рефакторинг, bug fixes, feature work
- Анализ и обновление инфраструктуры (Docker Compose, PostgreSQL, Redis)
- Чтение, исправление и расширение технической документации
- Синхронизация кода и документации

**Архитектурные паттерны:**
- Clean Architecture: Domain, Application, Infrastructure, Interfaces слои
- DDD: Entities, Value Objects, Aggregates, Repository Pattern
- CQRS-Lite: разделение команд и запросов
- Private fields в entities с factory functions
- Immutable Value Objects с валидацией

**Структура проекта:**
```
backend/
├── cmd/server/main.go          # Entry point, dependency injection
├── internal/
│   ├── domain/                 # Pure business logic
│   ├── application/            # Use cases, services
│   ├── infrastructure/         # DB, Redis, config
│   └── interfaces/api/         # HTTP handlers
└── migrations/                 # SQL миграции
```

**Команды разработки:**
```bash
cd backend
go test ./... -v                         # Unit тесты
go test -tags=integration ./...          # Интеграционные тесты
go test -race -coverprofile=coverage.out ./...
go run ./cmd/server                       # Запуск сервера
```

**Примеры промптов:**
- "Проанализируй backend Go этого проекта и исправь актуальные ошибки"
- "Оцени состояние кода и docker-окружения, затем обнови документацию"
- "Найди и поправь ошибки в `backend/internal/infrastructure/db/postgres/note_repo.go`"

---

### `knowledge-graph-docs-maintenance`
- **Фокус:** `README.md`, `docs/`, ADR, changelog, сопроводительная документация, `.github/`
- **Идеален для:** актуализации, оформления изменений и проверки корректности документации

**Основные задачи:**
- Обновление и актуализация `README.md`
- Формирование описаний изменений и changelog-подобных примечаний
- Поддержание ADR, архитектурной документации и конвенций
- Проверка корректности ссылок и структуры документации
- Оформление новых команд и workflow

**Типы задач:**
- Добавление новых разделов в README.md
- Создание инструкций по новым скриптам и командам
- Фиксация паттернов и правил в документации
- Создание/обновление ADR
- Генерация changelog-записей

**Примеры промптов:**
- "Оформи README.md разделом 'Как пользоваться агентами и командами'"
- "Проверь и обнови документацию по запуску periodic cleanup"
- "Создай краткую инструкцию по использованию новых Makefile и npm-скриптов"
- "Сгенерируй changelog для последних изменений"

**Ограничения:**
- Не изменяет функциональный код без явного запроса
- Работает только с документацией и описаниями
- Сохраняет текущее содержание, добавляя пояснения

---

### `knowledge-graph-testing`
- **Фокус:** Тесты всех уровней — Go unit/integration, frontend Vitest/Playwright/BDD, Python pytest
- **Идеален для:** написания, отладки и анализа тестов, работы с тестовой инфраструктурой и покрытием

**Основные задачи:**
- Написание и улучшение unit тестов (Go, TypeScript, Python)
- Создание и поддержка интеграционных тестов API
- Разработка E2E тестов с Playwright
- Написание BDD сценариев на Cucumber
- Анализ покрытия тестов и выявление пробелов
- Отладка failing тестов и исправление тестовой инфраструктуры

**Тестовая инфраструктура:**

*Backend (Go):*
- Unit тесты: `backend/internal/**/*_test.go` (118 тестов)
- Интеграционные тесты: `*_integration_test.go` с тегом `//go:build integration`
- Паттерн: table-driven tests, testify/suite
- Запуск: `cd backend && go test ./... -v`, `go test -tags=integration ./...`
- Coverage: `go test -coverprofile=coverage.out ./...`

*Frontend (TypeScript):*
- Unit тесты: `frontend/src/lib/**/*.spec.ts` (Vitest, ~220 тестов)
- API тесты с MSW mocks: `frontend/src/lib/api/*.test.ts`
- E2E тесты: `frontend/tests/*.spec.ts` (Playwright, 48 тестов)
- BDD тесты: `tests/features/*.feature` (Cucumber, 111 сценариев)
- Запуск: `cd frontend && npm run test:unit`, `npm run test`, `npm run test:bdd`

*NLP Service (Python):*
- Pytest тесты: `nlp-service/tests/*.py` (~15 тестов)
- Запуск: `cd nlp-service && pytest tests/ -v`

**Паттерны тестирования:**

*Go table-driven tests:*
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

*MSW mocks для frontend:*
```typescript
server.use(
    http.get('http://localhost:8080/api/v1/notes', () => HttpResponse.json({ notes: [] }))
);
```

*Cucumber Gherkin:*
```gherkin
Given I have test notes with connections
When I navigate to the graph view
Then I should see the notes displayed as celestial bodies
```

**Интеграционные паттерны:**
- SKIP_AUTH флаг: `window.__SKIP_AUTH__ = true` для обхода аутентификации
- Test database cleanup: `testutil.SetupTestDB()` для изолированной БД
- Cucumber Before/After hooks для очистки данных
- MSW response mocking для API unit тестов

**Примеры промптов:**
- "Добавь unit тесты для новой функции в Note entity"
- "Исправь падающие Playwright тесты в frontend/tests/graph.spec.ts"
- "Напиши BDD сценарий для новой фильтрации по типам заметок"
- "Проанализируй покрытие тестов и предложи где добавить тесты"
- "Настрой интеграционные тесты для нового handler"

---

### `knowledge-graph-integration`
- **Фокус:** Интеграция между backend API и frontend — mapping endpoints, DTO типы данных, middleware, интеграционные тесты
- **Идеален для:** синхронизации API контрактов, согласования типов данных, настройки аутентификации и работы со связями между слоями

**Основные задачи:**
- Синхронизация API endpoints между backend handlers и frontend API clients
- Согласование типов данных (DTO) между Go structs и TypeScript interfaces
- Настройка middleware (CORS, JWT, rate limiting, SKIP_AUTH)
- Написание и поддержка интеграционных тестов API
- Документирование API контрактов и ошибок
- Анализ соответствия frontend routes к backend endpoints

**Структура API интеграции:**

*Backend endpoints:*
```go
// Auth routes
v1.POST("/auth/register", authHandler.Register)
v1.POST("/auth/login", authHandler.Login)

// Notes
v1.POST("/notes", writeLimiter, noteHandler.Create)
v1.GET("/notes/:id", noteHandler.Get)
v1.PUT("/notes/:id", writeLimiter, noteHandler.Update)

// Graph
v1.GET("/notes/:id/graph", graphHandler.GetGraph)
v1.GET("/me/graph/cached", graphHandler.GetCachedGraph)
v1.GET("/me/graph/fresh", graphHandler.GetFreshGraph)
```

*Frontend API clients:*
```typescript
export async function getNotes(): Promise<Note[]> {
  const response = await api.get('v1/notes', { searchParams: { limit: 10000 } }).json<{ notes: Note[]; total: number; limit: number; offset: number }>();
  return response.notes;
}

export async function createNote(data: { title: string; content?: string; type?: string; email?: string }): Promise<Note> {
  return api.post('v1/notes', { json: data }).json();
}
```

*Frontend routes → Backend API mapping:*
```
frontend/src/routes/notes/new/+page.svelte → POST /api/v1/notes
frontend/src/routes/notes/[id]/+page.svelte → GET /api/v1/notes/:id
frontend/src/routes/notes/[id]/edit/+page.svelte → PUT /api/v1/notes/:id
frontend/src/routes/graph/+page.svelte → GET /api/v1/me/graph/cached
frontend/src/routes/graph/[id]/+page.svelte → GET /api/v1/notes/:id/graph
```

**DTO и типы данных:**

*Backend Go structs:*
```go
type NoteResponse struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Type      string    `json:"type"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

*Frontend TypeScript interfaces:*
```typescript
export interface Note {
  id: string;
  title: string;
  content: string;
  metadata: Record<string, any>;
  type?: string;
  created_at: string;
  updated_at: string;
}
```

**Middleware и аутентификация:**

*CORS настройка:*
```go
r.Use(func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
    c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With, X-Backend-Url")
})
```

*JWT Authentication:*
```go
jwtConfig := middleware.DefaultJWTConfig(jwtManager, tokenStore)
r.Use(middleware.JWTAuth(jwtConfig))
```

*SKIP_AUTH для тестов:*
- Frontend: `window.__SKIP_AUTH__ = true` в init script
- Backend: `middleware.SkipAuth(middleware.DefaultSkipAuthConfig(true))` при `cfg.SkipAuth`

**Интеграционные тесты:**

*Backend integration tests:*
```go
// *_integration_test.go с тегом //go:build integration
func (s *NoteHandlerIntegrationTestSuite) SetupSuite() {
    s.db, s.cleanup = testutil.SetupTestDB(s.T())
    s.router.POST("/notes", handler.Create)
}
```

*Frontend API integration tests (MSW):*
```typescript
server.use(
    http.get('http://localhost:8080/api/v1/notes', () => HttpResponse.json({ notes: [mockNote] }))
);
```

**API документация:**
- OpenAPI spec: `openAPI.yaml` в корне проекта
- API Errors: `docs/API_ERRORS.md` — коды ошибок и форматы
- Swagger UI: `/swagger/*any` endpoint

**Примеры промптов:**
- "Добавь новый endpoint для создания тегов и обнови frontend API client"
- "Обнови DTO в Note response после добавления нового поля"
- "Настрой CORS для нового домена frontend"
- "Исправь несоответствие типов между backend и frontend"
- "Добавь интеграционный тест для нового graph endpoint"
- "Обнови документацию API после изменений в handlers"

---

## Зачем нужны эти агенты

Агенты в этом репозитории не являются исполняемыми скриптами. Это метаданные, которые помогают выбрать правильный контекст при работе с проектом.

Они полезны, когда нужно:

- **Разделить задачи по области ответственности** (frontend / backend / docs / testing / integration)
- **Сохранить согласованность формата документации**
- **Не смешивать технические детали backend с UI-рекомендациями**
- **Задокументировать изменения и правила использования новых команд**
- **Обеспечить синхронизацию API контрактов между backend и frontend**
- **Поддерживать высокое качество тестового покрытия**
- **Следовать архитектурным паттернам и конвенциям проекта**

## Как ими пользоваться

1. **Найдите задачу или проблему**
2. **Выберите агента по области:**
   - Frontend/UI/UX → `knowledge-graph-frontend-svelte`
   - Backend/Infrastructure → `knowledge-graph-backend-go`
   - Документация/README/ADR → `knowledge-graph-docs-maintenance`
   - Тестирование → `knowledge-graph-testing`
   - Интеграция API/Backend-Frontend → `knowledge-graph-integration`
3. **Следуйте рекомендациям агента** при обновлении файлов

## Матрица выбора агента

| Тип задачи | Агент |
|------------|-------|
| Рефакторинг Svelte компонентов | `knowledge-graph-frontend-svelte` |
| Настройка Playwright/Vitest | `knowledge-graph-frontend-svelte` / `knowledge-graph-testing` |
| Backend Go разработка | `knowledge-graph-backend-go` |
| Docker/инфраструктура | `knowledge-graph-backend-go` |
| Обновление README | `knowledge-graph-docs-maintenance` |
| Создание ADR | `knowledge-graph-docs-maintenance` |
| Написание тестов | `knowledge-graph-testing` |
| Отладка падающих тестов | `knowledge-graph-testing` |
| Добавление API endpoint | `knowledge-graph-integration` |
| Синхронизация DTO типов | `knowledge-graph-integration` |
| Настройка CORS/auth middleware | `knowledge-graph-integration` |

## Команды для быстрого запуска

### Frontend
```bash
cd frontend
npm run dev                    # Dev server
npm run test:unit              # Vitest unit тесты
npm run test                   # Playwright E2E тесты
npm run test:bdd               # Cucumber BDD тесты
npm run build                  # Production build
```

### Backend
```bash
cd backend
go run ./cmd/server            # Запуск сервера
go test ./... -v               # Unit тесты
go test -tags=integration ./... # Интеграционные тесты
go test -coverprofile=coverage.out ./... # С покрытием
```

### NLP Service
```bash
cd nlp-service
uvicorn app.main:app --reload  # Dev server
pytest tests/ -v               # Тесты
```

### Полный стек
```bash
docker-compose up              # Все сервисы
docker-compose up postgres redis # Только БД и кэш
```

## Примеры задач по агентам

### Frontend задачи
- "Добавь правило для хранения временных тест-артефактов в frontend" — `knowledge-graph-frontend-svelte`
- "Проанализируй `frontend/src/components/GraphCanvas` и предложи упрощения для Svelte 5" — `knowledge-graph-frontend-svelte`
- "Настрой playwright для сохранения скриншотов в `frontend/test-results/temp/screenshots`" — `knowledge-graph-frontend-svelte`

### Backend задачи
- "Проанализируй backend Go этого проекта и исправь актуальные ошибки" — `knowledge-graph-backend-go`
- "Оцени состояние кода и docker-окружения, затем обнови документацию" — `knowledge-graph-backend-go`
- "Найди и поправь ошибки в `backend/internal/infrastructure/db/postgres/note_repo.go`" — `knowledge-graph-backend-go`

### Documentation задачи
- "Обнови архитектурную документацию и ADR по новому паттерну" — `knowledge-graph-backend-go` и `knowledge-graph-docs-maintenance`
- "Оформи README раздел с командами обслуживания и periodic cleanup" — `knowledge-graph-docs-maintenance`
- "Создай краткую инструкцию по использованию новых Makefile и npm-скриптов" — `knowledge-graph-docs-maintenance`

### Testing задачи
- "Напиши интеграционные тесты для нового graph endpoint" — `knowledge-graph-testing`
- "Добавь unit тесты для новой функции в Note entity" — `knowledge-graph-testing`
- "Проанализируй покрытие тестов и предложи где добавить тесты" — `knowledge-graph-testing`
- "Исправь падающие Playwright тесты в frontend/tests/graph.spec.ts" — `knowledge-graph-testing`

### Integration задачи
- "Добавь новый endpoint для создания тегов и обнови frontend API client" — `knowledge-graph-integration`
- "Исправь несоответствие типов между backend и frontend" — `knowledge-graph-integration`
- "Настрой CORS для нового домена frontend" — `knowledge-graph-integration`
- "Обнови документацию API после изменений в handlers" — `knowledge-graph-integration`

## Советы по использованию

1. **Будьте конкретны в промптах** — указывайте конкретные файлы и типы задач
2. **Используйте соответствующего агента** — это обеспечит правильный контекст
3. **Следуйте паттернам проекта** — агенты знают архитектурные конвенции
4. **Проверяйте команды** — все команды в документации проверены для проекта
5. **Обновляйте документацию** — при изменениях обновляйте соответствующие разделы
