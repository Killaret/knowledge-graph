# SPEC-AUDIT-1. Реестр: постановки против кода

Этап 0 (полнота + распределение) и этап A (вердикты по требованиям), Devin, 2026-09-28;
доработан после отклонения 27.09: добавлены 11 пропущенных файлов и повторяемая проверка полноты.
Распределение всех файлов `docs/tasks/` по этапам постановки.
Полнота сверяется с [`README.md`](README.md) (генерируемый индекс, в реестр не входит).
Файлы разборов (`*-review-findings.md`, `*-findings.md`) едут в этап своей постановки; у каждого проверяется,
разобраны ли незакрытые хвосты.

**Повторяемая проверка полноты** (сторож, добавлен 2026-09-27 после блокера 2 ревью):

```bash
node scripts/testing/check-spec-audit-1-register.mjs .
```

Скрипт сравнивает `docs/tasks/*.md` (минус `README.md` и сам реестр) с первой колонкой таблиц
этапов и краснит при расхождении в любую сторону.

Вердикты: `есть+тест` / `есть, без теста` / `сделано иначе` / `потеряно` / `не сделано`.

---

## Этап A. Граф и интерфейс

| Файл | Статус разбора |
|---|---|
| A-1-3d-readiness-signal.md | разобран 2026-09-28 — принято, все требования есть+тест |
| E2E-CANVAS-1-cockpit-canvas-controls-real-auth.md | разобран — принято |
| FE-COVERAGE-1-frontend-vitest5.md | разобран — принято |
| FE-DEPS-106-frontend-toolchain-update.md | разобран — принято |
| PUB-2-graph-view-mode.md | разобран — принято |
| PUB-2-review-findings.md | разобран — хвостов нет |
| SYNC-1-graph-loading-and-sync-review.md | разобран — A и A2 есть+тест; B/C бэклог по решению 69 |
| SYNC-1-review-findings.md | разобран — хвосты закрыты A2 |
| UI-DESIGN-1-app-design-review.md | разобран — хвост UI-LOAD-1 3D (бэклог, решение 69) |
| UI-DESIGN-1-review-findings.md | разобран — тот же хвост |
| UX-1-link-creation-and-canvas-refresh.md | разобран — не сделано, ждёт обсуждения UX (решение 67) |
| UX-2-500-error-page.md | разобран — принято; новая находка F-2 |
| UX-2-review-findings.md | разобран — хвостов нет |
| VIS-1-split-visual-baselines.md | разобран — разделение есть; открыт дефект F-1 |
| VIS-1-review-findings.md | разобран — находки закрыты в `2eb3216` |
| VIS-1-round2-review-findings.md | разобран — блокеры 1 и 3 устранены, блокер 2 жив (F-1) |

Вердикты: `есть+тест` — код и охраняющий тест на месте; `есть, без теста` — код есть,
охраны нет; `сделано иначе`; `не сделано` — в т.ч. осознанно отложено решением владельца.
«Принято» в колонке статуса = запись ревью в `AI_LOG.md` / архиве доски — это доказательство
живого прогона ревьюером, а не моё.

### A-1-3d-readiness-signal.md — принято (AI_LOG 2026-09-06, живой прогон ревьюера, `fc077f0`)

| # | Требование | Вердикт | Доказательство |
|---|---|---|---|
| A-1.1 | Одиночный рендер без запроса кадра | есть+тест | `features/graph-3d/lib/engine.ts:244` `renderOnce()`; визуальный спек ждёт `data-test-stable` |
| A-1.2 | В `disableAnimation`-ветке рендер после `simulateToStable`, затем `finishLoading` | есть+тест | `engine.ts:150-156` |
| A-1.3 | `finishLoading` после рендера кадра, а не до | есть+тест | `engine.ts:212-228` (renderOnce :223 → finishLoading :228) |
| A-1.4 | Признак теста — URL-параметр `stableRender`, не `process.env.VITEST` | есть+тест | `widgets/graph-3d-viewer/Graph3DViewer.svelte:49` |
| A-1.5 | Маркер `data-test-stable` | есть+тест | `Graph3DViewer.svelte:89` |
| A-1.6 | `OrbitControls.enableDamping` off при stableRender | есть+тест | `features/graph-3d/lib/scene.ts:56` |
| A-1.7 | Детерминизм: сид `Math.random`, reducedMotion, связи в публичном сидере | есть+тест | `tests/visual/*` `beforeEach`; `scripts/testing/seed-test-data.*` — 20% публичных |

### E2E-CANVAS-1-cockpit-canvas-controls-real-auth.md — принято (архив доски 2026-09-21)

| # | Требование | Вердикт | Доказательство |
|---|---|---|---|
| E2E-C.1 | readonly = интерактивен: pan/zoom/dblclick не отсекаются | есть+тест | `features/graph-interaction/event-bridge.ts`, `drag-and-drop.ts` (параметр readonly); юнит `zooms on wheel in readonly mode` — мутация красная (разбор в архиве доски) |
| E2E-C.2 | `GraphTopBar` вариант `floating` с canvas-контролами анониму | есть+тест | `features/graph-ui/GraphTopBar.svelte`; 7/7 real-auth на живом стеке |
| E2E-C.3 | Гард `dataKey === lastDataKey` без `simState.isRunning` | есть+тест | `widgets/graph-canvas/GraphCanvas.svelte` |
| E2E-C.4 | `data-testid="graph-empty-state"` | есть+тест | `routes/+page.svelte`; adversarial-тест пустого графа в `cockpit-canvas-controls.spec.ts` |

### FE-COVERAGE-1-frontend-vitest5.md — принято (AI_LOG 2026-09-12)

| # | Требование | Вердикт | Доказательство |
|---|---|---|---|
| FE-C.1 | Vitest 5 | есть+тест | `frontend/package.json`: `vitest ^5.0.0`, `@vitest/coverage-v8 ^5.0.0` |
| FE-C.2 | Порог 70% по всем четырём метрикам | есть+тест | `vitest.config.ts:92-96` (lines/functions/branches/statements = 70); прогон 1381/1381 зафиксирован в журнале |

### FE-DEPS-106-frontend-toolchain-update.md — принято (AI_LOG, PR #110 смержен)

| # | Требование | Вердикт | Доказательство |
|---|---|---|---|
| FE-D.1 | Совместимое обновление тулчейна | есть+тест | `package.json`: `vite ^8.3.0`, `kit ^2.70.3`; CI зелёный на PR #110 |
| FE-D.2 | TS 7 / ESLint 10 отложены осознанно | сделано иначе | `typescript ^5.9.3`, `eslint ^9.39.5` — отложено по peer-конфликтам, зафиксировано в постановке |

### PUB-2-graph-view-mode.md — принято (AI_LOG 2026-09-14, живой прогон 5 своих / 2 публичных)

| # | Требование | Вердикт | Доказательство |
|---|---|---|---|
| PUB-2.1-7 | Семь критериев: переключатель personal/community, умолчание по сессии, localStorage, запрос по режиму, мутация | есть+тест | `shared/stores/graph-view.svelte.ts:4-25`; мутация роняет 3 теста `shared/api/graph.test.ts`; живой прогон — `PUB-2-review-findings.md` |

### SYNC-1-graph-loading-and-sync-review.md — этап A принят, A2 на ревью, B/C в бэклоге

| # | Требование | Вердикт | Доказательство |
|---|---|---|---|
| SYNC-A.1 | Дельта от снимка клиента, `resync` без снимка | есть+тест | `graph-service` snapshot-ключи `snapshot:{user}:{hash}`; интеграционные `http_server_test.go`; живой прогон ревьюера — `SYNC-1-review-findings.md` |
| SYNC-A.2 | `removed_links` в дельте | есть+тест | `ComputeDelta`; мутация красная (findings) |
| SYNC-A.3 | События на всех путях записи + сторож | есть+тест | этап A: ручные `Publish*` + `check-graph-write-paths.mjs`; этап A2 заменил механизм на outbox |
| SYNC-A.4 | resync заменяет граф | есть+тест | `PreloadService` `seedGraph`; `real.test.ts` 31/31 |
| SYNC-A2.1-5 | outbox-таблица, декораторы, ретранслятор, запрет ручной публикации, манифест+CI | есть+тест | миграция `036_graph_outbox`, `infrastructure/outbox/`, `49c2de3`; **на ревью — приёмка не состоялась** |
| SYNC-B | Применение дельты по месту 2D/3D без перезапуска | не сделано | бэклог (решение 69 — в 1.0 вместе с SSE) |
| SYNC-C | SSE-доставка + переподключение | не сделано | бэклог (решение 69) |

### SYNC-1-review-findings.md — разбор закрыт

Оба хвоста этапа A (сторож не в CI; синхронизация черновика без события) перенесены в A2 и там закрыты:
сторож инвертирован и в `core-checks.tsv`+CI; синк черновика идёт через `noteRepo.Save` под обёрткой.
Незакрытых находок в файле нет.

### UI-DESIGN-1-app-design-review.md — принято как разбор (решения 64, 65)

| # | Требование | Вердикт | Доказательство |
|---|---|---|---|
| UI-D.1 | UI-PANELS-1 | есть+тест | принято 09-26, три мутации красные (`UI-DESIGN-1-review-findings.md`) |
| UI-D.2 | UI-QUICK-1 | есть+тест | принято 09-26, контраст 4,67:1 |
| UI-D.3 | UI-GRAPH-1 | есть+тест | принято 09-27 (подписи у топ-15 по связности, одно правило 2D/3D) |
| UI-D.4 | UI-LOAD-1 | есть+тест (2D) / не сделано (3D) | 2D принято 09-26; 3D-часть — бэклог после SYNC-1 (решение 69) — **открытый хвост** |

### UI-DESIGN-1-review-findings.md — разбор; хвост = UI-LOAD-1 3D (выше)

### UX-1-link-creation-and-canvas-refresh.md — бэклог (решение 67)

| # | Требование | Вердикт | Доказательство |
|---|---|---|---|
| UX-1.1-3 | Связь из правого меню; связь существующих заметок; канвас не пропадает на перезагрузке | не сделано | постановка ждёт обсуждения UX с владельцем; строка бэклога на доске |

### UX-2-500-error-page.md — принято (AI_LOG 2026-09-13)

| # | Требование | Вердикт | Доказательство |
|---|---|---|---|
| UX-2.1 | `+error.svelte` full-viewport, i18n, иллюстрация 5xx | есть+тест | `routes/+error.svelte`; `error-page.spec.ts` 4/4 (разбор `UX-2-review-findings.md`) |
| UX-2.2 | Баг `startsWith("/")` — все маршруты публичны | есть+тест | `shared/utils/route-match.ts`; мутация красная (findings) |
| UX-2.3 | Playwright-регрессия на 500 | есть+тест | `tests/error-500-page.spec.ts` + `routes/test/500/+page.server.ts` — **но см. F-2**: тестовый маршрут без гейта окружения |

### UX-2-review-findings.md — разбор закрыт, хвостов нет (F-2 — новая находка, не из этого разбора)

### VIS-1-split-visual-baselines.md — отклонено в раунде 2 (09-07); повторной сдачи нет

| # | Требование | Вердикт | Доказательство |
|---|---|---|---|
| VIS-1.1 | Разделение спек на anonymous/authenticated проекты | есть+тест | `playwright.config.ts:111-130` (`visual`, `visual-real-auth`); `tests/visual/visual-{anonymous,authenticated}.spec.ts` |
| VIS-1.2 | `SKIP_AUTH: "false"` в джобе visual-regression | есть+тест | `.github/workflows/main.yml:349` |
| VIS-1.3 | Оба проекта в прогоне | есть+тест | `main.yml:382` `--project=visual --project=visual-real-auth` |
| VIS-1.4 | Фикстура `PUBLIC_PERCENT` | есть+тест | `main.yml` seed-шаг; `seed-test-data.*` |
| VIS-1.5 | Блокер round2 «авторизованный проект без сессии» | есть+тест (структура) | `visual-real-auth`: `storageState` + dependency `setup-auth` (`playwright.config.ts:122-130`); живой прогон после round2 не зафиксирован |
| VIS-1.6 | Блокер round2 «`stableRender` портит query» | **не сделано** | `visual-anonymous.spec.ts:75`, `visual-authenticated.spec.ts:128,139` — дефект F-1 жив |
| VIS-1.7 | Блокер round2 «anon search 401 в эталоне» | есть+тест (устранён PUB-1) | `/api/v1/notes/search` в `SkipGETPaths` (`middleware/jwt.go:66`); `Handler.Search` — `uuid.Nil` → поиск по публичным (`note_handler.go:1713`); эталон надо переснять — старый содержит плашку ошибки |

Итог по файлу: разделение сделано и держится; открытый дефект один — F-1 (`stableRender` мёртв в трёх сценариях), эталоны после его починки переснять. Повторной сдачи этапа не было.

### VIS-1-review-findings.md (мой разбор `11cab1f`) — находки закрыты

Все три блокера отработаны Клодом в `2eb3216`: CI-2 восстановлена на доске (AI_LOG 09-07, ревью принято).

### VIS-1-round2-review-findings.md — разбор; блокеры 1–3 — статус в таблице VIS-1.5–1.7

## Этап B. Сбор и заметки

| Файл | Статус разбора |
|---|---|
| API-1-openapi-contract-and-handover.md | не начат |
| API-1-review-findings.md | не начат |
| BATCH-1-api-design.md | не начат |
| BATCH-1-review-findings.md | не начат |
| BATCH-DDD-1-validation.md | не начат |
| BATCH-TEST-1-strategy.md | не начат |
| COMET-1-event-reminder-fields.md | не начат |
| IMP-1-import-type-restrictions.md | не начат |
| IMP-1-review-findings.md | не начат |
| IMP-3-import-ghost-type-ux.md | не начат |
| IMP-3-review-findings.md | не начат |
| IMP-4-claude-review.md | не начат |
| IMP-4-import-recommendations-and-java-batch.md | не начат |
| IMP-4-review-findings.md | не начат |
| IMP-5-import-url-splitting-and-utf8.md | не начат |
| IMP-5-review-findings.md | не начат |
| IMP-6-import-nonhttp-scheme-batch-rejection.md | не начат |
| IMP-6-review-findings.md | не начат |
| IMP-7-import-title-only-fallback.md | не начат |
| IMP-8-import-failed-item-description.md | не начат |
| NOTE-DELETE-1-review-findings.md | не начат |
| NOTE-DELETE-1-soft-delete.md | не начат |
| NOTE-HEALTH-1-technical-and-user-health.md | не начат |
| NOTE-QUALITY-1-quality-loop.md | не начат |
| NOTE-QUALITY-1-review-findings.md | не начат |
| NOTE-TYPES-1-type-taxonomy-review.md | не начат |
| NOTE-TYPE-TAXONOMY.md | не начат |
| PROMISES-1-user-promises.md | не начат |
| URL-HEADING-1-findings-probe.md | не начат |
| URL-HEADING-1-heading-extraction.md | не начат |
| URL-HEADING-1-review-findings.md | не начат |

## Этап C. NLP, связи, рекомендации

| Файл | Статус разбора |
|---|---|
| CHUNK-1-measurement-findings.md | не начат |
| CHUNK-1-review-findings.md | не начат |
| CHUNK-1-structure-aware-chunker.md | не начат |
| CHUNK-PERF-1-long-paragraph.md | не начат |
| JAVA-HANDOVER-1-chunking-and-text-know-how.md | не начат |
| LINK-NEXT-1-automatic-and-own-link-types.md | не начат |
| LINK-TYPES-1-link-types-and-visuals.md | не начат |
| LINKS-1-findings.md | не начат |
| LINKS-1-review-findings.md | не начат |
| LINKS-1-wire-gamma-links.md | не начат |
| LINKS-2-manual-link-over-gamma.md | не начат |
| LINKS-2-review-findings.md | не начат |
| LINKS-3-batch-similarity-order.md | не начат |
| LINKS-3-review-findings.md | не начат |
| ORIGIN-1-born-from-relation.md | не начат |
| MODEL-1-embedding-model-measurement.md | не начат |
| MODEL-1-findings-raw.md | не начат |
| MODEL-1-review-findings.md | не начат |
| MODEL-1B-findings.md | не начат |
| MODEL-1B-measurement-gaps.md | не начат |
| MODEL-1B-public-findings.md | не начат |
| MODEL-1B-review-findings.md | не начат |
| MODEL-1B-separation-findings.md | не начат |
| MODEL-1B-separation-public.md | не начат |
| MODEL-2-e5-base-migration.md | не начат |
| NLP-2-review-findings.md | не начат |
| NLP-2-yake-replace-keybert-lemmatization.md | не начат |
| NLP-3-nlp-service-structure-review.md | не начат |
| NLP-4-normalization-measurement.md | не начат |
| NLP-4-note-logical-form-normalization.md | не начат |
| NLP-4-review-findings.md | не начат |
| P11-1-clustering-design-notes.md | не начат |
| P11-2-live-verification.md | не начат |
| P11-2-multilingual-embeddings.md | не начат |
| P11-2-review-findings.md | не начат |
| P11-3-keyword-normalization.md | не начат |
| P11-4-graph-clustering.md | не начат |
| RECO-1-recommendation-formula.md | не начат |
| W-1-eval-findings.md | не начат |
| W-1-link-weight-formula-validation.md | не начат |

## Этап D. Безопасность, данные, эксплуатация

| Файл | Статус разбора |
|---|---|
| A-1-auth-setup-review-findings.md | не начат |
| AUD-1-review-findings.md | не начат |
| AUD-2-data-isolation.md | не начат |
| AUD-2-review-findings.md | не начат |
| AUD-2-seeder-issue.md | не начат |
| AUD-3-token-transport.md | не начат |
| AUD-4-yandex-oauth-contract.md | не начат |
| AUD-5-perimeter-separation.md | не начат |
| AUD-6-bdd-reconnaissance.md | не начат |
| AUD-7a-enforce-boundaries.md | не начат |
| AUD-7b-lint-tests-and-coverage-denominator.md | не начат |
| AUD-10-handoff-loop.md | не начат |
| AUD-10-review-findings.md | не начат |
| AUTO-1-push-and-trigger.md | не начат |
| BACKUP-1-cloud-backup-setup.md | не начат |
| BACKUP-2-backup-location-single-source.md | не начат |
| BACKUP-2-review-findings.md | не начат |
| BACKUP-3-automatic-backups.md | не начат |
| BACKUP-3-review-findings.md | не начат |
| BACKUP-DIR-1-review-findings.md | не начат |
| CI-1-circular-dependency.md | не начат |
| CI-3-local-check-runner.md | не начат |
| CI-3-review-findings.md | не начат |
| CI-MAIN-1-review-findings.md | не начат |
| CLEAN-1-docker-cleanup-honesty.md | не начат |
| CLEAN-1-review-findings.md | не начат |
| CLEAN-2-ps1-encoding-and-backup-gate.md | не начат |
| CONFIG-1-review-findings.md | не начат |
| COVERAGE-1-align-backend-coverage-threshold.md | не начат |
| COVERAGE-1-review-findings.md | не начат |
| CSP-1-content-security-policy.md | не начат |
| CSP-1-review-findings.md | не начат |
| DB-POOL-1-postgres-pool-config.md | не начат |
| DB-POOL-1-review-findings.md | не начат |
| DEPENDABOT-25-yake-license.md | не начат |
| DEPENDABOT-79-nltk-vulnerability.md | не начат |
| DEPLOY-2-deploy-images-from-ci.md | не начат |
| DEPLOY-2-review-findings.md | не начат |
| DEPLOY-3-nlp-healthcheck-and-alpine-pin.md | не начат |
| DEPLOY-3-review-findings.md | не начат |
| DISK-1-everything-on-d.md | не начат |
| DISK-1-findings.md | не начат |
| DISK-1-review-findings.md | не начат |
| ENV-1-review-findings.md | не начат |
| GITHUB-SECURITY-1-triage.md | не начат |
| GORM-COMBINED-1-gorm-combined-update.md | не начат |
| LOG-1-zerolog-integration.md | не начат |
| MONGO-1-drafts-storage-discussion.md | не начат |
| PUB-1-anonymous-read-and-search.md | не начат |
| PUB-1-review-findings.md | не начат |
| PUB-3-rename-graph-endpoints.md | не начат |
| PUB-3-review-findings.md | не начат |
| REG-1-review-findings.md | не начат |
| REG-1-semantic-similarity-card-regression.md | не начат |
| REG-2-batch-similarity-integration-test.md | не начат |
| REG-2-review-findings.md | не начат |
| SEC-1-note-idor.md | не начат |
| SEC-1-review-findings.md | не начат |
| SEC-2-yandex-backup-token-rotation.md | не начат |
| SECURITY-1-handoff.md | не начат |

## Этап E. Процесс и инструменты

| Файл | Статус разбора |
|---|---|
| A-1-review-findings.md | не начат |
| A-3-review-findings.md | не начат |
| ARCHIVE-1-discussion-history.md | не начат |
| AUTHOR-1-commit-authorship-guard.md | не начат |
| AUTHOR-1-review-findings.md | не начат |
| AUTHOR-2-hook-activation.md | не начат |
| AUTHOR-2-review-findings.md | не начат |
| BOARD-1-handoff-retention.md | не начат |
| BOARD-1-review-findings.md | не начат |
| BOARD-2-board-size-and-backlog-policy.md | не начат |
| BOARD-2-review-findings.md | не начат |
| BOARD-3-board-archive.md | не начат |
| CHECK-ALL-1-runner-cannot-report-failure.md | не начат |
| CHECK-ALL-2-review-findings.md | не начат |
| CHECK-DECISIONS-2-review-findings.md | не начат |
| DECISIONS-1-decision-index-and-guard.md | не начат |
| DECISIONS-1-review-findings.md | не начат |
| DOC-AUDIT-1-documentation-inspection.md | не начат |
| DOC-AUDIT-1-review-findings.md | не начат |
| DOC-AUDIT-2-docs-vs-code.md | не начат |
| DOC-AUDIT-2-register.md | не начат |
| DOC-AUDIT-2-review-findings.md | не начат |
| DOC-REORG-1-implementation-notes.md | не начат |
| DOC-REORG-1-review-findings.md | не начат |
| DOCS-LINKS-1-review-findings.md | не начат |
| FACE-1-project-face-from-interview.md | не начат |
| GORDON-1-gordon-documents-review.md | не начат |
| GORDON-2-deployment-package-analysis.md | не начат |
| HOUSEKEEPING-1-review-findings.md | не начат |
| PROJECT-SKILLS-1-review-findings.md | не начат |
| PROTO-CRITERIA-1-review-findings.md | не начат |
| RELEASE-1-scope-1.0.md | не начат |
| RELEASE-TEST-1-manual-run-1.0.md | не начат |
| SPEC-AUDIT-1-specs-vs-code.md | не начат (сама постановка) |
| SPEC-AUDIT-1-review-findings.md | не начат (разбор этой задачи) |
| SPECS-1-review-findings.md | не начат |
| TASKS-INDEX-1-review-findings.md | не начат |
| TASKS-INDEX-1-task-discoverability.md | не начат |
| TEST-PORTS-1-review-findings.md | не начат |
| TRACE-1-decision-task-test.md | не начат |
| VERIFY-FINDING-MIRROR-1-review-findings.md | не начат |
| WORKTREE-1-agent-worktrees.md | не начат |
| WORKTREE-1-review-findings.md | разобран — доработка у Claude Code (ветка отцеплена → норма; сторож реестра в пост-ребейзный список) |

---

## Сводка по этапам

| Этап | файлов | есть+тест | без теста | иначе | потеряно | не сделано |
|---|---|---|---|---|---|---|
| A | 16 | 32 | 0 | 1 | 0 | 5 |
| B | 31 | — | — | — | — | — |
| C | 40 | — | — | — | — | — |
| D | 60 | — | — | — | — | — |
| E | 42 | — | — | — | — | — |
| **Σ** | **189** | | | | | |

Счётчики этапа A — по строкам таблиц выше; «не сделано» там — осознанные отсрочки решением
владельца (SYNC-B/C, UX-1, UI-LOAD-1 3D) и один открытый дефект VIS-1 round2 (F-1).
«Потерь» на этапе A не найдено — все принятые постановки стоят в коде на текущем HEAD.

Не распределено: `README.md` (генерируемый индекс) и `SPEC-AUDIT-1-register.md` (сам реестр).
Итого 189 файлов = все постановки и разборы каталога на 2026-09-28, включая постановку и разбор
SPEC-AUDIT-1. Полнота проверяется повторяемо: `node scripts/testing/check-spec-audit-1-register.mjs .`

---

## Дефекты, найденные при аудите (подтверждены ревьюером 2026-09-27)

Внесены в реестр по методу постановки: вердикт + повторяемое доказательство. Это не пункты
постановок, а находки в коде — финальное решение за владельцем вместе со списком «потеряно».

| # | Дефект | Вердикт | Доказательство | Этап-файл |
|---|---|---|---|---|
| F-1 | `stableRender` в визуальном тесте не включается: URL собирается как `"…?q=…" + "?stableRender=true"` — второй `?` попадает в значение `q` | дефект в тесте (стабильный режим мёртв в трёх сценариях) | `frontend/tests/visual/visual-anonymous.spec.ts:75`, `visual-authenticated.spec.ts:128,139`; `STABLE_RENDER = "?stableRender=true"` | VIS-1 |
| F-2 | Тестовые страницы доступны без проверки окружения: `/test/500` бросает 500 по `?trigger=500`, рядом `/test`, `/test/isolated-node`, `/test/link-pair` | дефект (незакрытая поверхность) | `frontend/src/routes/test/500/+page.server.ts:5-6`; `hooks.server.ts` не закрывает `/test` | UX-2 |
| F-3 | Импорт не пересчитывает рекомендации: `ProcessImportTask` ставит keywords/embedding/normalize/link-weights, но не `EnqueueRefreshRecommendations`; worker ставит refresh только из `generateGammaLinks`, которая выходит раньше при `len(created)==0` | дефект (у импортированной заметки без гамма-связей рекомендации не пересчитываются) | `application/import/service.go:591-602`; `queue/worker.go:242-244` (ранний выход), :248/:254 (refresh только при созданных связях); контроль — `note_handler.go` ставит refresh всегда | IMP-4 |
