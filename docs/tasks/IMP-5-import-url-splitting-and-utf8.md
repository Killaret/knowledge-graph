# IMP-5. Массовый импорт: разделение склеенных URL и корректная обработка UTF-8

Постановка для Devin. Порядок работы: [`../AI_AGENT_PROTOCOL.md`](../AI_AGENT_PROTOCOL.md).

Ставит пользователь/Claude Code, реализует Devin, проверяет Claude Code на живом Personal-стеке.

## Зачем

При массовом импорте закладок (`/import/bookmarks`) пользователь вставляет список URL, скопированный из чата/браузера. Если URL идут подряд без пробелов или переводов строк (`https://a.comhttps://b.com...`), фронтенд обрабатывает всё как один элемент и отправляет невалидный `url` — импорт падает ещё на preview.

Кроме того, при попытке импортировать страницы в неродной для UTF-8 кодировке (например, `windows-1251`) `ImportFetcher` либо получал кракозябры, либо при усечении длинного текста `text[:5000]` резал многобайтовую руну посередине. Это приводило к `ERROR: invalid byte sequence for encoding "UTF8"` и падению `INSERT`.

## Что выявлено вручную

- 13 URL, склеенных в одну строку: `https://animego.me/anime/...https://animego.me/anime/...`
- Результат первого импорта: `created 11`, `failed 2`.
  - `https://www.securitylab.ru/blog/personal/LifeHack/354349.php` — `invalid byte sequence for encoding "UTF8": 0xd0`.
  - `https://animego.me/anime/khot-ya-i-bezdarnaya-zlodeika-...-3481` — заголовок извлекался правильно, но его длина 223 байта > `Title` max 200.

## Что сделано

### 1. Frontend: разделение склеенных URL
- `frontend/src/shared/utils/extract-urls.ts` — новый утилитарный модуль, который:
  - делит строку по границам `http://`/`https://`;
  - сохраняет формат `Название | https://...` для одного URL на строку;
  - пропускает пустые строки и комментарии `#`;
  - дедуплицирует URL.
- `frontend/src/routes/import/bookmarks/+page.svelte` — заменён inline-`parseInput`/`parseLine` на вызов `extractURLs`.
- Регрессионные тесты: `frontend/src/shared/utils/extract-urls.test.ts` (6 сценариев).

### 2. Backend: `ImportFetcher`
- `backend/internal/infrastructure/web/import_fetcher.go`:
  - подключает `golang.org/x/net/html/charset` для преобразования ответа в UTF-8 по `Content-Type` или `<meta charset>`;
  - усечение текста идёт по рунам (`[]rune`), а не по байтам;
  - заголовок усекается до 200 байт на границе руны;
  - добавлена санитизация результата до валидного UTF-8.
- Регрессионные тесты: `backend/internal/infrastructure/web/import_fetcher_test.go` (конвертация `windows-1251`, рунное усечение 6000 рун, усечение длинного заголовка).

### 4. Frontend: импорт HTML-файла закладок
- Проблема: пользователь экспортировал закладки из Chrome (`bookmarks_11.09.2026.html`, 135 ссылок). `+page.svelte` разбирал их, но `buildPreview`/`startImport` отправляли все 135 элементов в `v1/import/bookmarks/preview`, а бэкенд `MaxBatchSize=50` отвечал `too many items`. Импорт молча падал.
- Исправления:
  - `frontend/src/shared/utils/extract-urls.ts` — `extractURLsFromHTML(html)` и `chunk(items, size)`.
  - `frontend/src/routes/import/bookmarks/+page.svelte` — `buildPreview` бьёт список на пачки по `MAX_IMPORT_BATCH_SIZE=50`, `startImport` запускает несколько `ImportTask` и агрегирует статусы в `taskStatus`.
  - `frontend/src/shared/api/import.ts` — экспорт `MAX_IMPORT_BATCH_SIZE=50`.
- Регрессионные тесты: `frontend/src/shared/utils/extract-urls.test.ts` (Netscape-HTML, теги в заголовках, `chunk`).

### 3. Backend: `BuildContent` и `buildBookmarkletContent`
- `backend/internal/application/import/service.go` — `BuildContent` усекает `text` до `maxContentLen - len(prefix)` байт, не разрывая рун.
- `backend/internal/interfaces/api/notehandler/note_handler.go` — дублирующий `buildBookmarkletContent` удалён, вместо него используется `importer.BuildContent`.
- Новый shared-модуль: `backend/internal/shared/textutil/utf8.go` (`TruncateToMaxBytes`, `TruncateToMaxRunes`, `SanitizeUTF8`).
- Регрессионные тесты:
  - `backend/internal/application/import/service_test.go`;
  - `backend/internal/shared/textutil/utf8_test.go`.

## Статус

- [x] Построение
- [x] Go unit-тесты (`go test ./...`) — зелёные
- [x] Frontend unit-тесты (`npm run test:unit`) — 1004/1004 зелёных
- [x] Пересборка Personal-стека (`docker compose -f docker-compose.personal.yml up -d --build`) — все сервисы healthy
- [x] Живая проверка `ImportFetcher` на animego и securitylab — заголовки и текст корректно декодируются и усекаются до лимитов
- [ ] Живая проверка повторного импорта склеенного списка и HTML-файла закладок — не проведена, Docker Desktop перестал запускаться после аварийного WSL-рестарта

## Примечание

Повторный импорт того же списка URL после пересборки должен дать `created 13`, `failed 0`. До пересборки старые контейнеры всё ещё содержат старую версию fetcher / парсера.

**2026-09-11:** HTML-импорт починен в коде и пересобран, но живую проверку нельзя завершить, пока Docker Desktop не восстановится.
