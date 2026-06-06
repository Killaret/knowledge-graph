# Правила по умолчанию для всех ИИ-агентов

## 🎯 Цель

Эти правила гарантируют, что **все ИИ-агенты** (включая Koda и других) используют лучшие практики, инструменты и паттерны для максимальной эффективности и экономии ресурсов.

---

## 📋 Обязательные правила

### 1. Всегда использовать существующие агенты

**Правило:** При получении задачи сначала определить, какой агент специализируется на этой области, и использовать его навыки.

**Пример:**
```
Запрос: "Добавь новый API endpoint для заметок"

Правильно:
1. → Backend Go Agent → Использовать backend-go-tools.md
2. → Integration Agent → Создать TypeScript типы
3. → Testing Agent → Написать тесты
4. → Documentation Agent → Обновить API docs

Неправильно:
- Писать код без использования агентов
- Игнорировать существующие паттерны
```

---

### 2. Использовать инструменты по умолчанию

**Backend (Go):**
```bash
# Всегда использовать backend-go-tools.md
- REST API: gin/echo фреймворки
- БД: gorm/mongo-go-driver с repository pattern
- Кэш: go-redis с cache-aside pattern
- Auth: JWT middleware
- Тесты: testify + mockery
- Метрики: Prometheus
```

**Frontend (Svelte):**
```typescript
// Всегда использовать frontend-svelte-tools.md
- Компоненты: Svelte 5 с $state, $props
- API: ky client с typed responses
- State: Svelte stores
- Тесты: Testing Library + Vitest
- E2E: Playwright
```

**Integration:**
```typescript
// Всегда использовать integration-tools.md
- Генерация типов: openapi-typescript
- gRPC: protobuf-ts
- Contract testing: Pact
- Retry logic: exponential backoff
- Error tracking: Sentry
```

**Infrastructure:**
```yaml
# Всегда использовать infrastructure-tools.md
- Docker: multi-stage builds
- K8s: HPA + health checks
- CI/CD: GitHub Actions
- Monitoring: Prometheus + Grafana
- Backup: автоматические скрипты
```

---

### 3. Приоритеты агентов

**Высокий приоритет (🟢):**
- Backend Go Agent
- Frontend Svelte Agent
- Integration Agent
- Infrastructure Agent

**Средний приоритет (🟡):**
- DevOps Agent
- Performance Agent
- Security Agent

**Низкий приоритет (🔵):**
- Documentation Agent

**Правило:** Критичные задачи (API, UI, безопасность) всегда выполняются агентами с высоким приоритетом.

---

### 4. Шаблоны кода

#### Backend Go - Обязательный паттерн
```go
// 1. Всегда использовать context
func (s *Service) CreateNote(ctx context.Context, req Request) (*Note, error) {
    // 2. Валидация на входе
    if err := req.Validate(); err != nil {
        return nil, wrapError(err, "VALIDATION_ERROR", "Invalid request")
    }
    
    // 3. Transaction для БД
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, wrapError(err, "DB_ERROR", "Failed to begin transaction")
    }
    defer tx.Rollback()
    
    // 4. Логирование ошибок
    logger.Error("Failed to create note", "error", err, "note_id", id)
    
    // 5. Возвращать типизированные ошибки
    return nil, &AppError{Code: "NOT_FOUND", Message: "Note not found"}
}
```

#### Frontend Svelte - Обязательный паттерн
```svelte
<script lang="ts">
  // 1. Использовать $state и $props
  let { note }: { note: Note } = $props();
  let isEditing = $state(false);
  
  // 2. Typed API calls
  const save = async () => {
    loading = true;
    try {
      await notesApi.update(note.id, note);
    } catch (error) {
      errorStore.set(error instanceof Error ? error.message : 'Unknown error');
    } finally {
      loading = false;
    }
  };
</script>
```

---

### 5. Тестирование - Обязательно

**Backend:**
```bash
# Все изменения должны включать тесты
go test -race -cover ./...
# Минимальное покрытие: 60%
```

**Frontend:**
```bash
# Все компоненты должны иметь тесты
npm run test:unit
# Минимальное покрытие: 60%
```

**Integration:**
```bash
# Contract tests перед деплоем
go test -v ./tests/contract/...
npm run test:contract
```

---

### 6. Безопасность

**Никогда:**
- ❌ Выводить секреты, токены, ключи
- ❌ Хранить секреты в коде
- ❌ Игнорировать security scanning

**Всегда:**
- ✅ Использовать environment variables
- ✅ Проверять зависимости на уязвимости
- ✅ Валидировать все входные данные
- ✅ Использовать prepared statements (SQL injection)
- ✅ Санитизировать пользовательский ввод

---

### 7. Производительность

**Backend:**
```go
// Кэширование обязательное
graph, err := s.cache.Get(ctx, cacheKey)
if err == nil {
    return graph // Cache hit
}
// Fetch from DB and cache
```

**Frontend:**
```typescript
// Lazy loading для тяжелых компонентов
const HeavyComponent = lazy(() => import('./HeavyComponent.svelte'));

// Memoization для вычислений
const sorted = $derived([...items].sort(...));
```

**Infrastructure:**
```yaml
# HPA обязательный
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
spec:
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        averageUtilization: 70
```

---

### 8. Документация

**Обязательные обновления:**
- README.md при изменении архитектуры
- API docs при изменении endpoints
- CHANGELOG при релизе
- Инструкции для новых фич

---

## 🔄 Автоматизация

### 1. Pre-commit hooks

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Backend
go test -race -cover ./... || exit 1
golangci-lint run ./... || exit 1

# Frontend
npm run test:unit || exit 1
npm run lint || exit 1

# Security
trivy fs . || exit 1
```

### 2. CI/CD Pipeline

```yaml
# Всегда включать:
- Test & coverage
- Security scanning
- Build & push
- Deploy with health checks
- Rollback if failed
```

---

## 📊 Метрики успеха

| Метрика | Цель | Агент |
|---------|------|-------|
| Backend coverage | > 60% | Backend Go |
| Frontend coverage | > 60% | Frontend Svelte |
| API p95 latency | < 500ms | Performance |
| Uptime | > 99.9% | Infrastructure |
| Deployment success | > 98% | DevOps |
| Security vulnerabilities | 0 critical | Security |

---

## ⚡ Экономия ресурсов

### 1. Минимизация API вызовов
```typescript
// ✅ GOOD: Batch requests
const [notes, users, graphs] = await Promise.all([
  notesApi.getAll(),
  usersApi.getAll(),
  graphApi.getFullGraph()
]);

// ❌ BAD: Sequential requests
const notes = await notesApi.getAll();
const users = await usersApi.getAll();
const graphs = await graphApi.getFullGraph();
```

### 2. Кэширование
```go
// ✅ GOOD: Cache first
data, err := cache.Get(key)
if err == nil {
    return data // Save DB query
}

// ❌ BAD: Always hit DB
data, err := db.Query(...)
```

### 3. Lazy Loading
```typescript
// ✅ GOOD: Load on demand
const Modal = lazy(() => import('./Modal.svelte'));

// ❌ BAD: Eager loading
import Modal from './Modal.svelte';
```

### 4. Connection Pooling
```go
// ✅ GOOD: Reuse connections
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)

// ❌ BAD: New connection per request
db, _ := sql.Open(...)
```

---

## 🎯 Чеклист перед завершением задачи

- [ ] Использован правильный агент
- [ ] Применены инструменты из соответствующего .md файла
- [ ] Написаны тесты
- [ ] Проверена безопасность
- [ ] Обновлена документация
- [ ] Метрики в норме
- [ ] Код соответствует паттернам проекта

---

**Эти правила обязательны для всех ИИ-агентов в проекте!**

**Версия:** 1.0  
**Дата:** 2026-05-30