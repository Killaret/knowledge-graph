# API Errors Documentation

> **Версия:** 1.1  
> **Дата:** Апрель 2026  
> **Статус:** Актуально для Backend API v1.1.0 (Unified REST API)

---

## Общий формат ответов

### Успешный ответ (200, 201)

```json
{
  "data": { ... },
  "message": "Опциональное сообщение"
}
```

### Ошибка (400, 404, 409, 500)

Все ошибки возвращаются в едином формате:

```json
{
  "code": "ERROR_CODE",
  "message": "Человекочитаемое описание",
  "details": [
    {
      "field": "имя_поля",
      "reason": "причина",
      "message": "описание ошибки",
      "received": "полученное значение",
      "expected": "ожидаемое значение"
    }
  ]
}
```

| Поле | Тип | Описание |
|------|-----|----------|
| `code` | string | Код ошибки (UPPER_SNAKE_CASE) |
| `message` | string | Человекочитаемое описание на русском |
| `details` | array | Детализация по полям (опционально) |

---

## HTTP Status Codes

| Код | Описание | Когда возникает |
|-----|----------|-----------------|
| **200** | OK | Успешный запрос |
| **201** | Created | Успешное создание ресурса |
| **204** | No Content | Успешное удаление без тела ответа |
| **400** | Bad Request | Ошибка валидации |
| **403** | Forbidden | Доступ запрещён (зарезервировано для будущей аутентификации) |
| **404** | Not Found | Ресурс не найден |
| **409** | Conflict | Конфликт данных (дубликат) |
| **500** | Internal Server Error | Внутренняя ошибка сервера |

---

## Коды ошибок

### VALIDATION_ERROR (400)
Ошибка валидации входных данных

### NOT_FOUND (404)
Ресурс не найден

### CONFLICT (409)
Конфликт данных (например, дубликат связи)

### FORBIDDEN (403)
Доступ запрещён (зарезервировано)

### UNAUTHORIZED (401)
Требуется аутентификация (зарезервировано)

### INTERNAL_ERROR (500)
Внутренняя ошибка сервера

---

## Ошибки Notes API

### POST /notes

**400 Bad Request - Validation Error**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
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
  "message": "Некорректные входные данные",
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
  "message": "Некорректные входные данные",
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
  "message": "Не удалось сохранить заметку"
}
```

---

### GET /notes/{id}

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Неверный формат UUID",
      "received": "not-a-uuid"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Заметка не найдена"
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Не удалось получить заметку"
}
```

---

### PUT /notes/{id}

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Неверный формат UUID",
      "received": "not-a-uuid"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Заметка не найдена"
}
```

**400 Bad Request - Validation Error**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
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
  "message": "Не удалось обновить заметку"
}
```

---

### DELETE /notes/{id}

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Неверный формат UUID"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Заметка не найдена"
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Не удалось удалить заметку"
}
```

---

## Ошибки Links API

### POST /links

**400 Bad Request - Validation Error**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
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
  "message": "Некорректные входные данные",
  "details": [
    {
      "field": "source_note_id",
      "reason": "invalid_format",
      "message": "Неверный формат UUID",
      "received": "invalid-uuid"
    }
  ]
}
```

**404 Not Found - Source note**
```json
{
  "code": "NOT_FOUND",
  "message": "Исходная заметка не найдена"
}
```

**404 Not Found - Target note**
```json
{
  "code": "NOT_FOUND",
  "message": "Целевая заметка не найдена"
}
```

**409 Conflict - Duplicate link**
```json
{
  "code": "CONFLICT",
  "message": "Конфликт данных",
  "details": [
    {
      "field": "link",
      "reason": "already_exists",
      "message": "Связь уже существует",
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
  "message": "Некорректные входные данные",
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
  "message": "Не удалось сохранить связь"
}
```

---

### GET /links/{id}

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Неверный формат UUID"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Связь не найдена"
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Не удалось получить связь"
}
```

---

### DELETE /links/{id}

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Неверный формат UUID"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Связь не найдена"
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Не удалось удалить связь"
}
```

---

### GET /notes/{id}/links

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Неверный формат UUID"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Заметка не найдена"
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Не удалось получить связь"
}
```

---

### DELETE /notes/{id}/links

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Неверный формат UUID"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Заметка не найдена"
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Не удалось удалить связь"
}
```

---

## Ошибки Search API

### GET /notes/search

**400 Bad Request - Missing query**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
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
  "message": "Не удалось выполнить поиск"
}
```

---

## Ошибки Graph API

### GET /notes/{id}/graph

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Неверный формат UUID"
    }
  ]
}
```

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Не удалось загрузить граф"
}
```

---

### GET /graph/all

**500 Internal Server Error**
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Не удалось загрузить граф"
}
```

---

## Ошибки Suggestions API

### GET /notes/{id}/suggestions

**400 Bad Request - Invalid UUID**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Некорректные входные данные",
  "details": [
    {
      "field": "id",
      "reason": "invalid_format",
      "message": "Неверный формат UUID"
    }
  ]
}
```

**404 Not Found**
```json
{
  "code": "NOT_FOUND",
  "message": "Заметка не найдена"
}
```

**202 Accepted** (рекомендации генерируются)
```json
{
  "suggestions": [],
  "message": "Recommendations are being generated"
}
```

---

## Обработка ошибок на клиенте

### JavaScript/TypeScript пример

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
          showUserMessage('Заметка не найдена');
          break;
        case 'VALIDATION_ERROR':
          console.error(`Validation error:`, error.details);
          const fieldErrors = error.details?.map(d => d.message).join(', ');
          showUserMessage(`Ошибки: ${fieldErrors}`);
          break;
        case 'CONFLICT':
          console.error(`Conflict: ${error.message}`);
          showUserMessage('Конфликт данных. Возможно, ресурс уже существует.');
          break;
        case 'INTERNAL_ERROR':
          console.error(`Server error: ${error.message}`);
          showUserMessage('Произошла ошибка сервера. Попробуйте позже.');
          break;
        default:
          console.error(`Unknown error: ${error.message}`);
      }
      
      return null;
    }
    
    const success = await response.json();
    return success.data; // Данные в поле data
  } catch (networkError) {
    console.error('Network error:', networkError);
    showUserMessage('Нет подключения к серверу');
    return null;
  }
}
```

---

### Go пример

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
    
    // Работа с success.Data
    return nil
}
```

---

## Rate Limiting

API реализует rate limiting через middleware. При превышении лимитов возвращается **429 Too Many Requests**.

### Лимиты

| Endpoint | Limit | Window |
|----------|-------|--------|
| Общие запросы | 100 | per minute |
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

Интерактивная документация API доступна по адресу:

```
http://localhost:8080/swagger/index.html
```

OpenAPI спецификация доступна по:

```
http://localhost:8080/openapi.yaml
```

---

## Логирование ошибок

Все ошибки логируются с:
- HTTP методом и URL
- Кодом ошибки и сообщением
- Деталями полей (при валидации)
- Request ID (для трассировки)
- Временной меткой

Пример лога:
```
[2026-04-27 22:30:45] ERROR: VALIDATION_ERROR
  Request: POST /api/notes
  Message: Некорректные входные данные
  Details: [{"field": "title", "reason": "required", "message": "Title is required"}]
  RequestID: req-abc-123-xyz
```
