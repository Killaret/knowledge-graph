# BATCH-API: дизайн batch-роутов для notes/links

## Статус

**2026-09-12.** Реализовано Devin. Пользовательские batch-роуты (`/notes/batch/create`, `/notes/batch/delete`) и import batch (`/import/batch`) добавлены, старый `POST /notes/batch` удалён, OpenAPI обновлено. Добавлены позитивные и жёстко негативные тесты на все три роута с проверкой мутаций в мок-репозиториях. В ходе тестирования найден и исправлен дефект: `/import/batch` не проверял, что `FindByID` вернул `nil`, и мог создавать связи на несуществующие заметки; пустой массив `notes` теперь возвращает 400. Ждёт ревью Claude Code / владельца, в первую очередь по контракту ссылок в `/import/batch`.

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
- Старый `POST /api/v1/notes/batch` удалён вместо deprecated.
- Фронтенд (`frontend/src/shared/api/notes.ts:deleteNotesBatch`) уже переехал на `v1/notes/batch/delete`.

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

## Реализация и покрытие тестами

- `POST /api/v1/notes/batch/create`: max 50, best-effort, валидация title/content/type, пустой массив → 400, ответ с `notes` и `failed`.
- `POST /api/v1/notes/batch/delete`: валидация UUID, 400 на пустой `ids`, 400 на malformed UUID, 404 при чужой заметке, 204 No Content, no-op для missing IDs.
- `POST /api/v1/import/batch`: синхронный best-effort, сначала notes, потом links, клиент может задать `id` заметки для ссылок внутри одного запроса; защита от перезаписи существующих `id`; пустой `notes` → 400.
- Исправлен дефект: связи не создаются, если `source_note_id` или `target_note_id` не существуют и не были созданы в том же запросе.
- Тесты: `backend/internal/interfaces/api/notehandler/note_handler_test.go`, `backend/internal/interfaces/api/notehandler/note_handler_import_test.go`, вспомогательный `mockLinkRepo` в `backend/internal/interfaces/api/notehandler/mock_repo.go`.
- Прогоны: `go test ./...`, `go vet ./...`, `npm run test:unit -- --run`, `npm run check`, `npm run lint` — зелёные.

## Открытые вопросы для обсуждения

1. **Контракт ссылок в `/import/batch` (ключевой):** внешний Java/source-text сервис не имеет UUID создаваемых заметок. Текущий вариант с клиентскими `id` работает, но неудобен для клиента. Возможные альтернативы:
   - ссылки по индексам массива `notes`;
   - ссылки по внешнему/source-идентификатору, который возвращается в ответе с маппингом на UUID;
   - упорядоченные операции ("create note, then create link ...") внутри одного batch.
   Решение требуется от Claude Code / владельца.
2. **Rate limit:** `writeLimiter` достаточно? Нужен ли отдельный лимит для batch?
3. **Авторизация import batch:** JWT + API key? Отдельный `X-API-Key` или тот же middleware?
4. **Типы импортируемых заметок:** ограничить `UI_TYPES` и `CelestialBody.UI_TYPES` как в IMP-1/IMP-3?

## TDD — предложение по процессу

Текущую batch-реализацию уже сделали, поэтому ретроспективно добавлены регрессионные тесты. Для будущих фич (особенно нового контракта ссылок) можно применить TDD:
1. Написать падающий тест на желаемый контракт/поведение.
2. Реализовать минимальный код, который заставляет тест проходить.
3. Отрефакторить, сохраняя зелёные тесты.
Это на обсуждение с Claude Code / владельцем.

## Связанные issue/PR

- Issue #37: <https://github.com/Killaret/knowledge-graph/issues/37>
- IMP-4: <https://github.com/Killaret/knowledge-graph/docs/tasks/IMP-4-import-recommendations-and-java-batch.md>
- `backend/cmd/server/router.go`: <ref_file file="D:\knowledge-graph\backend\cmd\server\router.go" />
- `backend/internal/interfaces/api/notehandler/note_handler.go`: <ref_file file="D:\knowledge-graph\backend\internal\interfaces\api\notehandler\note_handler.go" />
- `backend/internal/interfaces/api/notehandler/note_handler_test.go`: <ref_file file="D:\knowledge-graph\backend\internal\interfaces\api\notehandler\note_handler_test.go" />
- `backend/internal/interfaces/api/notehandler/note_handler_import_test.go`: <ref_file file="D:\knowledge-graph\backend\internal\interfaces\api\notehandler\note_handler_import_test.go" />
- `backend/openAPI.yaml`: <ref_file file="D:\knowledge-graph\backend\openAPI.yaml" />

## Следующее действие

Claude Code / владелец ревьюирует реализацию и тесты, решает контракт ссылок в `/import/batch` и даёт постановку по дальнейшей доработке. Devin внедряет после принятой постановки.
