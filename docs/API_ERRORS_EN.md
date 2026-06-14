да# API Errors Documentation

> **Version:** 1.1  
> **Date:** April 2026  
> **Status:** Current for Backend API v1.1.0 (Unified REST API)

---

## General Response Format

### Success Response (200, 201)

```json
{
  "data": { ... },
  "message": "Optional message"
}
```

### Error Response (400, 404, 409, 500)

All errors are returned in a unified format:

```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable description",
  "details": [
    {
      "field": "field_name",
      "reason": "reason",
      "message": "error description",
      "received": "received value",
      "expected": "expected value"
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `code` | string | Error code (UPPER_SNAKE_CASE) |
| `message` | string | Human-readable description |
| `details` | array | Field-level details (optional) |

---

## HTTP Status Codes

| Code | Description | When Occurs |
|------|-------------|-------------|
| **200** | OK | Successful request |
| **201** | Created | Successful resource creation |
| **204** | No Content | Successful deletion without response body |
| **400** | Bad Request | Validation error |
| **403** | Forbidden | Access denied (reserved for future authentication) |
| **404** | Not Found | Resource not found |
| **409** | Conflict | Data conflict (duplicate) |
| **500** | Internal Server Error | Internal server error |

---

## Error Codes

### VALIDATION_ERROR (400)
Input validation error

### NOT_FOUND (404)
Resource not found

### CONFLICT (409)
Data conflict (e.g., duplicate link)

### FORBIDDEN (403)
Access denied (reserved)

### UNAUTHORIZED (401)
Authentication required (reserved)

### INTERNAL_ERROR (500)
Internal server error

---

## Notes API Errors

### POST /notes

**400 Bad Request - Validation Error**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "title",
      "reason": "required",
      "message": "Title is required"
    }
  ]
}
```

**400 Bad Request - Title too long**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "title",
      "reason": "too_long",
      "message": "Title must not exceed 200 characters",
      "received": "very long title...",
      "expected": "max 200 chars"
    }
  ]
}
```

**400 Bad Request - Invalid note type**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "type",
      "reason": "invalid_value",
      "message": "Type must be one of: star, planet, comet, galaxy, asteroid, satellite, debris, nebula",
      "received": "invalid_type"
    }
  ]
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to save note"
}
```

---

### GET /notes/{id}

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Invalid UUID format",
      "received": "not-a-uuid"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Note not found"
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to retrieve note"
}
```

---

### PUT /notes/{id}

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Invalid UUID format",
      "received": "not-a-uuid"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Note not found"
}
```

**400 Bad Request - Validation Error**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "title",
      "reason": "invalid_value",
      "message": "Title is required",
      "received": ""
    }
  ]
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to update note"
}
```

---

### DELETE /notes/{id}

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Invalid UUID format"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Note not found"
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to delete note"
}
```

---

## Links API Errors

### POST /links

**400 Bad Request - Validation Error**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "source_note_id",
      "reason": "required",
      "message": "Source note ID is required"
    },
    {
      "field": "link_type",
      "reason": "invalid_value",
      "message": "Link type must be one of: reference, dependency, related, custom"
    }
  ]
}
```

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "source_note_id",
      "reason": "invalid_format",
      "message": "Invalid UUID format",
      "received": "invalid-uuid"
    }
  ]
}
```

**404 Not Found - Source note**
```json
{
  "code": "NOT_FOUND",
  "message": "Source note not found"
}
```

**404 Not Found - Target note**
```json
{
  "code": "NOT_FOUND",
  "message": "Target note not found"
}
```

**409 Conflict - Duplicate link**
```json
{
  "code": "CONFLICT",
  "message": "Data conflict",
  "details": [
    {
      "field": "link",
      "reason": "already_exists",
      "message": "Link already exists",
      "received": {
        "source_note_id": "uuid-1",
        "target_note_id": "uuid-2",
        "link_type": "reference"
      },
      "expected": "unique combination of source, target and type"
    }
  ]
}
```

**400 Bad Request - Invalid weight**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "weight",
      "reason": "out_of_range",
      "message": "Weight must be between 0.0 and 1.0",
      "received": 1.5,
      "expected": "0.0 - 1.0"
    }
  ]
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to save link"
}
```

---

### GET /links/{id}

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Invalid UUID format"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Link not found"
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to retrieve link"
}
```

---

### DELETE /links/{id}

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Invalid UUID format"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Link not found"
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to delete link"
}
```

---

### GET /notes/{id}/links

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Invalid UUID format"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Note not found"
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to retrieve links"
}
```

---

### DELETE /notes/{id}/links

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Invalid UUID format"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Note not found"
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to delete link"
}
```

---

## Search API Errors

### GET /notes/search

**400 Bad Request - Missing query**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "q",
      "reason": "required",
      "message": "Search query is required"
    }
  ]
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to perform search"
}
```

---

## Graph API Errors

### GET /notes/{id}/graph

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Invalid UUID format"
    }
  ]
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to load graph"
}
```

---

### GET /graph/all

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to load graph"
}
```

---

## Suggestions API Errors

### GET /notes/{id}/suggestions

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Invalid UUID format"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Note not found"
}
```

**202 Accepted** (recommendations being generated)
```json
{
  "suggestions": [],
  "message": "Recommendations are being generated"
}
```

---

## Client-Side Error Handling

### JavaScript/TypeScript Example

```typescript
interface FieldError {
  field: string;
  reason: string;
  message: string;
  received?: any;
  expected?: string;
}

interface ErrorResponse {
  code: string;
  message: string;
  details?: FieldError[];
}

async function handleApiCall() {
  try {
    const response = await fetch('/api/notes/invalid-id');
    
    if (!response.ok) {
      const error: ErrorResponse = await response.json();
      
      switch (error.code) {
        case 'NOT_FOUND':
          console.error(`Resource not found: ${error.message}`);
          showUserMessage('Note not found');
          break;
        case 'VALIDATION_ERROR':
          console.error(`Validation error:`, error.details);
          const fieldErrors = error.details?.map(d => d.message).join(', ');
          showUserMessage(`Errors: ${fieldErrors}`);
          break;
        case 'CONFLICT':
          console.error(`Conflict: ${error.message}`);
          showUserMessage('Data conflict. Resource may already exist.');
          break;
        case 'INTERNAL_ERROR':
          console.error(`Server error: ${error.message}`);
          showUserMessage('Server error occurred. Please try again later.');
          break;
        default:
          console.error(`Unknown error: ${error.message}`);
      }
      
      return null;
    }
    
    const success = await response.json();
    return success.data; // Data in data field
  } catch (networkError) {
    console.error('Network error:', networkError);
    showUserMessage('No connection to server');
    return null;
  }
}
```

---

### Go Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type FieldError struct {
    Field    string `json:"field"`
    Reason   string `json:"reason"`
    Message  string `json:"message"`
    Received any    `json:"received,omitempty"`
    Expected string `json:"expected,omitempty"`
}

type ErrorResponse struct {
    Code    string       `json:"code"`
    Message string       `json:"message"`
    Details []FieldError `json:"details,omitempty"`
}

type SuccessResponse struct {
    Data    any    `json:"data,omitempty"`
    Message string `json:"message,omitempty"`
}

func handleResponse(resp *http.Response) error {
    if resp.StatusCode >= 400 {
        var errResp ErrorResponse
        if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
            return fmt.Errorf("failed to decode error: %w", err)
        }
        
        switch errResp.Code {
        case "NOT_FOUND":
            return fmt.Errorf("not found: %s", errResp.Message)
        case "VALIDATION_ERROR":
            var fieldErrors []string
            for _, d := range errResp.Details {
                fieldErrors = append(fieldErrors, fmt.Sprintf("%s: %s", d.Field, d.Message))
            }
            return fmt.Errorf("validation error: %s - %v", errResp.Message, fieldErrors)
        case "CONFLICT":
            return fmt.Errorf("conflict: %s", errResp.Message)
        default:
            return fmt.Errorf("error %s: %s", errResp.Code, errResp.Message)
        }
    }
    
    var success SuccessResponse
    if err := json.NewDecoder(resp.Body).Decode(&success); err != nil {
        return fmt.Errorf("failed to decode success: %w", err)
    }
    
    // Work with success.Data
    return nil
}
```

---

## Rate Limiting

API implements rate limiting via middleware. When limits are exceeded, **429 Too Many Requests** is returned.

### Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| General requests | 100 | per minute |
| POST /notes | 30 | per minute |
| POST /links | 50 | per minute |
| PUT /notes/{id} | 20 | per minute |
| DELETE /notes/{id} | 20 | per minute |

### 429 Too Many Requests

```json
{
  "code": "RATE_LIMIT_EXCEEDED",
  "message": "Rate limit exceeded. Please try again later."
}
```

**Client-side handling:**
```go
if resp.StatusCode == http.StatusTooManyRequests {
    // Exponential backoff
    time.Sleep(time.Second * 5)
    return retryRequest()
}
```

---

## Swagger UI

Interactive API documentation is available at:

```
http://localhost:8080/swagger/index.html
```

OpenAPI specification is available at:

```
http://localhost:8080/openapi.yaml
```

---

## Error Logging

All errors are logged with:
- HTTP method and URL
- Error code and message
- Field details (on validation)
- Request ID (for tracing)
- Timestamp

Example log:
```
[2026-04-27 22:30:45] ERROR: VALIDATION_ERROR
  Request: POST /api/notes
  Message: Invalid input data
  Details: [{"field": "title", "reason": "required", "message": "Title is required"}]
  RequestID: req-abc-123-xyz
```
