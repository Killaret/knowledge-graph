# Ideas for Knowledge Graph Growth (Not Planned)

## Visualization and Interface
- [ ] **Graph clustering / fog of war** — grouping stars by regions when zoomed out, revealing when zoomed in (like Prezi).
- [ ] **Node and link appearance animation** — smooth appearance of "stars" when loading the graph.
- [ ] **Color coding links by weight** (thickness and color).
- [ ] **Keyboard shortcuts** — quick navigation, note creation, saving.

## Integrations
- [ ] **Calendar and messaging synchronization** — Google Calendar, Telegram, meeting reminders.
- [x] **Note import/export** (Markdown, JSON, Obsidian) — **planned in [Phase 5 roadmap](./ARCHITECTURE_ROADMAP.md#phase-5-obsidian-import-killer-feature)**.
- [ ] **Parser for loading text** — automatic creation of notes and links from large text (e.g., books).

## Functionality
- [ ] **Note versioning** (change history, diff).
- [ ] **Different coefficients for different link types** (reference, dependency, related).
- [ ] **GraphQL** — for complex aggregated queries (suggestions + recommendations + graph in one request).
- [ ] **Notification queues** — notifications for new links/recommendations.
- [ ] **Orchestration service (Saga)** — for distributed transactions (as it grows).

## Mobile and Cross-platform
- [ ] **Mobile app** (React Native / Flutter or PWA) — access knowledge base from phone.
- [ ] **Responsive design** — support for tablets and mobile browsers.

## Other
- [ ] **Music and sound effects** — optional theme for atmosphere.
- [ ] **Russian and English language support in NLP** (already partially available, can be expanded).
- [ ] **Fault tolerance and horizontal scaling** — service replicas (as load grows).
