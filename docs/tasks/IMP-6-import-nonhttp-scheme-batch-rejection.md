# IMP-6. Массовый импорт: превью падало с молчаливым 400 — два сложенных дефекта

Реализовано Devin, проверяет Claude Code на живом Personal-стеке.

## Что выявлено вручную

2026-09-18 владелец загрузил экспорт закладок `bookmarks_11.09.2026.html` (135 ссылок) на страницу массового импорта `/import/bookmarks`. Список URL распарсился, но при нажатии «Предпросмотр» визуально ничего не происходило. В логах Personal-бэкенда — повторные `400` на `POST /api/v1/import/bookmarks/preview`.

## Корень — два дефекта, сложенных друг на друга

### Дефект 1: logging-middleware съедал тело запроса ≥10 КБ

`internal/interfaces/api/middleware/logging.go` читал `c.Request.Body` целиком для логирования, но **восстанавливал его только при `len < 10000` байт**. При большем теле `ShouldBindJSON` получал пустой поток → `EOF` → `400 VALIDATION_ERROR`.

- Граница подтверждена эмпирически на живом `backend_personal`: тело 9999 байт → `200`, 10001 → `400 EOF`.
- Реальный batch 1 (items 1–50) содержит кириллические заголовки — в UTF-8 до 2 байт/символ, поэтому байтовый размер >10 КБ при ~9.3 КБ символов. Любой достаточно большой JSON-POST (>10 КБ) на **любой** endpoint падал так же — дефект шире импорта.
- Существующий `TestLoggingMiddlewareLargeBody` не ловил это: его хендлер не читал тело.

### Дефект 2: `binding:"url"` в batch-DTO отменял поэлементные ошибки

- Файл содержит javascript-букмарклет `KG Saver` (77-я запись, `HREF="javascript:(function(){...})();"`).
- HTML-парсер `extractURLsFromHTML` по дизайну не фильтрует схему — «preview/import will enforce policy» (комментарий в `extract-urls.test.ts`).
- Но DTO `importItem.URL` имел `binding:"required,url,max=16384"`: валидатор Gin `url` требует схему+хост и отверг `javascript:` **на этапе binding** — до того, как `Service.Preview` успел пометить элемент per-item ошибкой.

Задуманное поведение (per-item ошибки в превью) реализовано в сервисе, но было недостижимо из-за обоих слоёв.

## Что сделано

- `backend/internal/interfaces/api/middleware/logging.go` — body восстанавливается всегда; разбор для лога — только при `<10000` байт.
- `backend/internal/interfaces/api/notehandler/note_handler.go` — из `importItem.URL` убран валидатор `url` (осталось `required,max=16384`). Затрагивает `ImportBookmarksPreview` и `ImportBookmarks`.
- `backend/openAPI.yaml` — из схемы `ImportBookmarksItem.url` убран `format: uri`.
- Регрессионные тесты:
  - `TestLoggingMiddlewareRestoresLargeRequestBody` — тело >10 КБ доходит до хендлера целиком (адверсариал-фаза: на старом коде падает с 400).
  - `TestImportBookmarksPreview_NonHTTPScheme` — `javascript:` URL → `200` с per-item `error` (до фикса — `400`).

## Что НЕ тронуто и почему

- `bookmarkletRequest.URL` (`required,url`) — одиночный endpoint букмарклета; там не-URL должен честно отдавать 400.
- SSRF-политика: `IsAllowedURL` по-прежнему проверяется в `Service.maybeExtract` перед любым фетчем; `NormalizeURL` отвергает не-разрешённые схемы. Безопасность не ослаблена — поэлементная валидация и так была единственной настоящей проверкой.
- `extractURLsFromHTML` — фильтрация не добавлена, чтобы не-URL записи попадали в превью и пользователь видел их с пометкой ошибки, а не терял молча.
- `io.ReadAll` в middleware по-прежнему читает тело целиком в память (как было); отдельный вопрос — лимит размера запроса на уровне сервера/nginx, не входил.

## Проверка

- `go test ./internal/interfaces/api/middleware/ -run TestLoggingMiddleware` — все PASS; новый тест краснел на старом коде (адверсариал-фаза).
- `go test ./internal/interfaces/api/notehandler/ -run TestImportBookmarksPreview` — 3/3 PASS.
- Живой Personal-стек: прямой прогон всех трёх батчей реального файла — `200/200/200`; до фикса batch 1 давал `400 EOF`.

## Замечено попутно (не входит)

- `ImportBookmarksItem.url` в openAPI декларирует `maxLength: 2048`, а binding `max=16384` и фронт усекает до 16384 рун — рассинхрон лимита существовал до фикса.
- Любой JSON-POST >10 КБ на любой endpoint падал с 400 до этого фикса — стоит помнить при ретроспективе «странных» 400.
