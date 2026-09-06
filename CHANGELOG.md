# Changelog

Notable changes to this project, newest first. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

This file covers July 2026 onward. Earlier history lives in the git log.

## [Unreleased]

### Added

- Multilingual embedding foundation: `model_name` column in `note_embeddings`, model-filtered reads in
  `EmbeddingRepository` and graph-service, and `backend/cmd/embed-recompute` CLI for migrating all notes.
  The default target model is `paraphrase-multilingual-MiniLM-L12-v2` with a single `NLP_MODEL_NAME`
  environment variable across `nlp-service`, `backend` and `graph-service`.
- Agent working protocol, shared handoff board and transition log, so work passes between
  contributors through the repository instead of chat.
- `LICENSE` (MIT). The README had advertised MIT for months without the file, which legally
  meant all rights reserved.
- Guard that blocks destructive Docker commands against personal-stack volumes unless a fresh
  non-empty backup exists.
- CI drift guard for build configuration, after a silent revert of 3D fog densities went unnoticed.
- Seeded test user and an explicit `APP_ENV` profile for the isolated test stack.

### Changed

- The canonical regression cycle now reports the real status of every phase. Previously failing
  integration, E2E, BDD and visual tests only printed a warning and the run still ended with a
  success summary.
- Auto-commit removed from the regression cycle entirely.
- The 3D scene now signals readiness only after the first painted frame, so visual regression
  captures the scene instead of the loading screen.

### Fixed

- Yandex OAuth sign-in: login route and response contract.
- 3D fog densities restored after a configuration revert had hidden the graph geometry.

### Security

- OAuth token transport hardened: token removed from the query string, PKCE moved to `S256`,
  `state` validation detached from the PKCE flag.
- `SKIP_AUTH` restricted to the test profile. It had disabled per-owner data scoping, not just
  the login screen.
- Private responses no longer carry `Cache-Control: public`; authorized responses were cacheable
  by any shared cache on the path.

## 2026-08

### Added

- Cosmic Cockpit UI: four retractable panels, HUD, first-person mode.
- Adaptive 2D fog-of-war with a performance warning.
- Mass bookmark import with preview, asynchronous import and status polling; HTML content
  fetching for imported URLs.
- Bookmarklet MVP with its OpenAPI endpoint.
- Scheduled backups to Yandex Disk, daily and weekly, with a `BACKUP_ENABLED` toggle and
  event-driven backup triggers.
- Link weight recalculation as a queue task.
- Graph analytics endpoint: PageRank, connected components, top central nodes.
- `PUT /links/:id` — link update with validation.

### Changed

- Frontend moved to Feature-Sliced Design: components split into widgets and entities.
- Renderer and i18n monoliths split into focused modules; home-page logic extracted into a
  Svelte 5 runes module.
- 2D graph renderer performance pack, with regression tests.
- `NoteForm` and `TypeSelector` decoupled from the celestial-body model.

### Removed

- Obsolete Java `source-text-handler` service. Its specification is kept for the planned rewrite.

## 2026-07

### Added

- Dedicated graph-service: event-driven cache invalidation with a fallback model, auth and
  user-scoped filtering, analytics API (neighbors, path, recommendations), and a materialized
  closure view for graph traversal.
- Public note pool: publish and unpublish, filtered public graph.
- 3D graph rendered through graph-service layouts.

### Changed

- Three isolated stacks — development, personal and test — with an identity check between them.
- Canonical regression cycle over the isolated test stack.
- CI/CD service containers, timeouts and Docker health checks.
- CORS configuration moved to environment variables.
