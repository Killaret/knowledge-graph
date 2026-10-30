# IMP-6. Массовый импорт: превью падало с молчаливым 400 и таймаутом — три сложенных дефекта

Реализовано Devin, проверяет Claude Code на живом Personal-стеке.

## Что выявлено вручную

2026-09-18 владелец загрузил экспорт закладок `bookmarks_11.09.2026.html` (135 ссылок) на страницу массового импорта `/import/bookmarks`. Список URL распарсился, но при нажатии «Предпросмотр» визуально ничего не происходило, затем — падение по таймауту. В логах Personal-бэкенда — повторные `400`, затем `200` за 30.1 с с `broken pipe` (клиент уже отвалился).

## Корень — три дефекта, сложенных друг на друга

### Дефект 1: logging-middleware съедал тело запроса ≥10 КБ

`internal/interfaces/api/middleware/logging.go` читал `c.Request.Body` целиком для логирования, но **восстанавливал его только при `len < 10000` байт**. При большем теле `ShouldBindJSON` получал пустой поток → `EOF` → `400 VALIDATION_ERROR`.

- Граница подтверждена эмпирически на живом `backend_personal`: тело 9999 байт → `200`, 10001 → `400 EOF`.
- Реальный batch 1 (items 1–50) содержит кириллические заголовки — в UTF-8 до 2 байт/символ, поэтому байтовый размер >10 КБ при ~9.3 КБ символов. Любой JSON-POST >10 КБ на **любой** endpoint падал так же — дефект шире импорта.
- Существующий `TestLoggingMiddlewareLargeBody` не ловил это: его хендлер не читал тело.

### Дефект 2: `binding:"url"` в batch-DTO отменял поэлементные ошибки

- Файл содержит javascript-букмарклет `KG Saver` (77-я запись, `HREF="javascript:(function(){...})();"`).
- HTML-парсер `extractURLsFromHTML` по дизайну не фильтрует схему — «preview/import will enforce policy» (комментарий в `extract-urls.test.ts`).
- Но DTO `importItem.URL` имел `binding:"required,url,max=16384"`: валидатор Gin `url` требует схему+хост и отвергал бы такую запись **на этапе binding** — до того, как `Service.Preview` успел пометить элемент per-item ошибкой. *(Исправлено по ревью 2026-09-21: исходный текст утверждал, что валидатор отверг `javascript:` — проверено, `validator/v10` принимает `javascript:`, `chrome:`, `about:`, `file:`; отвергает он адреса **без схемы** (`www.example.com`). Реальный 400 на стенде давал дефект 1, не валидатор. Правка остаётся верной: проверка URL перенесена с binding на поэлементную ошибку — её охраняет `TestImportBookmarksPreview_SchemelessURL`.)*

### Дефект 3: превью фетчило страницы последовательно — таймаут клиента 30 с

`Service.Preview` вызывал `maybeExtract` по очереди: 50 элементов × фетч (per-fetch timeout 10 с в `import_fetcher.go`) = до ~8 мин худший случай. ky-клиент обрывает на 30 с (`client.ts timeout: 30000`) → `broken pipe`, хотя бэкенд отвечал 200.

- Важно: `extract_content` в превью — несущий флаг. `createBookmarksImport` не передаёт `options`, поэтому текст для импорта берётся из ответа превью; убрать фетч из превью нельзя без изменения протокола.

## Что сделано

- `backend/internal/interfaces/api/middleware/logging.go` — body восстанавливается всегда; разбор для лога — только при `<10000` байт.
- `backend/internal/interfaces/api/notehandler/note_handler.go` — из `importItem.URL` убран валидатор `url` (осталось `required,max=16384`). Затрагивает `ImportBookmarksPreview` и `ImportBookmarks`.
- `backend/internal/application/import/service.go` — `Preview` параллелит `maybeExtract` семафором `previewConcurrency = 10` (худший случай ~50 с < nginx 60 с), логика элемента вынесена в `previewItem`, порядок ответа сохранён.
- `frontend/src/shared/api/import.ts` — `previewBookmarks` `timeout: 120000` (фетч 50 страниц легально длиннее 30 с).
- `backend/openAPI.yaml` — из схемы `ImportBookmarksItem.url` убран `format: uri`.
- `frontend/src/shared/utils/i18n/messages/import.ts` — ru: «Одна страница (букмарклет)» → «Одна страница (bookmarklet)» — англицизм латиницей, просьба владельца.
- Регрессионные тесты:
  - `TestLoggingMiddlewareRestoresLargeRequestBody` — тело >10 КБ доходит до хендлера целиком (адверсариал-фаза: на старом коде падает с 400).
  - `TestImportBookmarksPreview_NonHTTPScheme` — `javascript:` URL → `200` с per-item `error` (документирует поведение; валидатор `javascript:` принимает, поэтому тест холостой к откату правки 2).
  - `TestImportBookmarksPreview_SchemelessURL` — `www.example.com` без схемы → `200` с per-item `error`; при возврате валидатора `url` в `importItem` падает (`400` на всю пачку). Добавлен по ревью 2026-09-21.

## Что НЕ тронуто и почему

- `bookmarkletRequest.URL` (`required,url`) — одиночный endpoint букмарклета; там не-URL должен честно отдавать 400.
- SSRF-политика: `IsAllowedURL` по-прежнему проверяется в `Service.maybeExtract` перед любым фетчем; `NormalizeURL` отвергает не-разрешённые схемы. Безопасность не ослаблена — поэлементная валидация и так была единственной настоящей проверкой.
- `extractURLsFromHTML` — фильтрация не добавлена, чтобы не-URL записи попадали в превью и пользователь видел их с пометкой ошибки, а не терял молча.
- `io.ReadAll` в middleware по-прежнему читает тело целиком (как было); отдельный вопрос — лимит размера запроса на уровне сервера/nginx, не входил.
- `ProcessImportTask` (воркер) оставлен последовательным — там асинхронная обработка, таймаутов клиента нет.

## Проверка

- `go test ./internal/interfaces/api/middleware/ -run TestLoggingMiddleware` — все PASS; новый тест краснел на старом коде (адверсариал-фаза).
- `go test ./internal/application/import/ ./internal/interfaces/api/notehandler/` — PASS.
- Живой Personal-стек: прямой прогон всех трёх батчей реального файла — `200/200/200`; до фикса batch 1 давал `400 EOF`, с `extract_content` — обрыв на 30 с.

## Замечено попутно (не входит)

- `ImportBookmarksItem.url` в openAPI декларирует `maxLength: 2048`, а binding `max=16384` и фронт усекает до 16384 рун — рассинхрон лимита существовал до фикса.
- Любой JSON-POST >10 КБ на любой endpoint падал с 400 до этого фикса — стоит помнить при ретроспективе «странных» 400.
- `createBookmarksImport` игнорирует `options` — протокол «экстракция только в превью» хрупкий: если пользователь отключит `extract_content`, импорт создаст заметки без текста. Вопрос на отдельную постановку.
