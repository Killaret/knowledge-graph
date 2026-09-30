# Roadmap

Where Knowledge Graph is going. What the project is and how to run it — [README](README.md).
What shipped — [CHANGELOG](CHANGELOG.md). The detailed plan behind each line below —
[docs/product/BACKLOG.md](docs/product/BACKLOG.md). Untested hypotheses — [docs/product/IDEAS.md](docs/product/IDEAS.md).

**Stage:** alpha. Single-user and local by design at the moment. A September 2026 audit found
issues that block any multi-user deployment — they are listed under *Now* and tracked in
[docs/archive/EXTERNAL_AUDIT_2026-09.md](docs/archive/EXTERNAL_AUDIT_2026-09.md).

## Now

Correctness of the foundation, before new surface area.

| Work | Why it comes first |
|---|---|
| Close the audit blockers: data isolation under `SKIP_AUTH`, seeded credentials in migrations, internal auth headers at the gateway, private responses marked cacheable | Each one blocks multi-user deployment. Status (DOC-AUDIT-2, 2026-09-26): header stripping done — both nginx configs null out `X-Internal-Auth`/`X-User-Id`; `SKIP_AUTH` is restricted to the seeded test user; authorized responses now send `Cache-Control: private` + `Vary: Authorization, Cookie` (`cacheControlMiddleware` in `router.go`). Still open: migration `019_add_test_user` inserts the known-password account into every database — it is not gated by `APP_ENV` |
| Make verification tell the truth: honest regression exit codes, a real 3D readiness signal, visual baselines that capture the scene | Done in substance: `test-a3-exit-codes.ps1`, `engine.isReady`/`data-test-stable` readiness marker, seeded-deterministic visual snapshots + Argos baselines |
| Enforce coverage thresholds in CI and make the orphaned BDD scenarios executable | Backend 70% gate is enforced in `_core-checks.yml` (`backend-coverage-total.py`); BDD runs via `scripts/run-bdd.cjs` |

## Next

| Work | Notes |
|---|---|
| Multilingual embeddings | The current model (`paraphrase-multilingual-MiniLM-L12-v2`, 384-dim) is already multilingual; the open work is the quality upgrade to e5-small/base — `docs/tasks/MODEL-2-e5-base-migration.md` |
| Keyword normalization | `/normalize` runs in the worker pipeline (`queue/worker.go`) — lemmatization quality and coverage remain open |
| Graph clustering | Communities over a hybrid metric — semantics as the base, existing links reinforcing it |
| Link types in the UI | Native picker and documentation for link semantics |
| Delta-update flicker | A race in the preload path makes the graph blink on refresh |

## Later

Honeycomb and orbital 3D layouts, zoomable navigation into a cluster, note archive and hygiene,
Obsidian import, PWA quick capture, periodic notes, sharing between users, external integrations.
Each is described in [docs/product/BACKLOG.md](docs/product/BACKLOG.md).

## Exploring

Guardians of a cluster, factory-line visualization, leaderboards and public universes, a social
layer, graph motion driven by graph metrics. These are hypotheses, not commitments —
[docs/product/IDEAS.md](docs/product/IDEAS.md).

### Phase 21: Task / Inbox Layer (Задачи и входящие) 🟡 Medium — Planned

**Priority:** 🟡 Medium
**Status:** ⏳ Planned
**Description:** Выделить дела/задачи из общих заметок. Сейчас для быстрых дел используются типы Dust и Asteroid, но нет явного отделения «знаний» от «действий».

- [ ] Добавить тип **Task (Сигнал)** — заметка-задача с дедлайном, статусом `done`, напоминанием.
- [ ] Использовать **Dust** как inbox для быстрых мыслей и «посмотреть завтра».
- [ ] **Asteroid** оставить для фрагментов, TODO, цитат.
- [ ] **Pulsar / Comet (periodic)** — для регулярных задач.
- [ ] UI: отдельный список «Задачи» с группировкой по срокам, чекбоксами, быстрым выполнением.
- [ ] Интеграция с Cosmic Notification System для напоминаний.

### Phase 22: Cluster Folders View (Кластеры как группы закладок) 🟢 Low — Hypothesis

**Priority:** 🟢 Low
**Status:** 💡 Idea
**Description:** Рассмотреть вариант отображения кластеров в виде групп/папок закладок, похожий на Opera Speed Dial / Workspaces. Группы — это кластеры, внутри — заметки/связи.

- [ ] Изучить UX браузерных групп вкладок (Opera, Vivaldi, Edge) и адаптировать к Knowledge Graph.
- [ ] Режим «Folders»: кластеры отображаются как визуальные плитки/папки, внутри — мини-граф или список заметок.
- [ ] Поддержка drag-and-drop заметок между кластерами.
- [ ] Быстрое сворачивание/разворачивание кластеров.
- [ ] Возможность закреплять кластеры, задавать цвет, порядок.
- **Гипотеза:** папочная организация снижает когнитивную нагрузку и упрощает навигацию для небольших графов.
- **Зависимости:** кластеризация, CelestialBody Semantic Model.

---

Planning currently lives in these files rather than in an issue tracker, which is why status has
to be read rather than queried. Moving the working plan to issues is itself on the list.
