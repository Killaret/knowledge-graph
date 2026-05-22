# Frontend Patterns — Knowledge Graph

Этот документ фиксирует реальные архитектурные и UI/UX паттерны фронтенда проекта, которые уже используются в коде и описаны в документации.

## 1. Основные архитектурные принципы

### 1.1 Компонентная архитектура на Svelte 5
- UI состоит из небольших переиспользуемых компонентов в `frontend/src/lib/components/`.
- Основные компоненты:
  - граф: `GraphCanvas.svelte`, `SmartGraph.svelte`, `Graph3D.svelte` (заморожен)
  - заметки: `NoteCard.svelte`, `NoteEditor.svelte`, `NoteSidePanel.svelte`
  - модальные окна: `Modal.svelte`, `ConfirmModal.svelte`, `CreateNoteModal.svelte`, `EditNoteModal.svelte`
  - общие элементы: `Button.svelte`, `SearchBar.svelte`, `Sidebar.svelte`, `ToastNotification.svelte`
- Паттерн «атомы → молекулы → организмы» реализуется через маленькие UI-компоненты и композицию.

### 1.2 Разделение ответственности
- `frontend/src/lib/api/` — внешний API-клиент и HTTP-слой.
- `frontend/src/lib/services/` — бизнес-логика, предзагрузка данных, side-effects.
- `frontend/src/lib/stores/` — глобальное состояние, реактивные хранилища.
- `frontend/src/lib/utils/` — утилиты и вспомогательные функции.
- `frontend/src/lib/mocks/` — повторно используемые мок-компоненты для тестового окружения.
- `frontend/src/lib/test-utils/` — хелперы для unit-тестов.

### 1.3 Централизованная конфигурация
- `frontend/src/lib/config.ts` читает `knowledge-graph.config.json` из корня проекта.
- Это единственный runtime-конфиг для фронтенда, значения которого управляются из `config/*.json` в корне.
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
- `frontend/src/lib/services/PreloadService.ts` реализует singleton `PreloadServiceClass`.
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
- Preload: `vitest` с config `vitest.config.preload.ts`
- Visual: `playwright test tests/visual/`

### 3.2 Мокирование API и окружения
- `vitest.config.ts` настраивает alias для тестов:
  - `$app/environment` → `src/lib/mocks/app/environment.ts`
  - `$app/navigation` → `src/lib/mocks/app/navigation.ts`
  - `$app/stores` → `src/lib/mocks/app/stores.ts`
  - `$lib` → `src/lib`
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
- Компоненты должны использовать `src/lib/api/*` для любых HTTP-запросов.
- Бизнес-логика предпочтительнее держать в `src/lib/services/*`, а не в `.svelte` файлах.
- Глобальное состояние должно быть оформлено через `src/lib/stores/*`.
- 3D-функциональность остаётся замороженной для стабильности.

### 5.2 Менее фундаментальные, но важные правила
- Использовать `Modal.svelte` как основу диалогов и избегать дублирующего modal-кода.
- Тестировать компоненты через `@testing-library/svelte` и MSW, избегая реального HTTP в unit-тестах.
- Стабилизировать тестовое окружение через mocks для анимаций, canvas и браузерных API.
- Для новых UI-компонентов следовать графовой тематике и единому стилю сообщений `Galactic Lexicon`.
- При изменении `frontend/src/lib/config.ts` проверять, что alias `$config$` остаётся доступным для Vitest.

## 6. Что сделать дальше
- Создать `frontend/test-results/temp/.gitignore` и настроить `frontend/.gitignore` на игнорирование temp-артефактов.
- Добавить в `frontend/README.md` раздел `Frontend patterns` с ссылкой на этот файл.
- При необходимости оформить `frontend/FRONTEND_RULES.md` как лёгкий свод правил для команды.
