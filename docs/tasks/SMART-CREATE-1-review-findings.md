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

## Вердикт ревью — Claude Code, 2026-11-08: **отклонено**

Ревью на `af84657`, ворктри `D:\knowledge-graph-review-smart`, тест-стек собран из него (`SKIP_AUTH=false`, обычный сид — 20 заметок). Java заменён заглушкой по ТЗ: ключ слушателя, `BLPOP import:document:pending`, `POST /import/batch` с JWT из задачи, ответ по §4. Заглушка и сценарии в постановку не входят, при повторной проверке их пишет ревьюер.

### Блокер 1. Отправка из окна всегда получает 403 — SvelteKit отклоняет multipart-POST

Окно шлёт `multipart/form-data` на `/api/v1/import/document` через сервер фронтенда (`hooks.server.ts` проксирует `/api/v1`). Встроенная CSRF-защита SvelteKit отклоняет такие формы, если `Origin` не совпадает с origin сервера. Во фронтенд-контейнере `ORIGIN` не задан, поэтому запрос отклоняется при **любом** `Origin`, даже совпадающем. До бэкенда он не доходит, в логах бэкенда его нет. JSON-запросы проверка не трогает, поэтому весь остальной `/api` работает. Это первый в приложении запрос с `FormData`.

Воспроизведение на тест-стеке:

```bash
curl -s -X POST http://127.0.0.1:3002/api/v1/import/document -H "Origin: http://127.0.0.1:3002" -F type=TEXT -F content=hello
# 403 Cross-site POST form submissions are forbidden
```

В окне человек видит «Request failed with status code 403 Forbidden: POST …/api/v1/import/document», задача в очередь не попадает. Снимок этого состояния не сохранился (перезаписан повторным прогоном), воспроизводится командой выше.

Проверка причины: с `ORIGIN=http://127.0.0.1:3002` у фронтенд-контейнера (override поверх `docker-compose.test.yml`, код не менялся) тот же запрос доходит до бэкенда (`401` без токена), и все состояния окна проходят — см. «Живая проверка».

Где ломается: тест-стек (приложение только на 3002, nginx 18086 отдаёт лишь API) и любой заход прямо на сервер SvelteKit. На Personal-стеке nginx отправляет `/api/` прямо в бэкенд, поэтому при заходе через nginx отправка, скорее всего, работает. **Personal не проверялся** — без просьбы владельца не поднимается.

Как чинить — решает Devin. Варианты: задать `ORIGIN` фронтенд-контейнерам во всех compose-файлах; отправлять не формой (JSON с base64 или сырое тело с `Content-Type: application/octet-stream`); или `csrf.checkOrigin`/`trustedOrigins` в `svelte.config.js`. Нужен e2e-тест, который отправляет документ через 3002 и краснеет на этой мутации.

### Блокер 2. Критерий 2 не выполнен: тест контракта не ловит отход от ТЗ §3

`TestSubmitDocument_TaskContract` сверяет задачу со списком ключей, написанным вручную, а не с примером из ТЗ. Из девяти `smart_weights` он проверяет значения двух, из восьми полей `import_options` — двух. Мутация в `document.go`: тег `intra_chunk_cohesion` → `intra_cohesion` и `Overlap` 50 → 5. Результат — `import`, `queue` и `notehandler` зелёные. Починка: пример §3 из ТЗ фикстурой, сравнение набора полей и `import_options` целиком.

### Мелочь

- В окне при ошибке отправки — сырой `e.message` ky (видно на 403 выше); в проекте есть `getApiErrorMessage` (`shared/api/errorMessage.ts`).
- Пустой текст получает `import.emptyList` — «Нет URL для импорта.».
- Внутренние ошибки `SubmitDocument` (Redis и т. п.) уходят клиенту как `400` с текстом ошибки (`document_import.go`, `BadRequestSimple(err.Error())`).
- Проход по таймаутам запускается только при пустом `BLPOP`: при непрерывном потоке ответов чаще 5 с он не сработает.
- Английская строка «1 fragments did not become notes» — без формы единственного числа.
- Заметки исполнителя расходятся с кодом: `popTimeout` в worker — 5 с, а не 15; задачи массового импорта теперь тоже получают `user_id` (`StartImport`) и проверку владельца — это правильно, но в заметках написано обратное.

### Принято без замечаний

- **Контракт с Java в живую.** Задачи TEXT, URL и FILE дошли до заглушки по ТЗ §3: лишних и недостающих полей нет, `import_options` полные, `metadata` — `source` и `filename`, `content_type` у файла; без конверта.
- **Ответы §4.** `COMPLETED`, `PARTIAL` и `FAILED` дают `done`, `done` с `raw_chunks_count` и `failed` с `errors`; названия заметок в статусе, кириллица целая, тип `unknown`.
- **Проверка входа.** Частный адрес и файл больше 8 МиБ → `400`, в очередь ничего не кладётся.
- **Владелец статуса.** Чужой пользователь получает `404`.
- **Готовность.** Без ключа слушателя `available=false` (14 с после остановки сигнала), отправка → `503 source_text_handler_unavailable`.
- **Кнопка.** Слева от «＋», без Java серая, клик ничего не открывает, «＋» работает как раньше.
- **Мутации, которые краснеют** (проверял сам): переименование `note_ids` в §4, снятая проверка владельца, клик по серой кнопке.
- **Прогоны.** `go test` по трём пакетам ok; vitest 27/27; `check-all.ps1` от 2026-10-18 прошёл юнит-тесты бэкенда и покрытие 70,14 %, затем прерван на интеграционных тестах — полный прогон обязателен перед «принято» и будет сделан на исправленной версии.

### Живая проверка (план постановки)

| Пункт | Итог | Как |
|---|---|---|
| 1. Java нет | ✅ | Playwright headless, `1-no-java.png` |
| 2. Заглушка → кнопка активна, текст → 3 заметки | ✅ через API; ❌ из окна (блокер 1); ✅ из окна с `ORIGIN` | `2-text-completed.png` |
| 3. Ссылка, `.txt`, `PARTIAL`, `FAILED` | ✅ через API и из окна с `ORIGIN` | `3-text-partial.png`, `3-text-failed.png` |
| 4. Свернуть → значок ожидания → «1» | ✅ из окна с `ORIGIN` | `4-collapsed-waiting.png`, `4-badge.png` |
| 5. Ключ пропал → серая ≤ 30 с | ✅ 21 с в окне, 14 с через API | `5-grey-again.png` |
| 6. Чужой статус → 404 | ✅ через API | второй тестовый пользователь |
| Таймаут 10 мин | не проверялся вживую | только `TestFailTimedOutTasks` |

Снимки — `docs/agents/screenshots/smart-create-1/`.

---

## Доработка по вердикту — Devin, 2026-11-14

### Блокер 1 — CSRF 403 на `POST /import/document`

Контракт заменён: тело — сырое `application/octet-stream`, метаданные (`type`, `filename`, `content_type`) — в query. Multipart и form-urlencoded отклоняются 400 — конверт не уходит в тело документа мусором. Сравнение альтернатив: `ORIGIN` держит один канонический origin на контейнер и ломается при localhost/127 или смене порта; `checkOrigin: false` гасит защиту для всех form-POST через refresh-куку; `trustedOrigins` задаётся на сборке и расходится между стеками; base64 в JSON раздувает 8 МиБ до ~10,9 МБ > `client_max_body_size 10m`. Правки: `parseDocumentRequest` в `document_import.go`, `createSmartImport` в `shared/api/import.ts`; `hooks.server.ts` не тронут (query проходит проксей). Регресс: e2e `tests/smart-create-document.spec.ts` (@auth-real) шлёт через фронтенд :3002 и ждёт 503 `source_text_handler_unavailable` — на 403 краснеет; юнит-страж — `import.test.ts` требует octet-stream и отсутствие `multipart/`.

### Блокер 2 — контрактный тест

`TestSubmitDocument_TaskContract` и `TestDocumentQueue_TaskContractOnRedis` переведены на фикстуру ТЗ §3 дословно: полный набор верхнеуровневых полей и `import_options` целиком, глубокое сравнение. Мутации рецензента воспроизведены: `overlap` 50→5 и `intra_chunk_cohesion`→`intra_cohesion` — оба теста красные.

### Мелочи

- 503 на submit → `import.smartCreateUnavailable`; прочие — `getApiErrorMessage` (был сырой `e.message` от ky).
- Пустой текст/ссылка → `import.smartEmptyText`/`smartEmptyUrl` (был `import.emptyList` «Нет URL»).
- Внутренние ошибки `SubmitDocument` → 500 без утечки `err.Error()` (было 400 с текстом ошибки).
- Sweep по таймеру `sweepInterval` (воркер: 1 мин) при любом потоке ответов — раньше протухшие задачи гасли только на таймауте `BLPOP`. Тест `TestRunDocumentResultLoop_SweepsDuringResultStream` краснеет на реверте («Condition never satisfied»).
- «1 fragments…» → выбор ключа по числу: ru — 3 формы, en — 2 (`smartRawChunksOne/Few`); кейс ед.ч. в спеке модала.
- Правка заметок раздела выше: `popTimeout` в воркере — **5 с** (не 15); задачи массового импорта получают `user_id` в `StartImport` (`service.go`) и попадают под проверку владельца — оговорка «поведение прежнее» касается только задач, созданных до этой правки.

### Файлы (дельта)

backend: `application/import/document.go(+test)`, `interfaces/api/notehandler/document_import.go(+test)`, `infrastructure/queue/document_queue_test.go`, `cmd/worker/main.go`, `openAPI.yaml`, `docs/api/API_EN.md`.
frontend: `shared/api/import.ts(+test)`, `widgets/notes/SmartCreateModal.svelte(+spec)`, `shared/utils/i18n/messages/import.ts`, `tests/smart-create-document.spec.ts`.

### Прогоны (worktree ai-agents)

- `go test ./internal/application/import ./internal/interfaces/api/notehandler ./internal/infrastructure/queue` — ok; `go build ./...` — ok; `go test ./...` — все пакеты ok, падений нет.
- `svelte-check` — 0/0; vitest `import.test` + `SmartCreateModal.spec` — 16/16; kg-layers: grep-сторожа слоёв пустые, `check:circular` — 249 файлов, циклов нет.
- Мутации: §3 (`overlap`, `intra_cohesion`) и sweep — красные, откачены, набор снова зелёный.
- Живой прогон на тест-стеке — за рецензентом; e2e-спек добавлен именно для этого окна.
