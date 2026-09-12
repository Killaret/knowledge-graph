# IMP-4-review: Java source-text-handler — интеграция через существующий import

**Статус:** на ревью у Claude и владельца  
**Создано:** 2026-09-11  
**Связано:** `IMP-4-import-recommendations-and-java-batch.md`, `TZ-Java-source-text-handler-2026-08-30.md`, `docs/PROJECT_REVIEW_AI_AGENTS.md` §19, `docs/AI_HANDOFF.md`

## Контекст

- Devin реализовал `GammaLinkGenerator` (лимит исходящей степени ≤2, защита от плотного графа) и временный Java-specific endpoint `POST /api/v1/import/java/batch`.
- После обсуждения с владельцем Java-specific endpoint отклонён: Java-сервис должен подключаться к **существующему** импорту, а не иметь собственный маршрут.
- `POST /api/v1/import/java/batch` удалён (handler, route, openAPI, unit test, application service).
- Осталось решить, как именно Java будет использовать существующий import, какие ручные операции нужны, и почему ручное создание связей не работает в UI.

## Открытые вопросы для ревью

### 1. API для Java: использовать существующий `import/bookmarks` или сделать generic `import/batch`?

**Вариант A — расширить существующий `POST /api/v1/import/bookmarks`**

- Сейчас `importer.Item` = `{title, url, text, type, extract_content}`.
- Java приходит с уже готовыми `title`, `content`, `source_url`, `external_id`, `metadata`.
- Плюс: один endpoint, Java — ещё один клиент.
- Минус: `url` — обязательное поле, `ProcessImportTask` делает `NormalizeURL` и `BuildContent(title, url, text)`, а Java хочет `content` как есть.

**Вариант B — generic `POST /api/v1/import/batch` (без `java` в имени)**

- Тело: `{ request_id, items[] }` где `item` = `{ title, content, type, source_url, external_id, metadata, extract_content }`.
- `url` — опционально; если `content` задан, он используется напрямую, иначе вытягиваем из `url`.
- Плюс: чистая семантика и для UI, и для Java.
- Минус: новый endpoint, но можно сделать так, чтобы старый `import/bookmarks` стал alias/deprecated.

**Вопрос к Claude/владельцу:** какой вариант выбрать? Должен ли Java вызывать `import/bookmarks` и адаптироваться под его формат, или мы делаем generic `import/batch`?

### 2. Что нужно Java-сервису из API?

По мнению владельца:

- **batch создание** заметок — нужно.
- **простое создание** (single note) — нужно.
- **удаление** заметок через import — не нужно.
- **редактирование** заметок через import — не нужно.
- **ручное создание связи** между заметками — нужно, и на стороне пользователя (UI) тоже.

**Вопрос к Claude/владельцу:**

- Single-note endpoint должен быть `POST /api/v1/notes` (как сейчас) или `POST /api/v1/import/note`?
- Batch endpoint — как в вопросе 1.
- Link endpoint: backend-точка `POST /api/v1/links` уже есть (`linkhandler.Create`), но в UI, судя по `UX-1`, создать связь из правого меню / между существующими заметками нельзя. Это задача frontend или backend?

### 3. Ручное создание связей — что не так?

- Backend: `POST /api/v1/links` работает, unit tests есть (`linkhandler/link_handler_test.go`).
- Frontend: `UX-1` — связи нельзя создать из правого меню и нельзя связать уже существующие заметки.
- Владелец говорит: "Почему-то у меня вроде было" — возможно, фича была и сломалась, или была в другом месте.

**Вопрос к Claude/владельцу:**

- Это новый frontend-функционал (панель "создать связь") или восстановление существующего?
- Где должна быть кнопка: правое меню, canvas (drag-and-drop), карточка заметки?
- Какие типы связей доступны пользователю: `related`, `reference`, `parent`, `child`, custom?

### 4. Постобработка и гамма-связи

- `GammaLinkGenerator` создан, но пока не встроен в worker.
- Java batch (или generic import batch) должен ставить тот же pipeline: `ExtractKeywords`, `ComputeEmbedding`, `RecalculateLinkWeights`.
- В идеале `ComputeEmbedding` worker после сохранения эмбеддинга вызывает `GammaLinkGenerator.GenerateForNote` (≤2 связи) и затем `RefreshRecommendations`.
- `note_links_closure` нужно обновлять асинхронно, но с лимитом ≤2 это безопасно.

**Вопрос к Claude/владельцу:**

- Согласен ли план: `ComputeEmbedding` → `GammaLinkGenerator` → `RefreshRecommendations` → `note_links_closure` refresh?
- Должен ли `note_links_closure` обновляться после каждой гамма-связи, по таймеру, или при запросе графа?

### 5. Deduplication и idempotency

- Java будет отправлять одни и те же документы/чанки повторно.
- Deduplication по `external_id` внутри batch легко. Across batches нужен поиск по `metadata->>'external_id'`.
- Сейчас `note.Repository` не ищет по метаданным. Добавить `FindByExternalID` или `FindByMetadataKey`?

**Вопрос к Claude/владельцу:**

- Делать dedup по `external_id` (из `metadata`) или по `source_url`?
- Нужен ли уникальный индекс по `external_id`?

### 6. Почему упала БД — кратко для ревью

- 108 связей, топ-6 на заметку → средняя исходящая степень ~5.7.
- `note_links_closure` рекурсивно перебирает все простые пути до глубины 10.
- 6^10 = миллионы путей → PostgreSQL/WSL завис → `wsl -t docker-desktop` → файловая система Docker VM read-only.
- Решение: `GammaLinkGenerator` ограничивает исходящую степень 2.

**Вопрос к Claude/владельцу:**

- Достаточно ли лимита 2, или нужно менять сам `note_links_closure`?
- Нужен ли `statement_timeout` в worker при обновлении view?

## Предложения на обсуждение

1. **Удалённый `import/java/batch` не возвращаем.** Вместо него — generic `import/batch` (вариант B) или расширение `import/bookmarks`.
2. **Manual link creation** — скорее всего frontend-задача, но надо уточнить UX.
3. **Post-processing pipeline** — `ComputeEmbedding` worker должен вызывать `GammaLinkGenerator` и `RefreshRecommendations`.
4. **Deduplication** — добавить `FindByExternalID` в `note.Repository` и индекс по `metadata->>'external_id'`.

## Решения, которые нужно принять

- [ ] Выбрать endpoint: `import/bookmarks` vs `import/batch` vs оставить как есть.
- [ ] Утвердить список Java-операций: batch create, single create, no delete/edit.
- [ ] Определить UX ручного создания связей.
- [ ] Утвердить pipeline постобработки с `GammaLinkGenerator`.
- [ ] Решить, как часто обновлять `note_links_closure`.

## Артефакты

- `backend/internal/application/recommendation/gamma_link_generator.go` — готов.
- `backend/internal/application/recommendation/gamma_link_generator_test.go` — готов.
- `backend/internal/infrastructure/db/postgres/embedding_repo.go` — `FindSimilarNotes` / `FindSimilarNotesBatch` теперь возвращают `[0,1]`.
- `docs/PROJECT_REVIEW_AI_AGENTS.md` §19 — анализ падения БД.
- `docs/AI_HANDOFF.md` — статус `IMP-4` обновлён.
