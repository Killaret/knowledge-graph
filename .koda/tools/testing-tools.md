# knowledge-graph-testing-tools

**Version:** 1.0  
**Purpose:** Comprehensive testing patterns for all layers  
**Status:** Active  
**Priority:** 🟢 High

---

## Overview

`testing-tools` provides comprehensive testing patterns for Knowledge Graph project across all layers and languages.

**Key Areas:**
- Go unit & integration tests
- Frontend Vitest & Playwright
- BDD with Cucumber
- Python pytest for NLP service
- Contract testing
- Test infrastructure & coverage

---

## Test Infrastructure Overview

### Current Coverage

| Layer | Framework | Count | Status |
|-------|-----------|-------|--------|
| Backend Unit | Go testing + testify | 118 | ✅ Active |
| Backend Integration | Go testing + testify/suite | 24 | ✅ Active |
| Frontend Unit | Vitest + Testing Library | ~220 | ✅ Active |
| Frontend E2E | Playwright | 48 | ✅ Active |
| BDD Scenarios | Cucumber | 111 | ✅ Active |
| NLP Service | pytest | ~15 | ✅ Active |

---

## Backend Testing (Go)

### 1. Unit Tests

#### Table-Driven Tests Pattern

```go
// backend/internal/domain/note_test.go
package domain

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestNote_NewNote(t *testing.T) {
    tests := []struct {
        name    string
        title   string
        content string
        typ     string
        wantErr bool
    }{
        {
            name:    "valid note",
            title:   "Test Note",
            content: "Test content",
            typ:     "star",
            wantErr: false,
        },
        {
            name:    "empty title",
            title:   "",
            content: "Test content",
            typ:     "star",
            wantErr: true,
        },
        {
            name:    "invalid type",
            title:   "Test",
            content: "Content",
            typ:     "invalid",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            note, err := NewNote(tt.title, tt.content, tt.typ)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.title, note.Title)
            }
        })
    }
}
```

#### Mock Repository Pattern

```go
// backend/internal/application/graph_service_test.go
package application

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock repository
type MockNoteRepository struct {
    mock.Mock
}

func (m *MockNoteRepository) Create(ctx context.Context, note *Note) error {
    args := m.Called(ctx, note)
    return args.Error(0)
}

func (m *MockNoteRepository) GetByID(ctx context.Context, id string) (*Note, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*Note), args.Error(1)
}

// Test with mock
func TestGraphService_GetNote(t *testing.T) {
    mockRepo := new(MockNoteRepository)
    service := NewGraphService(mockRepo)
    
    expectedNote := &Note{
        ID:        "123",
        Title:     "Test",
        Content:   "Content",
        CreatedAt: time.Now(),
    }
    
    mockRepo.On("GetByID", mock.Anything, "123").Return(expectedNote, nil)
    
    note, err := service.GetNote(context.Background(), "123")
    
    assert.NoError(t, err)
    assert.Equal(t, expectedNote, note)
    mockRepo.AssertCalled(t, "GetByID", mock.Anything, "123")
}
```

### 2. Integration Tests

#### Integration Test Setup

```go
// backend/internal/interfaces/api/note_handler_integration_test.go
//go:build integration

package api

import (
    "context"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/suite"
    _ "github.com/lib/pq"
    "your-project/testutil"
)

type NoteHandlerIntegrationTestSuite struct {
    suite.Suite
    router     *gin.Engine
    db         *sql.DB
    cleanup    func()
}

func (s *NoteHandlerIntegrationTestSuite) SetupSuite() {
    gin.SetMode(gin.TestMode)
    
    // Setup test database
    s.db, s.cleanup = testutil.SetupTestDB(s.T())
    
    // Setup router
    s.router = SetupRouter(s.db)
}

func (s *NoteHandlerIntegrationTestSuite) TearDownSuite() {
    s.cleanup()
}

func (s *NoteHandlerIntegrationTestSuite) TestCreateNote_Success() {
    reqBody := map[string]interface{}{
        "title":   "Integration Test",
        "content": "Test content",
        "type":    "star",
    }
    
    body, _ := json.Marshal(reqBody)
    req, _ := http.NewRequest("POST", "/api/v1/notes", 
        strings.NewReader(string(body)))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    s.router.ServeHTTP(w, req)
    
    s.Equal(201, w.Code)
    
    var response NoteResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    s.NotEmpty(response.ID)
    s.Equal("Integration Test", response.Title)
}

func TestNoteHandlerIntegrationTestSuite(t *testing.T) {
    suite.Run(t, new(NoteHandlerIntegrationTestSuite))
}
```

#### Test Database Utility

```go
// backend/testutil/testdb.go
package testutil

import (
    "database/sql"
    "testing"
    
    _ "github.com/lib/pq"
)

func SetupTestDB(t *testing.T) (*sql.DB, func()) {
    // Use test database
    dsn := "postgres://postgres:password@localhost:5432/knowledge_graph_test?sslmode=disable"
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }
    
    // Run migrations
    if err := RunMigrations(db); err != nil {
        t.Fatalf("Failed to run migrations: %v", err)
    }
    
    cleanup := func() {
        // Clean up test data
        db.Exec("TRUNCATE TABLE notes RESTART IDENTITY CASCADE")
        db.Close()
    }
    
    return db, cleanup
}
```

---

## Frontend Testing (TypeScript)

### 1. Unit Tests with Vitest

#### Component Tests

```typescript
// frontend/src/lib/components/NoteCard.test.ts
import { render, screen } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import NoteCard from './NoteCard.svelte';

describe('NoteCard', () => {
  it('renders note title', () => {
    const note = {
      id: '1',
      title: 'Test Note',
      content: 'Test content',
      type: 'star',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      metadata: {}
    };
    
    render(NoteCard, { props: { note } });
    
    expect(screen.getByText('Test Note')).toBeInTheDocument();
  });
  
  it('calls onNoteClick when clicked', async () => {
    const note = {
      id: '1',
      title: 'Test',
      content: 'Content',
      type: 'star',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      metadata: {}
    };
    
    const handleClick = vi.fn();
    
    render(NoteCard, { 
      props: { note, onNoteClick: handleClick } 
    });
    
    const card = screen.getByRole('article');
    await card.click();
    
    expect(handleClick).toHaveBeenCalledWith(note);
  });
});
```

#### Service Tests with MSW

```typescript
// frontend/src/lib/api/notes.test.ts
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest';
import { getNotes, createNote } from './notes';

const server = setupServer(
  http.get('http://localhost:8080/api/v1/notes', () => {
    return HttpResponse.json({
      notes: [
        {
          id: '1',
          title: 'Test Note',
          content: 'Content',
          type: 'star',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          metadata: {}
        }
      ],
      total: 1,
      limit: 10000
    });
  })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('notes API', () => {
  it('fetches notes successfully', async () => {
    const notes = await getNotes();
    
    expect(notes).toHaveLength(1);
    expect(notes[0].title).toBe('Test Note');
  });
  
  it('handles API errors', async () => {
    server.use(
      http.get('http://localhost:8080/api/v1/notes', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR' },
          { status: 500 }
        );
      })
    );
    
    await expect(getNotes()).rejects.toThrow('INTERNAL_ERROR');
  });
});
```

### 2. E2E Tests with Playwright

```typescript
// frontend/tests/graph.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Graph View', () => {
  test.beforeEach(async ({ page }) => {
    // Skip auth for tests
    await page.addInitScript(() => {
      window.__SKIP_AUTH__ = true;
    });
    
    await page.goto('/graph');
  });
  
  test('displays notes as graph nodes', async ({ page }) => {
    await expect(page.locator('.graph-node')).toHaveCount.greaterThan(0);
  });
  
  test('clicking node shows details', async ({ page }) => {
    const firstNode = page.locator('.graph-node').first();
    await firstNode.click();
    
    await expect(page.locator('.note-details')).toBeVisible();
  });
  
  test('can create new note from graph', async ({ page }) => {
    await page.click('[data-testid="create-note-btn"]');
    
    await page.fill('[data-testid="note-title"]', 'New Note');
    await page.fill('[data-testid="note-content"]', 'Content');
    await page.click('[data-testid="save-btn"]');
    
    await expect(page.locator('.note-details')).toBeVisible();
  });
});
```

### 3. BDD Tests with Cucumber

```gherkin
# tests/features/graph_view.feature
Feature: Graph View
  As a user
  I want to see my notes as a graph
  So that I can visualize connections between them

  Scenario: View graph with connected notes
    Given I have test notes with connections
    When I navigate to the graph view
    Then I should see the notes displayed as celestial bodies
    And I should see connection lines between related notes

  Scenario: Create connection from graph
    Given I am viewing the graph
    And I have two notes
    When I drag from one note to another
    Then a connection should be created between them
    And both notes should update to show the connection

  Scenario: Filter notes by type
    Given I have notes of different types
    When I filter by "star" type
    Then I should only see star notes in the graph
    And other notes should be hidden
```

```typescript
// tests/support/graph_steps.ts
import { Given, When, Then } from '@cucumber/cucumber';
import { expect } from '@playwright/test';

Given('I have test notes with connections', async function() {
  // Setup test data
  await this.api.post('/api/v1/notes', {
    title: 'Note 1',
    content: 'Content 1',
    type: 'star'
  });
  
  await this.api.post('/api/v1/notes', {
    title: 'Note 2',
    content: 'Content 2',
    type: 'star'
  });
  
  // Create connection
  await this.api.post('/api/v1/connections', {
    from_note_id: '1',
    to_note_id: '2',
    type: 'related'
  });
});

When('I navigate to the graph view', async function() {
  await this.page.goto('/graph');
});

Then('I should see the notes displayed as celestial bodies', async function() {
  await expect(this.page.locator('.graph-node')).toHaveCount.greaterThan(0);
});
```

---

## NLP Service Testing (Python)

```python
# nlp-service/tests/test_entity_extraction.py
import pytest
from app.services.entity_extractor import EntityExtractor

class TestEntityExtractor:
    @pytest.fixture
    def extractor(self):
        return EntityExtractor()
    
    def test_extract_person_name(self, extractor):
        text = "John Smith works at Google"
        entities = extractor.extract(text)
        
        assert any(e['type'] == 'PERSON' and e['text'] == 'John Smith' 
                  for e in entities)
        assert any(e['type'] == 'ORGANIZATION' and e['text'] == 'Google' 
                  for e in entities)
    
    def test_empty_input(self, extractor):
        entities = extractor.extract("")
        assert len(entities) == 0
    
    def test_special_characters(self, extractor):
        text = "Contact: john@example.com or +1-555-1234"
        entities = extractor.extract(text)
        
        assert any(e['type'] == 'EMAIL' for e in entities)
        assert any(e['type'] == 'PHONE' for e in entities)

# Run tests
# pytest tests/ -v
# pytest tests/ -v --cov=app
```

---

## Testing Commands

### Backend (Go)

```bash
# All unit tests
cd backend
go test ./... -v

# With race detector
go test ./... -race

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Integration tests only
go test -tags=integration ./... -v

# Specific package
go test ./internal/application/graph -v

# Table-driven tests only
go test -run TestNote_NewNote -v
```

### Frontend (TypeScript)

```bash
cd frontend

# Unit tests (Vitest)
npm run test:unit

# Unit tests with coverage
npm run test:unit:coverage

# E2E tests (Playwright)
npm run test

# E2E tests with UI mode
npm run test:ui

# BDD tests (Cucumber)
npm run test:bdd

# All tests
npm run test:all
```

### NLP Service (Python)

```bash
cd nlp-service

# All tests
pytest tests/ -v

# With coverage
pytest tests/ -v --cov=app

# Specific test file
pytest tests/test_entity_extraction.py -v

# With logging
pytest tests/ -v -s --log-cli-level=INFO
```

---

## Test Organization

### Backend Structure

```
backend/
├── internal/
│   ├── domain/
│   │   ├── note_test.go           # Unit tests
│   │   └── connection_test.go
│   ├── application/
│   │   ├── graph_service_test.go  # Unit tests
│   │   └── note_service_test.go
│   └── interfaces/api/
│       ├── note_handler_test.go   # Unit tests
│       └── note_handler_integration_test.go  # Integration tests
├── testutil/
│   └── testdb.go                  # Test utilities
└── cmd/server/
    └── main_test.go               # Integration tests
```

### Frontend Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── components/
│   │   │   └── NoteCard.test.ts   # Component tests
│   │   └── api/
│   │       └── notes.test.ts      # Service tests
│   └── routes/
│       └── +page.svelte.test.ts   # Route tests
├── tests/
│   └── graph.spec.ts              # E2E tests
└── tests/features/
    └── graph_view.feature         # BDD scenarios
```

---

## Best Practices

### 1. Test Naming

```go
// ✅ GOOD
func TestNote_NewNote_ValidTitle(t *testing.T)
func TestGraphService_GetNote_Cached(t *testing.T)

// ❌ BAD
func TestNote(t *testing.T)
func Test1(t *testing.T)
```

### 2. Test Isolation

```go
// ✅ Each test should be independent
func TestCreateNote(t *testing.T) {
    // Setup fresh data for each test
    db := setupTestDB(t)
    // ...
}

// ❌ Tests should not depend on each other
func TestCreateNote(t *testing.T) { /* ... */ }
func TestGetNote(t *testing.T) { 
    // Depends on TestCreateNote - BAD
}
```

### 3. Test Data Cleanup

```go
// ✅ Clean up after each test
func (s *TestSuite) TearDownTest() {
    s.db.Exec("TRUNCATE TABLE notes RESTART IDENTITY CASCADE")
}

// ❌ Don't leave test data
func (s *TestSuite) TestSomething(t *testing.T) {
    // No cleanup - BAD
}
```

### 4. Mock Only External Dependencies

```go
// ✅ Mock external services
mockRepo.On("GetByID", mock.Anything, "123").Return(note, nil)

// ❌ Don't mock internal logic
mockService.On("CreateNote", mock.Anything, mock.Anything).Return(note, nil)
```

### 5. Test Coverage Targets

```yaml
# Minimum coverage targets
- Backend: > 60%
- Frontend: > 70%
- NLP Service: > 80%
```

---

## Continuous Integration

```yaml
# .github/workflows/testing.yml
name: Testing

on: [push, pull_request]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      
      - name: Run tests
        run: |
          cd backend
          go test -race -cover ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '20'
          cache: 'npm'
      
      - name: Install dependencies
        run: |
          cd frontend
          npm ci
      
      - name: Run unit tests
        run: |
          cd frontend
          npm run test:unit:coverage
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '20'
      
      - name: Install dependencies
        run: |
          cd frontend
          npm ci
      
      - name: Install Playwright
        run: npx playwright install --with-deps
      
      - name: Run E2E tests
        run: |
          cd frontend
          npm run test
```

---

**Tools:** `testing-tools.md`  
**Coverage Target:** > 60% backend, > 70% frontend  
**Test Count:** 118+ unit, 24+ integration, 48 E2E, 111 BDD
