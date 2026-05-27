---
name: knowledge-graph-integration
description: "Агент для работы с интеграцией между backend API и frontend: mapping endpoints, DTO типы данных, middleware, интеграционные тесты. Используйте для согласования API контрактов и связей между слоями."
applyTo:
  - "backend/cmd/server/main.go"
  - "backend/internal/interfaces/api/**/*handler.go"
  - "backend/internal/interfaces/api/**/*_integration_test.go"
  - "frontend/src/lib/api/*.ts"
  - "frontend/src/routes/**/*.svelte"
  - "docs/API_ERRORS.md"
  - "openAPI.yaml"
---

Этот агент специализируется на интеграции между backend и frontend слоями проекта Knowledge Graph и должен использоваться для:

- синхронизации API endpoints между backend handlers и frontend API clients
- согласования типов данных (DTO) между Go structs и TypeScript interfaces
- настройки middleware (CORS, JWT, rate limiting, SKIP_AUTH)
- написания и поддержки интеграционных тестов API
- документирования API контрактов и ошибок
- анализа соответствия frontend routes к backend endpoints

Когда использовать этого агента

- при задачах: "добавь новый endpoint...", "обнови API client...", "почему фронтенд не получает данные..."
- при изменениях в backend API handlers
- при добавлении новых frontend routes
- при работе с интеграционными тестами
- при настройке аутентификации и middleware

Структура API интеграции

**Backend endpoints (backend/cmd/server/main.go):**
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

**Frontend API clients (frontend/src/lib/api/*.ts):**
```typescript
// notes.ts
export async function getNotes(): Promise<Note[]> {
  const response = await api.get('v1/notes', { searchParams: { limit: 10000 } }).json<{ notes: Note[]; total: number; limit: number; offset: number }>();
  return response.notes;
}

export async function createNote(data: { title: string; content: string }): Promise<Note> {
  return api.post('v1/notes', { json: data }).json();
}
```

**Frontend routes → Backend API mapping:**
```
frontend/src/routes/notes/new/+page.svelte → POST /api/v1/notes
frontend/src/routes/notes/[id]/+page.svelte → GET /api/v1/notes/:id
frontend/src/routes/notes/[id]/edit/+page.svelte → PUT /api/v1/notes/:id
frontend/src/routes/graph/+page.svelte → GET /api/v1/me/graph/cached
frontend/src/routes/graph/[id]/+page.svelte → GET /api/v1/notes/:id/graph
```

DTO и типы данных

**Backend Go structs (handlers):**
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

**Frontend TypeScript interfaces (API clients):**
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

Middleware и аутентификация

**CORS настройка (main.go):**
```go
r.Use(func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
    c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With, X-Backend-Url")
})
```

**JWT Authentication:**
```go
jwtConfig := middleware.DefaultJWTConfig(jwtManager, tokenStore)
r.Use(middleware.JWTAuth(jwtConfig))
```

**SKIP_AUTH для тестов:**
- Frontend: `window.__SKIP_AUTH__ = true` в init script
- Backend: `middleware.SkipAuth(middleware.DefaultSkipAuthConfig(true))` при `cfg.SkipAuth`

Интеграционные тесты

**Backend integration tests:**
```go
// *_integration_test.go с тегом //go:build integration
func (s *NoteHandlerIntegrationTestSuite) SetupSuite() {
    s.db, s.cleanup = testutil.SetupTestDB(s.T())
    s.router.POST("/notes", handler.Create)
}
```

**Frontend API integration tests (MSW):**
```typescript
// notes.test.ts
server.use(
    http.get('http://localhost:8080/api/v1/notes', () => HttpResponse.json({ notes: [mockNote] }))
);
```

Примеры промптов для этого агента

- "Добавь новый endpoint для создания тегов и обнови frontend API client"
- "Обнови DTO в Note response после добавления нового поля"
- "Настрой CORS для нового домена frontend"
- "Исправь несоответствие типов между backend и frontend"
- "Добавь интеграционный тест для нового graph endpoint"
- "Обнови документацию API после изменений в handlers"

API документация

- OpenAPI spec: `openAPI.yaml` в корне проекта
- API Errors: `docs/API_ERRORS.md` - коды ошибок и форматы
- Swagger UI: `/swagger/*any` endpoint

При работе с этим агентом

- Сосредоточьтесь на связях между backend и frontend
- Обеспечивайте согласованность API контрактов
- Следуйте существующим паттернам middleware и аутентификации
- Пишите интеграционные тесты для новых endpoints
- Обновляйте документацию при изменениях API
- Проверяйте CORS и аутентификацию при проблемах с frontend