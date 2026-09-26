# SPEC-AUDIT-1. Реестр: постановки против кода

Скелет (этап 0), Devin, 2026-09-28. Распределение всех файлов `docs/tasks/` по этапам постановки.
Полнота сверяется с [`README.md`](README.md) (генерируемый индекс, в реестр не входит).
Файлы разборов (`*-review-findings.md`, `*-findings.md`) едут в этап своей постановки; у каждого проверяется,
разобраны ли незакрытые хвосты.

Вердикты: `есть+тест` / `есть, без теста` / `сделано иначе` / `потеряно` / `не сделано`.

---

## Этап A. Граф и интерфейс

| Файл | Статус разбора |
|---|---|
| A-1-3d-readiness-signal.md | не начат |
| E2E-CANVAS-1-cockpit-canvas-controls-real-auth.md | не начат |
| FE-COVERAGE-1-frontend-vitest5.md | не начат |
| FE-DEPS-106-frontend-toolchain-update.md | не начат |
| PUB-2-graph-view-mode.md | не начат |
| PUB-2-review-findings.md | не начат |
| SYNC-1-graph-loading-and-sync-review.md | не начат |
| UI-DESIGN-1-app-design-review.md | не начат |
| UX-1-link-creation-and-canvas-refresh.md | не начат |
| UX-2-500-error-page.md | не начат |
| UX-2-review-findings.md | не начат |
| VIS-1-split-visual-baselines.md | не начат |
| VIS-1-review-findings.md | не начат |
| VIS-1-round2-review-findings.md | не начат |

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
| NOTE-QUALITY-1-quality-loop.md | не начат |
| NOTE-TYPE-TAXONOMY.md | не начат |
| URL-HEADING-1-findings-probe.md | не начат |
| URL-HEADING-1-heading-extraction.md | не начат |
| URL-HEADING-1-review-findings.md | не начат |

## Этап C. NLP, связи, рекомендации

| Файл | Статус разбора |
|---|---|
| CHUNK-1-measurement-findings.md | не начат |
| CHUNK-1-review-findings.md | не начат |
| CHUNK-1-structure-aware-chunker.md | не начат |
| LINKS-1-findings.md | не начат |
| LINKS-1-review-findings.md | не начат |
| LINKS-1-wire-gamma-links.md | не начат |
| LINKS-2-manual-link-over-gamma.md | не начат |
| LINKS-2-review-findings.md | не начат |
| LINKS-3-batch-similarity-order.md | не начат |
| LINKS-3-review-findings.md | не начат |
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
| DECISIONS-1-decision-index-and-guard.md | не начат |
| DECISIONS-1-review-findings.md | не начат |
| DOC-AUDIT-1-documentation-inspection.md | не начат |
| DOC-AUDIT-1-review-findings.md | не начат |
| DOC-AUDIT-2-docs-vs-code.md | не начат |
| DOC-AUDIT-2-register.md | не начат |
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
| SPEC-AUDIT-1-specs-vs-code.md | не начат (сама постановка) |
| SPECS-1-review-findings.md | не начат |
| TASKS-INDEX-1-review-findings.md | не начат |
| TASKS-INDEX-1-task-discoverability.md | не начат |
| TEST-PORTS-1-review-findings.md | не начат |
| VERIFY-FINDING-MIRROR-1-review-findings.md | не начат |
| WORKTREE-1-agent-worktrees.md | не начат |

---

## Сводка по этапам

| Этап | файлов | есть+тест | без теста | иначе | потеряно | не сделано |
|---|---|---|---|---|---|---|
| A | 14 | — | — | — | — | — |
| B | 25 | — | — | — | — | — |
| C | 33 | — | — | — | — | — |
| D | 60 | — | — | — | — | — |
| E | 37 | — | — | — | — | — |
| **Σ** | **169** | | | | | |

Не распределено: `README.md` (генерируемый индекс, не постановка). Итого 169 файлов =
все постановки и разборы каталога на 2026-09-28, включая саму постановку SPEC-AUDIT-1.
