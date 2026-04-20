# Статус тестов и покрытие

## Backend (Go)

### Юнит-тесты — 18 файлов, ~124 теста

| Пакет | Тест-файл | Кол-во тестов | Статус |
|-------|-----------|---------------|--------|
| `domain/graph` | `traversal_test.go` | 23 | ✅ |
| `domain/graph` | `traversal_integration_test.go` | 5 | ✅ |
| `domain/link` | `entity_test.go` | 2 | ✅ |
| `domain/link` | `value_objects_test.go` | 5 | ✅ |
| `domain/note` | `entity_test.go` | 2 | ✅ |
| `domain/note` | `value_objects_test.go` | 5 | ✅ |
| `application/graph` | `composite_loader_test.go` | 9 | ✅ |
| `application/recommendation` | `affected_notes_test.go` | 11 | ✅ |
| `application/recommendation` | `refresh_service_test.go` | 6 | ✅ |
| `infrastructure/db/postgres` | `embedding_repo_test.go` | 5 | ✅ |
| `infrastructure/db/postgres` | `link_repo_test.go` | 3 | ✅ |
| `infrastructure/db/postgres` | `note_repo_test.go` | 3 | ✅ |
| `infrastructure/db/postgres` | `recommendation_repo_test.go` | 16 | ✅ |
| `infrastructure/queue` | `tasks_test.go` | 5 | ✅ |
| `infrastructure/queue/tasks` | `recommendation_test.go` | 8 | ✅ |
| `interfaces/api/graphhandler` | `graph_handler_test.go` | 11 | ✅ |
| `interfaces/api/linkhandler` | `link_handler_test.go` | 1 | ✅ |
| `interfaces/api/notehandler` | `note_handler_test.go` | 4 | ✅ |

### CI статус
- **go test**: ✅ Запускается с `-race` и `-coverprofile`
- **golangci-lint**: ⚠️ Падает на `typecheck` — ложные ошибки на testify mock-объектах
- **Покрытие**: Загружается в Codecov (`codecov/codecov-action@v4`)

### Известные проблемы
1. `golangci-lint` v1.56.0 не корректно обрабатывает testify mock-методы (`m.Called`, `m.On`)
2. Нужно добавить `.golangci.yml` с отключением `typecheck` для тестов или перейти на `go vet`

---

## Frontend (Svelte)

### Юнит-тесты — 3 файла

| Компонент | Тест-файл | Статус |
|-----------|-----------|--------|
| `BackButton` | `BackButton.spec.ts` | ⚠️ Нужно проверить |
| `ConfirmModal` | `ConfirmModal.spec.ts` | ⚠️ Нужно проверить |
| `NoteCard` | `NoteCard.spec.ts` | ⚠️ Нужно проверить |

### Покрытие
- **Настройка**: ✅ Добавлено в `vitest.config.ts`
- **Провайдер**: v8
- **Репортеры**: text, json, html
- **Пороги**: lines 50%, functions 50%, branches 40%, statements 50%
- **Скрипт**: `npm run test:coverage`

### BDD тесты (Cucumber + Playwright) — 13 сценариев

| Feature | Сценарии | Статус | Проблема |
|---------|----------|--------|----------|
| `graph_2d_list.feature` | Toggle graph/list views | ❌ FAIL | `.note-card` не найдены |
| `graph_2d_list.feature` | Filter by type | ❌ FAIL | Таймаут на переходе в list view |
| `graph_2d_list.feature` | Search in list | ❌ FAIL | Таймаут на переходе в list view |
| `graph_3d_loading.feature` | Overlay disappears | ❌ FAIL | 500 на создании link |
| `graph_3d_loading.feature` | Fog clears | ❌ FAIL | 500 на создании link |
| `graph_3d_loading.feature` | Early interaction | ❌ FAIL | 500 на создании link |
| `note_management.feature` | — | ✅ PASS | — |
| `note_search.feature` | — | ✅ PASS | — |

### CI статус
- **test:unit**: ✅ Запускается
- **test:bdd**: ❌ 6/13 сценариев падают
- **test (playwright)**: ⚠️ Не запускается в ci-ai-agents.yml

---

## NLP Service (Python)

### Тесты
- ** pytest**: ✅ Запускается в `test-nlp` job
- **Кол-во тестов**: Неизвестно (нужно проверить)

---

## Рекомендации

### Срочно (блокирует CI)
1. ~~Исправить создание links (500 ошибка)~~ — ✅ Исправлено (колонки source_id/target_id)
2. ~~Исправить list view toggle~~ — ✅ Исправлено (синхронизация currentView)
3. Исправить golangci-lint — добавить `.golangci.yml` с исключением тестов

### Важно (улучшение качества)
1. Добавить больше unit-тестов для frontend (сейчас только 3 компонента)
2. Добавить интеграционные тесты для backend API
3. Поднять порог покрытия до 70% для backend
4. Настроить покрытие для BDD тестов

### Необходимые действия
```bash
# Frontend — установить зависимость покрытия
cd frontend && npm install

# Backend — добавить .golangci.yml
cd backend && touch .golangci.yml

# Проверить покрытие локально
cd frontend && npm run test:coverage
cd backend && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```
