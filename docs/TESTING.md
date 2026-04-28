# Руководство по тестированию Knowledge Graph

> **Версия:** 1.1  
> **Дата:** 28 апреля 2026  
> **Статус:** Актуально для текущего codebase  
> **Всего тестов:** ~496 (118 Go + 204 Frontend Unit + 48 Playwright + 111 BDD + 15 NLP)

---

## 📋 Содержание

1. [Обзор тестовой стратегии](#обзор-тестовой-стратегии)
2. [Backend тестирование](#backend-тестирование)
3. [Frontend тестирование](#frontend-тестирование)
4. [Интеграционное тестирование](#интеграционное-тестирование)
5. [E2E тестирование](#e2e-тестирование)
6. [Запуск тестов](#запуск-тестов)
7. [Отчёты и покрытие](#отчёты-и-покрытие)

---

## Обзор тестовой стратегии

### Уровни тестирования

```
┌─────────────────────────────────────────────────────────────┐
│                    E2E Tests                                 │
│     (Playwright + Cucumber - 48 + 111 = 159 тестов)       │
├─────────────────────────────────────────────────────────────┤
│              Integration Tests                              │
│     (Repository + API + Docker Compose)                    │
├─────────────────────────────────────────────────────────────┤
│                 Unit Tests                                  │
│    Backend (Go): 31 файлов, 118 тестовых функций           │
│    Frontend (TS): 18 файлов, 204 теста                     │
│    NLP (Python): 2 файла, ~15 тестов                     │
└─────────────────────────────────────────────────────────────┘
```

### Тестовая пирамида

| Уровень | Технологии | Покрытие | Время | Файлов |
|---------|------------|----------|-------|--------|
| **Unit** | Go testify, Vitest | 70% backend | < 10 сек | 51 |
| **Integration** | Go + Postgres, Playwright | Repositories, API | ~ 2 мин | - |
| **E2E** | Playwright + Cucumber | Полный сценарий | ~ 5 мин | 24 |

---

## Backend тестирование

### Структура тестов

**Всего: 31 файл, 118 тестовых функций**

```
backend/
├── internal/
│   ├── domain/                              # 6 файлов, 19 функций
│   │   ├── note/
│   │   │   ├── entity_test.go              # 2 теста
│   │   │   └── value_objects_test.go       # 3 теста
│   │   ├── link/
│   │   │   ├── entity_test.go              # 2 теста
│   │   │   └── value_objects_test.go       # 3 теста
│   │   └── graph/
│   │       ├── traversal_test.go           # 7 тестов
│   │       └── traversal_integration_test.go # 2 теста
│   ├── application/                         # 3 файла, 7 функций
│   │   ├── graph/
│   │   │   └── composite_loader_test.go    # 2 теста
│   │   └── recommendation/
│   │       ├── affected_notes_test.go      # 3 теста
│   │       └── refresh_service_test.go     # 2 теста
│   ├── infrastructure/                      # 14 файлов, 62 функции
│   │   ├── db/postgres/                     # 12 файлов, 44 функции
│   │   │   ├── embedding_repo_test.go      # 5 тестов
│   │   │   ├── link_repo_test.go           # 3 теста
│   │   │   ├── link_repo_unit_test.go      # 18 тестов
│   │   │   ├── link_repo_integration_test.go # 1 тест
│   │   │   ├── note_repo_test.go           # 3 теста
│   │   │   ├── note_repo_unit_test.go      # 12 тестов
│   │   │   ├── note_repo_integration_test.go # 1 тест
│   │   │   ├── recommendation_repo_test.go # 6 тестов
│   │   │   ├── tag_repo_integration_test.go # 1 тест
│   │   │   └── user_repo_integration_test.go # 1 тест
│   │   ├── nlp/
│   │   │   └── client_test.go              # 9 тестов
│   │   └── queue/                           # 2 файла, 6 функций
│   │       ├── tasks_test.go               # 3 теста
│   │       └── tasks/recommendation_test.go # 3 теста
│   └── interfaces/                          # 6 файлов, 14 функций
│       └── api/
│           ├── common/validation/
│           │   └── validators_test.go        # 9 тестов
│           ├── graphhandler/
│           │   ├── graph_handler_test.go     # 3 теста
│           │   └── graph_handler_integration_test.go # 1 тест
│           ├── linkhandler/
│           │   ├── link_handler_test.go      # 1 тест
│           │   └── link_handler_integration_test.go # 1 тест
│           ├── notehandler/
│           │   ├── note_handler_test.go      # 4 теста
│           │   └── note_handler_integration_test.go # 1 тест
│           └── taghandler/
│               └── tag_handler_integration_test.go # 1 тест
└── internal/config/
    └── config_test.go                        # 5 тестов
```

### Domain Layer Tests

#### Запуск

```bash
cd backend

# Все unit тесты
go test ./internal/domain/... -v

# Конкретный пакет
go test ./internal/domain/note -v

# С покрытием
go test ./internal/domain/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

#### Примеры тестов

**Note Entity** (`entity_test.go`):
```go
func TestNote_Create(t *testing.T) {
    note, err := note.Create("Test Title", "Test Content")
    require.NoError(t, err)
    assert.NotEmpty(t, note.ID)
    assert.Equal(t, "Test Title", note.Title.Value())
    assert.WithinDuration(t, time.Now(), note.CreatedAt, time.Second)
}

func TestNote_UpdateTitle(t *testing.T) {
    note, _ := note.Create("Old", "Content")
    err := note.UpdateTitle("New Title")
    require.NoError(t, err)
    assert.Equal(t, "New Title", note.Title.Value())
    assert.True(t, note.UpdatedAt.After(note.CreatedAt))
}
```

**Graph Traversal** (`traversal_test.go`):
```go
func TestTraversal_BFS(t *testing.T) {
    loader := newMockNeighborLoader()
    traversal := graph.NewTraversal(loader, graph.MAXStrategy)
    
    suggestions, err := traversal.GetSuggestions(context.Background(), "note-1", 3)
    
    require.NoError(t, err)
    assert.Len(t, suggestions, 3)
    assert.Equal(t, "note-2", suggestions[0].NoteID) // Highest weight
}
```

### Application Layer Tests

**Composite Loader** (`composite_loader_test.go`):
```go
func TestCompositeLoader_Combine(t *testing.T) {
    loader := graph.NewCompositeLoader(explicitLoader, embeddingLoader, 0.7, 0.3)
    
    suggestions, err := loader.LoadSuggestions(ctx, "note-1", 5)
    
    require.NoError(t, err)
    // Verify weighted combination: 0.7 * explicit + 0.3 * semantic
}
```

### Infrastructure Layer Tests

**Repository Integration** (требует PostgreSQL):
```bash
# Запуск с Docker Compose
docker-compose up -d postgres

# Интеграционные тесты
go test ./internal/infrastructure/db/postgres/... -v -tags=integration
```

### Interface Layer Tests

**HTTP Handlers**:
```bash
# Handler tests
go test ./internal/interfaces/api/... -v

# С моками репозиториев
go test ./internal/interfaces/api/notehandler -v -run TestCreateNote
```

---

## Frontend тестирование

### Структура тестов

**Unit тесты: 22 файла, ~220 тестов**
**E2E тесты: 10 файлов, 48 тестов**
**BDD тесты: 3 файла, 13 сценариев**

```
frontend/
├── src/lib/
│   ├── components/                      # 18 component tests (.spec.ts)
│   │   ├── BackButton.spec.ts              # Back button component
│   │   ├── ConfirmModal.spec.ts            # Confirmation modal
│   │   ├── CreateNoteModal.spec.ts         # Create note modal
│   │   ├── EditNoteModal.spec.ts           # Edit note modal
│   │   ├── FloatingControls.spec.ts        # Floating UI controls
│   │   ├── Graph3D.spec.ts                 # 3D graph component
│   │   ├── GraphCanvas.interactions.spec.ts # Canvas interactions
│   │   ├── GraphCanvas.links.spec.ts       # Link rendering
│   │   ├── GraphCanvas.node-types.spec.ts  # Node type rendering
│   │   ├── GraphCanvas.rendering.spec.ts   # Canvas rendering
│   │   ├── LazyGraph3D.spec.ts             # Lazy-loaded 3D
│   │   ├── LinkCreator.spec.ts             # Link creation UI
│   │   ├── NoteCard.spec.ts                # Note card component
│   │   ├── NoteEditor.spec.ts              # Note editor
│   │   ├── NoteSidePanel.spec.ts           # Side panel
│   │   ├── SearchBar.spec.ts               # Search component
│   │   ├── SmartGraph.spec.ts              # Smart graph features
│   │   └── TagSelector.spec.ts             # Tag selection UI
│   ├── api/                            # 3 API client tests (.test.ts)
│   │   ├── notes.test.ts              # Note API client (12KB)
│   │   ├── graph.test.ts              # Graph API client (6.5KB)
│   │   └── links.test.ts              # Links API client (8.3KB)
│   └── utils/                          # 1 utility test (.test.ts)
│       └── deviceCapabilities.test.ts # Device capability detection
├── tests/                               # Playwright E2E tests
│   ├── home-page.spec.ts               # Homepage tests
│   ├── notes.spec.ts                   # Note CRUD E2E
│   ├── graph-3d.spec.ts                 # 3D graph E2E
│   ├── graph-3d-modules.spec.ts         # 3D module tests
│   ├── camera-position.spec.ts          # Camera navigation
│   ├── type-filters.spec.ts             # Type filtering
│   ├── progressive-rendering.spec.ts  # Progressive rendering
│   ├── performance/
│   │   └── graph-3d-performance.spec.ts # Performance tests
│   ├── visual/
│   │   └── visual-regression.spec.ts    # Visual regression
│   └── features/                        # BDD scenarios
│       ├── graph_2d_list.feature        # 5 scenarios
│       ├── graph_interaction.feature    # 5 scenarios
│       └── graph_3d_loading.feature     # 3 scenarios
└── tests/ (корень проекта)              # Общие BDD тесты
    └── features/                        # 11 файлов, 98 сценариев
        ├── local_3d_graph.feature       # 13 scenarios
        ├── full_3d_graph.feature        # 12 scenarios
        ├── camera_navigation.feature    # 10 scenarios
        ├── type_filters.feature         # 10 scenarios
        ├── search_and_discovery.feature # 9 scenarios
        ├── celestial_body_types.feature # 8 scenarios
        ├── graph_view.feature           # 8 scenarios
        ├── link_types.feature           # 8 scenarios
        ├── note_management.feature      # 8 scenarios
        ├── graph_navigation.feature     # 6 scenarios
        └── import_export.feature        # 6 scenarios
```

**Конвенция именования:**
- `.spec.ts` — тесты компонентов (Vitest + Testing Library)
- `.test.ts` — тесты API клиентов и утилит (Vitest)
- Сторы (`stores`) тестируются через компонентные тесты, отдельных файлов нет

### E2E тесты (Playwright)

#### Запуск

```bash
cd frontend

# Установка браузеров
npx playwright install chromium

# Все тесты
npm run test

# Только UI режим
npx playwright test --ui

# Определённый файл
npx playwright test notes.spec.ts

# Отладка
npx playwright test --debug
```

#### Тестовые файлы

**Note CRUD** (`notes.spec.ts`):
```typescript
test('create note', async ({ page }) => {
  await page.goto('/');
  await page.click('[data-testid="create-note-btn"]');
  await page.fill('[data-testid="title-input"]', 'Test Note');
  await page.fill('[data-testid="content-input"]', 'Test content');
  await page.click('[data-testid="save-btn"]');
  
  await expect(page.locator('[data-testid="note-card"]')).toContainText('Test Note');
});
```

**3D Graph** (`graph-3d.spec.ts`):
```typescript
test('3D graph renders with WebGL', async ({ page }) => {
  await page.goto('/graph/3d/note-123');
  
  // Wait for canvas
  const canvas = page.locator('canvas');
  await expect(canvas).toBeVisible();
  
  // Verify WebGL context
  const webglSupported = await page.evaluate(() => {
    const canvas = document.querySelector('canvas');
    return canvas && !!canvas.getContext('webgl2');
  });
  expect(webglSupported).toBe(true);
});
```

**Progressive Rendering** (`progressive-rendering.spec.ts`):
```typescript
test('fog animation completes', async ({ page }) => {
  await page.goto('/graph/3d/note-123');
  
  // Check initial fog state
  const initialOpacity = await page.evaluate(() => 
    window.getComputedStyle(document.body).getPropertyValue('--fog-opacity')
  );
  expect(initialOpacity).toBe('0.9');
  
  // Wait for animation
  await page.waitForTimeout(2000);
  
  // Verify fog cleared
  const finalOpacity = await page.evaluate(() => 
    document.querySelector('[data-testid="stats-bar"]')?.textContent
  );
  expect(finalOpacity).toContain('Nodes: 10');
});
```

### Unit тесты (Vitest + jsdom)

```bash
# Установка
npm install -D vitest @testing-library/svelte jsdom

# Запуск
npx vitest

# С покрытием
npx vitest run --coverage
```

**Three.js Modules** (`graph-3d-modules.spec.ts`):
```typescript
import { describe, it, expect, vi } from 'vitest';
import { setupScene } from '$lib/three/core/sceneSetup';

describe('sceneSetup', () => {
  it('initializes scene with fog', () => {
    const { scene } = setupScene();
    expect(scene.fog).toBeDefined();
    expect(scene.fog.density).toBe(0.02);
  });
});
```

### Тесты сохранения связей (Link Preservation)

Тесты проверяют, что связи между заметками корректно отображаются и не теряются при различных операциях с 3D графом.

#### Сценарии тестирования

**1. Сохранение связей при переключении режимов просмотра**
```gherkin
Scenario: Links remain visible when switching from local to full graph
  Given a note "Hub Note" has 3 related notes
  When I navigate to local 3D view for "Hub Note"
  And I click the "Show all notes" toggle
  Then all existing links remain visible without flickering
  And the stats bar shows link count greater than 0
```

**2. Сохранение связей при зуме камеры**
```gherkin
Scenario: Links persist during camera zoom operations
  Given a note "Zoom Test" has 2 related notes
  When I navigate to "/graph/3d/{zoomTestId}"
  And I zoom in on the graph
  Then the links remain connected to their nodes
  And no links appear disconnected or floating
```

**3. Сохранение связей при вращении камеры**
```gherkin
Scenario: Links persist during camera rotation
  Given a note "Rotate Test" has 2 related notes
  When I navigate to "/graph/3d/{rotateTestId}"
  And I rotate the camera 90 degrees around the graph
  Then the links remain connected to their nodes
  And the links rotate with the nodes
```

**4. Проверка на дублирование связей**
```gherkin
Scenario: Links are not duplicated when switching views multiple times
  Given a note "Switch Test" has 2 related notes
  When I navigate to "/graph/3d/{switchTestId}"
  And I record the initial link count
  And I toggle "Show all notes" twice
  Then the link count matches the initial recorded count
  And no duplicate links are present in the graph
```

#### Реализация тестов

Файл: `tests/features/step_definitions/progressive-graph-steps.ts`

```typescript
// Camera zoom steps
When('I zoom in on the graph', async function(this: ITestWorld) {
  const canvas = this.page.locator('canvas').first();
  const box = await canvas.boundingBox();
  if (box) {
    await canvas.click({ position: { x: box.width / 2, y: box.height / 2 } });
    await this.page.mouse.wheel(0, -500);
    await this.page.waitForTimeout(500);
  }
});

// Link preservation verification
Then('the links remain connected to their nodes', async function(this: ITestWorld) {
  const canvas = this.page.locator('canvas').first();
  await expect(canvas).toBeVisible();
  const box = await canvas.boundingBox();
  expect(box).not.toBeNull();
});

Then('no links appear disconnected or floating', async function(this: ITestWorld) {
  const statsBar = this.page.locator('.stats-bar').first();
  await expect(statsBar).toBeVisible();
  const statsText = await statsBar.textContent();
  const linkMatch = statsText!.match(/(\d+)\s*links?/i);
  if (linkMatch) {
    expect(parseInt(linkMatch[1], 10)).toBeGreaterThan(0);
  }
});
```

#### Запуск тестов

```bash
# Запуск тестов связей
cd tests
npx cucumber-js --tags "@link-preservation"

# Запуск всех 3D тестов
npx cucumber-js features/local_3d_graph.feature features/full_3d_graph.feature
```

---

## NLP Service тестирование (Python)

### Структура

```
nlp-service/
├── tests/
│   ├── test_api.py              # ~8 тестов (FastAPI endpoints)
│   └── test_nlp_utils.py        # ~6 тестов (NLP функции)
├── app/
│   ├── main.py                  # FastAPI приложение
│   ├── nlp_utils.py             # NLP утилиты
│   └── models.py                # Pydantic модели
└── requirements.txt             # Зависимости
```

### Запуск

```bash
cd nlp-service

# Установка зависимостей
pip install -r requirements.txt

# Запуск всех тестов
pytest tests/ -v

# Запуск конкретного файла
pytest tests/test_api.py -v
pytest tests/test_nlp_utils.py -v

# С покрытием
pytest tests/ --cov=app --cov-report=html
```

### Тесты

**API Tests** (`test_api.py`):
- `TestHealthEndpoint` - Проверка health check
- `TestKeywordsEndpoint` - Тесты извлечения ключевых слов
- `TestEmbeddingsEndpoint` - Тесты генерации эмбеддингов

**NLP Utils Tests** (`test_nlp_utils.py`):
- `TestKeywordExtraction` - Извлечение ключевых слов
- `TestEmbeddingModel` - Работа с эмбеддингами

---

## Интеграционное тестирование

### Docker Compose Integration

```bash
# Полный стек для тестирования
docker-compose -f docker-compose.yml -f docker-compose.test.yml up -d

# Запуск интеграционных тестов
cd backend && go test ./... -tags=integration -v
```

### API Contract Testing

```bash
# Проверка OpenAPI спецификации
npm install -D @redocly/cli

npx @redocly/cli lint backend/openAPI.yaml
npx @redocly/cli stats backend/openAPI.yaml
```

---

## E2E тестирование (Cucumber BDD)

### Структура

```
tests/
├── features/
│   ├── graph_navigation.feature      # Навигация по графу
│   ├── graph_view.feature            # 2D/3D режимы
│   ├── full_3d_graph.feature         # Полный 3D граф (все заметки)
│   ├── local_3d_graph.feature        # Локальный 3D граф (одна заметка + связи)
│   ├── note_management.feature       # CRUD операции
│   ├── search_and_discovery.feature   # Поиск
│   └── import_export.feature          # Импорт/экспорт
│
└── features/step_definitions/
    ├── graph_steps.ts                # Шаги для графа
    ├── progressive-graph-steps.ts    # Шаги для 3D графа (fog, camera, links)
    ├── note_steps.ts                 # Шаги для заметок
    └── common_steps.ts               # Общие шаги
```

### Feature-файлы 3D графа

**`full_3d_graph.feature`** (9 сценариев):
- Переход на полный 3D граф с главной страницы
- Отображение всех заметок
- Загрузка без спиннера (progressive loading)
- Туманность при загрузке
- Центрирование камеры
- **Сохранение связей при зуме**
- **Сохранение связей при вращении**
- Корректное отображение связей с множеством узлов

**`local_3d_graph.feature`** (13 сценариев):
- Переход из деталей заметки в 3D
- Переход с главной при выделении
- Одиночная заметка с туманом
- Прогрессивная загрузка
- Переключатель "Показать все заметки"
- **Сохранение связей при переключении режимов**
- **Сохранение связей при зуме камеры**
- **Сохранение связей при вращении камеры**
- **Проверка на дублирование связей**
- **Корректность связей после прогрессивной загрузки**

### Запуск Cucumber

```bash
# Все BDD тесты
npm run test:cucumber

# С определённым тегом
CUCUMBER_TAGS="@smoke" npm run test:cucumber

# HTML отчёт
npm run test:cucumber:report
```

### Пример Feature

```gherkin
Feature: Note Management
  As a user
  I want to create, read, update and delete notes
  So that I can manage my knowledge

  @smoke
  Scenario: Create a new note
    Given I am on the main page
    When I click the create note button
    And I enter "Test Note" as the title
    And I enter "Test content" as the content
    And I click the save button
    Then I should see "Test Note" in the note list

  Scenario: Delete a note
    Given I have created a note with title "To Delete"
    When I open the note "To Delete"
    And I click the delete button
    And I confirm the deletion
    Then I should not see "To Delete" in the note list
```

---

## Запуск тестов

### Полная проверка (все уровни)

```bash
# 1. Backend unit tests
cd backend && go test ./... -v

# 2. Frontend E2E
cd frontend && npm run test

# 3. BDD Cucumber
cd tests && npm run test:cucumber

# 4. Health checks
./scripts/health-check.sh
```

### CI/CD Pipeline

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test ./... -race -coverprofile=coverage.out
      - uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '20'
      - run: npm ci
      - run: npx playwright install
      - run: npm run test
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: playwright-report
          path: playwright-report/

  e2e:
    runs-on: ubuntu-latest
    needs: [backend, frontend]
    steps:
      - uses: actions/checkout@v3
      - run: docker-compose up -d
      - run: sleep 30  # Wait for services
      - run: npm run test:cucumber
```

---

## Отчёты и покрытие

### Backend покрытие

```bash
cd backend

# Генерация отчёта
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Просмотр
open coverage.html

# Проверка порога (минимум 70%)
go-test-coverage -coverprofile=coverage.out -threshold=70
```

### Frontend покрытие

```bash
cd frontend

# E2E покрытие (через Playwright trace)
npx playwright test --trace on

# Открытие отчёта
npx playwright show-report
```

### Текущее покрытие (апрель 2026)

| Компонент | Покрытие | Тесты | Статус |
|-----------|----------|-------|--------|
| **Backend Domain** | ~85% | 19 | ✅ Отлично |
| **Backend Application** | ~75% | 7 | ✅ Хорошо |
| **Backend Infrastructure** | ~60% | 62 | ✅ Хорошо |
| **Backend Interface** | ~70% | 14 | ✅ Хорошо |
| **Frontend Unit** | ~60% | ~220 | ✅ Отлично |
| **Frontend E2E** | N/A | 48 | ✅ Отлично |
| **BDD Scenarios** | N/A | 111 | ✅ Отлично |
| **NLP Python** | ~80% | ~15 | ✅ Отлично |
| **Итого** | - | **~512** | ✅ |

### Необходимые дополнительные тесты

- [ ] **Worker Integration Tests** - Redis queue + task processing
- [ ] **Load Tests** - k6 или Artillery для API нагрузки
- [ ] **Security Tests** - OWASP ZAP сканирование
- [ ] **Contract Tests** - Pact для API контрактов

---

## Полезные команды

```bash
# Быстрый чек
make test              # Все тесты
make test-backend      # Только backend
make test-frontend     # Только frontend E2E
make test-cucumber     # Только BDD

# Отладка
make test-debug        # С отладочной информацией
make test-watch        # Watch mode

# Отчёты
make coverage          # Покрытие
make report            # HTML отчёт
```
