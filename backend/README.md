# Knowledge Graph Backend

REST API backend for Knowledge Graph application with unified response format.

## Features

- **Unified API Responses** — Consistent `SuccessResponse` and `ErrorResponse` formats
- **RESTful Design** — Resource-based URLs with proper HTTP methods
- **Detailed Error Messages** — Machine-readable codes with human-readable Russian messages
- **Swagger UI** — Interactive API documentation
- **Field-level Validation** — Detailed validation errors with `field`, `reason`, `message`, `received`, `expected`

## Quick Start

```bash
# Run server
go run ./cmd/server

# Run worker
go run ./cmd/worker
```

## API Documentation

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **OpenAPI Spec**: http://localhost:8080/openapi.yaml

## Response Format

### Success (200, 201)

```json
{
  "data": { ... },
  "message": "Ресурс успешно создан"
}
```

### Error (400, 404, 409, 500)

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
  "details": [
    {
      "field": "title",
      "reason": "required",
      "message": "Title is required",
      "received": "",
      "expected": "non-empty string"
    }
  ]
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Invalid input data |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Data conflict (duplicate) |
| `FORBIDDEN` | 403 | Access denied (reserved) |
| `INTERNAL_ERROR` | 500 | Server error |

## Project Structure

```
backend/
├── cmd/
│   ├── server/          # HTTP server
│   └── worker/          # Background worker
├── internal/
│   ├── application/     # Use cases
│   ├── domain/          # Business logic
│   ├── infrastructure/  # DB, Redis, NLP
│   └── interfaces/
│       └── api/
│           ├── common/          # Unified response structures
│           ├── notehandler/     # Note handlers
│           ├── linkhandler/     # Link handlers
│           └── graphhandler/    # Graph handlers
├── migrations/          # SQL migrations
└── openAPI.yaml        # API specification
```

## Handlers

All handlers use unified response helpers from `internal/interfaces/api/common`:

```go
// Success responses
apicommon.JSON(c, 200, data)
apicommon.JSONWithMessage(c, 201, data, apicommon.MsgResourceCreated)

// Error responses
apicommon.BadRequest(c, []apicommon.FieldError{
    apicommon.NewFieldError("title", apicommon.ReasonRequired, "Title is required"),
})
apicommon.NotFound(c, "Заметка")
apicommon.Conflict(c, details)
apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedSaveNote)
```

## Build

```bash
go build ./...
```

## Test

```bash
go test ./...
```

## Documentation

See `../docs/API_ERRORS.md` for detailed error handling documentation.
