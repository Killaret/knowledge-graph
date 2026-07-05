# Ideas for Knowledge Graph Growth (Not Planned)

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
- [ ] **Note import/export** (Markdown, JSON, Obsidian) — planned for [Phase 4](./ROADMAP.md#phase-4-explorer-update-featureexplorer-update).
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
