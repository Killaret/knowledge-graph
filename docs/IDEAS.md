# Ideas and Hypotheses

Nothing here is a commitment. These are directions worth exploring, kept apart from the
[backlog](BACKLOG.md) precisely so that planned work and speculation are never mistaken for one
another. An idea moves to the backlog only after it has a stated hypothesis, a way to validate
it, and a reason to exist beyond "would be nice".

## 🧪 Experimental & Ideas (Гипотезы и эксперименты)

Идеи, требующие проработки, прототипирования и проверки гипотез. Могут быть переведены в активные фазы после validation.

### 📥 Legacy manual-test backlog (archived from `docs/MANUAL_TEST_ISSUES.md`)

**Priority:** 🟡 Medium/🟢 Low (не блокирует релизы)
**Status:** 💡 Idea / needs re-triage
**Description:** `docs/MANUAL_TEST_ISSUES.md` был снимком ручного тестирования от 28 июля 2026 (real-auth режим, до перехода на Cosmic Cockpit layout) и накопил дубли нумерации и вперемешку решённые/открытые пункты. Файл удалён; ниже — пункты, которые на момент архивации выглядели ещё нерешёнными и не покрыты текущими чек-листами (`docs/MANUAL_TEST_CHECKLIST_COCKPIT.md`, `docs/MANUAL_TEST_CHECKLISTS_RU.md`). Требуют повторной проверки на актуальной сборке, а не немедленного фикса.

- Изменение email в профиле не сохраняется после reload (ProfileEditor / backend `UpdateMe` — возможна проблема с транзакцией или кэшированием).
- Нет toast-уведомления об успешном сохранении email.
- Страница `/settings` не существует (404, нет пункта в UI и хоткея).
- Внешний вид страницы профиля требует доработки (без уточнения, что конкретно).
- Некоторые UI-пункты того снимка (кнопка logout в `SidebarWidget`, `TypeSelector` не влезает на экран, кнопка "Назад" на странице входа) относятся к дизайну до перехода на Cosmic Cockpit — актуальность нужно перепроверить на текущей вёрстке, а не переносить как есть.

### Phase 13: Factory Line (Производственная линия) 🟢 Low — Hypothesis

**Priority:** 🟡 Medium (после кластеризации и сот)
**Status:** 💡 Idea
**Description:** Режим визуализации графа как производственной цепочки (Factorio / Shapez).

- Заметки отображаются как «станки» с входами и выходами.
- Связи — конвейеры с анимацией движения (анимация только для связей в текущем viewport, fallback — статические стрелки).
- Типы заметок мапятся на роли: `raw` (сырьё / черновик), `processor` (обработчик / анализ), `product` (продукт / вывод). Источник роли — существующее поле `note.type` или новый тег `factoryRole`.
- Обрыв цепи подсвечивается как «брак».
- Метрика продуктивности: число завершённых цепочек.
- **Гипотеза:** метафора завода мотивирует замыкать цепи и вести гигиену заметок.
- **MVP:** режим отображения с прямоугольными узлами и анимированными стрелками.
- **Зависимости:** нет (отдельный режим визуализации, но требуется ролевая классификация заметок).
- **Validation:** >30% новых связей замыкаются в завершённые цепочки за месяц; режим используется ≥1 раз в неделю активными пользователями.

### Phase 14: Semantic Guardians (Семантические стражи) 🟢 Low — Hypothesis

**Priority:** 🟢 Low (после кластеризации и архива)
**Status:** 💡 Idea
**Target:** Mobile (PWA first, native later) — desktop показывает состояние стражей, но не является основным интерфейсом взаимодействия.
**Description:** Пассивная геймификация гигиены знаний через персонифицированного стража кластера.

- Каждый кластер получает стража — семантическую фигурку, отражающую суть кластера (герой, бог, автор, предмет). Пользователь называет стража сам.
- Страж привязан к кластеру и может быть перемещён между кластерами (drag-and-drop, touch-оптимизировано для мобильного).
- Сила стража зависит от свежести заметок в кластере, общего количества связей, завершённости кластера и наличия архивных заметок.
- **Волны забвения** пассивно атакуют кластер, если он не обновлялся M дней (ежедневно/еженедельно). Страж автоматически отбивает атаку (авто-анимация), если кластер активен.
- Push-уведомления: «Волна забвения приближается к кластеру "Философия". Достоевский готов защищаться».
- Перемещение стража = переприоритизация: кластер-донор тускнеет, кластер-акцептор усиливается.
- **Семантическая обратная связь:** страж, названный пользователем («Достоевский», «Проект Альфа»), служит напоминанием о заброшенной теме.
- **Гипотеза:** персонифицированные стражи (особенно в мобильном приложении с push-уведомлениями) создают эмоциональную связь с заметками и мотивируют возвращаться к заброшенным темам.
- **MVP:** ручное именование стража кластера, drag-and-drop между кластерами, одна фигурка-силуэт, анимация волны забвения; мобильный PWA-интерфейс (`features/guardians/` + `widgets/guardians-viewer/`) с push-уведомлениями.
- **Зависимости:** Phase 11 (Galactic Clusters), Phase 15 (Archive & Note Hygiene), NLP-сервис для будущей авто-генерации имён.
- **Validation:** рост возвратов к «забытым» кластерам на ≥20% в течение месяца; ≥30% активных пользователей дают имя хотя бы одному стражу.

### Phase 15: Archive & Note Hygiene (Архив и гигиена заметок) 🟡 Medium — Planned

**Priority:** 🟡 Medium (перед Semantic Guardians)
**Status:** ⏳ Planned
**Description:** Механика автоматической архивации неиспользуемых заметок.

- Заметка, не обновлявшаяся N дней, с низкой активностью связей и не помеченная как важная, автоматически помечается как «забытая».
- Забытые заметки тускнеют в графе, связи прерываются.
- При достижении критического порога заметка уходит в **Архив** (скрывается из основного графа).
- Архив — отдельный слой графа, не влияющий на основную навигацию. Режим просмотра архива.
- Пользователь может вручную архивировать/разархивировать заметки.
- После Phase 14 страж кластера может защищать заметки в кластере от архивации.
- Метрика «здоровья графа»: процент активных заметок к общему числу.
- **Гипотеза:** автоматическая архивация снижает когнитивную нагрузку и поддерживает граф в актуальном состоянии.
- **MVP:** ASYNQ scheduled task для пометки «забытых» заметок, UI-переключатель в архив, endpoint для ручной архивации.
- **Зависимости:** нет (базовая механика над существующей моделью заметок); интеграция со стражами (Phase 14) — опционально.
- **Validation:** ≥70% автоматически архивированных заметок не разархивируются вручную в течение месяца; пользователи отмечают, что граф стал «чище» (опрос/фидбек).

### Phase 16: Leaderboards & Public Universes 🟢 Low — Hypothesis

**Priority:** 🟢 Low (после социальных фич)
**Status:** 💡 Idea
**Description:** Публичный рейтинг пользователей по активности в графе плюс публичная страница «вселенной» пользователя.

- Backend: агрегация метрик (число заметок, связей, streak) с кэшем в Redis.
- Топ по количеству заметок/связей/streak, обновление раз в N часов (не realtime).
- Публичная страница вселенной `routes/u/[username]` — переиспользует существующий публичный граф (`/api/v1/graph/public`).
- Приватность: пользователь может выключить показ в лидерборде, оставив публичный граф видимым (или наоборот).
- **Гипотеза:** видимость прогресса и чужих «вселенных» повышает вовлечённость и мотивирует держать граф в порядке.
- **MVP:** топ-N лидерборд по количеству заметок (`routes/leaderboard`) + публичная страница вселенной, переиспользующая публичный граф.
- **Зависимости:** публичные заметки ✅ (готово), 🎮 Gamification (очки/ачивки, сейчас в бэклоге).
- **Validation:** ≥15% пользователей с публичными заметками просматривают лидерборд минимум раз в неделю.

### Phase 17: Social Sharing (Пересылка заметок) 🟡 Medium — Planned

**Priority:** 🟡 Medium
**Status:** ⏳ Planned
**Description:** Отправка заметки другому пользователю по username, с входящим Inbox у получателя.

- Backend: таблица `note_shares` (`note_id`, `sender_id`, `recipient_id`, `status`, `created_at`).
- Endpoints: `POST /api/v1/shares` (создать шар), `GET /api/v1/shares/inbox` (список входящих).
- Frontend: кнопка «Поделиться» на заметке (поиск получателя по username), Inbox-панель (вписывается в правую панель Cosmic Cockpit), анимация «входящей кометы» для новых заметок.
- Разрешения: отправить можно любую свою заметку; получатель либо копирует заметку себе, либо просто просматривает (решить на MVP-этапе).
- **MVP:** шаринг по username + список входящих с принять/отклонить.
- **Зависимости:** поиск пользователя по username, модель прав на заметки.
- **Related:** входит в кандидаты TD-3 (Full SSE) для мгновенного уведомления получателя вместо появления при следующей загрузке списка.

### Phase 18: Social Layer (Чат и гильдии) 🟢 Low — Hypothesis

**Priority:** 🟢 Low (последняя фаза — нужна набранная аудитория)
**Status:** 💡 Idea
**Description:** Личные сообщения между пользователями, в перспективе — гильдии/группы вокруг общих графов.

- Backend: таблицы `messages`, `conversations`.
- Транспорт: реализуется поверх SSE-хаба из TD-3 (`GET /api/v1/events/stream`), с fallback на polling при недоступности SSE.
- Frontend: `features/chat/` + виджет чата, интеграция с Cosmic Cockpit панелями.
- Гильдии (группы, общий доступ к части графа) — отложить до валидации личных сообщений на реальных пользователях.
- **Гипотеза:** прямое общение вокруг общих тем/заметок удерживает пользователей и создаёт сетевой эффект.
- **MVP:** 1:1 личные сообщения без гильдий.
- **Зависимости:** Phase 17 (Social Sharing), желательно TD-3 (Full SSE) для доставки в реальном времени.

### Phase 19: Periodic Notes (Периодические заметки) 🟡 Medium — Idea

**Priority:** 🟡 Medium
**Status:** 💡 Idea
**Description:** Тип заметки Pulsar с периодом повторения и чеклистом — для повторяющихся тем/рутин (обзоры, привычки, регулярные ревью).

- Backend: поля `period_days`, `checklist` (JSONB) в `notes`.
- Уведомление о просрочке чеклиста — кандидат на TD-3 (Full SSE) для realtime, с fallback на обычный toast/почту.
- Frontend: `ChecklistEditor.svelte`, визуальный индикатор просрочки на узлах типа Pulsar.
- **MVP:** ручное поле `period_days` + чеклист с подсветкой просрочки; без авто-планирования повторов.
- **Зависимости:** нет (расширяет существующую модель заметок).

### 💡 DB-backed Runtime Configuration 🟢 Low — Idea

**Priority:** 🟢 Low  
**Status:** 💡 Idea  
**Description:** Store rarely-changed runtime tunables in a `system_config` table and expose a protected admin API for CRUD + manual reload.

- **Scope:**
  - Runtime tunables only: rate limits, graph/recommendation/pagination limits, password policy, UI thresholds.
  - No secrets (JWT_SECRET, SMTP passwords, OAuth credentials remain env/JSON only).
- **API:**
  - `GET /api/v1/admin/system-config` — list all entries.
  - `POST /api/v1/admin/system-config` — create a key/value pair.
  - `PATCH /api/v1/admin/system-config/:key` — update value.
  - Protected by `STATIC_API_KEY` (or `RequireAdmin()` later).
- **Runtime effect:** load from DB at startup; manual reload endpoint to refresh in-memory `config.Config`.
- **Validation:** deferred to implementation; consider a strict key/type registry.
- **Risks:** drift between `knowledge-graph.config.json` and DB; caching/reload semantics across backend instances and worker.
- **MVP:** migration + repository + service + handler + reload endpoint + integration tests.
- **Dependencies:** existing `user_settings` pattern, `RequireAdmin` middleware, `CacheClient`.
- **Validation criteria:** changing a rate limit via API and reload affects live requests without restart; no invalid keys accepted after validation is implemented.

### 💡 Graph Motion Driven by Graph Metrics & Clusters 🟢 Low — Idea

**Priority:** 🟡 Medium (after Phase 11 Galactic Clusters)  
**Status:** 💡 Idea  
**Description:** Move beyond randomized per-node motion and tie visual animation parameters to graph structure so clusters and hubs are visually distinct.

- **Scope:**
  - Rotation speed, particle density, corona activity and glow frequency should depend on:
    - node degree (hubs pulse/spin faster or slower),
    - distance from graph center,
    - cluster membership (`cluster_id` from graph-service),
    - note type / anomaly flag.
  - Keep the current fast-fix randomization as a fallback when structural data is unavailable.
- **Runtime data needed:**
  - `cluster_id` in `/graph/full` response (Phase 11).
  - Per-node degree and/or `closeness` pre-computed by graph-service or derived from links.
  - Optional: semantic "mass" from embeddings/NLP.
- **Visual mappings (examples):**
  - High-degree nodes: slower, heavier rotation; stronger glow.
  - Peripheral nodes: faster, smaller particles.
  - Nodes in the same cluster: share a subtle color sway or orbital direction bias.
  - Comets with many outbound links: longer, more active tails.
- **MVP:** add `degree`, `cluster_id`, `distance_from_center` fields to graph payload; read them in `GraphCanvas` and bias the existing `animation.ts`/`particle-system.ts` parameters.
- **Dependencies:** Phase 11 (Galactic Clusters), graph-service API enrichment, possibly Phase 16 (DB-backed config) for tunable mapping coefficients.
- **Risks:** over-engineering motion can hurt performance on 500+ nodes; needs graceful degradation.
- **Validation criteria:** watching the graph, users can visually guess which nodes are hubs/central vs peripheral; motion no longer feels like a uniform screensaver.

_Add future ideas above the next section break._

---


## Visualization and Interface
- [ ] **Graph clustering / fog of war** — grouping stars by regions when zoomed out, revealing when zoomed in (like Prezi).
- [x] **Node and link appearance animation** — smooth appearance of "stars" when loading the graph. **Implemented in simulation.ts (fade-in animation).**
- [x] **Color coding links by weight** (thickness and color). **Implemented in renderer.ts (gamma-link color coding, dashed lines).**
- [ ] **Keyboard shortcuts** — quick navigation, note creation, saving.
- [ ] **🚀 Spaceship Navigator** — interactive starship that follows cursor, points at nodes, drifts idly (feature/explorer-update)
  - 3 modes: FOLLOW_CURSOR, POINT_AT_NODE, IDLE_DRIFT
  - Tooltips from GalacticLexicon
  - Segmented node loading waves
  - Camera zoom animation (0.5→1.0)

## Integrations
- [ ] **Calendar and messaging synchronization** — Google Calendar, Telegram, meeting reminders.
- [ ] **Note import/export** (Markdown, JSON, Obsidian) — planned for [Phase 4](BACKLOG.md#phase-4-cosmic-navigator-spaceship).
- [ ] **Parser for loading text** — automatic creation of notes and links from large text (e.g., books).

## Functionality
- [ ] **Note versioning** (change history, diff).
- [x] **Different coefficients for different link types** (reference, dependency, related). **Implemented (link_type field with weights).**
- [ ] **GraphQL** — for complex aggregated queries (suggestions + recommendations + graph in one request).
- [ ] **Notification queues** — notifications for new links/recommendations.
- [ ] **Orchestration service (Saga)** — for distributed transactions (as it grows).
- [ ] **🏆 Points & Cosmetics System** — gamification with points, daily bonuses, and shop (feature/explorer-update)
  - Points for actions: CreateNote(+1), CreateLink(+2), PublishNote(+5)
  - Daily bonus streaks: 2-day +2, 7-day +10
  - Cosmetics shop: spaceship skins, engines, trails, satellites
  - API: `/users/me/points`, `/cosmetics`, `/cosmetics/buy`

## Mobile and Cross-platform
- [ ] **Mobile app** (React Native / Flutter or PWA) — access knowledge base from phone.
- [ ] **Responsive design** — support for tablets and mobile browsers.

## Other
- [ ] **Music and sound effects** — optional theme for atmosphere.
- [ ] **Russian and English language support in NLP** (already partially available, can be expanded).
- [ ] **Fault tolerance and horizontal scaling** — service replicas (as load grows).
