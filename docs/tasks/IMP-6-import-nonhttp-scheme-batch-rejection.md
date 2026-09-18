# IMP-6. Массовый импорт: одна не-http(s) ссылка в файле закладок роняла весь батч с молчаливым 400

Реализовано Devin, проверяет Claude Code на живом Personal-стеке.

## Что выявлено вручную

2026-09-18 владелец загрузил экспорт закладок `bookmarks_11.09.2026.html` (135 ссылок) на страницу массового импорта `/import/bookmarks`. Список URL распарсился, но при нажатии «Предпросмотр» визуально ничего не происходило. В логах Personal-бэкенда — повторные `400` на `POST /api/v1/import/bookmarks/preview`.

## Корень

- Файл содержит javascript-букмарклет `KG Saver` (77-я по счёту запись, `HREF="javascript:(function(){...})();"`).
- HTML-парсер `extractURLsFromHTML` по дизайну не фильтрует схему — «preview/import will enforce policy» (комментарий в `extract-urls.test.ts`).
- Но DTO `importItem.URL` имел `binding:"required,url,max=16384"`: валидатор Gin `url` требует схему+хост и отверг `javascript:` **на этапе binding** — до того, как `Service.Preview` успел пометить элемент per-item ошибкой.
- Итог: батч 51–100 целиком уходил в 400, фронт ставил `status="error"` с generic-сообщением — снаружи «ничего не происходит».

Задуманное поведение (per-item ошибки в превью) реализовано в сервисе, но было недостижимо из-за binding-валидации.

## Что сделано

- `backend/internal/interfaces/api/notehandler/note_handler.go` — из `importItem.URL` убран валидатор `url` (осталось `required,max=16384`). Затрагивает `ImportBookmarksPreview` и `ImportBookmarks`.
- `backend/openAPI.yaml` — из схемы `ImportBookmarksItem.url` убран `format: uri`.
- Регрессионный тест `TestImportBookmarksPreview_NonHTTPScheme` в `note_handler_import_test.go`: `javascript:` URL → `200` с per-item `error` (до фикса — `400`).

## Что НЕ тронуто и почему

- `bookmarkletRequest.URL` (`required,url`) — одиночный endpoint букмарклета; там не-URL должен честно отдавать 400.
- SSRF-политика: `IsAllowedURL` по-прежнему проверяется в `Service.maybeExtract` перед любым фетчем; `NormalizeURL` отвергает не-разрешённые схемы. Безопасность не ослаблена — поэлементная валидация и так была единственной настоящей проверкой.
- `extractURLsFromHTML` — фильтрация не добавлена, чтобы не-URL записи попадали в превью и пользователь видел их с пометкой ошибки, а не терял молча.

## Проверка

- `go test ./internal/interfaces/api/notehandler/ -run TestImportBookmarksPreview` — 3/3 PASS (новый тест краснел бы до фикса: binding давал 400).
- Живой Personal-стек: `backend_personal` пересобран с фиксом, импорт 135 закладок владельца — см. логи.

## Замечено попутно (не входит)

- `ImportBookmarksItem.url` в openAPI декларирует `maxLength: 2048`, а binding `max=16384` и фронт усекает до 16384 рун — рассинхрон лимита существовал до фикса.
