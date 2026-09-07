# ТЗ для Java-сервиса `source-text-handler`
 
**Дата:** 2026-08-30  
**Статус:** согласовано, ожидает реализации
 
## 1. Назначение и границы
 
`source-text-handler` — Java 17 микросервис для парсинга, чанкинга и первичной обработки внешних документов.
 
**Делает:**
- Забирает из Redis plain JSON-задачу (документ, URL, текст).
- Парсит содержимое (Apache Tika).
- Проводит **двухпроходный анализ**:
  - Pass 1 — строит структурную карту документа.
  - Pass 2 — оценивает когезию, знаки препинания, ключевые слова.
- Для "умных" чанков — **сам создаёт заметки** в Go backend по HTTP.
- Для "сырых" чанков — **не создаёт заметки**, возвращает их в `ImportResult.raw_chunks[]`.
- Возвращает в Redis plain JSON-результат обработки.
 
**Не делает:**
- Не использует NLP/MLP/LLM (OpenNLP, Stanford, Lucene и т.д.).
- Не создаёт связи (`links`).
- Не управляет статусами задач в БД.
- Не знает про asynq, outbox, SSE, polling.
- Не логирует JWT и сырые документы.
 
**Go backend** отвечает за:
- приём `ImportResult`,
- обновление `import_tasks`,
- доработку `raw_chunks` (LLM/NLP),
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
    "max_chunk_length": 8000,
    "max_retries": 3,
    "create_links": false,
    "smart_threshold": 70,
    "smart_weights": {
      "has_heading": 25,
      "section_boundary": 10,
      "intra_chunk_cohesion": 20,
      "punctuation_completeness": 15,
      "chunk_ending_weight": 15,
      "length_ok": 10,
      "keyword_trend_stable": 5,
      "punctuation_balance": 5,
      "metadata_present": 5
    }
  },
  "metadata": {
    "filename": "doc.pdf",
    "source": "telegram"
  }
}
```
 
- `event_id` — уникальный ID задачи, для idempotency.
- `correlation_id` — сквозной ID, возвращается в `ImportResult`.
- `jwt` — access token пользователя, используется в `Authorization: Bearer <jwt>`.
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
      "quality_score": 45,
      "metrics": {
        "has_heading": false,
        "section_boundary": false,
        "intra_chunk_cohesion": 0.72,
        "cohesion_prev": 0.18,
        "cohesion_next": 0.15,
        "punctuation_completeness": 0.85,
        "chunk_ending_weight": 1.0,
        "sentences_complete_ratio": 1.0,
        "ends_at_paragraph": true,
        "length_chars": 8500,
        "length_ok": true,
        "title_quality_score": 40,
        "keyword_density": 0.12,
        "keyword_trend": "stable",
        "repetition_score": 0.3,
        "content_clean": 0.9,
        "starts_clean": true,
        "punctuation_balance": true,
        "metadata_present": true,
        "top_keywords": ["когезия", "java", "алгоритм"]
      }
    }
  ],
  "links": [],
  "errors": [],
  "processed_at": "2026-08-30T12:00:00Z"
}
```
 
Правила:
- `COMPLETED` — все чанки smart, `raw_chunks` пуст.
- `PARTIAL` — есть и `note_ids`, и `raw_chunks`.
- `FAILED` — ни одна smart-заметка не создана или документ не распарсился.
- `note_ids` — ID заметок, созданных Java (smart-чанки).
- `raw_chunks` — чанки, которые Java не смогла сделать "умными"; Go backend дорабатывает их позже.
- `links` — **всегда пустой массив**. Java не создаёт связи.
- `processed_at` — ISO-8601.
 
## 5. Двухпроходная система Smart/Raw
 
### Pass 1. Структурная карта (Structure Map)
 
Java парсит документ и строит дерево:
 
```
Document
├── Sections[]
│   ├── heading: "..."
│   ├── heading_level: 1..6
│   ├── metadata: {page, source, ...}
│   └── Paragraphs[]
│       ├── sentences[]
│       │   ├── text
│       │   ├── tokens[]
│       │   ├── keywords[]
│       │   └── ending_punctuation
│       └── keywords[]
```
 
Используем:
- **Apache Tika** — извлечение текста и структуры (headings, paragraphs, lists, tables).
- **`java.text.BreakIterator`** (JDK) — разбивка на предложения.
- **Регулярные выражения** — токенизация, очистка.
- **Стоп-лист** (RU + EN) — в ресурсах сервиса.
 
### Pass 2. Когезионная карта (Coherence Map)
 
Для каждого чанка Java считает:
- `intra_chunk_cohesion` — связность предложений внутри чанка.
- `cohesion_prev` / `cohesion_next` — пересечение keywords с соседями.
- `punctuation_completeness` — завершённость предложений.
- `chunk_ending_weight` — вес последнего знака препинания.
- `keyword_trend` — стабильность/рост/падение частоты ключевых слов.
- `repetition_score` — уникальность 3-грамм (борьба с boilerplate).
 
## 6. Разбивка длинных секций
 
### Порядок
1. **По знакам препинания.** Ищем точку, где предложение заканчивается на `.`, `!`, `?` (вес 1.0), следующее начинается с заглавной буквы / нового абзаца, и `cohesion_next` < 0.3.
2. **По падению `cohesion`.** Если знаков препинания нет, ищем место, где `cohesion_next` между предложениями резко падает (< 0.2).
3. **По `max_chunk_length`.** Если нет ни знаков, ни падения cohesion — режем по `max_chunk_length`.
4. **Если секция совсем без знаков препинания** — RAW, разбирает Go/LLM.
 
## 7. Веса для знаков препинания
 
| Знак | Вес | Смысл |
|---|---|---|
| `.` | 1.0 | Идея завершена полностью |
| `!` | 1.0 | Завершено с эмоцией |
| `?` | 1.0 | Вопрос завершён |
| `;` | 0.7 | Закончен подпункт / часть идеи |
| `:` | 0.5 | Вводится перечисление / пояснение |
| `,` | 0.1 | Идея продолжается |
| `—` / `–` | 0.3 | Отступление, ещё не конец |
| `(` / `[` | 0.2 | Начало вставки |
| `)` / `]` | 0.8 | Конец вставки (если скобки сбалансированы) |
| `«` / `»` / `"` | 0.5 | Граница цитаты |
| `...` | 0.4 | Намёк на продолжение |
| нет знака | 0.0 | Обрыв |
 
## 8. Формула smart-скоринга
 
```
score =
  has_heading * w_has_heading +
  section_boundary * w_section_boundary +
  intra_chunk_cohesion * 20 +
  punctuation_completeness * 15 +
  chunk_ending_weight * 15 +
  length_ok * 10 +
  keyword_trend_stable * 5 +
  punctuation_balance * 5 +
  metadata_present * 5
  + penalties
 
penalties:
  -15, если repetition_score < 0.5
  -15, если чанк заканчивается на `,` 
```
 
- `score >= smart_threshold` (по умолчанию 70) → **SMART**.
- `score < smart_threshold` → **RAW**.
 
## 9. Hard rules — сразу RAW
 
Если выполняется любое:
- `content` пустой.
- `content.length > 10000`.
- `title` пустой **и** не получается сгенерировать.
- Контент — не текст (одни цифры/спецсимволы).
- Парсер вернул мусор.
 
## 10. Title generation
 
Java пытается сгенерировать title в порядке:
1. Заголовок из документа.
2. First sentence, если короткий (5–20 слов) и не из стоп-слов.
3. First 100 chars (mark `title_quality_score` low).
 
## 11. HTTP-контракт с Go backend
 
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
 
## 12. SSRF-защита / валидация URL
 
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
 
## 13. Логирование и безопасность
 
- Использовать SLF4J/Logback, убрать `System.err.println`.
- Не логировать `jwt`, base64-контент, сырые документы.
- Redis внутри Docker network, не публично.
- Graceful shutdown (origin уже есть, сохранить).
 
## 14. Парсинг / чанкинг
 
- `HybridChunker` не должен терять первый короткий чанк.
- `TextDocumentParser.parseFromUrl` должен корректно обрабатывать URL (или делегировать в Tika URL parser).
- `LinkDetector` — либо реализовать, либо оставить `links: []`. Связи не создаются.
- Chunk size не должен превышать Go content limit (10000 символов).
- Только **Apache Tika + JDK** — без OpenNLP, Stanford, Lucene и т.д.
 
## 15. Docker и инфраструктура
 
- `Dockerfile` обновить под `maven-shade-plugin` → `target/app.jar`.
- Сервис должен иметь `/health` на порту 8081.
- `REDIS_URI` и `BACKEND_URL` из env.
- Позже — добавить в `docker-compose.yml`, `docker-compose.personal.yml`, `docker-compose.test.yml`.
 
## 16. Тесты
 
- WireMock: `201 Created`, `/api/v1`, `Authorization`, `data.id`.
- Unit: `HybridChunker`, `smart_score`, `TikaDocumentParser`, URL validation, `ImportDocumentHandler`, `BreakIterator` sentence splitting, keyword extraction.
- Integration: Redis inbound/outbound с plain JSON.
- Запуск: `mvn test` должен проходить.
 
## 17. Что НЕ делает сервис (повтор)
 
- Не обновляет статусы `pending`/`processing`/`done`.
- Не знает про asynq, outbox, SSE, polling.
- Не создаёт связи.
- Не обрабатывает `raw_chunks` — это зона Go backend.
- Не отправляет результат напрямую на фронт.
- Не использует внешние API синонимов / LLM.
