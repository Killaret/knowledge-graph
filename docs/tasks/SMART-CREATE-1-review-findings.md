# SMART-CREATE-1 — разбор реализации (Devin, 2026-10-18)

Исполнительские заметки к постановке
[`SMART-CREATE-1-document-button-via-source-text-handler.md`](SMART-CREATE-1-document-button-via-source-text-handler.md).
Вердикт ревью пишет Claude Code — здесь только доказательная база.

## Что сделано

### Бэкенд

- `internal/application/import/document.go` — порт `DocumentQueue`
  (alive/push/BLPOP-ответы), `SubmitDocument`, `DocumentAvailability`,
  `GetTaskStatusForUser`, `RunDocumentResultLoop`, `ApplyDocumentResult`,
  `failTimedOutTasks`.
- `internal/infrastructure/queue/document_queue.go` — адаптер `*redis.Client`
  → порт: `EXISTS source-text-handler:alive`, `RPUSH import:document:pending`,
  `BLPOP import:responses:pending` (timeout → `ErrResultTimeout`).
- `internal/interfaces/api/notehandler/document_import.go` —
  `GET /api/v1/import/document/availability` → `{available: bool}`;
  `POST /api/v1/import/document` (multipart, `type=TEXT|URL|FILE`).
- Статус документ-задач живёт в кеше (`import:task:<id>`, TTL 1 ч, как массовый
  импорт), поле `kind="document"` + `user_id` владельца. `task_id` ответа =
  `correlation_id` задачи.
- `GET /api/v1/import/:task_id/status` переведён на `GetTaskStatusForUser`:
  задачи с `user_id` отдаются только владельцу, чужие — 404. Задачи массового
  импорта без `user_id` сохраняют прежнее поведение (обратная совместимость —
  см. «Решения и ограничения»).
- Читатель ответов запускается в `cmd/worker` как горутина
  `RunDocumentResultLoop(ctx, popTimeout=15s, taskTimeout=10m)`; в `cmd/server`
  очередь тоже собирается (нужна для push), читатель там не стартует.
- Таймаут — чисто Go-сторона: по каждому BLPOP-таймауту `failTimedOutTasks`
  сканирует `import:task:*` и гасит pending-документы старше 10 минут в
  `failed`/`reason="timeout"`. Критерий 11.

### Фронтенд

- `GraphTopBar.svelte` — кнопка `data-testid="smart-create-button"` непосредственно
  перед `create-note-button` в `right-cluster`; рендерится только при
  `isAuthenticated && onSmartCreate`. Опрос `GET /import/document/availability`
  каждые 15 с (первый — сразу при маунте) — активность в пределах 15 с после
  появления heartbeat и до 30 с после исчезновения (период 15 с + ключ EX 15).
  `aria-disabled` + класс `--disabled` + tooltip «недоступен»; клик по недоступной
  — no-op (но клик по `pending`/`unread` открывает модал — индикаторы кликабельны).
- `SmartCreateModal.svelte` — три плитки (текст/URL/файл), состояния
  `pending → done (Создано заметок: N + заголовки) | failed (+errors) | timeout`.
  Сворачивание → индикатор `smart-create-waiting`; по завершении — бейдж «1»
  (`smart-create-badge`), повторный клик открывает результат.
- Проводка: `home-page.svelte.ts` (состояние/аксессоры) → `GraphPageShell` /
  `CosmicCockpitLayout` → `GraphTopBar`; модал смонтирован в `routes/+page.svelte`
  и `routes/graph/+page.svelte` рядом с существующей `CreateNoteModal`.
- `shared/api/import.ts` — `getDocumentAvailability`, `submitDocument`
  (multipart через ky, `FormData`), `getImportStatus`.
- i18n `import.smartCreate*` — кнопка, недоступность, плитки, статусы, ошибки.

## Прогоны

- `go test ./...` (backend) — 57 пакетов ok, FAIL нет.
- `svelte-check` — 0 errors / 0 warnings.
- `npm run check:circular` — чисто (249 файлов).
- Vitest: `GraphTopBar.spec.ts` 20/20, `SmartCreateModal.spec.ts` 7/7,
  связанные наборы (`widgets/graph-page`, `routes/graph`, `graph-ui`) 59/59.
- Границы слоёв: domain без framework-импортов; FSD-проверки shared/entities
  пустые; `infrastructure/queue` → `application/import` — порт внутри, адаптер
  снаружи (направление внутрь, норма).

## Мутации (по критериям приёмки)

| Мутация | Краснеет |
|---|---|
| Убрана проверка `IsListenerAlive` в `SubmitDocument` | `TestSubmitDocument_NoListener` + `TestImportDocument_NoListener` |
| `user_id` → `userId` в payload | `TestSubmitDocument_TaskContract`, `TestDocumentQueue_TaskContractOnRedis` |
| `PARTIAL`: `RawChunksCount` занулён | `TestApplyDocumentResult_Partial` |
| Убрана проверка владельца в `GetTaskStatusForUser` | `TestGetTaskStatusForUser_Foreign` |
| Убрана проверка возраста в sweep-таймауте | `TestFailTimedOutTasks` |
| В лог unknown-task добавлен `string(payload)` | `TestApplyDocumentResult_LogsNoSecrets` |
| Убран SSRF-фильтр `IsAllowedURL` | `TestImportDocument_RejectsPrivateURL` |
| Убран лимит файла (оба сравнения) | `TestImportDocument_FileOverLimit` |
| Кнопка перенесена ПОСЛЕ «＋» (внутрь его `{#if}`) | «renders the smart create button before the create note button» + 3 смежных |
| `aria-disabled` → всегда `false` | «is grey and inert while the document handler is unavailable» |

Все мутации откачены; после каждого отката набор снова зелёный.

## Решения и ограничения

1. **Обратная совместимость статусов массового импорта.** Задача требует 404
   чужому для документ-задач. `GetTaskStatusForUser` отдаёт 404 любой задаче с
   `user_id`, кроме владельца. Старые bookmark-задачи `user_id` не имеют — для
   них поведение прежнее (читаемость всем авторизованным), иначе сломался бы
   существующий массовый импорт.
2. **Лимит файла 8 MiB** — в ТЗ лимита нет; взят ниже `client_max_body_size 10m`
   nginx, чтобы multipart не резался прокси. Зафиксировано в openAPI.
3. **`kind`/`user_id`/`notes`/`raw_chunks_count`/`errors`/`reason`/`created_at`
   добавлены в `ImportTaskStatus` как опциональные поля** — существующие
   клиенты bookmark-статуса не затронуты (omitempty).
4. **JWT в payload §3** берётся из Authorization-заголовка запроса и кладётся
   в задачу как есть. Обновление токена перед сдачей в Java не реализовано —
   открытый вопрос владельца (Java держит один токен на весь пакет `/import/batch`,
   длинный разбор Tika может пережить TTL). Записано в постановке как дефолт
   «не делать».
5. **Логи:** в читателе логируются только correlation_id/служебные поля; JWT и
   тело документа в логи не пишутся — `TestApplyDocumentResult_LogsNoSecrets`.
6. **Heartbeat-писатель — на стороне Java** (дополнение к ТЗ для Alximac, ключ
   `source-text-handler:alive`, `EX 15`, рефреш ≤ 5 с). Без него кнопка серая и
   `POST` отвечает 503 — это контрактная задумка, не дефект.
7. **Живая проверка на тест-стеке не выполнялась** — по протоколу верификация
   живого стека за рецензентом. Заглушка слушателя для прогона: писать ключ,
   `BLPOP import:document:pending`, отвечать в `import:responses:pending`
   фикстурами §4.

## Файлы

Бэкенд: `application/import/document.go(+test)`, `infrastructure/queue/document_queue.go(+test)`,
`interfaces/api/notehandler/document_import.go(+test)`, правки в `service.go`,
`note_handler.go`, `cmd/server/{main,router}.go`, `cmd/worker/main.go`,
`openAPI.yaml`, `docs/api/API_EN.md`.
Фронт: `features/graph-ui/GraphTopBar.svelte(+spec)`, `widgets/notes/SmartCreateModal.svelte(+spec)`,
`features/home-page/home-page.svelte.ts`, `widgets/cosmic-cockpit/CosmicCockpitLayout.svelte`,
`widgets/graph-page/GraphPageShell.svelte`, `routes/+page.svelte`, `routes/graph/+page.svelte`,
`shared/api/import.ts`, `shared/utils/i18n/messages/import.ts`.
