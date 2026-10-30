# IMP-4: единый конвейер рекомендаций для импорта и batch-метод для Java source-text-handler

**Статус:** новая — постановка от владельца, ждёт обсуждения и приоритета.  
**Дата:** 2026-09-10  
**Источник:** ручной прогон Personal-стека и обсуждение Java source-text-handler.

## Суть

Сейчас импорт закладок и массовый импорт проходят **разные** post-processing-пути:

- Ручное создание (`POST /notes`) и bookmarklet (`POST /notes/bookmarklet`) вызывают `enqueueRecommendationTasks()` и обновляют `note_recommendations`.
- Массовый импорт (`ProcessImportTask` в `backend/internal/application/import/service.go`) ставит в очередь `ExtractKeywords`, `ComputeEmbedding` и `RecalculateLinkWeights`, но **не ставит** `RefreshRecommendations`.

Из-за этого заметки, созданные массовым импортом, не участвуют в том же конвейере рекомендаций, что остальные.

## Требования

### 1. Единый конвейер для всех путей создания заметки

Любой способ появления заметки (ручная, букмарклет, массовый импорт, Java source-text-handler) должен запускать:

- `ExtractKeywords`
- `ComputeEmbedding`
- `RecalculateLinkWeights`
- `RefreshRecommendations`

Точки кода:
- `backend/internal/interfaces/api/notehandler/note_handler.go` (`Create`, `Bookmarklet`)
- `backend/internal/application/import/service.go` (`ProcessImportTask`)
- `backend/internal/infrastructure/queue/task_queue.go` / `Enqueue*`

### 2. Batch-метод создания заметок

Java `source-text-handler` обрабатывает URL/документ, извлекает несколько «умных» чанков и должен уметь быстро создать несколько заметок.

- Go backend должен предоставить `POST /api/v1/notes/batch`.
- Тело:
  ```json
  {
    "notes": [
      { "title": "...", "content": "...", "type": "unknown", "metadata": {...} }
    ]
  }
  ```
- Ответ: `201 Created` с массивом `data[].id`.
- До появления batch Java создаёт по одному через `POST /api/v1/notes`.
- При batch-создании backend должен запускать тот же конвейер, что и при одиночном создании.

### 3. Java source-text-handler — контроль рекомендаций

После создания заметок (одиночно либо batch) Java-сервис должен **проверить**, что backend поставил задачи на обновление:

- В ответе на `POST /notes` / `POST /notes/batch` или через `GET /import/:id/status` backend возвращает либо `recommendations_task_queued`, либо список enqueued task IDs.
- Либо Java делает `GET /notes/:id/suggestions` после небольшой задержки и убеждается, что рекомендации посчитались (даже пустые — с `X-Recommendations-Source`).

### 4. Разделение ответственности

- **Java** — только извлечение, чанкинг, языковая детекция, `content_clean`, `raw_chunks`, `smart_score`, формирование `title`/`content` для заметок.
- **Go backend** — приём `ImportResult`, обновление `import_tasks`, создание/обновление `raw_chunks`, persist notes, постобработка (keywords, embeddings, recommendations, links).

### 5. Регрессия

- Тест: массовый импорт → все заметки имеют `note_embeddings`, `note_keywords` и `note_recommendations` (или хотя бы запущены соответствующие задачи).
- Тест: `POST /notes/batch` → массив `data[].id` + задания в очереди.
- Тест: Java source-text-handler с WireMock проверяет, что `data[].id` не пустые и ответ содержит признак post-processing.

## Связанные файлы

- `backend/internal/application/import/service.go`
- `backend/internal/interfaces/api/notehandler/note_handler.go`
- `backend/internal/infrastructure/queue/worker.go`
- `TZ-Java-source-text-handler-2026-08-30.md` §10 «HTTP-контракт с Go backend» и §18 «URL-импорт"

## Что сделано 2026-09-11

- Вручную заполнены `note_recommendations` (108) и `links` (18) из `note_embeddings` по косинусному сходству (top-6 для рекомендаций, top-1 для связей), `REFRESH MATERIALIZED VIEW note_links_closure`.
- Проверено на живом API: `GET /notes/:id/suggestions` возвращает 6 кандидатов со score; `GET /notes/:id/graph` возвращает узлы и связи.
- Обнаружен и исправлен дефект пустого `title` в precomputed-ветке `GetSuggestions` (`note_handler.go:1087`), тесты обновлены.
- Проанализировано падение Docker: `note_links_closure` с рекурсивным CTE взрывается на плотных графах. Добавлен `GammaLinkGenerator` (`backend/internal/application/recommendation/gamma_link_generator.go`), который строго лимитирует исходящую степень гамма-связей и покрыт юнит-тестами.
- Исправлен `EmbeddingRepository.FindSimilarNotes` / `FindSimilarNotesBatch`: теперь возвращают нормализованный `[0,1]` score.
- Java-specific endpoint (`POST /api/v1/import/java/batch`) реализован и затем **удалён** по решению владельца: Java будет подключаться к существующему импорту, а не иметь собственный маршрут. Подробные вопросы ревью вынесены в [`IMP-4-claude-review.md`](IMP-4-claude-review.md).
- Docker Desktop восстановлен владельцем после перезагрузки.