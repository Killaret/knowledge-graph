# BATCH-API: дизайн batch-роутов для notes/links

## Статус

**2026-09-12.** Реализовано Devin. Пользовательские batch-роуты (`/notes/batch/create`, `/notes/batch/delete`) и import batch (`/import/batch`) добавлены, старый `POST /notes/batch` удалён, OpenAPI обновлено. Ждёт ревью Claude Code / владельца.

## Контекст

- Issue #37 (`Add bulk method`) требует batch-метод для Java/source-text handler и массового импорта.
- Сейчас `POST /api/v1/notes/batch` — это **удаление** пачки заметок, что сбивает с толку.
- Нет синхронного batch-создания заметок и связей.
- `POST /api/v1/import/bookmarks` — асинхронный импорт только закладок.

## Пользовательские batch-роуты (`/api/v1/notes`)

Цель: довести `NoteBatch` до ума — два явных POST-метода, оба с массивом в теле.

### 1. `POST /api/v1/notes/batch/create`

Создаёт пачку заметок **синхронно**.

**Request:**
```json
{
  "notes": [
    {
      "title": "...",
      "content": "...",
      "type": "star",
      "metadata": {}
    }
  ]
}
```

**Лимиты:**
- max 50 заметок за запрос;
- title max 200, content max 50000, type из `oneof`.

**Response 201:**
```json
{
  "notes": [
    {
      "id": "...",
      "title": "...",
      "content": "...",
      "type": "...",
      "metadata": {},
      "is_public": false,
      "created_at": "...",
      "updated_at": "..."
    }
  ],
  "failed": []
}
```

**Response 400 (валидация):**
```json
{
  "errors": [
    {"index": 0, "field": "title", "reason": "required", "message": "..."}
  ]
}
```

**Атомарность:** при обсуждении выбрать один из вариантов:
- **A. All-or-nothing** — любая ошибка валидирует весь запрос, ничего не создаётся.
- **B. Best-effort** — создаются валидные, возвращается `failed: []` с ошибками по индексам. Требует явного определения, что делать с постобработкой (keywords, embeddings, рекомендации) для частичного набора.

**Постобработка:** для каждой созданной заметки запускать тот же конвейер, что и при `POST /notes`:
- `EnqueueExtractKeywords`
- `EnqueueComputeEmbedding`
- `EnqueueRecalculateLinkWeights`
- `enqueueRecommendationTasks`
- публикация `NoteCreated`
- инвалидация `graphCache` по `userID` один раз в конце.

### 2. `POST /api/v1/notes/batch/delete`

**Явное** удаление пачки заметок.

**Request:**
```json
{
  "ids": ["uuid-1", "uuid-2"]
}
```

**Поведение:**
- Валидирует все UUID.
- Проверяет владение (SEC-1): если есть чужая заметка — 404, ничего не удаляется.
- Возвращает 204 No Content (как сейчас) **или** 200 с `{ "deleted": [...], "not_found": [...] }`.

**Backward compatibility:**
- Старый `POST /api/v1/notes/batch` остаётся как alias на `delete` до перехода фронтенда, но помечается `deprecated`.
- Фронтенд (`frontend/src/shared/api/notes.ts:deleteNotesBatch`) потом переехать на `notes/batch/delete`.

## Import batch-роуты (`/api/v1/import`)

Цель: дать Java/source-text handler и другим внешним источникам синхронно создать пачку заметок и связей, **без delete**.

### Вариант A — один объединённый роут

`POST /api/v1/import/batch`

**Request:**
```json
{
  "notes": [
    {
      "title": "...",
      "content": "...",
      "type": "star",
      "is_public": false,
      "metadata": {},
      "source_url": "https://..."
    }
  ],
  "links": [
    {
      "source_note_id": "...",
      "target_note_id": "...",
      "link_type": "reference",
      "weight": 0.8
    }
  ]
}
```

**Порядок:**
1. Создать все заметки.
2. Создать связи, указывающие на успешно созданные заметки.

**Response 200:**
```json
{
  "created_notes": [{"id": "...", "title": "..."}],
  "created_links": [{"id": "..."}],
  "failed_notes": [{"index": 0, "error": "..."}],
  "failed_links": [{"index": 0, "error": "..."}]
}
```

**Плюсы:** один вызов для Java, атомарность логики импорта в одном месте.
**Минусы:** смешаны notes и links, сложнее валидация/ответ.

### Вариант B — два отдельных роута

- `POST /api/v1/import/batch/notes`
- `POST /api/v1/import/batch/links`

**Плюсы:** проще, REST-чистота, легче тестировать.
**Минусы:** Java делает два вызова; клиент должен сам сохранять созданные `note_id` между вызовами.

### Рекомендация Devin

**Вариант A** (`POST /api/v1/import/batch`) — потому что Java-источник, похоже, хочет передать один payload с заметками и связями, а не вести state между двумя запросами.

## Открытые вопросы для обсуждения

1. **Атомарность create/delete batch:** all-or-nothing или best-effort?
2. **Rate limit:** `writeLimiter` достаточно? Нужен ли отдельный лимит для batch?
3. **Постобработка при batch-create:** запускать keywords/embeddings для каждой заметки сразу или одну фоновую `RefreshRecommendations`?
4. **Ответ delete batch:** оставить 204 или вернуть JSON с результатами?
5. **Backward compatibility:** оставить `POST /notes/batch` на delete или сразу ломать и обновить frontend?
6. **Import batch:** вариант A или B?
7. **Авторизация import batch:** JWT + API key? Отдельный `X-API-Key` или тот же middleware?
8. **Типы импортируемых заметок:** ограничить `UI_TYPES` и `CelestialBody.UI_TYPES` как в IMP-1/IMP-3?

## Связанные issue/PR

- Issue #37: <https://github.com/Killaret/knowledge-graph/issues/37>
- IMP-4: <https://github.com/Killaret/knowledge-graph/docs/tasks/IMP-4-import-recommendations-and-java-batch.md>
- `backend/cmd/server/router.go`: <ref_file file="D:\knowledge-graph\backend\cmd\server\router.go" />
- `backend/internal/interfaces/api/notehandler/note_handler.go`: <ref_file file="D:\knowledge-graph\backend\internal\interfaces\api\notehandler\note_handler.go" />
- `backend/openAPI.yaml`: <ref_file file="D:\knowledge-graph\backend\openAPI.yaml" />

## Следующее действие

Claude Code проводит ревью дизайна, закрывает открытые вопросы и готовит постановку. Devin реализует после принятой постановки.
