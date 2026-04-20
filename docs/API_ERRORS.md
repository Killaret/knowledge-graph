# API Errors Documentation

> **Версия:** 1.0  
> **Дата:** Апрель 2026 (актуально на момент написания)  
> **Статус:** Актуально для Backend API v1.0.0

## Общий формат ошибок

Все ошибки API возвращаются в едином формате:

```json
{
  "error": "error_code",
  "message": "Human-readable error description",
  "detail": "Optional detailed information",
  "code": 404
}
```

| Поле | Тип | Описание |
|------|-----|----------|
| `error` | string | Код ошибки (snake_case) |
| `message` | string | Человекочитаемое описание |
| `detail` | string | Дополнительные детали (опционально) |
| `code` | integer | HTTP статус код |

---

## HTTP Status Codes

| Код | Описание | Когда возникает |
|-----|----------|-----------------|
| **200** | OK | Успешный GET/PUT запрос |
| **201** | Created | Успешное создание ресурса |
| **204** | No Content | Успешное удаление без тела ответа |
| **400** | Bad Request | Невалидный запрос (валидация) |
| **404** | Not Found | Ресурс не найден |
| **500** | Internal Server Error | Внутренняя ошибка сервера |

---

## Ошибки Notes API

### GET /notes/{id}

**404 Not Found**
```json
{
  "error": "note_not_found",
  "message": "Note with ID 'abc-123' not found",
  "code": 404
}
```

---

### POST /notes

**400 Bad Request - Missing title**
```json
{
  "error": "validation_error",
  "message": "Title is required",
  "detail": "Field 'title' cannot be empty",
  "code": 400
}
```

**400 Bad Request - Title too long**
```json
{
  "error": "validation_error", 
  "message": "Title exceeds maximum length",
  "detail": "Maximum length is 200 characters",
  "code": 400
}
```

**400 Bad Request - Content too long**
```json
{
  "error": "validation_error",
  "message": "Content exceeds maximum length", 
  "detail": "Maximum length is 10000 characters",
  "code": 400
}
```

---

### PUT /notes/{id}

**404 Not Found**
```json
{
  "error": "note_not_found",
  "message": "Note with ID 'abc-123' not found",
  "code": 404
}
```

---

### DELETE /notes/{id}

**404 Not Found**
```json
{
  "error": "note_not_found",
  "message": "Note with ID 'abc-123' not found",
  "code": 404
}
```

---

## Ошибки Links API

### POST /links

**400 Bad Request - Missing source**
```json
{
  "error": "validation_error",
  "message": "Source note ID is required",
  "detail": "Field 'source_note_id' is required",
  "code": 400
}
```

**400 Bad Request - Missing target**
```json
{
  "error": "validation_error",
  "message": "Target note ID is required",
  "detail": "Field 'target_note_id' is required", 
  "code": 400
}
```

**404 Not Found - Source note**
```json
{
  "error": "source_note_not_found",
  "message": "Source note with ID 'abc-123' not found",
  "code": 404
}
```

**404 Not Found - Target note**
```json
{
  "error": "target_note_not_found",
  "message": "Target note with ID 'xyz-789' not found",
  "code": 404
}
```

**400 Bad Request - Invalid weight**
```json
{
  "error": "validation_error",
  "message": "Weight must be between 0.0 and 1.0",
  "detail": "Received: 1.5",
  "code": 400
}
```

---

### GET /links/{id}

**404 Not Found**
```json
{
  "error": "link_not_found",
  "message": "Link with ID 'link-123' not found",
  "code": 404
}
```

---

### DELETE /links/{id}

**404 Not Found**
```json
{
  "error": "link_not_found",
  "message": "Link with ID 'link-123' not found",
  "code": 404
}
```

---

## Ошибки Search API

### GET /notes/search

**400 Bad Request - Missing query**
```json
{
  "error": "validation_error",
  "message": "Search query is required",
  "detail": "Parameter 'q' is missing",
  "code": 400
}
```

**400 Bad Request - Query too short**
```json
{
  "error": "validation_error",
  "message": "Search query too short",
  "detail": "Minimum length is 2 characters",
  "code": 400
}
```

---

## Ошибки Graph API

### GET /notes/{id}/graph

**404 Not Found**
```json
{
  "error": "note_not_found",
  "message": "Note with ID 'abc-123' not found",
  "code": 404
}
```

---

### GET /graph/all

**500 Internal Server Error**
```json
{
  "error": "internal_error",
  "message": "Failed to load graph data",
  "detail": "Database connection error",
  "code": 500
}
```

---

## Ошибки Suggestions API

### GET /notes/{id}/suggestions

**404 Not Found**
```json
{
  "error": "note_not_found",
  "message": "Note with ID 'abc-123' not found",
  "code": 404
}
```

---

## Внутренние ошибки (500)

### Общие ошибки сервера

```json
{
  "error": "internal_error",
  "message": "Internal server error",
  "detail": "An unexpected error occurred. Please try again later.",
  "code": 500
}
```

### Ошибки базы данных

```json
{
  "error": "database_error",
  "message": "Database operation failed",
  "detail": "Connection timeout",
  "code": 500
}
```

### Ошибки NLP сервиса

```json
{
  "error": "nlp_service_error",
  "message": "NLP service unavailable",
  "detail": "Failed to connect to NLP service at http://nlp:5000",
  "code": 500
}
```

---

## Обработка ошибок на клиенте

### JavaScript/TypeScript пример

```typescript
async function handleApiCall() {
  try {
    const response = await fetch('/api/notes/invalid-id');
    
    if (!response.ok) {
      const error = await response.json();
      
      switch (error.code) {
        case 404:
          console.error(`Resource not found: ${error.message}`);
          showUserMessage('Заметка не найдена');
          break;
        case 400:
          console.error(`Validation error: ${error.detail}`);
          showUserMessage('Проверьте введённые данные');
          break;
        case 500:
          console.error(`Server error: ${error.message}`);
          showUserMessage('Произошла ошибка. Попробуйте позже.');
          break;
        default:
          console.error(`Unknown error: ${error.message}`);
      }
      
      return null;
    }
    
    return await response.json();
  } catch (networkError) {
    console.error('Network error:', networkError);
    showUserMessage('Нет подключения к серверу');
    return null;
  }
}
```

### Go пример

```go
if resp.StatusCode == http.StatusNotFound {
    var apiErr struct {
        Error   string `json:"error"`
        Message string `json:"message"`
    }
    json.NewDecoder(resp.Body).Decode(&apiErr)
    return fmt.Errorf("%s: %s", apiErr.Error, apiErr.Message)
}
```

---

## Логирование ошибок

Все ошибки логируются с:
- HTTP методом и URL
- Кодом ошибки и сообщением
- Request ID (для трассировки)
- Временной меткой

Пример лога:
```
[2026-04-15 10:30:45] ERROR: note_not_found
  Request: GET /api/notes/invalid-id
  Message: Note with ID 'invalid-id' not found
  RequestID: req-abc-123-xyz
```
