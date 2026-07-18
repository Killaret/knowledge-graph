# Frontend Patterns — Knowledge Graph

Этот документ фиксирует реальные архитектурные и UI/UX паттерны фронтенда проекта, которые уже используются в коде и описаны в документации.

## 1. Основные архитектурные принципы

### 1.1 Компонентная архитектура на Svelte 5

- UI состоит из небольших переиспользуемых компонентов в `frontend/src/components/`.
- Структура Atomic Design:
  - `frontend/src/components/atoms/` — `Button.svelte`, `Input.svelte`, `Badge.svelte`, `Icon.svelte`
  - `frontend/src/components/molecules/` — `SearchBar.svelte`, `NoteCard.svelte`, `TagList.svelte`
  - `frontend/src/components/organisms/` — `GraphCanvas.svelte`, `NoteSidePanel.svelte`, `Sidebar.svelte`, `NoteEditor.svelte`
- Основные компоненты:
  - граф: `GraphCanvas.svelte`, `Graph3D.svelte` (заморожен)
  - заметки: `NoteCard.svelte`, `NoteEditor.svelte`, `NoteSidePanel.svelte`, `CreateNoteModal.svelte`, `EditNoteModal.svelte`
  - модальные окна: `Modal.svelte`, `ConfirmModal.svelte`
  - общие элементы: `Button.svelte`, `SearchBar.svelte`, `Sidebar.svelte`, `ToastNotification.svelte`
- Паттерн «атомы → молекулы → организмы» реализуется через маленькие UI-компоненты и композицию.

### 1.2 Разделение ответственности (FSD + Atomic Design)

- `frontend/src/shared/api/` — внешний API-клиент и HTTP-слой.
- `frontend/src/shared/services/` — бизнес-логика, предзагрузка данных, side-effects.
- `frontend/src/shared/stores/` — глобальное состояние, реактивные хранилища.
- `frontend/src/shared/utils/` — утилиты и вспомогательные функции.
- `frontend/src/shared/mocks/` — повторно используемые моки окружения SvelteKit для Vitest.
- `frontend/src/shared/test-utils/` — хелперы для unit-тестов.
- `frontend/src/shared/types/` — shared TypeScript-типы.
- `frontend/src/shared/styles/` — общие стили и CSS-переменные.
- `frontend/src/shared/config/` — единый runtime-конфиг.
- `frontend/src/components/` — UI-компоненты по Atomic Design.
- `frontend/src/features/` — feature-модули (`graph-interaction/`, `graph-forms/`).
- `frontend/src/shared/lib/domain/` — Value Objects и чистые доменные модели, не зависящие от UI или Canvas (`CelestialBody`).
- `frontend/src/shared/lib/graph/` — вспомогательные графовые утилиты/рендерер (legacy-остаток после FSD); адаптеры рендеринга внедряются из этих модулей, но сами доменные данные хранятся в `domain/`.

### 1.2.1 Domain Value Object — CelestialBody

- Визуальные параметры узлов графа (тип, label, emoji, цвет, glow-цвет, радиусы, масса, скорость вращения, смещение гравитации) централизованы в `frontend/src/shared/lib/domain/celestial-body.ts`.
- `CelestialBody` — immutable Value Object с `static readonly` экземплярами (`STAR`, `PLANET`, `COMET`, `GALAXY`, `NEBULA`, `ASTEROID`, `SATELLITE`, `BLACKHOLE`, `MOON`, `DEBRIS`, `DUST`, `TECHNICAL`, `UNKNOWN`) и четырьмя аномалиями (`REALITY_RIFT`, `CHROMATIC_MAW`, `VOID_WHISPER`, `COSMIC_ABOMINATION`).
- `CelestialBody.fromString(type)` используется для безопасного, регистронезависимого разрешения типа с fallback на `UNKNOWN`.
- Canvas-функции отрисовки подключаются к `CelestialBody` из `GraphCanvas/renderer.ts` через поле `drawFunction`, чтобы доменный объект оставался чистым от Canvas-зависимостей.
- UI-компоненты (`TypeSelector`, `NoteCard`, `NoteSidePanel`, `overlay.svelte`, `+page.svelte`) получают label/emoji/color из `CelestialBody`, исключая дублирование хардкода.

### 1.3 Централизованная конфигурация

- `frontend/src/shared/config/config.ts` читает `knowledge-graph.config.json` из корня проекта.
- `frontend/src/lib/config.ts` удалён; единый runtime-конфиг — `frontend/src/shared/config/`.
- Значения управляются из `config/*.json` в корне.
- Примеры разделов:
  - `frontend.graph['2d']`, `frontend.graph['3d']`
  - `frontend.api.default_limit`
  - `frontend.test.*`
  - `frontend.achievements.poll_interval_ms`

### 1.4 State management и SSR-safe паттерны

- Хранилища реализованы через `$state`-объекты и экспорт функций-аксессоров.
- `auth.svelte.ts` содержит:
  - инициализацию состояния через `localStorage` только в браузере (`browser` из `$app/environment`)
  - методы `login`, `logout`, `register`, `refreshAccessToken`, `isAuthenticated`
  - поддержку `SKIP_AUTH` для тестов через window флаг, query-параметр и localStorage
- Паттерн: разделение логики состояния и UI-компонентов.

### 1.5 Сервисы и фоновые процессы

- `frontend/src/shared/services/PreloadService.ts` реализует singleton `PreloadServiceClass`.
- Паттерн: preload + cache TTL
  - фоновый preload публичного графа (`getFullGraphData`)
  - authenticated preload (`getCachedGraph`, `getFreshGraph`)
  - хранение данных с `timestamp` и `ttl`
- Этот сервис повышает отзывчивость app при входе и переключении между страницами.

### 1.6 Архитектура маршрутов SvelteKit

- Маршруты реализованы через `frontend/src/routes/`:
  - `/` → `+page.svelte`
  - `/graph` → `graph/+page.svelte`
  - `/graph/3d` → `graph/3d/+page.svelte`
  - `/notes/:id` → `notes/[id]/+page.svelte`
  - `/search` → `search/+page.svelte`
- Структура маршрутов соответствует документированному архитектурному обзору.

## 2. UI/UX паттерны

### 2.1 Graph-first визуальный стиль

- Проект ориентирован на графовую визуализацию: основной экран `GraphCanvas`, визуальные ноды, связи и фильтры.
- UI строится вокруг графа как центрального элемента и множества вспомогательных панелей.

### 2.2 Прогрессивная загрузка и отзывчивость

- Используется предзагрузка данных и lazy-loading:
  - `LazyGraph3D.svelte` — ленивая загрузка 3D-графа (заморожена)
  - `PreloadService` — предзагрузка данных до первой активности
- UI должен работать корректно со статусами загрузки, ошибки и пустого состояния.

### 2.3 Общая стилистика и тематизация

- `Galactic Lexicon` отвечает за themed messaging:
  - `standard` и `galactic` режимы
  - локализация `ru` / `en`
  - используется в `ToastNotification`, `ApiErrorDisplay`, `ShareModal` и других компонентах
- Паттерн: единая система тональности сообщений и статусов.

### 2.4 Модальные окна и состояние подтверждения

- Базовый `Modal.svelte` используется как основа для всех диалогов.
- Паттерн: единственная точка контроля модалок + централизованное управление видимостью через `ui.ts`.

### 2.5 Доступность и взаимодействие

- В проекте документировано требование проверять aria-атрибуты, фокусную навигацию и контрастность.
- Тесты UI-функций ориентированы на доступность и стабильность даже при mocked browser APIs.

### 2.6 Стабильность визуальных тестов

- Для тестов в `vitest-setup.ts` смокируются:
  - `Element.prototype.animate`
  - `requestAnimationFrame` / `cancelAnimationFrame`
  - `ResizeObserver`
  - `HTMLCanvasElement.getContext('2d')`
  - `Three.js` рендер и `CSS2DRenderer`
  - модули `GraphCanvas/animation.ts`
- Этот паттерн позволяет запускать UI-тесты в jsdom стабильно.

## 3. Тестовая инфраструктура

### 3.1 Разделение тестовых слоёв

- Unit: `vitest` / `@testing-library/svelte` → `npm run test:unit`
- E2E: `playwright` → `npm run test`
- BDD: `cucumber` → `npm run test:cucumber`
- Visual: `playwright test --project=visual`

### 3.2 Мокирование API и окружения

- `vitest.config.ts` и `svelte.config.js` настраивают alias:
  - `$app/*` → `src/shared/mocks/app/*`
  - `$shared/*` → `src/shared/*`
  - `$components/*` → `src/components/*`
  - `$features/*` → `src/features/*`
  - `$lib` — удалён (legacy `src/lib` больше не существует)
  - `$config$` → `../knowledge-graph.config.json`
- `vitest-setup.ts` запускает MSW и задаёт глобальные fallback-обработчики.
- Тесты покрывают API-слой, компоненты, auth, graph и визуальные состояния.

### 3.3 Playwright и auth bypass

- `playwright.config.ts` использует:
  - `baseURL` из `FRONTEND_URL`
  - `trace: on-first-retry`
  - `forbidOnly` и `retries` для CI
  - `SKIP_AUTH` через query/localStorage/window для обхода аутентификации
- Проект сохраняет минимальную Docker-ориентированную конфигурацию в отдельных `playwright.config.*.ts` файлах.

## 4. Политика временных данных тестов

- Временные артефакты: `frontend/test-results/temp/`
- Золотые эталоны: `frontend/test-results/baseline/`
- Разрешается хранить:
  - скриншоты `frontend/test-results/temp/screenshots/`
  - логи `frontend/test-results/temp/logs/`
  - отчёты Playwright/Vitest в temp
- Не очищать `baseline/` автоматически без явного обновления эталонов.
- CI очищает только `temp/` между прогонками.

## 5. Свод правил

### 5.1 Фундаментальные правила

- Все runtime-параметры фронтенда берутся из `knowledge-graph.config.json`.
- Компоненты должны использовать `src/shared/api/*` (alias `$shared/api/*`) для любых HTTP-запросов.
- Бизнес-логика предпочтительнее держать в `src/shared/services/*` или `src/features/*`, а не в `.svelte` файлах.
- Глобальное состояние должно быть оформлено через `src/shared/stores/*` (runes-based `.svelte.ts` модули).
- UI-компоненты живут в `src/components/` (atoms/molecules/organisms) и не должны содержать доменную бизнес-логику.
- 3D-функциональность остаётся замороженной для стабильности.

### 5.2 Менее фундаментальные, но важные правила

- Использовать `Modal.svelte` как основу диалогов и избегать дублирующего modal-кода.
- Тестировать компоненты через `@testing-library/svelte` и MSW, избегая реального HTTP в unit-тестах.
- Стабилизировать тестовое окружение через mocks для анимаций, canvas и браузерных API.
- Для новых UI-компонентов следовать графовой тематике и единому стилю сообщений `Galactic Lexicon`.
- При изменении `frontend/src/shared/config/config.ts` проверять, что alias `$config$` остаётся доступным для Vitest.

## 6. Что сделать дальше

- Создать `frontend/test-results/temp/.gitignore` и настроить `frontend/.gitignore` на игнорирование temp-артефактов.
- Добавить в `frontend/README.md` раздел `Frontend patterns` с ссылкой на этот файл.
- При необходимости оформить `frontend/FRONTEND_RULES.md` как лёгкий свод правил для команды.
