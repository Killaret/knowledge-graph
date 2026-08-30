# ТЗ для Java-сервиса `source-text-handler`

**Дата:** 2026-08-22  
**Статус:** согласовано, ожидает реализации

## 1. Назначение и границы

`source-text-handler` — Java 17 микросервис для парсинга, чанкинга и первичной обработки внешних документов.

**Делает:**
- Забирает из Redis plain JSON-задачу (документ, URL, текст).
- Парсит содержимое (Apache Tika, PDF, HTML, текст и т.д.).
- Разбивает документ на чанки.
- Для "умных" чанков — **сам создаёт заметки** в Go backend по HTTP.
- Для "сырых" чанков — **не создаёт заметки**, возвращает их в `ImportResult.raw_chunks[]`.
- Возвращает в Redis plain JSON-результат обработки.

**Не делает:**
- Не создаёт связи (`links`).
- Не управляет статусами задач в БД.
- Не знает про asynq, outbox, SSE, polling.
- Не логирует JWT и сырые документы.

**Go backend** отвечает за:
- приём `ImportResult`,
- обновление `import_tasks`,
- доработку `raw_chunks` асинхронно,
- создание связей (если будет решено).

## 2. Ветки

- **Главная Java-ветка:** `origin/java-source-text-handler`.
- **Устаревшая:** `origin/sourceTextHandler`.
- **Интеграция с основным проектом:** `feat/2d-adaptive-fog` (локальные `Dockerfile`, `README` отсюда переносим).
- При мерже в `feat/2d-adaptive-fog` не сбрасывать frontend-изменения.

## 3. Входящий контракт (Go backend → Java)

Redis list: `import:document:pending`  
Формат: plain JSON, без `AsynqTaskEnvelope`.

```json
{
  "event_id": "uuid-task-1",
  "correlation_id": "uuid-correlation-1",
  "user_id": "uuid-user-1",
  "jwt": "eyJhbGci...",
  "type": "FILE|URL|TEXT",
  "content": "base64-string | URL | plain text",
  "content_type": "application/pdf",
  "import_options": {
    "chunk_size": 500,
    "overlap": 50,
    "min_chunk_length": 100,
    "max_retries": 3,
    "create_links": false
  },
  "metadata": {
    "filename": "doc.pdf",
    "source": "telegram"
  }
}
```

- `event_id` — уникальный ID задачи, для idempotency.
- `correlation_id` — сквозной ID, возвращается в `ImportResult`.
- `jwt` — access token пользователя, используется во всех HTTP-запросах.
- `type` — `FILE`, `URL`, `TEXT`.
- `content` — base64 для `FILE`, URL для `URL`, текст для `TEXT`.
- `content_type` — обязателен для `FILE`.

## 4. Исходящий контракт (Java → Go backend)

Redis list: `import:responses:pending`  
Формат: plain JSON.

```json
{
  "correlation_id": "uuid-correlation-1",
  "event_id": "uuid-task-1",
  "user_id": "uuid-user-1",
  "status": "COMPLETED|PARTIAL|FAILED",
  "note_ids": ["uuid-note-1", "uuid-note-2"],
  "raw_chunks": [
    {
      "index": 3,
      "content": "сырой текст ...",
      "metadata": {
        "filename": "doc.pdf",
        "page": 12,
        "source": "file"
      },
      "quality_score": 35
    }
  ],
  "links": [],
  "errors": [],
  "processed_at": "2026-08-22T12:00:00Z"
}
```

Правила:
- `COMPLETED` — все чанки обработаны, smart-заметки созданы, `raw_chunks` пуст.
- `PARTIAL` — часть smart-заметок создана, часть в `raw_chunks`.
- `FAILED` — ни одна заметка не создана, документ не распарсился.
- `note_ids` — ID заметок, созданных Java (smart-чанки).
- `raw_chunks` — чанки, которые Java не смогла сделать "умными"; Go backend дорабатывает их позже.
- `links` — **всегда пустой массив**. Java не создаёт связи.
- `processed_at` — ISO-8601.

## 5. Smart vs Raw — критерии

Для каждого чанка Java считает `smart_score` 0–100.

### Smart-чанк (создаём заметку)

| Критерий | Баллы | Описание |
|---|---|---|
| Есть заголовок | +30 | Tika выделила heading, или first sentence выглядит как тезис (5–15 слов). |
| Граница по смыслу | +20 | Чанк заканчивается на границе абзаца/раздела, не рубит предложение. |
| Длина ок | +20 | 200–8000 символов. |
| Title качественный | +20 | Title отличается от content, first sentence информативен. |
| Metadata есть | +10 | `page`, `section`, `source`, `filename` заполнены. |

Если `smart_score >= 70` — Java создаёт заметку.

### Raw-чанк (не создаём, отдаём в Go)

Если:
- нет заголовка,
- контент > 10000 символов без разбивки,
- граница рубит предложение,
- Java не уверена в title,
- документ плохо структурирован.

В `raw_chunks[]` передаётся:
- `index` — порядковый номер,
- `content` — сырой текст,
- `metadata` — всё, что смогла извлечь,
- `quality_score` — число 0–69.

## 6. HTTP-контракт с Go backend

### Base URL

Читается из переменной окружения `BACKEND_URL`.

- Local: `http://localhost:9000/api/v1`
- Docker network: `http://backend:8080/api/v1`

### Заголовки

- `Authorization: Bearer <jwt>` — обязательно, берётся из задачи.
- `Content-Type: application/json`.

### Создание одной заметки

`POST /api/v1/notes`

```json
{
  "title": "...",
  "content": "...",
  "type": "unknown",
  "metadata": {
    "event_id": "...",
    "correlation_id": "...",
    "chunk_index": 0,
    "filename": "doc.pdf",
    "page": 12,
    "source": "file"
  }
}
```

Важно:
- `title` — обязательно, max 255 символов.
- `content` — максимум 10000 символов.
- Go возвращает `201 Created` + `{"data": {"id": "..."}, ...}`.
- Java парсит `data.id`.

### Batch создание заметок

Когда Go backend добавит `POST /api/v1/notes/batch` (create), Java должна уметь его использовать:

```json
{
  "notes": [
    { "title": "...", "content": "...", "type": "unknown", "metadata": {...} },
    { "title": "...", "content": "...", "type": "unknown", "metadata": {...} }
  ]
}
```

До появления batch Java создаёт заметки по одному через `POST /api/v1/notes`.

### Связи

Java **не вызывает** `POST /api/v1/links`. Связи создаются позже на стороне Go backend или пользователем.

## 7. SSRF-защита / валидация URL

Перед скачиванием URL обязательна проверка. Запрещены:
- схема не `http`/`https` (`file://`, `ftp://`, `javascript:`, пустая),
- пустой хост,
- `localhost`,
- loopback, private, link-local IP:
  - `127.0.0.0/8`
  - `10.0.0.0/8`
  - `172.16.0.0/12`
  - `192.168.0.0/16`
  - `169.254.0.0/16`
  - `fe80::/10`
- нераспарсиваемые URL.

Референс: `backend/internal/application/import/service.go` (`IsAllowedURL`).

## 8. Логирование и безопасность

- Использовать SLF4J/Logback, убрать `System.err.println`.
- Не логировать `jwt`, base64-контент, сырые документы.
- Redis внутри Docker network, не публично.
- Graceful shutdown (origin уже есть, сохранить).

## 9. Парсинг / чанкинг

- `HybridChunker` не должен терять первый короткий чанк.
- `TextDocumentParser.parseFromUrl` должен корректно обрабатывать URL (или делегировать в Tika URL parser).
- `LinkDetector` — либо реализовать, либо оставить `links: []`. Связи всё равно не создаются.
- Chunk size не должен превышать Go content limit (10000 символов).

## 10. Docker и инфраструктура

- `Dockerfile` обновить под `maven-shade-plugin` → `target/app.jar`.
- Сервис должен иметь `/health` на порту 8081.
- `REDIS_URI` и `BACKEND_URL` из env.
- Позже — добавить в `docker-compose.yml`, `docker-compose.personal.yml`, `docker-compose.test.yml`.

## 11. Тесты

- WireMock: `201 Created`, `/api/v1`, `Authorization`, `data.id`.
- Unit: `HybridChunker`, `smart_score`, `TikaDocumentParser`, URL validation, `ImportDocumentHandler`.
- Integration: Redis inbound/outbound с plain JSON.

## 12. Что НЕ делает сервис (повтор)

- Не обновляет статусы `pending`/`processing`/`done`.
- Не знает про asynq, outbox, SSE, polling.
- Не создаёт связи.
- Не обрабатывает `raw_chunks` — это зона Go backend.
- Не отправляет результат напрямую на фронт.
