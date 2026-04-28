# Статус тестов и покрытие

> **Обновлено:** 28 апреля 2026
> **Всего тестов:** ~496 (118 Go + 204 Frontend Unit + 48 Playwright + 111 BDD + 15 NLP)

## Backend (Go)

### Юнит-тесты — 31 файл, 118 тестовых функций

| Пакет | Тест-файл | Кол-во функций | Статус |
|-------|-----------|----------------|--------|
| `domain/graph` | `traversal_test.go` | 7 | ✅ |
| `domain/graph` | `traversal_integration_test.go` | 2 | ✅ |
| `domain/link` | `entity_test.go` | 2 | ✅ |
| `domain/link` | `value_objects_test.go` | 3 | ✅ |
| `domain/note` | `entity_test.go` | 2 | ✅ |
| `domain/note` | `value_objects_test.go` | 3 | ✅ |
| `application/graph` | `composite_loader_test.go` | 2 | ✅ |
| `application/recommendation` | `affected_notes_test.go` | 3 | ✅ |
| `application/recommendation` | `refresh_service_test.go` | 2 | ✅ |
| `infrastructure/db/postgres` | `embedding_repo_test.go` | 5 | ✅ |
| `infrastructure/db/postgres` | `link_repo_test.go` | 3 | ✅ |
| `infrastructure/db/postgres` | `link_repo_unit_test.go` | 18 | ✅ |
| `infrastructure/db/postgres` | `link_repo_integration_test.go` | 1 | ✅ |
| `infrastructure/db/postgres` | `note_repo_test.go` | 3 | ✅ |
| `infrastructure/db/postgres` | `note_repo_unit_test.go` | 12 | ✅ |
| `infrastructure/db/postgres` | `note_repo_integration_test.go` | 1 | ✅ |
| `infrastructure/db/postgres` | `recommendation_repo_test.go` | 6 | ✅ |
| `infrastructure/db/postgres` | `tag_repo_integration_test.go` | 1 | ✅ |
| `infrastructure/db/postgres` | `user_repo_integration_test.go` | 1 | ✅ |
| `infrastructure/nlp` | `client_test.go` | 9 | ✅ |
| `infrastructure/queue` | `tasks_test.go` | 3 | ✅ |
| `infrastructure/queue/tasks` | `recommendation_test.go` | 3 | ✅ |
| `interfaces/api/common/validation` | `validators_test.go` | 9 | ✅ |
| `interfaces/api/graphhandler` | `graph_handler_test.go` | 3 | ✅ |
| `interfaces/api/graphhandler` | `graph_handler_integration_test.go` | 1 | ✅ |
| `interfaces/api/linkhandler` | `link_handler_test.go` | 1 | ✅ |
| `interfaces/api/linkhandler` | `link_handler_integration_test.go` | 1 | ✅ |
| `interfaces/api/notehandler` | `note_handler_test.go` | 4 | ✅ |
| `interfaces/api/notehandler` | `note_handler_integration_test.go` | 1 | ✅ |
| `interfaces/api/taghandler` | `tag_handler_integration_test.go` | 1 | ✅ |
| `internal/config` | `config_test.go` | 5 | ✅ |

### CI статус
- **go test**: ✅ Запускается с `-race` и `-coverprofile`
- **golangci-lint**: ✅ Конфигурирован в `.golangci.yml`
- **Покрытие**: Загружается в Codecov (`codecov/codecov-action@v4`)

### Конфигурация `.golangci.yml`
- Отключен `typecheck` для тестов (ложные ошибки на testify mocks)
- Исключены файлы `*_test.go` из проверки `typecheck`

---

## Frontend (Svelte)

### Юнит-тесты — 22 файла, ~220 тестов

#### Component Tests (.spec.ts) — 18 файлов

| Компонент | Тест-файл | Статус |
|-----------|-----------|--------|
| `BackButton` | `BackButton.spec.ts` | ✅ |
| `ConfirmModal` | `ConfirmModal.spec.ts` | ✅ |
| `CreateNoteModal` | `CreateNoteModal.spec.ts` | ✅ |
| `EditNoteModal` | `EditNoteModal.spec.ts` | ✅ |
| `FloatingControls` | `FloatingControls.spec.ts` | ✅ |
| `Graph3D` | `Graph3D.spec.ts` | ✅ |
| `GraphCanvas/interactions` | `GraphCanvas.interactions.spec.ts` | ✅ |
| `GraphCanvas/links` | `GraphCanvas.links.spec.ts` | ✅ |
| `GraphCanvas/node-types` | `GraphCanvas.node-types.spec.ts` | ✅ |
| `GraphCanvas/rendering` | `GraphCanvas.rendering.spec.ts` | ✅ |
| `LazyGraph3D` | `LazyGraph3D.spec.ts` | ✅ |
| `LinkCreator` | `LinkCreator.spec.ts` | ✅ |
| `NoteCard` | `NoteCard.spec.ts` | ✅ |
| `NoteEditor` | `NoteEditor.spec.ts` | ✅ |
| `NoteSidePanel` | `NoteSidePanel.spec.ts` | ✅ |
| `SearchBar` | `SearchBar.spec.ts` | ✅ |
| `SmartGraph` | `SmartGraph.spec.ts` | ✅ |
| `TagSelector` | `TagSelector.spec.ts` | ✅ |

#### API & Utils Tests (.test.ts) — 4 файла

| Модуль | Тест-файл | Статус |
|--------|-----------|--------|
| `api/notes` | `notes.test.ts` | ✅ |
| `api/graph` | `graph.test.ts` | ✅ |
| `api/links` | `links.test.ts` | ✅ |
| `utils/deviceCapabilities` | `deviceCapabilities.test.ts` | ✅ |

**Примечание:** Сторы (`stores`) тестируются через компонентные тесты, отдельных файлов нет — логика живёт в компонентах.

### Playwright E2E тесты — 10 файлов, 48 тестов

| Файл | Содержимое |
|------|-----------|
| `camera-position.spec.ts` | Тесты навигации камеры |
| `graph-3d-modules.spec.ts` | Модульные тесты 3D графа |
| `graph-3d.spec.ts` | Интеграционные тесты 3D графа |
| `home-page.spec.ts` | Тесты домашней страницы |
| `notes.spec.ts` | CRUD тесты заметок |
| `progressive-rendering.spec.ts` | Тесты прогрессивного рендеринга |
| `type-filters.spec.ts` | Тесты фильтров по типам |
| `performance/graph-3d-performance.spec.ts` | Performance тесты |
| `visual/visual-regression.spec.ts` | Визуальная регрессия |

### BDD тесты (Cucumber)

**Корневые тесты (tests/features/):** 11 файлов, 98 сценариев
- `local_3d_graph.feature` (13 сценариев)
- `full_3d_graph.feature` (12)
- `camera_navigation.feature` (10)
- `type_filters.feature` (10)
- `search_and_discovery.feature` (9)
- `celestial_body_types.feature` (8)
- `graph_view.feature` (8)
- `link_types.feature` (8)
- `note_management.feature` (8)
- `graph_navigation.feature` (6)
- `import_export.feature` (6)

**Frontend тесты (frontend/tests/features/):** 3 файла, 13 сценариев
- `graph_2d_list.feature` (5)
- `graph_interaction.feature` (5)
- `graph_3d_loading.feature` (3)

**Step definitions:** 6 файлов в `tests/features/step_definitions/`

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

### CI статус
- **test:unit**: ✅ 204 теста проходят
- **test:bdd**: ⚠️ Требуются работающий backend и БД

---

## NLP Service (Python)

### Тесты — 2 файла, ~15 тестов

| Файл | Содержимое |
|------|-----------|
| `tests/test_api.py` | Тесты FastAPI endpoints (~8 тестов) |
| `tests/test_nlp_utils.py` | Тесты NLP функций (~6 тестов) |

### Запуск
```bash
cd nlp-service
pytest tests/ -v
```

---

## Сводная таблица

| Категория | Файлов | Тестов | Статус |
|-----------|--------|--------|--------|
| **Go Unit** | 31 | 118 | ✅ |
| **Frontend Unit** | 18 | 204 | ✅ |
| **Playwright E2E** | 10 | 48 | ⏭️ Требует БД |
| **BDD Scenarios** | 14 | 111 | ⏭️ Требует БД |
| **NLP Python** | 2 | ~15 | ✅ |
| **Итого** | **75** | **~496** | ✅ |

## Рекомендации

### Для запуска E2E/BDD тестов
```bash
# 1. Запустить инфраструктуру
docker-compose up -d postgres redis

# 2. Запустить backend
cd backend && go run ./cmd/server/main.go

# 3. Запустить frontend (dev mode)
cd frontend && npm run dev

# 4. Запустить E2E тесты
cd frontend && npm run test
```

### Проверка локально
```bash
# Frontend Unit
cd frontend && npm run test:unit

# Frontend с покрытием
cd frontend && npm run test:coverage

# Backend
cd backend && go test -race -coverprofile=coverage.out ./...

# NLP
cd nlp-service && pytest tests/
```
