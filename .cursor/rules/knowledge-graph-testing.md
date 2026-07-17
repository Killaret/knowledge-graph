# Cursor Rule: knowledge-graph-testing

Testing standards for all layers: Go unit/integration, Svelte/Vitest, and
Playwright E2E. Coverage target: **>60%** per package.

---

## Go Unit Tests — Table-Driven with testify/require

**Always use `require` for fatal assertions** (stops test immediately on
failure). Use `assert` only for non-fatal checks where you want to see all
failures.

```go
// backend/internal/domain/note/entity_test.go — real pattern
func TestNewNote(t *testing.T) {
    tests := []struct {
        name         string
        title        string
        content      string
        noteType     string
        wantErr      bool
        expectedType string
    }{
        {
            name:         "valid note defaults to star type",
            title:        "Test Title",
            content:      "Test Content",
            noteType:     "",
            wantErr:      false,
            expectedType: "star",
        },
        {
            name:    "empty title returns error",
            title:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            title, err := note.NewTitle(tt.title)
            if tt.wantErr {
                require.Error(t, err)   // require — stops immediately
                return
            }
            require.NoError(t, err)

            content, _ := note.NewContent(tt.content)
            metadata, _ := note.NewMetadata(map[string]interface{}{})
            n := note.NewNote(title, content, tt.noteType, metadata)

            assert.Equal(t, tt.expectedType, n.Type())
            assert.NotEqual(t, uuid.Nil, n.ID())
        })
    }
}
```

---

## Naming Conventions

```
unit test:         <file>_test.go        (same package)
unit test (black): <file>_unit_test.go   (external package)
integration test:  <file>_integration_test.go
```

Test function names:
```
TestTypeName_MethodName        → TestNote_UpdateTitle
TestFunctionName               → TestNewNote
TestFunctionName/case_name     → sub-tests via t.Run
```

---

## testcontainers-go for Integration Tests

```go
// backend/internal/infrastructure/db/postgres/note_repo_integration_test.go
//go:build integration

package postgres_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestNoteRepository_Integration(t *testing.T) {
    ctx := context.Background()

    // Spin up a real Postgres with pgvector image
    container, err := postgres.RunContainer(ctx,
        postgres.WithImage("pgvector/pgvector:pg16"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    require.NoError(t, err)
    t.Cleanup(func() { container.Terminate(ctx) })

    dsn, err := container.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)

    db, err := db.Connect(dsn)
    require.NoError(t, err)

    // Run migrations
    err = postgres.RunMigrations(db, "../../../../backend/migrations")
    require.NoError(t, err)

    repo := postgres.NewNoteRepository(db, nil)

    // Table-driven test cases using real Postgres
    title, _ := note.NewTitle("Integration Test")
    content, _ := note.NewContent("Body")
    metadata, _ := note.NewMetadata(map[string]interface{}{})
    n := note.NewNote(title, content, "star", metadata)

    require.NoError(t, repo.Save(ctx, n))
    found, err := repo.FindByID(ctx, n.ID())
    require.NoError(t, err)
    require.NotNil(t, found)
    require.Equal(t, "Integration Test", found.Title().String())
}
```

Run integration tests:
```bash
go test ./... -tags=integration -v
```

---

## miniredis for Redis Unit Tests

```go
// backend/internal/application/cache/graph_cache_test.go
import "github.com/alicebob/miniredis/v2"

func TestGraphCache(t *testing.T) {
    mr := miniredis.RunT(t)           // auto-cleanup on test end
    rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

    cache := cache.NewGraphCache(rdb)
    ctx := context.Background()

    data := cache.GraphData{
        Nodes: []cache.GraphNode{{ID: "1", Title: "A"}},
    }
    err := cache.CacheUserGraph(ctx, "user-1", data)
    require.NoError(t, err)

    got, hit, err := cache.GetCachedUserGraph(ctx, "user-1")
    require.NoError(t, err)
    require.True(t, hit)
    require.Len(t, got.Nodes, 1)
}
```

---

## Coverage Requirements

```bash
# Per-package coverage
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -html=coverage.out -o coverage.html

# Enforce minimum (add to CI)
TOTAL=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
if (( $(echo "$TOTAL < 60" | bc -l) )); then
  echo "Coverage $TOTAL% is below 60% threshold"
  exit 1
fi
```

---

## Vitest Unit Tests (Frontend)

```typescript
// frontend/src/shared/stores/auth.svelte.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock external modules
vi.mock('$shared/api/auth', () => ({
  login: vi.fn().mockResolvedValue({
    access_token: 'test-token',
    refresh_token: 'refresh-token'
  })
}));

describe('auth store', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('returns false when not authenticated', () => {
    expect(isAuthenticated()).toBe(false);
  });
});
```

```bash
cd frontend
npx vitest run              # run once
npx vitest run --coverage   # with Istanbul coverage
```

---

## @testing-library/svelte Component Tests

```typescript
// frontend/src/components/molecules/NoteCard.spec.ts
import { render, screen } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import NoteCard from './NoteCard.svelte';

describe('NoteCard', () => {
  it('renders title from props', () => {
    render(NoteCard, {
      props: {
        note: { id: '1', title: 'My Note', content: 'Body text' },
        onDelete: vi.fn()
      }
    });
    expect(screen.getByText('My Note')).toBeInTheDocument();
  });
});
```

---

## Playwright E2E Tests

**⚠️ IMPORTANT:** Always use the isolated test stack for E2E and BDD testing. Never run E2E tests against dev or personal stacks.

### Isolated Testing Model

The project uses an isolated testing model where dev and personal stacks are stopped during testing to prevent resource conflicts and ensure accurate test results.

### Test Stack Usage

```bash
# Start isolated test stack
.\scripts\start-test.ps1              # Windows
./scripts/start-test.sh               # Linux/Mac

# Seed test data
.\scripts\seed-test-data.ps1          # Windows
./scripts\seed-test-data.sh           # Linux/Mac

# Run E2E tests against test stack
cd frontend && npx playwright test

# Stop and destroy test stack
.\scripts\stop-test.ps1               # Windows
./scripts/stop-test.sh                # Linux/Mac
```

### Full Regression Cycle

```bash
# Full regression cycle (24 steps with isolated model)
.\scripts\run-full-test-cycle.ps1      # Windows
./scripts/run-full-test-cycle.sh       # Linux/Mac
```

The full regression cycle:
1. Captures dev stack state snapshot
2. Stops dev and personal stacks
3. Runs all tests on isolated test stack
4. Compares dev stack state before/after testing
5. Verifies dev/personal identity
6. Auto-commits if all checks pass

### E2E Test Example

```typescript
// frontend/e2e/notes.spec.ts
import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  // Skip auth in E2E by injecting localStorage flag (dev mode only)
  await page.addInitScript(() => {
    localStorage.setItem('__SKIP_AUTH__', 'true');
  });
});

test('creates a note and sees it in graph', async ({ page }) => {
  await page.goto('http://localhost:3002/notes/new');  // Test stack URL
  await page.fill('[name="title"]', 'Playwright Test Note');
  await page.fill('[name="content"]', 'E2E test content');
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/\/notes\/.+/);
});
```

Config: `frontend/cucumber.mjs` (BDD runner), `playwright.config.ts`.

### Test Stack URLs

- **Frontend:** http://localhost:3002
- **Backend API:** http://localhost:8083
- **PostgreSQL:** localhost:5434
- **Redis:** localhost:6381
- **MongoDB:** localhost:27018

---

## go-sqlmock for Repository Unit Tests

```go
// backend/internal/infrastructure/db/postgres/note_repo_unit_test.go
import sqlmock "github.com/DATA-DOG/go-sqlmock"

func TestNoteRepo_FindByID_NotFound(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)

    gormDB, _ := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
    repo := NewNoteRepository(gormDB, nil)

    mock.ExpectQuery(`SELECT .* FROM "notes"`).
        WillReturnError(gorm.ErrRecordNotFound)

    result, err := repo.FindByID(context.Background(), uuid.New())
    require.NoError(t, err)   // FindByID returns (nil, nil) on not found
    require.Nil(t, result)
}
```

---

## Anti-Patterns

```go
// ❌ assert instead of require for fatal checks
assert.NoError(t, err)
assert.NotNil(t, entity)  // test continues even if entity is nil → panic next line

// ❌ No sub-test names (makes failures unreadable)
for _, tt := range tests {
    // Missing: t.Run(tt.name, func(t *testing.T) { ... })
}

// ❌ Real database in unit test (slow, requires running Postgres)
db, _ := db.Connect(os.Getenv("DATABASE_URL"))  // use testcontainers or sqlmock
```

```typescript
// ❌ Testing implementation details instead of behavior
expect(component.internalState).toBe(true)  // test DOM/output, not internals

// ❌ Missing mock cleanup between tests
// Always use beforeEach to reset mocks: vi.clearAllMocks()
```
