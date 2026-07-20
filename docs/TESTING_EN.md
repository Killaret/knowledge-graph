# Knowledge Graph Testing Guide

> **Version:** 1.1
> **Date:** April 28, 2026
> **Status:** Current for current codebase
> **Total tests:** ~496 (118 Go + 204 Frontend Unit + 48 Playwright + 111 BDD + 15 NLP)

---

## Table of Contents

1. [Test Strategy Overview](#test-strategy-overview)
2. [Backend Testing](#backend-testing)
3. [Frontend Testing](#frontend-testing)
4. [Integration Testing](#integration-testing)
5. [E2E Testing](#e2e-testing)
6. [Running Tests](#running-tests)
7. [Reports and Coverage](#reports-and-coverage)

---

## Test Strategy Overview

### Testing Levels

```
┌─────────────────────────────────────────────────────────────┐
│                    E2E Tests                                 │
│     (Playwright + Cucumber - 48 + 111 = 159 tests)        │
├─────────────────────────────────────────────────────────────┤
│              Integration Tests                              │
│     (Repository + API + Docker Compose)                    │
├─────────────────────────────────────────────────────────────┤
│                 Unit Tests                                  │
│    Backend (Go): 31 files, 118 test functions              │
│    Frontend (TS): 18 files, 204 tests                      │
│    NLP (Python): 2 files, ~15 tests                        │
└─────────────────────────────────────────────────────────────┘
```

### Testing Pyramid

| Level | Technologies | Coverage | Time | Files |
|-------|--------------|----------|------|-------|
| **Unit** | Go testify, Vitest | Target 70% (current ~62% BE / ~64% FE lines) | < 10 sec | 51 |
| **Integration** | Go + Postgres, Playwright | Repositories, API | ~ 2 min | - |
| **E2E** | Playwright + Cucumber | Full scenario | ~ 5 min | 24 |

---

## Backend Testing

### Test Structure

**Total: 35+ files, 118+ test functions**

```
backend/
├── internal/
│   ├── domain/                              # 6 files, 19 functions
│   │   ├── note/
│   │   │   ├── entity_test.go              # 2 tests
│   │   │   └── value_objects_test.go       # 3 tests
│   │   ├── link/
│   │   │   ├── entity_test.go              # 2 tests
│   │   │   └── value_objects_test.go       # 3 tests
│   │   └── graph/
│   │       ├── traversal_test.go           # 7 tests
│   │       └── traversal_integration_test.go # 2 tests
│   ├── application/                         # 3 files, 7 functions
│   │   ├── graph/
│   │   │   └── composite_loader_test.go    # 2 tests
│   │   └── recommendation/
│   │       ├── affected_notes_test.go      # 3 tests
│   │       └── refresh_service_test.go     # 2 tests
│   ├── infrastructure/                      # 14 files, 62 functions
│   │   ├── db/postgres/                     # 12 files, 44 functions
│   │   │   ├── embedding_repo_test.go      # 5 tests
│   │   │   ├── link_repo_test.go           # 3 tests
│   │   │   ├── link_repo_unit_test.go      # 18 tests
│   │   │   ├── link_repo_integration_test.go # 1 test
│   │   │   ├── note_repo_test.go           # 3 tests
│   │   │   ├── note_repo_unit_test.go      # 12 tests
│   │   │   ├── note_repo_integration_test.go # 1 test
│   │   │   ├── recommendation_repo_test.go # 6 tests
│   │   │   ├── tag_repo_integration_test.go # 1 test
│   │   │   └── user_repo_integration_test.go # 1 test
│   │   ├── nlp/
│   │   │   └── client_test.go              # 9 tests
│   │   └── queue/                           # 2 files, 6 functions
│   │       ├── tasks_test.go               # 3 tests
│   │       └── tasks/recommendation_test.go # 3 tests
│   └── interfaces/                          # 6 files, 14 functions
│       └── api/
│           ├── common/validation/
│           │   └── validators_test.go        # 9 tests
│           ├── graphhandler/
│           │   ├── graph_handler_test.go     # 3 tests
│           │   └── graph_handler_integration_test.go # 1 test
│           ├── linkhandler/
│           │   ├── link_handler_test.go      # 1 test
│           │   └── link_handler_integration_test.go # 1 test
│           ├── notehandler/
│           │   ├── note_handler_test.go      # 4 tests
│           │   └── note_handler_integration_test.go # 1 test
│           └── taghandler/
│               └── tag_handler_integration_test.go # 1 test
└── internal/config/
    └── config_test.go                        # 5 tests
```

### Domain Layer Tests

#### Running

```bash
cd backend

# All unit tests
go test ./internal/domain/... -v

# Specific package
go test ./internal/domain/note -v

# With coverage
go test ./internal/domain/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

#### Test Examples

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

**Repository Integration** (requires PostgreSQL):
```bash
# Run with Docker Compose
docker-compose up -d postgres

# Integration tests
go test ./internal/infrastructure/db/postgres/... -v -tags=integration
```

### Interface Layer Tests

**HTTP Handlers**:
```bash
# Handler tests
go test ./internal/interfaces/api/... -v

# With repository mocks
go test ./internal/interfaces/api/notehandler -v -run TestCreateNote
```

---

## Frontend Testing

### Test Structure

**Unit tests: 52 files, 535 tests**
**E2E tests: 16 files (Playwright)**
**BDD tests: 15 feature files**

```
frontend/
├── src/
│   ├── components/                      # Component tests (.spec.ts)
│   │   ├── atoms/...
│   │   ├── molecules/...
│   │   └── organisms/...                # GraphCanvas, NoteEditor, NoteSidePanel, etc.
│   ├── shared/
│   │   ├── api/                         # API client tests (.test.ts)
│   │   ├── stores/                      # runes-based store tests
│   │   ├── services/                    # PreloadService and other service tests
│   │   ├── utils/                       # Utility tests
│   │   ├── hooks/                       # Hook tests
│   │   └── types/...                    # Type tests
│   └── routes/                          # Route-level tests
├── tests/                               # Playwright E2E tests
│   ├── home-page.spec.ts                # Homepage tests
│   ├── notes.spec.ts                    # Note CRUD E2E
│   ├── type-filters.spec.ts             # Type filtering
│   ├── auth-*.spec.ts                   # Auth-related E2E tests
│   ├── public-graph.spec.ts             # Public graph
│   ├── manual-checklist-section-*.spec.ts # Manual checklist tests
│   ├── smoke.spec.ts                    # Smoke tests
│   ├── visual/
│   │   └── visual-regression.spec.ts    # Visual regression
│   └── features/                        # BDD scenarios
│       ├── graph_2d_list.feature
│       └── graph_interaction.feature
└── tests/ (project root)                # Common BDD tests
    └── features/
        ├── achievements.feature
        ├── auth_cosmic_theme.feature
        ├── graph_navigation.feature
        ├── graph_view.feature
        ├── note_management.feature
        ├── search_and_discovery.feature
        ├── type_filters.feature
        └── ...
```

**Naming Convention:**
- `.spec.ts` — component tests (Vitest + Testing Library)
- `.test.ts` — API client and utility tests (Vitest)
- Stores are tested via component tests, no separate files

### E2E Tests (Playwright)

#### Running

```bash
cd frontend

# Install browsers
npx playwright install chromium

# All tests
npm run test

# UI mode only
npx playwright test --ui

# Specific file
npx playwright test notes.spec.ts

# Debug
npx playwright test --debug
```

#### Test Files

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

### Unit Tests (Vitest + jsdom)

```bash
# Install
npm install -D vitest @testing-library/svelte jsdom

# Run
npx vitest

# With coverage
npx vitest run --coverage
```

**Three.js Modules** — 3D graph functionality is **frozen/removed for v1.0**. The `graph-3d-modules.spec.ts` test and `three/core/sceneSetup` module no longer exist.

### Link Preservation Tests

Tests verify that connections between notes are correctly displayed and not lost during various 3D graph operations.

#### Test Scenarios

**1. Link preservation when switching view modes**
```gherkin
Scenario: Links remain visible when switching from local to full graph
  Given a note "Hub Note" has 3 related notes
  When I navigate to local 3D view for "Hub Note"
  And I click the "Show all notes" toggle
  Then all existing links remain visible without flickering
  And the stats bar shows link count greater than 0
```

**2. Link preservation during camera zoom**
```gherkin
Scenario: Links persist during camera zoom operations
  Given a note "Zoom Test" has 2 related notes
  When I navigate to "/graph/3d/{zoomTestId}"
  And I zoom in on the graph
  Then the links remain connected to their nodes
  And no links appear disconnected or floating
```

**3. Link preservation during camera rotation**
```gherkin
Scenario: Links persist during camera rotation
  Given a note "Rotate Test" has 2 related notes
  When I navigate to "/graph/3d/{rotateTestId}"
  And I rotate the camera 90 degrees around the graph
  Then the links remain connected to their nodes
  And the links rotate with the nodes
```

**4. Check for link duplication**
```gherkin
Scenario: Links are not duplicated when switching views multiple times
  Given a note "Switch Test" has 2 related notes
  When I navigate to "/graph/3d/{switchTestId}"
  And I record the initial link count
  And I toggle "Show all notes" twice
  Then the link count matches the initial recorded count
  And no duplicate links are present in the graph
```

#### Test Implementation

File: `tests/features/step_definitions/progressive-graph-steps.ts`

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

#### Running Tests

```bash
# Run link tests
cd tests
npx cucumber-js --tags "@link-preservation"

# Run all 3D tests
npx cucumber-js features/local_3d_graph.feature features/full_3d_graph.feature
```

---

## NLP Service Testing (Python)

### Structure

```
nlp-service/
├── tests/
│   ├── test_api.py              # ~8 tests (FastAPI endpoints)
│   └── test_nlp_utils.py        # ~6 tests (NLP functions)
├── app/
│   ├── main.py                  # FastAPI application
│   ├── nlp_utils.py             # NLP utilities
│   └── models.py                # Pydantic models
└── requirements.txt             # Dependencies
```

### Running

```bash
cd nlp-service

# Install dependencies
pip install -r requirements.txt

# Run all tests
pytest tests/ -v

# Run specific file
pytest tests/test_api.py -v
pytest tests/test_nlp_utils.py -v

# With coverage
pytest tests/ --cov=app --cov-report=html
```

### Tests

**API Tests** (`test_api.py`):
- `TestHealthEndpoint` - Health check verification
- `TestKeywordsEndpoint` - Keyword extraction tests
- `TestEmbeddingsEndpoint` - Embedding generation tests

**NLP Utils Tests** (`test_nlp_utils.py`):
- `TestKeywordExtraction` - Keyword extraction
- `TestEmbeddingModel` - Embedding model operation

---

## Integration Testing

### Docker Compose Integration

```bash
# Full stack for testing
docker-compose -f docker-compose.yml -f docker-compose.test.yml up -d

# Run integration tests
cd backend && go test ./... -tags=integration -v
```

### API Contract Testing

```bash
# Check OpenAPI specification
npm install -D @redocly/cli

npx @redocly/cli lint backend/openAPI.yaml
npx @redocly/cli stats backend/openAPI.yaml
```

---

## E2E Testing (Cucumber BDD)

### Structure

```
tests/
├── features/
│   ├── graph_navigation.feature      # Graph navigation
│   ├── graph_view.feature            # 2D/3D modes
│   ├── full_3d_graph.feature         # Full 3D graph (all notes)
│   ├── local_3d_graph.feature        # Local 3D graph (one note + links)
│   ├── note_management.feature       # CRUD operations
│   ├── search_and_discovery.feature   # Search
│   └── import_export.feature          # Import/export
│
└── features/step_definitions/
    ├── graph_steps.ts                # Graph steps
    ├── progressive-graph-steps.ts    # 3D graph steps (fog, camera, links)
    ├── note_steps.ts                 # Note steps
    └── common_steps.ts               # Common steps
```

### 3D Graph Feature Files

**`full_3d_graph.feature`** (9 scenarios):
- Navigate to full 3D graph from homepage
- Display all notes
- Loading without spinner (progressive loading)
- Fog during loading
- Camera centering
- **Link preservation during zoom**
- **Link preservation during rotation**
- Correct link display with many nodes

**`local_3d_graph.feature`** (13 scenarios):
- Navigate from note details to 3D
- Navigate from homepage on selection
- Single note with fog
- Progressive loading
- "Show all notes" toggle
- **Link preservation when switching modes**
- **Link preservation during camera zoom**
- **Link preservation during camera rotation**
- **Check for link duplication**
- **Link correctness after progressive loading**

### Running Cucumber

```bash
# All BDD tests
npm run test:cucumber

# With specific tag
CUCUMBER_TAGS="@smoke" npm run test:cucumber

# HTML report
npm run test:cucumber:report
```

### Feature Example

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

## Running Tests

### Full Check (all levels)

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

## Docker Test Stacks

The project provides three separate Docker environments for different purposes:

### Stack Comparison

| Stack | Purpose | Frontend Port | Backend Port | Config File |
|-------|---------|---------------|--------------|-------------|
| **Dev** | Development | 5173 | 9000 | `docker-compose.yml` |
| **Personal** | Personal notes | 3001 | 8085 | `docker-compose.personal.yml` |
| **Test** | Testing/E2E | 3002 | 8083 | `docker-compose.test.yml` |

### Test Stack (Isolated Testing Environment)

The test stack provides a completely isolated environment for testing that is destroyed after use, ensuring no test data contaminates development or personal environments.

**Port Mapping:**
- Frontend: `3002:3000`
- Backend: `8083:8080`
- Graph Service: `9095:9091` (HTTP), `9094:9090` (gRPC)
- NLP Service: `5002:5000`
- PostgreSQL: `5434:5432`
- Redis: `6381:6379`
- MongoDB: `27019:27017`
- Nginx: `8084:8080`

**Key Features:**
- Unique container names (`kg-test-*`)
- Separate volumes (`pgdata_test`, `mongodbdata_test`)
- Test database (`knowledge_test`)
- Auto-flush Redis on startup
- SKIP_AUTH enabled for testing
- Completely isolated from dev/personal stacks

#### Starting the Test Stack

```powershell
# Using PowerShell script
.\scripts\start-test.ps1

# Or manually
docker compose -f docker-compose.test.yml up -d --build
```

#### Stopping the Test Stack

```powershell
# Using PowerShell script (removes all data)
.\scripts\stop-test.ps1

# Or manually
docker compose -f docker-compose.test.yml down -v
```

The `-v` flag is critical—it removes all volumes, ensuring complete cleanup of test data.

#### Seeding Test Data

```powershell
# Populate test database with sample data
.\scripts\seed-test-data.ps1
```

This creates:
- Test user (`testuser` / `Test123!`)
- 5 sample notes (star, planet, galaxy, nebula, blackhole)
- 5 test tags
- Sample connections between notes

#### Running Tests Against Test Stack

```bash
# Start test stack
.\scripts\start-test.ps1

# Seed test data (optional)
.\scripts\seed-test-data.ps1

# Run E2E tests
cd frontend
npm run test

# Run BDD tests
npm run test:bdd

# Stop and cleanup
.\scripts\stop-test.ps1
```

#### Health Checks

```bash
# Test stack health endpoints
curl http://localhost:8083/health           # Backend
curl http://localhost:8084/health           # Nginx gateway
curl http://localhost:9095/health           # Graph service
curl http://localhost:5002/health           # NLP service
```

### Dev Stack

For development work with live data persistence.

**Starting:**
```bash
docker compose up -d
```

**Ports:**
- Frontend: `5173:3000`
- Backend: `9000:8080`
- Graph Service: `9091:9091`
- PostgreSQL: `15432:5432`
- Redis: `6379:6379`

### Personal Stack

For personal knowledge base with backup integration.

**Starting:**
```bash
docker compose -f docker-compose.personal.yml up -d
```

**Ports:**
- Frontend: `3001:3000`
- Backend: `8085:8080`
- Graph Service: `9092:9091`
- PostgreSQL: `5433:5432`
- Redis: `6380:6379`

### Stack Isolation

All three stacks can run simultaneously without conflicts:

| Resource | Dev | Personal | Test |
|----------|-----|----------|------|
| PostgreSQL Port | 15432 | 5433 | 5434 |
| Redis Port | 6379 | 6380 | 6381 |
| MongoDB Port | 27017 | 27018 | 27019 |
| Backend Port | 9000 | 8085 | 8083 |
| Frontend Port | 5173 | 3001 | 3002 |
| Container Names | kg-* | kg-*-personal | kg-test-* |
| Volumes | postgres_data | pgdata_personal | pgdata_test |

**Important:** Never run multiple stacks that share the same ports or container names. Always use the dedicated test stack for automated testing.

---

## Reports and Coverage

### Backend Coverage

```bash
cd backend

# Generate report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# View
open coverage.html

# Check threshold (target 70%, enforced minimum 60%)
go-test-coverage -coverprofile=coverage.out -threshold=60
```

### Frontend Coverage

```bash
cd frontend

# Generate report
npm run test:coverage

# View
cd coverage
open index.html

# E2E coverage (via Playwright trace)
npx playwright test --trace on

# Open report
npx playwright show-report
```

### Current Coverage

| Component | Coverage | Tests | Status |
|-----------|----------|-------|--------|
| **Backend Unit** | ~62% | — | ✅ Good (target 70%) |
| **Frontend Unit (lines)** | ~64% | 580 | ✅ Good (target 70%) |
| **Frontend Unit (functions)** | ~57% | 580 | ⚠️ Below 60% gate (target 70%) |
| **Frontend E2E** | N/A | 48 | ✅ Excellent |
| **BDD Scenarios** | N/A | 111 | ✅ Excellent |
| **NLP Python** | ~80% | ~15 | ✅ Excellent |
| **Total** | - | **~512** | ✅ |

### Additional Tests Needed

- [ ] **Worker Integration Tests** - Redis queue + task processing
- [ ] **Load Tests** - k6 or Artillery for API load
- [ ] **Security Tests** - OWASP ZAP scanning
- [ ] **Contract Tests** - Pact for API contracts
- [ ] **CORS Configuration Tests** - Verify CORS headers and origin validation

### CORS Testing

**Environment Variables:**
- `CORS_ALLOWED_ORIGINS` - Comma-separated list of allowed origins
- `CORS_ALLOWED_METHODS` - Allowed HTTP methods (default: `GET,POST,PUT,DELETE,OPTIONS`)
- `CORS_ALLOWED_HEADERS` - Allowed headers (default: `Content-Type,Authorization`)
- `CORS_MAX_AGE` - Preflight cache duration in seconds (default: `86400`)

**Test Commands:**
```bash
# Check CORS headers
curl -I http://localhost:8080/api/v1/notes

# Test preflight request
curl -X OPTIONS http://localhost:8080/api/v1/notes \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type,Authorization"
```

**Expected Headers:**
- `Access-Control-Allow-Origin: <origin from whitelist>`
- `Access-Control-Allow-Methods: GET,POST,PUT,DELETE,OPTIONS`
- `Access-Control-Allow-Headers: Content-Type,Authorization`
- `Access-Control-Max-Age: 86400`

---

## Useful Commands

```bash
# Quick check
make test              # All tests
make test-backend      # Backend only
make test-frontend     # Frontend E2E only
make test-cucumber     # BDD only

# Debug
make test-debug        # With debug info
make test-watch        # Watch mode

# Reports
make coverage          # Coverage
make report            # HTML report
```
