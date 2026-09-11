# Manual Test Feedback — ai-agents branch

> Accumulator for issues found during manual testing of the test stack.
> When enough items pile up, we will triage them into:
> - **urgent fixes** (do now)
> - **roadmap** (planned feature work)
> - **ideas/backlog** (hypotheses / nice-to-have)
>
> **Test steps live in [`MANUAL_TEST_CHECKLIST_COCKPIT.md`](MANUAL_TEST_CHECKLIST_COCKPIT.md)**
> (layout/panels, filters, view switching, 2D motion/readability, first-person,
> visual theme, note detail page). This file only tracks the **findings log**
> below — add an entry here when a checklist item fails, referencing the
> section number (e.g. "Cockpit §4").
>
> **Environment:** `http://127.0.0.1:3002` (test stack, SKIP_AUTH=true)
> **Last rebuild:** see git log for `fix(frontend): desynchronize 2D graph node motion and particles`, `fix(frontend): reduce 2D graph overlaps and improve readability`, `fix(frontend): improve first-person exit and panel handles`, and `feat(frontend): add Altered Carbon cockpit gradient text theme`

---

## Findings

### Urgent fixes

<!-- Things that block release or make the app unusable. We fix these immediately. -->

- **Case:** Mass import `/import/bookmarks` preview returns 400 with long bookmark titles or URLs
- **What:** Pasting a bookmark list or dropping a `bookmarks.html` file that contained a URL >2048 characters or a title >200 characters caused the whole preview batch to fail with `VALIDATION_ERROR`. The UI only showed the generic `import.error` text, so the real reason was hidden.
- **Expected:** Preview returns per-item errors, long titles/URLs are accepted up to safe limits, and the user sees the validation message.
- **Actual:** Backend DTO `binding:"max=2048"` on `URL` and `binding:"max=200"` on `Title` rejected the entire request; `note.NewTitle` counted bytes instead of runes, so Cyrillic titles >100 characters also failed after passing the DTO; the UI replaced the structured error with `t("import.error")`.
- **Hotfix applied:**
  - `backend/internal/domain/note/value_objects.go` — `NewTitle` and `NewContent` now count runes (`utf8.RuneCountInString`).
  - `backend/internal/infrastructure/web/import_fetcher.go` — title truncation now uses `TruncateToMaxRunes` (was bytes).
  - `backend/internal/interfaces/api/notehandler/note_handler.go` — `importItem.URL` limit raised to 16384, `importItem.Title` limit raised to 1000 (validated to 200 runes by the service); `bookmarkletRequest.URL` limit raised to 16384.
  - `frontend/src/shared/utils/extract-urls.ts` — truncates titles to 200 runes and URLs to 16384 runes, decodes `&amp;` and numeric HTML entities in titles/URLs.
  - `frontend/src/routes/import/bookmarks/+page.svelte` — shows the first backend `details.message` instead of a generic error on preview/import failure.
- **Regression tests:** `backend/internal/domain/note/value_objects_test.go` (Cyrillic 200/201 runes), `backend/internal/infrastructure/web/import_fetcher_test.go` (title truncation by runes), `frontend/src/shared/utils/extract-urls.test.ts` (long title/URL truncation, HTML entity decoding).
- **Status:** backend rebuilt, frontend rebuilt from local build due to `npm ci` file-lock issues in Docker; verified on Personal stack that a 2100-character URL and a 250-character Cyrillic title no longer cause 400, and long URL imports complete successfully.
- **Screenshot / Logs:** Preview response for `https://example.com/<2100 chars>` / `а`×250: `{"data":{"items":[{"error":"title too long (max 200 characters)","is_new":true,...}]}}`; import task for 2100-char URL: `status: done`, `created: 1`.

- **Case:** Note detail page shows no related notes despite links existing
- **What:** A note with outgoing and/or incoming `links` rows showed an empty "Related notes" section. The graph service returned the link but did not return the linked note, so the UI could not render a card.
- **Expected:** `GET /graph-service/api/v1/graph/note/:id?depth=1` returns both the note and its linked neighbors in both directions.
- **Actual:** `loadRooted` in `services/graph-service/internal/db/postgres_client.go` started the CTE at `level = 1` and used `nodes.level < $2`. For `depth=1`, the recursive step never ran, so only the root note was included. The link was loaded in `loadLinksByNoteIDs`, but the target node was missing, and incoming links were not traversed at all.
- **Hotfix applied:**
  - `services/graph-service/internal/db/postgres_client.go` — starts CTE at `level = 0`, traverses both `source` and `target` directions in a single recursive step (the CTE may only reference itself once), and selects `DISTINCT id, title, type`.
  - Cleared all `graph-service:note:*` Redis cache keys on Personal so the new query is used.
- **Regression tests:** `go test -count=1 ./...` in `services/graph-service` passes; manual API call for a test note with two-directional `related` links returns `total_nodes: 2`, `total_links: 2`.
- **Status:** graph-service container rebuilt and healthy; Personal cache cleared; note detail should now display related notes.
- **Screenshot / Logs:** `GET /graph-service/api/v1/graph/note/aebeb115-...?depth=1` → `{"data":{"nodes":[{"id":"b96eeb43-...","title":"Note Two"}, {"id":"aebeb115-...","title":"Note One"}],"links":[{"source":"aebeb115-...","target":"b96eeb43-..."},{"source":"b96eeb43-...","target":"aebeb115-..."}]}}`.

- **Case:** Cockpit §4 (2D Graph readability)
- **What:** 2D graph looks like a tangled mess of overlapping nodes and links; hard to read which node is which.
- **Expected:** Nodes and links are spaced out enough to distinguish individual elements.
- **Actual:** Everything overlaps, looks like a solid "hairball".
- **Hotfix applied:**
  - Increased initial placement circle radius from 30% to 45% of viewport.
  - Link distance 100 → `120 * densityFactor`, link strength 0.2 → 0.35.
  - Many-body charge -100 → `-180 * densityFactor`.
  - Collision radius 25 → `35 * densityFactor`.
  - Slower alpha decay (0.1 → 0.05) and 200 → 300 warmup ticks so the layout has time to spread.
- **Status:** container rebuilt, ready for re-test under Cockpit §4.
- **Screenshot / Logs:** —

- **Case:** Cockpit §3 / §8 (3D view only shows labels)
- **What:** 3D graph loads but only HTML labels are visible; node spheres and links are not seen, so the scene looks like floating text on a black background.
- **Expected:** Nodes appear as colored geometry and links are visible; fog is light enough that the graph is not hidden.
- **Actual:** WebGL spheres were too dim (MeshStandardMaterial under distance lighting) and fog density was heavy enough to blend nodes into the background.
- **Hotfix applied:**
  - Switched node material to `MeshBasicMaterial` with the celestial body glow color so nodes are bright regardless of scene lighting.
  - Added a `fitScale` in `NodeManager` that scales node size with the graph bounding radius so nodes stay visible at the camera distance used by auto-zoom.
  - Reduced label background opacity and blur so spheres are not completely covered by label boxes.
  - Tuned fog presets in `knowledge-graph.config.json` and the default config (birth: 0.01 → 0.0006, nebula: 0.005 → 0.0003) so fog adds atmosphere without hiding geometry.
  - The engine now sets fog to the final light density immediately after the synchronous simulation warm-up.
- **Status:** container rebuilt, ready for re-test. See Cockpit §8 checklist.
- **Screenshot / Logs:** —

- **Case:** Cockpit §4 (isolated 2D nodes fly far away)
- **What:** Some disconnected or weakly connected nodes drift far from the main cluster, making the graph zoom out to a tiny center and forcing the user to zoom back in.
- **Expected:** All nodes remain within a comfortable viewport without extreme outliers.
- **Actual:** The 2D force simulation had no strong inward pull; the existing bounding force only kicked in at a large radius.
- **Hotfix applied:**
  - Tightened the 2D bounding radius from `0.55` to `0.4` of the smaller viewport dimension.
  - Increased the bounding-force strength to 1.0 so outliers are pulled back more quickly.
- **Status:** container rebuilt, ready for re-test. Dense center clusters are still a readability issue until semantic clustering is implemented.
- **Screenshot / Logs:** —

- **Case:** Cockpit §5 (first-person mode)
- **What:** There is no obvious way to exit first-person mode: the floating exit button is invisible until you hover it, there is no Esc hotkey, and no hint about how to get out.
- **Expected:** User always sees how to exit first-person mode, can use Esc, and gets a clear label on the exit control.
- **Actual:** Once first-person is active, UI panels disappear and the exit button is hidden; users feel trapped.
- **Hotfix applied:**
  - `CockpitFirstPersonButton` is now always visible and shows "Exit first-person (Esc)".
  - `CosmicCockpitLayout` listens to the `Escape` key and exits first-person mode.
  - HUD button still toggles first-person; its own label already changes between enter/exit.
- **Status:** container rebuilt, ready for re-test.
- **Screenshot / Logs:** —

- **Case:** Cockpit §1 (panel handles)
- **What:** After auto-expanding the right/bottom panel, the arrow handle stays visible even though the panel is already open.
- **Expected:** The pull-handle (arrow) is only visible when the panel is collapsed.
- **Actual:** Arrows remain on screen when the panel is expanded, creating visual noise.
- **Hotfix applied:** `CockpitPanel` now renders the handle only when `!isOpen`.
- **Status:** container rebuilt, ready for re-test.
- **Screenshot / Logs:** —

- **Case:** Cockpit §1 (panel handles, left/top — found while implementing gradient text)
- **What:** The right/bottom handle fix above only covered the right/bottom handle block. `CockpitPanel.svelte` renders a **second**, separate handle block for left/top panels that still lacked the `!isOpen` guard — same bug, different code path.
- **Expected:** Left/top pull-handles also disappear once their panel is open.
- **Actual:** Left/top arrows stayed visible even when expanded.
- **Fix applied:** `{#if position === "left" || position === "top"}` → `{#if (position === "left" || position === "top") && !isOpen}`, and dropped the now-dead `class:open={isOpen}` binding to match the right/bottom block.
- **Regression test:** `CockpitPanel.spec.ts` — `shows the pull-handle for %s when collapsed` / `hides the pull-handle for %s once the panel is open`, parametrized over all four positions (`right`, `bottom`, `left`, `top`).
- **Status:** fixed, covered by unit tests, container rebuilt.
- **Screenshot / Logs:** —

- **Case:** Real-auth E2E suite (`chromium-real-auth`) — 7 stable failures
- **What:** During the full regression cycle, and again on an isolated re-run against the same real-auth build, 7 of 27 tests fail identically:
  - `cockpit-canvas-controls.spec.ts` — `top-bar-fog` click is intercepted by `.right-cluster` (60s timeout); the zoom/transform "keeps the canvas visible" check fails.
  - `floating-auth-panel.spec.ts` — after login through the floating panel, `graph-stats` stays `" 10 nodes · 20 links"` for the whole 20s poll (graph never reloads).
  - `note-creation-flows.spec.ts` (×3) and `child-note-flows.spec.ts` — a note created via the floating `+` button, the `N` hotkey ghost form, or the child-note panel never appears in list view (`[data-testid="note-title"]` not found within 20s).
- **Expected:** All `chromium-real-auth` tests pass on the seeded real-auth test stack.
- **Actual:** `20 passed / 7 failed`, identical on both runs — a stable defect, not flakiness.
- **Status:** reproduced on the isolated test stack (`SKIP_AUTH=false`). Unrelated to P11-2: no frontend/auth code changed in this session. Needs triage.
- **Screenshot / Logs:** `frontend/test-results/*chromium-real-auth*/error-context.md`; Playwright `line` reporter output for the `chromium-real-auth` project.

- **Case:** P11-2 multilingual embeddings — live verification on the isolated test stack
- **What:** Stack raised from scratch and seeded with 30 notes / 20 links. Verified the offline model load, embedding dimensions, cross-language similarity and the full recompute path.
- **Expected:** Offline start with `HF_HUB_OFFLINE=1`, 384 dimensions in both languages, same-meaning cross-language pairs clearly above unrelated ones, recompute restoring every vector to the current model.
- **Actual:** All four hold. `docker run --network none -e HF_HUB_OFFLINE=1` loaded the model from the baked-in cache and returned 384 dimensions. Similarity: 0.9652 / 0.9897 / 0.9787 for same-meaning RU↔EN pairs against 0.5110 / 0.5984 / 0.5729 for unrelated ones. Three rows marked `all-MiniLM-L6-v2` were found by `embed-recompute -dry-run` (exactly 3), enqueued, recomputed by the workers, and returned to `paraphrase-multilingual-MiniLM-L12-v2` at 384 dimensions; `note_recommendations` held 30 rows.
- **Status:** task rejected on a different criterion — the graph-service model filter is covered by no executed test and its only integration test fails with `column "model_name" does not exist`. See [`tasks/P11-2-review-findings.md`](tasks/P11-2-review-findings.md).
- **Screenshot / Logs:** `go test -tags=integration -p=1 -count=1 ./...` in `backend` — 51 packages, exit 0; `go test -tags=integration -run TestPostgresClient ./internal/db/` in `services/graph-service` — FAIL with SQLSTATE 42703; mutation output for `FindSimilarNotes` recorded in the findings file.

- **Case:** Mass import of concatenated URLs + Cyrillic pages
- **What:** Pasting a string like `https://a.comhttps://b.com...` into `/import/bookmarks` produced one invalid item and no notes; when a long Cyrillic page was fetched, `ImportFetcher` truncated text at byte 5000 and split a multi-byte UTF-8 sequence, causing `invalid byte sequence for encoding "UTF8"` and a long anime title (223 bytes) exceeded the 200-byte `Title` limit.
- **Expected:** UI splits concatenated URLs, backend decodes non-UTF-8 pages, and content/title truncation never breaks UTF-8 or fails the whole note.
- **Actual:** Initial import of 13 URLs: 11 created, 2 failed.
- **Hotfix applied:**
  - `frontend/src/shared/utils/extract-urls.ts` splits at `http://`/`https://` boundaries, preserves `title | url` lines, and deduplicates.
  - `backend/internal/infrastructure/web/import_fetcher.go` uses `golang.org/x/net/html/charset` for encoding detection, truncates text by runes, and truncates title to 200 bytes at a valid rune boundary.
  - `backend/internal/application/import/service.go` `BuildContent` now truncates by rune-safe bytes; `backend/internal/interfaces/api/notehandler/note_handler.go` reuses `BuildContent` instead of a duplicate.
  - `backend/internal/shared/textutil/utf8.go` introduced helpers `TruncateToMaxBytes`, `TruncateToMaxRunes`, `SanitizeUTF8`.
- **Regression tests:** `frontend/src/shared/utils/extract-urls.test.ts` (6 cases), `backend/internal/infrastructure/web/import_fetcher_test.go` (3 cases), `backend/internal/application/import/service_test.go` (multi-byte `BuildContent`), `backend/internal/shared/textutil/utf8_test.go`.
- **Status:** fixed in code, unit tests green (`go test ./...` and `npm run test:unit` 999/999); Personal stack rebuilt; live `ImportFetcher` fetches for animego and securitylab return valid UTF-8 and truncated titles.
- **Screenshot / Logs:** Import task `145edf5c-c142-496f-99ab-6fbf4b5f7af1` on Personal stack: total 13, created 11, failed 2; worker log `ERROR: invalid byte sequence for encoding "UTF8": 0xd0`; second anime URL title byte length 223 > 200.

- **Case:** HTML bookmarks file import fails for files with more than 50 links
- **What:** Dropping `bookmarks_11.09.2026.html` (135 links exported from Chrome) on `/import/bookmarks` produced no notes because the frontend sent all 135 items to `v1/import/bookmarks/preview` and the backend rejected the batch with `too many items` (`MaxBatchSize=50`).
- **Expected:** The UI parses the Netscape-format HTML, splits the list into backend-sized batches, previews each batch, and starts import tasks for each batch with a single aggregated progress bar.
- **Actual:** Import error; no notes created from the HTML file.
- **Hotfix applied:**
  - `frontend/src/shared/utils/extract-urls.ts` — `extractURLsFromHTML(html)` parses `<A HREF="...">` tags (case-insensitive), strips nested HTML from titles, and `chunk(items, size)` splits arrays.
  - `frontend/src/routes/import/bookmarks/+page.svelte` — `buildPreview` and `startImport` iterate in `MAX_IMPORT_BATCH_SIZE=50` chunks; `pollStatus` aggregates multiple `ImportTaskStatus` into one progress object.
  - `frontend/src/shared/api/import.ts` — exports `MAX_IMPORT_BATCH_SIZE=50`.
- **Regression tests:** `frontend/src/shared/utils/extract-urls.test.ts` (Netscape HTML, title tag stripping, `chunk`).
- **Status:** fixed in code; `npm run check` clean, `npm run lint` clean, `npm run test:unit` 1004/1004; Personal frontend rebuilt.
- **Screenshot / Logs:** `bookmarks_11.09.2026.html`: 135 `<A HREF="...">` links extracted by `extractURLsFromHTML`; `npm run test:unit` output `110 passed`; `svelte-check` 0 errors.

- **Case:** Manual recalculation of recommendations and graph links for existing notes
- **What:** `note_recommendations` and `links` were empty, so `GetSuggestions` and graph view returned nothing. `embed-recompute -post` only enqueues `RefreshRecommendations`/`RecalculateLinkWeights` and `RefreshService` returns 0 candidates when `links` and `note_links_closure` are empty.
- **Expected:** Recommendations and graph appear for existing notes; `GetSuggestions` includes correct `title` for each candidate.
- **Actual:** After `embed-recompute -post`: `note_recommendations` = 0, `links` = 0. Manually inserted semantic-based `note_recommendations` (108) and `links` (18, top-1 per note, `link_type='related'`, `source_type='gamma'`); `REFRESH MATERIALIZED VIEW note_links_closure`. `GetSuggestions` and `GetGraph` returned data, but `GetSuggestions` precomputed branch had empty `title`.
- **Hotfix applied:**
  - One-off SQL (`seed_semantic_safe.sql`) inserted top-6 `note_recommendations` and top-1 `links` from `note_embeddings` cosine similarity.
  - `backend/internal/interfaces/api/notehandler/note_handler.go` now loads `title` for each precomputed recommendation (same as semantic fallback branch).
  - `backend/internal/interfaces/api/notehandler/handler_unit_test.go` updated `TestGetSuggestions_Precomputed`, `PrecomputedStale`, `LimitParam` to expect `FindByID` for the recommended note.
  - `backend/internal/application/recommendation/gamma_link_generator.go` added: limits outgoing gamma links to `MaxGammaOutDegree`, skips self-loops and duplicates, supports batch generation. Covered by unit tests.
  - `backend/internal/infrastructure/db/postgres/embedding_repo.go` fixed `FindSimilarNotes` / `FindSimilarNotesBatch` to return a normalized `[0,1]` score.
- **Regression tests:** `go test ./...` — all packages pass; `go vet` clean; `npm run test:unit` 110 files / 1004 tests pass; `npm run check` 0 errors. Java-specific `POST /api/v1/import/java/batch` реализован и затем удалён; openAPI-контракт синхронизирован с текущим роутером. Детали Java-интеграции вынесены на ревью Claude: `docs/tasks/IMP-4-claude-review.md`.
- **Status:** data fixed in Personal DB; title fix in code and unit tests green; backend container **not redeployed** because Docker Desktop stopped starting with "unable to start".
- **Screenshot / Logs:** `note_recommendations` 108; `links` 18; `note_links_closure` 28; `GET /api/v1/notes/8fc562d1-05f7-47af-9751-c26667bded28/suggestions` → 200 with 6 candidates and scores; `GET .../graph?depth=3` → 200 with 4 nodes and 4 links.

- **Case:** Docker Desktop unable to start after WSL restart
- **What:** The first attempt to refresh `note_links_closure` with 108 dense semantic links (top-6) caused an exponential recursive path explosion and hung the `psql` / WSL VM. `wsl -t docker-desktop` was used to stop the hung VM; after that Docker Desktop engine reports `read-only file system` and fails to start.
- **Expected:** Docker Desktop starts and `kg-*-personal` containers come back healthy.
- **Actual:** `docker ps` returns `Error response from daemon: Docker Desktop is unable to start`; `docker desktop start`/`restart` hang; `docker desktop diagnose` finishes but engine remains down.
- **What to do:** Restart Windows or Docker Desktop. Do **not** reset Docker Desktop to factory defaults (that would erase named volumes including `pgdata_personal`). If it still fails, the `docker-desktop-data` VHDX may need `e2fsck` from a second WSL distro.
- **Status:** Devin cannot recover Docker Desktop from CLI; owner action required.
- **Screenshot / Logs:** `docker ps` → `Docker Desktop is unable to start`; `docker desktop diagnose` bundle `C:\Users\89209\AppData\Local\Temp\E33C7E28-...\20260911110956.zip`.

### Roadmap items

<!-- Real feature work that is understood and has clear value. -->

- **Case:** Personal stack — разделение между «связи» (`links`), «граф» и «семантическая похожесть» (`note_embeddings`).
- **What:** Ранее была рискованная трактовка: казалось, что без `links` невозможно получить семантических кандидатов. Проверка показала, что семантическое fallback (`FindSimilarNotes` по `note_embeddings`) **реализовано** и работает независимо от `links`. Реальные проблемы:
  1. `FindSimilarNotes` возвращал `score: 0` из-за несовпадения SQL-псевдонима (`similarity`) с именем поля GORM-структуры (`Score`) — **исправлено**.
  2. `GetSuggestions` (`traversal_service.go:118-188`) использует веса `alpha=1.0, beta=0.0, gamma=0.0` по умолчанию и жёстко кодирует `semanticScore = 0.0`, поэтому вёс каждого кандидата — это нормализованный BFS-вес по `EmbeddingLoader`, а **не** чистая косинусная близость.
  3. `RefreshService` и graph-service пока не сохраняют embedding-only кандидатов в `note_recommendations` / `note_links_closure`, но карточка (`GET /notes/:id/suggestions`) их уже видит через fallback.
- **Expected:**
  - Карточка показывает семантически похожие заметки **без** необходимости иметь `links`.
  - `score` должен отражать реальную смешанную метрику (alpha*graph + beta*semantic + gamma*keyword), а не нормализованный BFS-вес.
  - Mass import должен запускать тот же конвейер, что и ручное создание (`RefreshRecommendations` в том числе).
- **Actual:**
  - Карточка теперь возвращает похожую заметку с непустым заголовком и `score: 1`.
  - `score: 1` вводит в заблуждение, потому что при единственном кандидате BFS-нормализация даёт `1.0` независимо от реальной близости.
  - `links`, `note_recommendations`, `note_links_closure` — по-прежнему пусты; автоматических связей нет.
- **Triage:**
  - `FindSimilarNotes` / title fallback — **исправлено** и **пересобрано** на Personal-стеке.
  - Доработка `TraversalService` для честного `semanticScore` — задача на будущее, см. `MANUAL_TEST_FEEDBACK.md` §«Pure-semantic vs graph-normalized score».
  - Mass import + batch endpoint — `IMP-4` в `docs/AI_HANDOFF.md`.
  - Java-очистка контента (`TZ-Java-source-text-handler-2026-08-30.md`, §18–19) — предпосылка качественных эмбеддингов.
- **Status:** частично исправлено, детали и UX-наблюдения задокументированы.
- **Screenshot / Logs:** `GET /api/v1/notes/47fa746d-fa7d-4c10-acdd-c7da1167a689/suggestions?limit=6` → 200 с `title` и `score: 1`; SQL: `SELECT count(*) FROM links;` → 0, `SELECT count(*) FROM note_recommendations;` → 0.

### Ideas / backlog

<!-- Hypotheses, nice-to-haves, or "explore later" items. -->

- **Case:** Cockpit §1 (visual depth)
- **What:** Cockpit panels look flat 2D. User wants a 2.5D / pseudo-3D effect where panel ends look voluminous, as if they really go into the screen depth, increasing immersion in the canvas.
- **Expected:** Panels have subtle 3D extrusion / bevel / perspective so they feel like physical surfaces receding into space.
- **Actual:** Panels are flat rectangles.
- **Proposed idea:**
  - Add CSS `perspective` to the cockpit container and small `rotateX`/`rotateY` transforms on panel bodies.
  - Use layered borders with gradient shadows on panel edges to simulate thickness.
  - Animate the depth subtly on hover/focus.
  - Keep it optional and ensure it does not hurt click targets or accessibility.
- **Triage:** backlog / roadmap candidate (Phase 11 or UI-polish phase). Not a bug.
- **Screenshot / Logs:** —

- **Case:** Cockpit §6 (Altered Carbon theming — implemented)
- **What:** Original proposal wanted a distinct "Altered Carbon" neon look via new `--cockpit-*` CSS variables, plus static two-color gradient text.
- **Review notes (before implementing):**
  - The project already has an "Allotropic Carbon" theme (`--carbon-*`, `--color-info`, `--color-glow-purple` in `shared/styles/global.css`) applied site-wide — a parallel `--cockpit-*` set would have duplicated it.
  - `CockpitFrame.svelte` already implements grid + stars + corner bolts + cyan/magenta border gradient, just with hardcoded rgba instead of variables.
  - Proposed target components `CosmicHUD`/`CosmicNotificationCenter`/`CosmicToast` do not exist under those names; actual files are `CockpitHUD.svelte` and the app-wide `widgets/notification/ToastNotification.svelte` (not cockpit-scoped, left untouched).
- **What was implemented** (revised prompt, `docs/PROMPT_COCKPIT_ALTERED_CARBON_THEME.md`, now removed after landing):
  - `shared/styles/global.css`: cockpit-scoped aliases (`--cockpit-accent` → `--color-info`, `--cockpit-accent-2` → `--color-comet`, `--cockpit-gradient-text` → `--carbon-gradient-primary`, etc.) — no new hues.
  - `.cockpit-gradient-text` utility class: animated moving gradient text via `background-position` + `@keyframes cockpit-gradient-shift`, phase offset via `--cockpit-text-delay` (desynchronized per instance, not lockstep). `.cockpit-gradient-text--static` modifier for data-critical text. `prefers-reduced-motion` disables the animation.
  - Applied animated gradient text to decorative headings: `CockpitPanel` panel titles (delay per position), `CockpitFirstPersonButton` exit label, `CockpitLeftPanel` section headings (delay per section).
  - Applied **static** gradient text only to `CockpitHUD`'s cluster name; numeric metrics (notes/links/FPS/health/sync) stay plain solid-color text for readability.
  - Replaced hardcoded `#2dd4bf`/`rgba(45, 212, 191, ...)` with `var(--cockpit-accent, ...)` in `CockpitPanel`, `CockpitFirstPersonButton`, `CockpitHUD`, and `CockpitFrame`'s corner bolts.
- **Regression tests:** `CockpitPanel.spec.ts`, `CockpitFirstPersonButton.spec.ts`, `CockpitHUD.spec.ts` (20 new tests, all passing). `npm run check` — 0 errors. Full `npm run test:unit` — 858/858 passing.
- **Status:** implemented, tested, checklist updated (Cockpit §6). Ready for manual re-test.
- **Screenshot / Logs:** —

---

## Verification

- **Case:** A-1 / 3D visual regression sensitivity
- **What:** Verified that the 3D visual test now produces a stable, deterministic element screenshot and that changing the fog `density_final` produces a visibly different image.
- **Expected:** Two consecutive runs with the same config produce identical 1280×720 crops; raising `birth.density_final` from 0.0006 to 0.02 produces a clearly different frame.
- **Actual:** Baseline and second-run crops are 1240×680 and pixel-identical across runs. The dense-fog crop differs by 114 528 / 843 200 pixels (13.58 %). Differences are concentrated in the background fog and label dimming; node positions remain identical.
- **Screenshot / Logs:** `docs/assets/a1-3d-visual-regression/3d-baseline.png`, `docs/assets/a1-3d-visual-regression/3d-fog-dense.png`, `frontend/test-results/visual-visual-regression-V-df9ff--3D-Graph---renders-3D-view-visual/argos/visual/3d-graph-view.png`; `npm run test:unit` 987/987; `npm run check` clean.

---

## Verification

- **Case:** AUD-4 / Yandex OAuth login route contract
- **What:** Verified that `/api/v1/auth/yandex/login` returns JSON `{"url": "..."}` with all required OAuth parameters, that the old `/api/v1/auth/yandex` path is no longer registered, and that `YANDEX_CLIENT_ID` is wired through to the test backend container.
- **Expected:** Old path (`/api/v1/auth/yandex`) returns an error or 404/401; new path returns `200` with a URL pointing to `https://oauth.yandex.com/authorize` and containing `client_id`, `response_type=code`, `state`, `scope`, `code_challenge` and `code_challenge_method=S256`.
- **Actual:** Before fix: `GET /api/v1/auth/yandex/login` returned `404 Not Found`; `GET /api/v1/auth/yandex` returned `501 Not Implemented` (no client id configured). After fix and rebuild with `YANDEX_CLIENT_ID=test-yandex-client-id`: `GET /api/v1/auth/yandex/login` returns `200 OK` and JSON `{"url":"https://oauth.yandex.com/authorize?client_id=test-yandex-client-id&code_challenge=...&code_challenge_method=S256&response_type=code&scope=login%3Aemail+login%3Ainfo&state=..."}`. `GET /api/v1/auth/yandex` returns `401` (route removed, falls through to JWT middleware).
- **Screenshot / Logs:** `curl -s -D - http://127.0.0.1:18083/api/v1/auth/yandex/login` (before: `404`, after: `200`); `go test ./...` green; `go vet ./...` clean; `npm run test:unit` 987/987; `npm run check` 0 errors; `npm run lint` no new warnings.

## Verification

- **Case:** Playwright real-auth setup for visual regression
- **What:** Verified that `tests/setup/auth.setup.ts` can log in as the seeded `testuser` and persist a `storageState` file, and that the new `visual-real-auth` project runs `@visual` tests against the real-auth test stack.
- **Expected:** `setup-auth` project completes; `visual-real-auth` loads the authenticated page and a sample `@visual` test passes without `__SKIP_AUTH__` injection.
- **Actual:**
  - `npx playwright test --project=setup-auth` → `1 passed` in 6.3s.
  - `npx playwright test --project=visual-real-auth --grep "Home page - default view"` → `2 passed` (setup + test) in 9.8s.
  - `frontend/tests/setup/.auth/testuser.json` was created and used by the `visual-real-auth` project.
- **Screenshot / Logs:** Playwright `line` reporter output for `setup-auth` and `visual-real-auth` projects.

## Verification

### P11-2 — multilingual embeddings (pre-review)

- **Scope:** schema migration `030_add_embedding_model_name`, model-filtered `EmbeddingRepository` and graph-service `GetEmbeddings`, `NLP_MODEL_NAME` wiring, `embed-recompute` CLI.
- **Date:** 2026-09-06
- **Agent:** Devin
- **Environment:** local Go build; integration test container (`testcontainers-go` + Postgres 16 + pgvector).
- **Tests executed:**
  - `cd backend && go build ./...` → `exit 0`
  - `cd backend && go test ./...` → `ok` (all packages)
  - `cd backend && go test -tags=integration ./internal/infrastructure/db/postgres -run=TestEmbeddingRepository_ModelFiltering` → `ok`
  - `cd services/graph-service && go build ./...` → `exit 0`
  - `cd services/graph-service && go test ./...` → `ok` (all packages)
- **Findings:** None; the model-filtering integration test confirms that `FindSimilarNotes` only compares vectors with the same `model_name` and `FindNoteIDsMissingModel` correctly flags notes that still carry the old model vector.

## Verification

### P11-2 — multilingual embeddings (live test stack)

- **Scope:** live verification on the isolated test stack after a clean rebuild (`docker-compose.test.yml`, model `paraphrase-multilingual-MiniLM-L12-v2`).
- **Date:** 2026-09-07
- **Agent:** Devin
- **Tests executed:**
  - `docker compose -f docker-compose.test.yml up -d --build --wait` → all 8 `kg-test-*` containers healthy.
  - `GET http://127.0.0.1:15002/health` → `{"status":"healthy","model_loaded":true}`.
  - `POST /embed` («кошка сидит на окне») → 384-dim vector.
  - `POST /similarity` RU↔EN («кошка сидит на окне» / "a cat sits on the window") → `0.9897`; unrelated RU pair → `0.5447`.
  - DB: `schema_migrations` latest = `030`; `note_embeddings.model_name` NOT NULL, no default; all seeded rows carry `paraphrase-multilingual-MiniLM-L12-v2`.
  - `go run ./cmd/embed-recompute -dry-run` → `0` missing; after marking one row `all-MiniLM-L6-v2` → exactly `1` missing (`785e5934-…`); after restoring → `0`.
  - `scripts/testing/run-full-test-cycle.ps1 -SkipManual` → all phases pass except the pre-existing `chromium-real-auth` failures recorded in Findings above.
- **Findings fixed during this pass:**
  - `nlp-service/entrypoint.sh` and the Dockerfile preloaded the model via `SentenceTransformer(name)`, which fills `~/.cache/torch` — not the HF hub cache (`$HF_HOME/hub`) that `nlp_utils.py` reads via `snapshot_download`. The runtime then re-downloaded the model on every fresh cache. Switched both paths to `snapshot_download(repo_id='sentence-transformers/<model>', cache_dir=$HF_HOME/hub)` and set `HF_HUB_DISABLE_XET=1` after CDN read timeouts.
  - `run-full-test-cycle.ps1` set `SKIP_AUTH=true` globally, leaking into `go test` (`SKIP_AUTH` is only allowed with `APP_ENV=test`) and falsely detecting dev/personal as running because the test stack shares the `knowledge-graph` compose project — it then tried to restore (i.e. start) both stacks. `SKIP_AUTH` is now scoped to the test-stack start, and stack detection uses exact container names.
  - Default `all-MiniLM-L6-v2` values remained in compose files, `backend/.env.example`, both Go configs, both repo fallbacks, the NLP Dockerfile ARG and `nlp_utils.py` — all switched to `paraphrase-multilingual-MiniLM-L12-v2`. Remaining `all-MiniLM-L6-v2` mentions are intentional: migration `030` marks historical rows, task docs describe the old state, and tests use it as the "other model" fixture.
  - `cleanup-docker.ps1/.sh` could remove unused volumes (including Personal volumes once their containers were stopped). Volumes are now preserved by default; `-RemoveVolumes`/`--remove-volumes` removes only anonymous dangling volumes, never anything matching `personal` or the protected label. A raw tarball backup of the three Personal volumes was taken beforehand (`backups/personal-volumes-raw-2026-09-06-215805-*.tar.gz`).

## Verification

### Auth session restore (A-1 auth setup blocker, fixed)

- **Scope:** `initAuth` wiped `kg_auth_session` at startup via `setApiKey(null)` before `hasSessionHint()` was consulted, so `visual-real-auth` pages rendered the anonymous public view.
- **Date:** 2026-09-07
- **Agent:** Devin
- **Tests executed:**
  - New unit test `should attempt cookie refresh when only the session hint survives` in `auth.svelte.test.ts` — green on the fix, **red on the buggy code** (mutation check: `refreshTokens` never called).
  - `npm run test:unit` → 988/988; `npm run check` → 0 errors.
  - Live probe on the isolated test stack (`SKIP_AUTH=false`, locally built `node build` on :3002): `setup-auth` produced `storageState`, then a fresh context loaded `/` — calls were `/api/v1/auth/refresh` → `/api/v1/users/me` → `graph-service/api/v1/graph/full?limit=100`; `localStorage.kg_auth_session` remained `1`; "Sign in" absent.
  - `npx playwright test --project=setup-auth` → 1 passed; `--project=visual-real-auth --grep "Home page - default view"` → 2 passed.
- **Screenshot / Logs:** probe output above (request list + `HINT: 1`, `SIGN_IN_COUNT: 0`, `RESULT: SESSION-RESTORED`); Playwright `line` reporter output.
- **Note:** the 7 `chromium-real-auth` failures above are a **separate** defect — those tests inject `__ACCESS_TOKEN__` and do not use the cookie/session-hint path.

### AUD-5 public/internal graph-service perimeter

- **Scope:** forged `X-Internal-Auth` plus `X-User-Id` could cross the browser-facing nginx listener and select another user's graph.
- **Date:** 2026-09-07
- **Agent:** Devin
- **Before:** isolated test-stack, old graph-service image and nginx pass-through mutation; identical request through `http://127.0.0.1:18086/graph-service/api/v1/graph/full?limit=100` returned `HTTP/1.1 200 OK` and the seeded private graph.
- **After:** the same URL and headers returned `HTTP/1.1 401 Unauthorized`; `Access-Control-Allow-Headers` contains only `Authorization, Content-Type`.
- **Internal channel:** direct test-network-equivalent request on `127.0.0.1:19091` with the configured internal token and user ID returned 100 nodes, proving explicitly trusted server-to-server delegation still works.
- **Browser / SSR:** fresh real-auth storage state rendered `/graph/3d` with `100 nodes` and two successful `graph/full` responses; the SvelteKit server proxy returned 20 nodes from the public endpoint.
- **Security headers:** nginx `/health` returned `X-Content-Type-Options: nosniff`, `X-Frame-Options: SAMEORIGIN`, `Referrer-Policy: strict-origin-when-cross-origin`; `Server: nginx` contains no version.
- **Automated regression:** HTTP and gRPC tests reject a forged user header when trust is disabled. Mutation back to unconditional HTTP trust makes `TestAuthMiddlewareIgnoresInternalUserHeaderByDefault` fail with `expected 401 ... got 200`.
- **Screenshot / Logs:** before/after `curl -D -` output, live node counts, mutation output, and test command output recorded in the implementing session; no screenshot applies to this transport-level finding.

## Findings (Personal stack, 2026-09-10)

### Semantic "similar notes" section in the note card

- **Case:** Personal stack — semantic suggestions returned a single result with `score: 0` and empty `title`; only the `note_id` was correct.
- **What:** `GET /api/v1/notes/:id/suggestions` reached the semantic fallback (`FindSimilarNotes`). The SQL query used an expression alias `similarity` while `ORDER BY similarity DESC` referenced it, but the GORM raw scan into `SimilarNote.Score` could not map `similarity` → `Score`, so `Score` was always `0`. In addition the `note_handler.go` semantic fallback did not load the neighbor title.
- **Expected:** Semantic fallback returns real cosine-derived `score` (scaled 0..1) and the neighbor's `title`.
- **Actual:** Response was `{"suggestions":[{"note_id":"...","title":"","score":0}]}`, which the UI rendered as `0.000` with an empty link.
- **Proposed fix / idea:**
  1. Rename the SQL column to `score` and the `ORDER BY` to `score DESC` in `backend/internal/infrastructure/db/postgres/embedding_repo.go` (both `FindSimilarNotes` and `FindSimilarNotesBatch`, plus the `BatchSimilarNote` `gorm` tag).
  2. Load the title for each neighbor in `backend/internal/interfaces/api/notehandler/note_handler.go:1140-1156`.
- **Status:** Исправлено и пересобрано на Personal-стеке. Проверка: `GET /api/v1/notes/47fa746d-.../suggestions` теперь возвращает `{"note_id":"fbe3ed50-...","title":"Ателье колдовских колпаков / Tongari Boushi no Atelier","score":1}` с `X-Recommendations-Source: graph-service`.
- **Screenshot / Logs:**
  - До: `{"suggestions":[{"note_id":"...","title":"","score":0}]}`.
  - После: `{"suggestions":[{"note_id":"fbe3ed50-...","title":"Ателье колдовских колпаков / Tongari Boushi no Atelier","score":1}]}`.

### Pure-semantic vs graph-normalized score

- **Case:** Personal stack — `score: 1` for the only candidate even though the real cosine similarity is low.
- **What:** `backend/cmd/server/main.go:213` builds `TraversalService` with `NewTraversalService`, which defaults to graph-only weights `alpha=1.0, beta=0.0, gamma=0.0` and ignores `cfg.RecommendationAlpha/Beta/Gamma`. `backend/internal/domain/graph/traversal_service.go:160` hardcodes `semanticScore = 0.0`, so the embedding similarity is treated only as a graph edge weight and then `NormalizeWeights` maps the single best edge to `1.0`.
- **Expected:** The score shown in the note card should reflect the **actual** embedding/keyword combination, not a graph-BFS normalization. With only two notes, the score should still be the real semantic closeness, not `1.0`.
- **Actual:** With two notes in the graph, the suggestion score is always `1.0`, because there is only one candidate and the `BFSNormalize` step divides by the max weight.
- **Proposed fix / idea:**
  - Use `NewTraversalServiceWithWeights(compositeLoader, depth, decay, aggregation, normalize, cfg.RecommendationAlpha, cfg.RecommendationBeta, cfg.RecommendationGamma)` in `main.go`.
  - Compute `semanticScore` per candidate from the embedding similarity (already loaded via `EmbeddingLoader` or `FindSimilarNotes`) instead of hardcoding `0.0`.
  - Consider whether `GetSuggestions` should bypass the graph-service/closure fallback and use pure `FindSimilarNotes` when `links` is empty.
- **Triage:** Связано с `IMP-4` и Java-очисткой контента. Для двух заметок карточка теперь работает, но цифра `1.000` вводит в заблуждение. Обсудить приоритет с владельцем.

### UX observations (from owner)

- **Case:** Personal stack — right-click / side menu link creation, linking existing notes, canvas flicker on edit/delete.
- **What:**
  1. Связи нельзя создать из правого/контекстного меню; нужно открывать карточку заметки.
  2. Нет возможности связать две уже существующие заметки.
  3. При удалении или изменении заметки канвас на некоторое время полностью пропадает.
- **Expected:** Удобное создание связей из любого вида (граф, правое меню, карточка); возможность связать существующие; канвас не исчезает при операциях с заметками.
- **Actual:** Связь только из карточки; нет связывания существующих; канвас пустой после изменения/удаления.
- **Triage:** Задача `UX-1` заведена в `docs/AI_HANDOFF.md` и `docs/tasks/UX-1-link-creation-and-canvas-refresh.md`.

### Mass import does not enqueue recommendation refresh

- **Case:** Personal stack — mass import vs manual/bookmarklet creation.
- **What:** `backend/internal/application/import/service.go:ProcessImportTask` enqueues `ExtractKeywords`, `ComputeEmbedding` and `RecalculateLinkWeights`, but **not** `RefreshRecommendations`. Manual creation and bookmarklet call `enqueueRecommendationTasks`.
- **Expected:** Любой путь создания заметки (mass import, bookmarklet, ручное) должен запускать один и тот же конвейер post-processing, включая `RefreshRecommendations`.
- **Actual:** Mass import не обновляет `note_recommendations`.
- **Triage:** Задача `IMP-4` заведена в `docs/AI_HANDOFF.md` и `docs/tasks/IMP-4-import-recommendations-and-java-batch.md`.

### REG-1: регрессионный тест на семантическое сходство в карточке заметки

- **Scope:** `GET /api/v1/notes/:id/suggestions` — семантический fallback (`note_embeddings`) должен возвращать непустые `title` и `score > 0`.
- **Date:** 2026-09-10
- **Agent:** Devin
- **Defects found manually:**
  1. `EmbeddingRepository.FindSimilarNotes` использовал SQL-алиас `similarity`, а GORM-структура `SimilarNote` читает `score`; `ORDER BY score` падал, `Score` был `0`.
  2. `NoteHandler.GetSuggestions` в семантическом fallback не вызывал `repo.FindByID` для загрузки `title`.
- **Fix applied:**
  - alias `similarity` → `score` в `FindSimilarNotes`/`FindSimilarNotesBatch`/`BatchSimilarNote` (`backend/internal/infrastructure/db/postgres/embedding_repo.go`);
  - `FindByID` + `Title` в `NoteHandler.GetSuggestions` (`backend/internal/interfaces/api/notehandler/note_handler.go`).
- **Regression coverage:**
  - `backend/internal/interfaces/api/notehandler/note_handler_semantic_integration_test.go` (`//go:build integration`) поднимает pgvector-БД, создаёт две заметки с эмбеддингами и проверяет `200` с `X-Recommendations-Source: semantic`, `note_id`, `title` и `score > 0`.
  - `TestGetSuggestions_SemanticFallback` в `handler_unit_test.go` теперь проверяет, что заголовок попадает в ответ.
- **Mutation checks:**
  - Откачено `as score` → `as similarity`: тест падает с `suggestion.Score == 0`.
  - Откачено `FindByID` для `title`: тест падает с `actual ""` vs `expected "Tongari Boushi no Atelier"`.
- **Commands:** `go vet ./...`, `go test ./...`, `go test -tags=integration ./internal/infrastructure/db/postgres/`, `go test -tags=integration -run TestNoteHandlerSemanticIntegrationSuite ./internal/interfaces/api/notehandler/` — зелёные.
- **Triage:** Задача `REG-1` заведена в `docs/AI_HANDOFF.md` и `docs/tasks/REG-1-semantic-similarity-card-regression.md`; статус **на ревью**.

## How to add a finding

Create a new bullet under the right section with:

```markdown
- **Case:** 1 / 2 / 3 / 4 / other
- **What:** concise description
- **Expected:** what you expected
- **Actual:** what happened
- **Screenshot / Logs:** attach if possible
- **Proposed fix / idea:** optional
```

### API-1 Swagger UI renders the hand-written spec

- **Scope:** after lowering `backend/openAPI.yaml` to `openapi: 3.0.3`, the embedded Swagger UI must render the contract instead of «Unable to render this definition»; the generated `backend/docs` package, its blind import and its `swag init`/COPY lines were removed.
- **Date:** 2026-09-08
- **Agent:** Devin
- **Live check (test stack, rebuilt `kg-test-backend` image):** `GET http://127.0.0.1:18083/openapi.yaml` -> 200, 75922 bytes, `openapi: 3.0.3`; `GET /swagger/index.html` -> 200; `GET /swagger/doc.json` -> 500 (the generated spec is gone, the UI does not read it).
- **Browser check (headless Chromium via Playwright):** `http://127.0.0.1:18083/swagger/index.html` — HTTP 200, 0 `.errors-wrapper`/`.error` blocks, **72 rendered operations**; title «Knowledge Graph API 1.1.0 OAS3».
- **Screenshot / Logs:** [`assets/api-1/swagger-render.png`](assets/api-1/swagger-render.png); console output `HTTP: 200 | errorBlocks: 0 | operations rendered: 72`.
