# Mass Import — Test Plan

Test cases for the `/api/v1/import/bookmarks` batch-import feature (open tabs, `bookmarks.html`, plain URL list). Covers backend, frontend, asynq worker and manual regression scenarios.

## Unit Tests (backend)

| # | Case | Expected |
|---|------|----------|
| 1 | `BuildContent` still works for single capture | Returns `## [title](url)\n\ntext`, capped at the domain `Content` limit (50 000 runes; the old 10 000-byte cut was removed by URL-HEADING-1). |
| 2 | Parse `bookmarks.html` (Netscape Bookmark File) | Extracts all `A` tags with `HREF` and `text`, skips folders. |
| 3 | Parse plain-text URL list | Accepts one URL per line and `title | url` format, ignores empty lines. |
| 4 | Normalize URLs | Strips fragments, lower-cases scheme, rejects `file://` and private IP ranges. |
| 5 | Dedup by URL | Only the first occurrence of a URL is kept; second is skipped with a reason. |
| 6 | Truncate content before `NewContent` | Title + URL + extracted text never exceeds the domain `Content` limit. |
| 7 | `web.ImportFetcher` extracts a structured page (URL-HEADING-1 stage A) | Returns `ExtractedPage`: title + title candidates, normalized outline, Markdown body, `noise_dropped`, `related_links`, `sections_dropped`; `script`/`style`/nav/footer/aside are removed. |
| 7a | Content container selection | `<main>`/`<article>`/`role=main` first; class fallback skips chrome (`modal-content`, menus); a class-matched fragment holding none of the document headings falls back to `<body>`; BEM modifiers like `has-sidebar` are not noise. |
| 7b | Title candidates and suffix stripping | `title_candidates` = h1s → `og:title` → `<title>` minus site suffix (` — /–/-/|/·/:` separators, compared against `og:site_name`, `application-name`, host and its label) → last URL segment; stage-A rule picks the first 3–200-rune candidate that is not the site name; `title_source = "rule"`. |
| 7c | Section budget | `IMPORT_CONTENT_MAX_RUNES` (default 20 000) truncates by whole sections/blocks — never inside a list or a fenced code block — and appends `_(обрезано: N разделов не вошли)_`; `metadata.import_truncated.sections_dropped` is set. |
| 7d | Golden snapshot suite | 19 local HTML snapshots under `backend/internal/infrastructure/web/testdata/urlheading/` + `expected.json`; titles and outline subsequences must match — any miss is a regression. |
| 7e | Charset handling | Declared charset is honored; an uncertain sniff (`<meta charset="">`, no HTTP charset) keeps UTF-8 instead of decoding through windows-1252. |
| 7f | `metadata.related_links` | Links from nav/aside/footer/page-level headers are stored (max 20, http(s) only), never rendered into the note body. |
| 8 | `web.ImportFetcher` HTTP errors | 4xx/5xx responses and context cancellation return an error. |
| 9 | `importer.Service.maybeExtract` URL policy | `file://`, `localhost`, and private IP addresses are rejected before fetching. |

## Integration Tests (backend)

| # | Case | Expected |
|---|------|----------|
| 7 | `POST /api/v1/import/bookmarks/preview` with 3 valid URLs | Returns 200 with `items` array containing `title`, `url`, `text`, `type`, `is_new`. |
| 7a | Preview with `extract_content` | Items carry `title_candidates`, `outline`, `noise_dropped`, `title_source` ("rule"); the UI offers a candidates dropdown and a collapsible outline. |
| 8 | Preview with an unreachable URL | The item is marked as `failed` and has a short error message. |
| 9 | Preview with a URL already in the user's notes | `is_new: false`, `existing_note_id` set. |
| 10 | `POST /api/v1/import/bookmarks` with approved items | Returns 202 with `task_id`; asynq task `import:bookmarks` is enqueued. |
| 11 | Worker processes the task | Creates notes for all non-duplicate items, skips duplicates, stores errors. |
| 12 | `GET /api/v1/import/:task_id/status` | Returns `pending`, `processing`, `done` or `failed` with `total`, `created`, `skipped`, `failed` counters. |
| 13 | Worker updates progress in Redis | Progress increments as each item is processed; status endpoint reflects current values. |

## E2E Tests (Playwright, SKIP_AUTH)

| # | Case | Steps | Expected |
|---|------|-------|----------|
| 14 | Import a plain list of URLs | Open `/import/bookmarks`, paste 3 URLs, click preview. | Table with 3 rows appears; types default to `asteroid`. |
| 15 | Edit type and remove an item | In preview, change type of row 2, remove row 3, click import. | 202 response, task created, only 2 rows submitted. |
| 16 | Poll and see success | Wait for status, then go to `/notes`. | New notes appear in the list with correct titles. |
| 17 | Import `bookmarks.html` | Drag a Netscape bookmark file into the drop zone. | File parsed, preview table shows bookmarks. |
| 18 | Duplicate handling | Import the same list twice. | Second import reports `skipped: 2`. |
| 19 | Fetch page title and text | Paste bare URLs, enable «Fetch page title and text for each URL», click preview. | Table shows titles extracted from HTML and `text` column contains body snippets. |

## Manual / Regression Tests

| # | Case | Steps | Expected |
|---|------|-------|----------|
| 19 | Browser extension: all open tabs | Click extension icon, select all tabs, import. | Task created; all tabs appear as notes. |
| 20 | 50+ URLs at once | Paste 50 URLs and import. | Endpoint returns 202, worker completes without timeout, status shows progress. |
| 21 | URL with large extracted text | Fetch a page with body > `IMPORT_CONTENT_MAX_RUNES`. | Text truncated at section boundaries; `metadata.import_truncated` reports dropped sections; no 500 error. |
| 22 | Invalid URLs only | Paste only `not-a-url` lines. | Validation error, no task created. |
| 23 | Auth: not authenticated | Call endpoint without cookie/token. | 401 Unauthorized. |

## Security / Edge Cases

- SSRF: reject `localhost`, `127.0.0.1`, `10.0.0.0/8`, `192.168.0.0/16`, `file://`.
- Max batch size: 50 items per request; return 400 if exceeded.
- Fetch timeout: 10 seconds per URL; failed items are recorded, not retried.
- Content limit: final note content must not exceed the domain `Content` limit (10 000).
- Type validation: each item type must be in the allowed celestial-body enum.
