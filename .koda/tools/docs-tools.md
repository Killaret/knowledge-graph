# knowledge-graph-docs-tools

**Version:** 1.0  
**Purpose:** Documentation maintenance and generation  
**Status:** Active  
**Priority:** 🟡 Medium

---

## Overview

`docs-tools` provides patterns and guidelines for maintaining project documentation, README files, ADRs, and changelogs.

**Key Areas:**
- README.md maintenance
- ADR (Architecture Decision Records)
- Changelog generation
- API documentation
- Code comments and inline docs
- Documentation consistency

---

## Documentation Structure

### Current Organization

```
docs/
├── AGENTS.md                    # Agent descriptions (Copilot)
├── AGENT_SETUP_GUIDE.md         # Agent setup instructions
├── BEFORE_AFTER_COMPARISON.md   # Before/after comparisons
├── AGENT_VISIBILITY.md          # Agent visibility in chats
├── API_ERRORS.md                # API error codes
└── ADR/                         # Architecture Decision Records
    ├── 001-use-clean-architecture.md
    ├── 002-use-postgresql.md
    └── ...
```

---

## README.md Maintenance

### Structure Template

```markdown
# Project Name

Short description (1-2 sentences)

## Features

- ✅ Feature 1
- ✅ Feature 2
- ✅ Feature 3

## Quick Start

```bash
docker-compose up
```

## Architecture

[Architecture diagram and description]

## Documentation

- [Backend](docs/backend.md)
- [Frontend](docs/frontend.md)
- [API](docs/api.md)

## Agents

This project uses AI agents for development assistance:

- `backend-go` - Backend development
- `frontend-svelte` - Frontend development
- `integration` - API integration
- [More agents...]

## Commands

### Backend
```bash
cd backend
go test ./... -v
```

### Frontend
```bash
cd frontend
npm run test:unit
```

## Contributing

[Contributing guidelines]

## License

[License info]
```

### Update Checklist

When updating README.md:

- [ ] Check all links are valid
- [ ] Verify code examples work
- [ ] Update version numbers
- [ ] Add new features
- [ ] Remove deprecated information
- [ ] Check formatting and structure

---

## ADR (Architecture Decision Records)

### ADR Template

```markdown
# ADR 001: Use Clean Architecture

**Status:** Accepted  
**Date:** 2026-05-24  
**Deciders:** [Team members]

## Context

[What is the issue that prompted this decision?]

We need a clear separation of concerns between business logic, 
infrastructure, and interface layers to ensure maintainability 
and testability.

## Decision

We will use Clean Architecture with the following layers:

```
backend/
├── internal/
│   ├── domain/           # Pure business logic
│   ├── application/      # Use cases
│   ├── infrastructure/   # External systems
│   └── interfaces/       # API, CLI, etc.
```

## Rationale

1. **Separation of Concerns**: Each layer has a clear responsibility
2. **Testability**: Domain layer can be tested without infrastructure
3. **Maintainability**: Changes in one layer don't affect others
4. **Flexibility**: Easy to swap implementations (e.g., database)

## Consequences

### Positive

- ✅ Clear boundaries between layers
- ✅ Easier to write unit tests
- ✅ Better code organization
- ✅ Easier onboarding for new developers

### Negative

- ⚠️ More boilerplate code
- ⚠️ Learning curve for team
- ⚠️ Initial setup complexity

## Alternatives Considered

1. **MVC Pattern**: Less clear separation, harder to test
2. **Hexagonal Architecture**: Similar benefits, different terminology
3. **Simple Layered Architecture**: Less flexibility

## References

- [Clean Architecture by Robert Martin](https://www.amazon.com/Clean-Architecture-Craftsmans-Software-Structure/dp/0134494164)
- [Domain-Driven Design](https://en.wikipedia.org/wiki/Domain-driven_design)
```

### ADR Naming Convention

```
ADR 001: [Action] [Subject]

Examples:
- ADR 001: Use Clean Architecture
- ADR 002: Choose PostgreSQL as Database
- ADR 003: Implement JWT Authentication
```

### ADR Status Values

- **Draft**: Proposed, under review
- **Accepted**: Approved, should be implemented
- **Rejected**: Not chosen
- **Deprecated**: No longer valid
- **Superseded**: Replaced by another ADR

---

## Changelog Generation

### Changelog Format

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- New feature X
- New endpoint /api/v1/xyz

### Changed
- Improved performance of Y
- Updated dependency Z

### Deprecated
- Old API endpoint /api/v1/old

### Removed
- Deprecated feature W

### Fixed
- Bug fix for issue #123

### Security
- Fixed vulnerability in authentication

## [1.0.0] - 2026-05-30

### Added
- Initial release
- User authentication
- Note management API
- Graph visualization

### Changed
- Improved error handling
- Updated UI design

### Fixed
- Fixed memory leak in graph renderer
- Fixed race condition in cache

## [0.9.0] - 2026-05-27

### Added
- Docker Compose setup
- CI/CD pipeline

### Changed
- Refactored database layer
```

### Changelog Categories

- **Added**: New features
- **Changed**: Changes in existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security improvements

### Generating Changelog from Git

```bash
# Generate changelog from git commits
git log --oneline --since="2026-05-01" > changelog.tmp

# Group by type
grep -i "feat:" changelog.tmp | sed 's/^/- /' > features.md
grep -i "fix:" changelog.tmp | sed 's/^/- /' > fixes.md
grep -i "security:" changelog.tmp | sed 's/^/- /' > security.md
```

---

## API Documentation

### OpenAPI Specification

```yaml
# openAPI.yaml
openapi: 3.0.3
info:
  title: Knowledge Graph API
  description: API for managing notes and graph connections
  version: 1.0.0

paths:
  /api/v1/notes:
    get:
      summary: List all notes
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            default: 100
        - name: offset
          in: query
          schema:
            type: integer
            default: 0
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  notes:
                    type: array
                    items:
                      $ref: '#/components/schemas/Note'
                  total:
                    type: integer
                  limit:
                    type: integer
                  offset:
                    type: integer
        '401':
          $ref: '#/components/responses/Unauthorized'
        '500':
          $ref: '#/components/responses/InternalServerError'
    
    post:
      summary: Create a new note
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateNoteRequest'
      responses:
        '201':
          description: Note created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Note'
        '400':
          $ref: '#/components/responses/BadRequest'

components:
  schemas:
    Note:
      type: object
      properties:
        id:
          type: string
          format: uuid
        title:
          type: string
        content:
          type: string
        type:
          type: string
          enum: [star, highlight, link]
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
    
    CreateNoteRequest:
      type: object
      required:
        - title
        - content
        - type
      properties:
        title:
          type: string
          minLength: 1
          maxLength: 255
        content:
          type: string
        type:
          type: string
          enum: [star, highlight, link]
  
  responses:
    Unauthorized:
      description: Unauthorized
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    
    BadRequest:
      description: Bad request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
```

### API Errors Documentation

```markdown
# API Errors

All API errors follow this format:

```json
{
  "code": "ERROR_CODE",
  "message": "Human readable message",
  "details": {
    "field": "Additional details"
  }
}
```

## Error Codes

### Authentication

- `UNAUTHORIZED` - No authentication provided
- `INVALID_TOKEN` - Token is invalid or expired
- `TOKEN_EXPIRED` - Token has expired

### Validation

- `VALIDATION_ERROR` - Request validation failed
- `INVALID_INPUT` - Input data is invalid
- `MISSING_FIELD` - Required field is missing

### Resource

- `NOT_FOUND` - Resource not found
- `ALREADY_EXISTS` - Resource already exists
- `CONFLICT` - Resource conflict

### System

- `INTERNAL_ERROR` - Internal server error
- `DATABASE_ERROR` - Database operation failed
- `SERVICE_UNAVAILABLE` - Service temporarily unavailable
```

---

## Code Documentation

### Go Documentation

```go
// Package domain contains core business logic and entities.
// 
// This package should have no dependencies on external systems.
// All business rules and validation logic belong here.
//
// Example:
//
//	note, err := domain.NewNote("Title", "Content", "star")
//	if err != nil {
//	    log.Fatal(err)
//	}
package domain

// Note represents a knowledge note in the graph.
// It can be a star (main idea), highlight (important excerpt),
// or link (reference to external content).
type Note struct {
    // ID is a unique identifier for the note (UUID format)
    ID string
    
    // Title is the note title (1-255 characters)
    Title string
    
    // Content is the note content (max 10000 characters)
    Content string
    
    // Type is the note type: star, highlight, or link
    Type NoteType
    
    // CreatedAt is when the note was created
    CreatedAt time.Time
    
    // UpdatedAt is when the note was last updated
    UpdatedAt time.Time
}

// NewNote creates a new Note with validation.
// Returns an error if:
//   - title is empty or exceeds 255 characters
//   - content is empty
//   - type is not valid (star, highlight, link)
func NewNote(title, content string, typ string) (*Note, error) {
    // Implementation
}
```

### TypeScript Documentation

```typescript
/**
 * Note entity representing a knowledge note in the graph.
 * 
 * Can be one of three types:
 * - `star`: Main idea or concept
 * - `highlight`: Important excerpt or quote
 * - `link`: Reference to external content
 * 
 * @example
 * ```typescript
 * const note: Note = {
 *   id: '123e4567-e89b-12d3-a456-426614174000',
 *   title: 'My Note',
 *   content: 'Note content',
 *   type: 'star',
 *   created_at: new Date().toISOString(),
 *   updated_at: new Date().toISOString(),
 *   metadata: {}
 * };
 * ```
 */
export interface Note {
  /** Unique identifier (UUID format) */
  id: string;
  
  /** Note title (1-255 characters) */
  title: string;
  
  /** Note content (max 10000 characters) */
  content: string;
  
  /** Note type */
  type: 'star' | 'highlight' | 'link';
  
  /** Creation timestamp (ISO 8601) */
  created_at: string;
  
  /** Last update timestamp (ISO 8601) */
  updated_at: string;
  
  /** Additional metadata */
  metadata: Record<string, any>;
}

/**
 * Creates a new note with validation.
 * 
 * @param title - Note title (required, 1-255 chars)
 * @param content - Note content (required)
 * @param type - Note type (star, highlight, or link)
 * @returns Promise resolving to created Note
 * @throws {ValidationError} If validation fails
 * 
 * @example
 * ```typescript
 * const note = await createNote('Title', 'Content', 'star');
 * ```
 */
export async function createNote(
  title: string,
  content: string,
  type: NoteType
): Promise<Note> {
  // Implementation
}
```

---

## Documentation Commands

### Generate Documentation

```bash
# Backend Go docs
cd backend
godoc -http=:6060  # View at http://localhost:6060

# TypeScript docs
cd frontend
npx typedoc --out docs/api

# API docs from OpenAPI
docker run -p 8080:8080 -v $(pwd):/tmp/swagger swaggerapi/swagger-ui
```

### Validate Documentation

```bash
# Check markdown formatting
markdownlint docs/*.md

# Check links
lychee docs/*.md

# Validate OpenAPI spec
swagger validate openAPI.yaml
```

---

## Documentation Checklist

### New Feature Documentation

- [ ] Update README.md with new feature
- [ ] Add API endpoint to OpenAPI spec
- [ ] Create ADR if architectural decision
- [ ] Add inline code comments
- [ ] Update API_ERRORS.md if new error codes
- [ ] Add examples to documentation
- [ ] Update changelog

### Documentation Review

- [ ] All links work
- [ ] Code examples are current
- [ ] Version numbers updated
- [ ] No typos or grammatical errors
- [ ] Consistent formatting
- [ ] Screenshots up to date

---

## Best Practices

### 1. Keep Documentation Current

```markdown
✅ GOOD: Documentation matches code

// Code
func CreateNote(title string) (*Note, error) {
    if title == "" {
        return nil, ErrEmptyTitle
    }
}

// Docs
// CreateNote creates a new note.
// Returns ErrEmptyTitle if title is empty.
```

```markdown
❌ BAD: Outdated documentation

// Code (changed)
func CreateNote(title string, content string) (*Note, error) {
    // ...
}

// Docs (old)
// CreateNote creates a note with just a title.
```

### 2. Use Examples

```markdown
✅ GOOD: With examples

## Usage

```bash
# Create a note
curl -X POST http://localhost:8080/api/v1/notes \
  -H "Content-Type: application/json" \
  -d '{"title":"My Note","content":"Content","type":"star"}'
```

### 3. Document Errors

```markdown
✅ GOOD: Error documentation

### Errors

| Code | Description |
|------|-------------|
| 400 | Invalid input data |
| 401 | Not authenticated |
| 404 | Note not found |
| 500 | Server error |
```

### 4. Version Documentation

```markdown
✅ GOOD: Versioned docs

## API v1 (Current)
[Documentation]

## API v0 (Deprecated)
[Documentation - will be removed in v2.0]
```

---

## Tools

### Markdown Linting

```yaml
# .markdownlintrc
{
  "default": true,
  "MD013": false,  # Line length
  "MD033": false,  # Inline HTML
  "MD007": {
    "indent": 2
  }
}
```

### Link Checker

```bash
# Install
npm install -g markdown-link-check

# Run
find docs -name "*.md" -exec markdown-link-check {} \;
```

---

**Tools:** `docs-tools.md`  
**Coverage:** All public APIs documented  
**Update Frequency:** With every release
