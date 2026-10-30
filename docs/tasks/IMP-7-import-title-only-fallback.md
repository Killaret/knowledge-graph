# IMP-7. Массовый импорт: для упавших на фетче ссылок нет опции «только заголовок»

Реализовано Devin, проверяет Claude Code на живом Personal-стеке.

## Что выявлено вручную

2026-09-18 владелец импортировал `bookmarks_18.09.2026.html` (113 http-ссылок). В превью 31 ссылка получила `error` (фетч не удался: антибот, логин-страницы, SPA, мёртвые сайты) — и была **безальтернативно исключена** из импорта: фильтр `is_new && !error`, контролы строки заблокированы, счётчик их не считал.

Среди отвалившихся — почти все «книжные» закладки (LitRes, baza-knig ×3, a.kniga.me, ozon, forex-ofsite, twirl API Book) и манга-читалки (mintmanga ×2, readmanga, animego). Владельцу пришлось просить забрать их вручную — Devin создал 12 заметок напрямую через `/api/v1/import/bookmarklet`.

Формулировка владельца: «не было опции взять просто наименование — дать альтернативу сохранить просто из наименования».

## Причина

Фронтенд жёстко выкидывал error-элементы на трёх уровнях:

- `startImport`: `.filter((it) => it.is_new && !it.error)`;
- `importableCount`: тот же фильтр — кнопка «Импортировать N» их не считала;
- разметка: `disabled={!item.is_new || !!item.error}` на title/type-контролах — отредактировать заголовок упавшей строки было нельзя.

Бэкенд дефекта не имел: `ProcessImportTask` создаёт заметку из title+URL при пустом `text` (`BuildContent`), а `maybeExtract` не фетчит, когда `Title` заполнен и `extract_content` не задан. Title-only элемент проходит штатный путь как есть.

## Что сделано

- `frontend/src/shared/utils/import-preview-items.ts` (новый) — `isImportable(item)` и `toImportItems(items)`: предикат `is_new && (!error || title_only)`, маппинг в payload с `text: ""` для title-only строк.
- `frontend/src/shared/api/import.ts` — `ImportPreviewItem.title_only?: boolean` — клиентский флаг, в запрос не уходит (маппинг явный).
- `frontend/src/routes/import/bookmarks/+page.svelte`:
  - `toggleTitleOnly(index)`; `startImport` и `importableCount` через новые хелперы;
  - у error-строк с `is_new` — чекбокс «только заголовок» / «title only»; включённая строка становится полностью редактируемой (заголовок, тип) и попадает в импорт без текста;
  - `disabled` у контролов заменён на `!isImportable(item)` — единый предикат с фильтром импорта;
  - стиль `.title-only-toggle`.
- `frontend/src/shared/utils/i18n/messages/import.ts` — `import.titleOnly` (en/ru).
- `frontend/src/shared/utils/import-preview-items.test.ts` — 7 тестов: предикат по всем состояниям строки, маппинг, пустой `text` для title-only.

## Что НЕ тронуто и почему

- Бэкенд (`Preview`, `ProcessImportTask`, DTO) — изменений не требуется: title-only элемент — это обычный элемент с пустым `text`, заметка создаётся штатно. Проверено на очереди: 10 элементов с пустым текстом уже ждут обработки.
- Повторный фетч title-only элементов при импорте не нужен и не происходит: `maybeExtract` срабатывает только при `extract_content` или пустом `Title`.
- Чекбокс показывается только для `is_new` error-строк — дубликатам «только заголовок» не нужен, они и так пропускаются.

## Проверка

- `npx vitest run src/shared/utils/import-preview-items.test.ts` — 7/7 PASS.
- `npx svelte-check` — 0 errors, 0 warnings.
- Personal: `frontend_personal` пересобран, на стеке проверить превью с error-строками и галку.

## Следующий шаг

- IMP-8 (бэклог): поле описания «что это» у title-only строк + бейдж «стаб» для элементов с тривиальным текстом — `tasks/IMP-8-import-failed-item-description.md`.

## Остаточные вопросы

- Дубль twirl API Book: в очереди импорта он уже есть (нормализация срезала `#fragment`, текст пустой) — и вручную создана отдельная заметка с фрагментом в URL. Решить на аудите.
- 19 ссылок из файла не взяты вручную (postman web, корп. gitlab/cloud, hh, adzuna/jooble, deepseek-чат, liveworksheets ×3, puzzle-english ×2, docs.id.itmo.pro, bpmaward, reyohoho, логины bingx/bybit) + 2 дубликата (wooordhunt, englspace). Владельцу решить, нужны ли — после выката фикса их можно добрать галкой «только заголовок» при повторном импорте (дедуп по нормализованному URL сработает на уже взятые).
