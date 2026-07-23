---
name: Testing Rules
alwaysApply: false
globs: ["**/*_test.go", "**/*.test.ts", "**/*.spec.ts", "tests/**", "**/*_test.py"]
description: Testing patterns - Go table-driven tests, Vitest, Playwright E2E, pytest, coverage >60%
---

# Testing Rules

## Coverage Requirement

**Minimum coverage: >60% for all modules.** This is mandatory and non-negotiable.

## Go — Table-Driven Tests with Testify

### Unit Test Pattern

```go
package note

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

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
            name:         "valid note with star type",
            title:        "Test Title",
            content:      "Test Content",
            noteType:     "star",
            wantErr:      false,
            expectedType: "star",
        },
        {
            name:         "empty type defaults to star",
            title:        "Test",
            content:      "Content",
            noteType:     "",
            wantErr:      false,
            expectedType: "star",
        },
        {
            name:    "empty title returns error",
            title:   "",
            content: "Content",
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            title, err := NewTitle(tt.title)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            content, _ := NewContent(tt.content)
            n := NewNote(title, content, tt.noteType, nil)
            assert.Equal(t, tt.expectedType, n.Type())
            assert.NotEmpty(t, n.ID())
        })
    }
}
```

### Repository Mock Pattern

```go
// Mock implementation for testing use cases
type MockNoteRepository struct {
    notes map[uuid.UUID]*note.Note
    err   error
}

func NewMockNoteRepository() *MockNoteRepository {
    return &MockNoteRepository{notes: make(map[uuid.UUID]*note.Note)}
}

func (m *MockNoteRepository) Save(ctx context.Context, n *note.Note) error {
    if m.err != nil {
        return m.err
    }
    m.notes[n.ID()] = n
    return nil
}

func (m *MockNoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
    if m.err != nil {
        return nil, m.err
    }
    n, ok := m.notes[id]
    if !ok {
        return nil, note.ErrNoteNotFound
    }
    return n, nil
}
```

### Integration Test Pattern

```go
//go:build integration

func TestNoteRepository_Integration(t *testing.T) {
    db := setupTestDB(t) // Creates test database, runs migrations
    defer cleanupTestDB(t, db)

    repo := postgres.NewNoteRepository(db, nil)
    ctx := context.Background()

    t.Run("save and find by ID", func(t *testing.T) {
        title, _ := note.NewTitle("Integration Test")
        content, _ := note.NewContent("Test content")
        n := note.NewNote(title, content, "star", nil)

        err := repo.Save(ctx, n)
        require.NoError(t, err)

        found, err := repo.FindByID(ctx, n.ID())
        require.NoError(t, err)
        assert.Equal(t, n.ID(), found.ID())
        assert.Equal(t, "Integration Test", found.Title().String())
    })
}
```

## Frontend — Vitest + @testing-library/svelte

### Component Test Pattern

```typescript
import { render, screen, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import Button from '$components/atoms/Button.svelte';

describe('Button', () => {
  it('renders with correct variant class', () => {
    render(Button, { props: { variant: 'danger' } });
    const btn = screen.getByRole('button');
    expect(btn).toHaveClass('danger');
  });

  it('calls onClick handler', async () => {
    const onClick = vi.fn();
    render(Button, { props: { onClick } });
    await fireEvent.click(screen.getByRole('button'));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it('is disabled when disabled prop is true', () => {
    render(Button, { props: { disabled: true } });
    expect(screen.getByRole('button')).toBeDisabled();
  });
});
```

### Store Test Pattern

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock dependencies
vi.mock('$shared/api/auth', () => ({
  login: vi.fn(),
  refreshTokens: vi.fn(),
}));

describe('auth store', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset store state
  });

  it('sets currentUser after successful login', async () => {
    const mockUser = { id: '1', email: 'test@example.com', username: 'test' };
    vi.mocked(authApi.login).mockResolvedValue({ access_token: 'token', refresh_token: 'refresh' });
    vi.mocked(usersApi.getProfile).mockResolvedValue(mockUser);

    await login('test@example.com', 'password');
    expect(currentUser()).toEqual(mockUser);
  });
});
```

### API Client Test Pattern

```typescript
import { describe, it, expect, vi } from 'vitest';
import { getNotes, createNote } from '$shared/api/notes';

vi.mock('$shared/api/client');

describe('notes API', () => {
  it('fetches notes with pagination', async () => {
    const mockResponse = { notes: [{ id: '1', title: 'Test' }], total: 1 };
    vi.mocked(client.get).mockReturnValue({ json: () => Promise.resolve(mockResponse) });

    const result = await getNotes(10, 0);
    expect(result.notes).toHaveLength(1);
    expect(result.total).toBe(1);
  });
});
```

## Playwright — E2E Tests

**⚠️ IMPORTANT:** Always use the isolated test stack for E2E and BDD testing. Never run E2E tests against dev or personal stacks.

### Isolated Testing Model

The project uses an isolated testing model where dev and personal stacks are stopped during testing to prevent resource conflicts and ensure accurate test results.

### Test Stack Usage

```bash
# Start isolated test stack
.\scripts\testing\start-test.ps1              # Windows
./scripts/testing/start-test.sh               # Linux/Mac

# Seed test data
.\scripts\testing\seed-test-data.ps1          # Windows
./scripts\testing\seed-test-data.sh           # Linux/Mac

# Run E2E tests against test stack
cd frontend && npx playwright test

# Stop and destroy test stack
.\scripts\testing\stop-test.ps1               # Windows
./scripts/testing/stop-test.sh                # Linux/Mac
```

### Full Regression Cycle

```bash
# Full regression cycle (24 steps with isolated model)
.\scripts\testing\run-full-test-cycle.ps1      # Windows
./scripts/testing/run-full-test-cycle.sh       # Linux/Mac
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
import { test, expect } from '@playwright/test';

test.describe('Graph Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:3002/graph');  // Test stack URL
    await page.waitForLoadState('networkidle');
  });

  test('displays graph canvas', async ({ page }) => {
    const canvas = page.locator('[data-testid="graph-canvas"]');
    await expect(canvas).toBeVisible();
  });

  test('creates a new note from graph', async ({ page }) => {
    await page.click('[data-testid="create-note-btn"]');
    await page.fill('[data-testid="note-title-input"]', 'E2E Test Note');
    await page.fill('[data-testid="note-content-input"]', 'Created via Playwright');
    await page.click('[data-testid="save-note-btn"]');
    await expect(page.locator('text=Note created successfully')).toBeVisible();
  });
});
```

### Test Stack URLs

- **Frontend:** http://localhost:3002
- **Backend API:** http://localhost:18083
- **PostgreSQL:** localhost:15434
- **Redis:** localhost:16381
- **MongoDB:** localhost:27019
- **NLP:** localhost:15002
- **Graph service:** localhost:9095

## Python — pytest

```python
import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app

@pytest.fixture
async def client():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as c:
        yield c

@pytest.mark.asyncio
async def test_health_endpoint(client):
    response = await client.get("/health")
    assert response.status_code == 200
    assert response.json()["status"] == "healthy"

@pytest.mark.asyncio
async def test_extract_keywords(client):
    response = await client.post(
        "/extract_keywords",
        json={"text": "Machine learning and neural networks", "top_n": 5}
    )
    assert response.status_code == 200
    data = response.json()
    assert len(data["keywords"]) <= 5
    assert all("keyword" in kw and "weight" in kw for kw in data["keywords"])
```

## Commands

```bash
# Go tests
cd backend && go test ./... -v                    # All tests
cd backend && go test ./... -cover                # With coverage
cd backend && go test -tags=integration ./...     # Integration tests

# Frontend tests
cd frontend && npm run test:unit                  # Vitest
cd frontend && npm run test:unit -- --coverage    # With coverage
cd frontend && npx playwright test               # E2E

# NLP tests
cd nlp-service && pytest tests/ -v               # All tests
cd nlp-service && pytest --cov=app tests/        # With coverage
```

## Rules

1. Every new feature MUST have tests
2. Use `require` for fatal assertions, `assert` for non-fatal (Go)
3. Test names should describe the behavior, not the implementation
4. Mock external dependencies (DB, Redis, HTTP clients)
5. Integration tests use build tag `//go:build integration`
6. Always test error paths, not just happy paths
7. **Stop dev/personal stacks before E2E/BDD/regression** — running all stacks simultaneously causes Docker instability and Windows `localhost` → `::1` Playwright failures
8. On Windows, use `http://127.0.0.1:3002` / `http://127.0.0.1:18083` or rebuild the test frontend with `VITE_API_URL=http://127.0.0.1:18083`
9. Ensure `.env` contains `JWT_SECRET` and DB passwords matching existing `postgres_data` / `pgdata_personal` volumes

### Windows Playwright/BDD URL workaround

```powershell
$env:FRONTEND_URL = "http://127.0.0.1:3002"
$env:BACKEND_URL = "http://127.0.0.1:18083"
npx playwright test --project=chromium-skip-auth
```

### Rebuild test frontend for IPv4

```bash
docker compose -f docker-compose.test.yml build --build-arg VITE_API_URL=http://127.0.0.1:18083 frontend-test
```
