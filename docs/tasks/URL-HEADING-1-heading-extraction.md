# URL-HEADING-1: Извлечение заголовка и структуры из h1–h6

**2026-09-14 — обсуждение.** Владелец предложил гипотезу: структура HTML-заголовков (`h1`–`h6`) даёт более надёжное название и содержимое для импортированных/обогащённых URL, чем плоский `<title>` и `innerText`.

## Зачем

Сейчас `backend/internal/infrastructure/web/import_fetcher.go` берёт `title` из `<title>`, а текст — плоским `innerText` со вставкой `\n` на блочных тегах. На многих сайтах:

- `<title>` содержит мусор вида `«Статья — Блог компании ... — ещё что-то»`.
- `h1` обычно ближе к смысловому названию.
- `h2`–`h6` дают готовую структуру документа, которую можно превратить в Markdown-описание.

## Гипотеза владельца

1. Брать `h1` (или склеенные `h1`, если их несколько) как кандидат в `title`.
2. Проверять кандидата в NLP-модуле (semantic similarity к keywords/содержимому, оценка «заголовочности»).
3. Остальную структуру (`h2` и далее) компоновать как документ в `content`:
   - `h2` → раздел/абзац-заголовок;
   - текст под `h2` и до `h3` — как параграф;
   - `h3` — как подзаголовок и т.д.
4. Сохранять иерархию в Markdown.

## Что есть сейчас

- `ImportFetcher.Extract(rawURL) (title, text, error)` <`backend/internal/infrastructure/web/import_fetcher.go`>.
- `BuildContent(title, url, text)` строит Markdown `## [title](url)\n\n{text}` <`backend/internal/application/import/service.go`>.
- Тесты: `backend/internal/infrastructure/web/import_fetcher_test.go`.

## Предварительный план

1. Изменить `ImportFetcher` так, чтобы он возвращал не только `title` и `text`, но и структурированный `outline` (`[]Heading{Level, Text, Children}`).
2. Добавить `HeadingExtractor` / `extractHeadings(doc)`.
3. Сделать `title` выбираемым из нескольких кандидатов:
   - `<title>`;
   - первый `h1`;
   - склеенные `h1` (через разделитель ` — ` или ` | `);
   - fallback — URL.
4. NLP-сервис: новый endpoint или расширение существующего, который получает кандидатов и текст, возвращает лучший `title` и, возможно, `summary`.
5. `BuildContent` получает `outline` и строит Markdown с `#`, `##`... и текстом под ними.
6. Сохранять лимиты (title ≤ 200 runes, content ≤ 50000 runes) и рунную обрезку.

## Открытые вопросы

1. **Несколько `h1`**: склеивать, брать первый, или выбирать по площади/порядку в `<main>`/`article`?
2. **`<title>` vs `h1`**: когда `<title>` лучше? Возможно, `<title>` использовать как fallback, если `h1` слишком короткий/мусорный.
3. **NLP-проверка**: что именно сервис должен проверять? Semantic similarity между `title` и текстом? Отсев «сайтовых» шаблонов (`Home`, `Menu`)?
4. **Содержимое под заголовками**: как быть с `ul`/`ol`, `table`, `pre`/`code`? Превращать в Markdown-списки/блоки?
5. **Фильтрация шума**: как убрать заголовки из `nav`, `aside`, `footer`? Сейчас они исключаются на уровне видимого текста, но `h2`-`h6` в боковых панелях всё равно пролезают.
6. **Ограничение глубины**: в какой момент обрезать outline, чтобы не превысить `maxContentLen`?
7. **Обогащение существующих заметок**: должна ли новая логика применяться только к импорту или и к ручному созданию с `source_url`?

## Связанное

- `docs/tasks/NOTE-QUALITY-1-quality-loop.md` — может входить в pipeline обогащения.
- `docs/tasks/IMP-5-import-url-splitting-and-utf8.md` — предыдущий раунд работы над `ImportFetcher`.
- `backend/internal/infrastructure/web/import_fetcher.go`
- `backend/internal/application/import/service.go`
- `backend/internal/interfaces/api/notehandler/note_handler.go`
- `nlp-service/app/nlp_utils.py`
