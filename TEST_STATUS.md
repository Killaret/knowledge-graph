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
- **golangci-lint**: ✅ Конфигурирован в `.golangci.yml`
- **Покрытие**: Загружается в Codecov (`codecov/codecov-action@v4`)

### Конфигурация `.golangci.yml`
- Отключен `typecheck` для тестов (ложные ошибки на testify mocks)
- Исключены файлы `*_test.go` из проверки `typecheck`

---

## Frontend (Svelte)

### Юнит-тесты — 22 файла, 204 теста

| Компонент/Модуль | Тест-файл | Статус |
|------------------|-----------|--------|
| `BackButton` | `BackButton.spec.ts` | ✅ |
| `ConfirmModal` | `ConfirmModal.spec.ts` | ✅ |
| `CreateNoteModal` | `CreateNoteModal.spec.ts` | ✅ |
| `EditNoteModal` | `EditNoteModal.spec.ts` | ✅ |
| `FloatingControls` | `FloatingControls.spec.ts` | ✅ |
| `Graph3D` | `Graph3D.spec.ts` | ✅ |
| `GraphCanvas` | `GraphCanvas.svelte.spec.ts` | ✅ |
| `GraphCanvas/links` | `GraphCanvas.links.spec.ts` | ✅ |
| `GraphCanvas/node-types` | `GraphCanvas.node-types.spec.ts` | ✅ |
| `GraphCanvas/rendering` | `GraphCanvas.rendering.spec.ts` | ✅ |
| `GraphCanvas/simulation` | `GraphCanvas.simulation.spec.ts` | ✅ |
| `GraphCanvas/zoom-pan` | `GraphCanvas.zoom-pan.spec.ts` | ✅ |
| `LazyGraph3D` | `LazyGraph3D.spec.ts` | ✅ |
| `LinkCreator` | `LinkCreator.spec.ts` | ✅ |
| `NoteCard` | `NoteCard.spec.ts` | ✅ |
| `NoteSidePanel` | `NoteSidePanel.spec.ts` | ✅ |
| `SearchBar` | `SearchBar.spec.ts` | ✅ |
| `SmartGraph` | `SmartGraph.spec.ts` | ✅ |
| `api/notes` | `notes.spec.ts` | ✅ |
| `api/graph` | `graph.spec.ts` | ✅ |
| `stores/notes` | `notes.spec.ts` | ✅ |
| `utils/deviceCapabilities` | `deviceCapabilities.spec.ts` | ✅ |

### Покрытие (последний запуск)
- **Настройка**: ✅ Добавлено в `vitest.config.ts`
- **Провайдер**: v8
- **Репортеры**: text, json, html
- **Lines**: 59.86%
- **Functions**: 80.14%
- **Branches**: ~45%
- **Statements**: ~58%
- **Пороги**: lines 50%, functions 50%, branches 40%, statements 50%
- **Скрипт**: `npm run test:coverage`

### BDD тесты (Cucumber + Playwright)
- 13 сценариев в `tests/features/`
- Статус: частично требуют обновления

### CI статус
- **test:unit**: ✅ 204 теста проходят
- **test:bdd**: ⚠️ Требуют обновления

---

## NLP Service (Python)

### Тесты
- **pytest**: ✅ Запускается в `test-nlp` job
- **Кол-во тестов**: Неизвестно (нужно проверить)

---

## Рекомендации

### Важно (улучшение качества)
1. Поднять порог покрытия lines до 70% для frontend
2. Добавить интеграционные тесты для backend API
3. Обновить BDD тесты для актуального UI
4. Настроить покрытие для BDD тестов

### Проверка локально
```bash
# Frontend
cd frontend && npm run test:unit

# Frontend с покрытием
cd frontend && npm run test:coverage

# Backend
cd backend && go test -race -coverprofile=coverage.out ./...

# Backend покрытие в HTML
cd backend && go tool cover -html=coverage.out
```
