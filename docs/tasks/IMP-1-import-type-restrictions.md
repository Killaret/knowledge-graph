# IMP-1. Ограничение типов заметок в импорте закладок

Постановка для Devin. Порядок работы: [`../AI_AGENT_PROTOCOL.md`](../AI_AGENT_PROTOCOL.md).

Ставит пользователь/Claude Code, реализует Devin, проверяет Claude Code на живом Personal-стеке.

## Зачем

В массовом импорте закладок (`/import/bookmarks`) селектор типа заметки показывает все возможные типы, включая `technical`, `unknown` и аномалии (`reality_rift`, `chromatic_maw`, `void_whisper`, `cosmic_abomination`). В обычном создании заметки (модалка, форма призрака) используется `CelestialBody.UI_TYPES`, который исключает не-UI типы. Импорт нарушает это правило.

## Что выяснено в коде

1. `frontend/src/routes/import/bookmarks/+page.svelte:22-39` содержит жёстко закодированный массив `noteTypes`, в который включены все типы.
2. `frontend/src/entities/shared/model/celestial-body.ts:517` экспортирует `CelestialBody.UI_TYPES`, отфильтрованный по `isUi`.
3. `frontend/src/widgets/notes/CreateNoteModal.svelte:133` и `frontend/src/features/graph-ui/modals.svelte:72` уже используют `CelestialBody.UI_TYPES` для селектора.
4. `POST /api/v1/import/bookmarks` и `POST /api/v1/import/bookmarklet` (OpenAPI `BookmarkletRequest`/`ImportBookmarksItem`) разрешают все типы, включая не-UI. Это позволяет обойти ограничение через прямой вызов API.

## Что сделать

### 1. UI массового импорта
- Заменить hardcoded `noteTypes` в `frontend/src/routes/import/bookmarks/+page.svelte` на `CelestialBody.UI_TYPES.map((b) => b.type)`.
- Убедиться, что i18n-лейблы `filter.type.<type>` по-прежнему рендерятся корректно.

### 2. Единообразие в single import
- Проверить, есть ли в `/import` (bookmarklet landing) возможность выбора типа. Если нет — оставить `asteroid` по умолчанию; не нужно добавлять селектор без запроса владельца.

### 3. Backend-защита от обхода (опционально, обсудить с владельцем)
- Добавить в `BookmarkletRequest`/`ImportBookmarksItem` и `CreateNoteRequest` валидацию, что переданный `type` входит в пользовательский набор (`CelestialBody.UI_TYPES`), либо приводить неизвестный/запрещённый тип к `asteroid`.
- `technical` должен оставаться возможным для системных сценариев (например, Knowledge Core) — но не для обычного пользовательского импорта.

### 4. Тесты
- Юнит-тест в `celestial-body.test.ts` уже проверяет `UI_TYPES`; дополнить или добавить тест, что `import/bookmarks` использует тот же набор.
- Регрессионный E2E/Playwright: открыть `/import/bookmarks`, проверить, что `technical` и аномалии отсутствуют в `<select>`.

## Критерии приёмки

1. В `/import/bookmarks` селектор типа не содержит `technical`, `unknown`, `reality_rift`, `chromatic_maw`, `void_whisper`, `cosmic_abomination`.
2. Доступны все типы из `CelestialBody.UI_TYPES`.
3. (Если пункт 3 принят) Прямой `POST /api/v1/import/bookmarks` с `type: technical` отклоняется или заменяется на `asteroid`, а не создаёт заметку типа `technical` от имени пользователя.
4. Линтер/типчек и `npm run test:unit` проходят.
5. Ручная проверка на Personal-стеке: импорт закладки с типом `asteroid` работает, `technical` в выпадающем списке не виден.

## Что приложить к результату

Скриншот `/import/bookmarks` с открытым селектором типа, подтверждающий отсутствие запрещённых типов. Плюс `curl` с негативным кейсом, если реализован backend-гвард.

## Связанное

- `docs/tasks/NOTE-TYPE-TAXONOMY.md` — актуальный состав `UI_TYPES` и порядок типов; перед реализацией IMP-1 убедиться, что `CelestialBody.UI_TYPES` уже синхронизирован с таксономией.
