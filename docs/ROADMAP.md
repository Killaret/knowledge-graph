# Knowledge Graph Roadmap v2.0

**⚠️ DEPRECATED:** This file has been moved to the project root.  
**📍 New Location:** [ROADMAP.md](../ROADMAP.md)

Please use the updated roadmap in the project root for the latest development plans.

---

## 📋 Единый Roadmap (Legacy)

| Приоритет | Фаза | Задачи | Статус |
|-----------|------|--------|--------|
| 🔴 Critical | Стабилизация Personal | Исправить проксирование, идентичные данные dev/personal, сравнить отрисовку графа, бэкапы в Яндекс.Диск | 🔄 В процессе |
| 🔴 Critical | Интеграционные тесты | Исправить циклический импорт, ошибки типов, конструкторы handler'ов | ⏳ Промт готов |
| 🔴 High | Импорт/Экспорт | JSON/Markdown/CSV импорт и экспорт, Obsidian-совместимость | ⏳ Промт готов |
| 🔴 High | Космический Навигатор | SVG-звездолёт, режимы следования за курсором, сегментированная загрузка узлов, тултипы из лексикона | ⏳ Промт готов |
| 🟡 Medium | Очки и Кастомизация | Таблицы user_points, user_cosmetics, магазин скинов, ежедневные бонусы | ⏳ Промт готов |
| 🟡 Medium | Активация Аномалий | Подключить drawRealityRift, drawChromaticMaw, drawVoidWhisper, drawCosmicAbomination для типа unknown | ⏳ Микро-промт готов |
| 🟢 Low | Мягкое внедрение FSD | Выделить shared/, entities/, настроить алиасы, обновить импорты, документировать | ✅ Готово |
| 🟢 Low | Улучшение Связей | Тултипы при наведении на связь, цветовое кодирование gamma-связей, анимация появления связей | ✅ Готово |
| 🟢 Low | Публичный Пул | Публикация заметок, лайки, закладки, форк | ⏳ Промт готов |
| 🟢 Low | Обработчик Пылинкок | Асинхронная NLP-обработка dust-заметок, предложение типа и тегов, DustInboxPanel | ⏳ Промт готов |
| 🟢 Low | Галактический Лексикон + Достижения | i18n, SSE-уведомления, seed-достижений | ⏳ Промт готов |
| 🟢 Low | Импорт из Obsidian | Парсинг Markdown, YAML frontmatter, [[wikilinks]] → связи | ⏳ Промт готов |
| 🟢 Low | Hierarchical Clustering | Multi-level keyword clustering, spatial precomputation, zoom-aware rendering | ⏳ План в ARCHITECTURE_ROADMAP.md |
| 🟢 Low | Мультиарендность (SaaS) | RLS, tenants, RBAC, миграция данных | ⏳ План в ARCHITECTURE_SUMMARY.md |
| 🟢 Low | Мобильная версия / PWA | Адаптивная вёрстка, offline-режим | ⏳ Бэклог |
| 🟢 Low | Интеграция с внешними API | Pocket, Readwise, Twitter | ⏳ Бэклог |

---

## 🎯 Product Development Strategy

This plan focuses on turning Knowledge Graph into a reliable and useful product for real-world use, as opposed to the architectural SaaS migration plan (see `ARCHITECTURE_ROADMAP.md`).

---

## 🎯 Phase 1: Stability & Personal Use (1–2 days)
**Goal:** Turn the system into a reliable personal tool

### Tasks

#### 🔧 Audit & Bug Fixes
- [ ] Run full system audit (already done, critical issues found)
- [ ] **CRITICAL:** Fix backend compilation errors
  - Update redis client v9 API (MaxConnAge → ConnMaxIdleTime, IdleTimeout → ConnMaxLifetime)
  - Fix undefined function `quit` → `Quit`
  - Update sql.DBStats API (MaxIdleConns field removed)
- [ ] **CRITICAL:** Remove real OAuth token from code
  - Revoke token in Yandex OAuth
  - Replace with placeholder in knowledge-graph.config.json
  - Add .env variable BACKUP_YANDEX_OAUTH_TOKEN
- [ ] Fix mocks in frontend tests (getCachedGraph, getFreshGraph)
- [ ] Fix mock types in backend tests

#### 💾 Backup Verification
- [ ] Ensure Yandex.Disk backup scheme works
- [ ] Test backup restoration
- [ ] Test scheduled automatic backups
- [ ] Verify data integrity after restoration

#### 📝 Start Using Notes
- [ ] Create personal project diary in Knowledge Graph
- [ ] Daily usage to identify UX issues
- [ ] Document interface rough spots
- [ ] Collect performance and usability feedback

### Definition of Done
- [ ] All critical bugs fixed
- [ ] System compiles and runs successfully
- [ ] Backups verified and working automatically
- [ ] Personal use started, initial feedback collected

---

## 🚀 Phase 2: "Share" — MVP for External Users (3–5 days)
**Goal:** Make the project accessible to other users

### Tasks

#### 🌐 Public Notes Pool
- [ ] Implement public sharing system
  - Generate public links to notes
  - View public notes without authentication
  - Likes and bookmarks for public content
  - Copy public notes to your own
- [ ] Search in public notes pool
- [ ] Public user profiles
- [ ] Content moderation (flags, reports)

#### 📚 Deployment Instructions
- [ ] Write clear `docs/DEPLOY.md`
  - System requirements
  - Docker Compose one-command deployment
  - Environment variable configuration
  - Troubleshooting common issues
- [ ] Update `README.md` with quick start
- [ ] Add FAQ section
- [ ] Create video tutorial (optional)

#### 🧪 Deployment Testing
- [ ] Deploy project on clean machine/VPS
- [ ] Test all functionality in fresh environment
- [ ] Test fresh installation
- [ ] Check production configuration
- [ ] Load testing basic functionality

#### 👶 Onboarding
- [ ] Add welcome "dust particles" (quick notes)
  - "Welcome to Knowledge Graph"
  - "How to get started"
  - "Graph examples"
- [ ] Create interactive UI tour
- [ ] Add hints and tooltips
- [ ] Create example notes of different types
- [ ] Contextual help for first-time users

### Definition of Done
- [ ] Public sharing functionality fully works
- [ ] Any developer can deploy with one command
- [ ] Deployment tested on clean machine
- [ ] New users can quickly onboard

---

## 💎 Phase 3: Killer Features for Growth (1–2 weeks)
**Goal:** Add features that will attract new audience

### Tasks

#### 📥 Obsidian Import
- [ ] Support Obsidian Markdown format
- [ ] Import links (wikilinks)
- [ ] Convert Obsidian graphs to Knowledge Graph format
- [ ] Preserve metadata and tags
- [ ] Batch import entire vaults
- [ ] Video tutorial for import

#### 🌐 Landing Page & SEO
- [ ] Create beautiful landing page
  - Project description
  - Screenshots and demos
  - Features and benefits
  - Call-to-action
- [ ] SEO optimization
  - Meta tags, Open Graph
  - Structured data
  - Sitemap.xml
  - Robots.txt
- [ ] Traffic analytics
- [ ] A/B testing landing page

#### 🔗 Basic Integrations
- [ ] Telegram bot for quick note creation
  - /new command to create note
  - Automatic keyword extraction
  - Send to personal graph
- [ ] Webhook for integrations
  - Zapier/Make integrations
  - GitHub issues → notes
  - Email → notes
- [ ] Browser extension
  - Save pages as notes
  - Text selection → new note

### Definition of Done
- [ ] Obsidian import works flawlessly
- [ ] Landing page attracts traffic
- [ ] Telegram bot functional and documented
- [ ] Basic integrations available

---

## 🚀 Phase 4: Explorer Update (feature/explorer-update)
**Goal:** Add gamification, navigation, and data portability

#### 🚀 Космический Навигатор (Spaceship Navigator)
- [ ] Создать `SpaceshipNavigator.svelte` (SVG, анимация двигателей, ~40x40px)
- [ ] Реализовать 3 режима: FOLLOW_CURSOR, POINT_AT_NODE, IDLE_DRIFT
- [ ] Интеграция с GraphCanvas (координаты мыши, клики по узлам)
- [ ] Тултипы от GalacticLexicon (создание заметок, связей, достижений)
- [ ] Эффект загрузки: патрульный облёт, камера 0.5→1.0
- [ ] Сегментированная загрузка узлов (волны 0-200ms, 200-500ms, 500-800ms)
- [ ] Unit-тест на переключение режимов
- [ ] Визуальный тест (скриншот графа с корабликом)

#### 🏆 Система очков и кастомизации
- [ ] Миграции БД: `user_points`, `point_transactions`, `user_cosmetics`
- [ ] API: `GET /users/me/points`, `GET /cosmetics`, `POST /cosmetics/buy`, `GET /users/me/cosmetics`
- [ ] Начисление очков: CreateNote(+1), CreateLink(+2), QuickCapture(+1), PublishNote(+5)
- [ ] Ежедневные бонусы (Asynq): streak 2-й день +2, 7-й день +10
- [ ] Компонент `CosmeticsShop.svelte` (Modal, Button, ApiErrorDisplay)
- [ ] Применение скинов в SpaceshipNavigator (user_cosmetics)
- [ ] Unit-тест на PointService
- [ ] Интеграционный тест: покупка предмета

#### 📥 Импорт/Экспорт заметок
- [ ] Сервис `ImportExportService` (JSON, Markdown, CSV)
- [ ] API: `POST /import`, `GET /export?format=json|markdown|csv`, `GET /notes/{id}/export?format=md`
- [ ] Markdown-формат (Obsidian-совместимость): YAML frontmatter, [[wikilinks]]
- [ ] UI: кнопки Импорт/Экспорт в профиле/левой панели
- [ ] Unit-тест на ImportFromJSON
- [ ] Unit-тест на ExportToMarkdown
- [ ] Интеграционный тест: экспорт → импорт → сравнение

---

## 📋 Backlog (someday)

### Multi-tenancy (SaaS)
- [ ] Full SaaS architecture (see `ARCHITECTURE_ROADMAP.md`)
- [ ] Single sign-on for teams
- [ ] Access control management
- [ ] Billing and subscriptions

### Mobile Version
- [ ] React Native app
- [ ] Offline mode
- [ ] Push notifications
- [ ] Desktop synchronization

### 3D Graph Return
- [ ] 3D rendering optimization
- [ ] VR/AR support
- [ ] Interactive 3D scenes
- [ ] Performance improvements

### AI Features
- [ ] Automatic note summarization
- [ ] Idea and connection generation
- [ ] Voice-to-text for notes
- [ ] Image recognition for visual content

### Collaboration
- [ ] Real-time note editing
- [ ] Comments and discussions
- [ ] Version history with diff
- [ ] Activity streams

---

## 💎 Most Important Right Now

### Priority Actions for Today:

1. **Start keeping your diary in Knowledge Graph**
   - This is the best way to find UX issues
   - Daily usage will reveal real needs
   - Collect performance feedback

2. **Fix critical bugs**
   - Compilation errors (blocking usage)
   - Security vulnerability (OAuth token)
   - Tests for stability

3. **Verify backups**
   - Ensure data is safe
   - Test restoration
   - Automate backups

### Success Metrics

**Phase 1 (1-2 days):**
- ✅ System works stably for personal use
- ✅ Backups verified and reliable
- ✅ List of UX issues collected from personal experience

**Phase 2 (3-5 days):**
- ✅ Other users can use the system
- ✅ Deployment documented and automated
- ✅ Onboarding helps new users

**Phase 3 (1-2 weeks):**
- ✅ New features attract audience
- ✅ Integrations expand capabilities
- ✅ Project grows and develops

---

## 📊 Progress Tracking

### Current Status: **Phase 1 - 40% Complete, Phase 2 - 30% Complete**

| Phase | Status | Progress | Priority |
|-------|--------|----------|----------|
| Phase 1: Stability | 🔄 In Progress | 40% | 🔴 Critical |
| Phase 2: MVP for Users | 🔄 In Progress | 30% | 🟡 High |
| Phase 3: Killer Features | ⏸️ Waiting | 0% | 🟢 Medium |
| Phase 4: Explorer Update | 🔄 In Progress | 10% | 🟡 High |
| Backlog | ⏸️ Waiting | 0% | 🔵 Low |

### Near-term Goals (next 7 days):

1. **Fix critical bugs** (day 1-2)
2. **Set up reliable backups** (day 2-3)
3. **Start personal use** (day 1-7, continuous)
4. **Collect UX feedback** (day 3-7)

### Long-term Goals (next 30 days):

1. **Complete Phase 1** (day 1-2)
2. **Launch Phase 2** (day 3-7)
3. **Start Phase 3** (day 8-21)
4. **First public users** (day 14-30)

---

## 🔄 Regular Updates

This roadmap will be updated as tasks are completed and priorities change. Updates will be published in CHANGELOG.md.

**Last updated:** 2026-07-04
**Next review:** 2026-07-11

---

*Roadmap v2.0 — объединённый трекер задач с приоритетами и статусами*