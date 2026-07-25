# Дополнительный чек-лист ручного тестирования: ai-agents / 3D-рефакторинг

**Ветка:** `ai-agents` / `feature/graph3d-unfreeze`  
**Базовый коммит:** `025777e`  
**Фокус:** только доработки ветки и их взаимодействие с остальной системой.  
**Окружение:** isolated test stack (`scripts/testing/start-test.ps1`).

> Этот чек-лист **дополняет** [MANUAL_TEST_CHECKLISTS_RU.md](MANUAL_TEST_CHECKLISTS_RU.md), а не заменяет его.

---

## Перед стартом

- [ ] `scripts/testing/stop-test.ps1` — убедиться, что dev/personal стеки остановлены.
- [ ] `scripts/testing/start-test.ps1` — test stack поднят.
- [ ] `scripts/testing/seed-test-data.ps1` — 100 notes + 60 links созданы.
- [ ] `curl http://localhost:8083/health` → `{"status":"ok"}`.
- [ ] `curl http://localhost:9095/health` → `OK` (graph-service, порт из `docker-compose.test.yml`).
- [ ] `curl http://localhost:3002` → HTTP 200.
- [ ] Открыть DevTools Network + Console.

---

## 1. FSD-рефакторинг 3D-графа

### Ленивая загрузка и WebGL
- [ ] Открыть `http://localhost:3002` в **2D** виде. В Network не должен загружаться бандл `Graph3DScene` (только `Graph3DViewer` shell, ~небольшой).
- [ ] Переключиться в **3D**. Должен динамически загрузиться `features/graph-3d/ui/Graph3DScene.svelte`.
- [ ] Включить эмуляцию отсутствия WebGL (DevTools → Sensors / или отключить WebGL в about:config). Появляется user-facing ошибка с `data-testid="graph-3d-error"`.
- [ ] В Console нет необработанных ошибок `THREE`/`WebGL` при инициализации.

### Структура фичи
- [ ] Код 3D-графа находится в `features/graph-3d/`, виджет — `widgets/graph-3d-viewer/`.
- [ ] `components/organisms/GraphCanvas3D.svelte` удалён (не используется).
- [ ] `shared/stores/graph.svelte.ts` импортируется из `routes/+page.svelte`, `routes/graph/3d/+page.svelte` и `routes/graph/3d/[id]/+page.svelte` без циклических зависимостей.

---

## 2. Переключение видов и shared graphStore

- [ ] На главной (`/`) кликнуть **3D** в `FloatingControls` — `graphStore.currentView` становится `"3d"`.
- [ ] Кликнуть **List** — отображается список заметок.
- [ ] Кликнуть **Graph** — возвращается 2D canvas.
- [ ] Переключение между видами не сбрасывает `graphStore.selectedNodeId` (если нода была выбрана).
- [ ] Выбрать ноду в **2D** — открывается `NoteSidePanel`. Переключиться в **3D** — та же нода подсвечена (`selectedNodeId` синхронизирован).
- [ ] В **3D** кликнуть другую ноду — `graphStore.selectedNodeId` обновляется, `NoteSidePanel` открывается.
- [ ] Вернуться в **2D** — выделена новая нода.

---

## 3. Graph-service интеграция и LayoutProvider

### Полный 3D-граф
- [ ] Перейти на `/graph/3d`. В Network виден запрос к `graph-service`:
  - `GET http://localhost:8083/graph-service/api/v1/graph/full` (через SvelteKit proxy) **или** напрямую `http://localhost:9095/api/v1/graph/full`.
- [ ] Ответ содержит `nodes` с полями `id`, `title`, `type`, `x`, `y`, `z`.
- [ ] Узлы отрисовываются в 3D, камера авто-фитит сцену.
- [ ] Повторное открытие `/graph/3d` при наличии кэша в Redis (`graph-service:full:public` / `:userID`) отдаётся быстрее (по логам graph-service `Cache hit`).

### Note-centric 3D-граф
- [ ] Перейти на `/graph/3d/<note-id>` (например, кликом "Open in 3D" из `NoteSidePanel` при наличии такой ссылки).
- [ ] Запрос: `GET /api/v1/graph/note/<id>?depth=3&layout=3d`.
- [ ] Центральная нода — запрошенная заметка; depth=3 ограничивает соседей.
- [ ] Параметр `layout=3d` не кэшируется под тем же ключом, что и 2D (проверить Redis).

### Главная страница в 3D
- [ ] На `/` переключить в **3D**. Граф строится из `filteredGraphData`, которая приходит из `getFullGraphData()` (`+page.svelte`) — запрос выполняется один раз.
- [ ] Узлы имеют координаты `x/y/z` от graph-service (в Console `[+page] First 5 raw nodes` должны содержать `x`, `y`, `z`).
- [ ] Если позиции пришли, `Graph3DEngine` делает `warmStartTicks=10` вместо 80 (быстрый старт).
- [ ] На главной **2D** и **List** используется тот же единый запрос к `getFullGraphData()` через `getGraphWithPreload()`, а fallback-нормализация и повторный `loadGraphData()` удалены.
- [ ] После создания/изменения/удаления заметки `+page.svelte` сначала пробует `/api/v1/graph/delta?last_hash=` через `PreloadService.updateWithDelta()`, и только при отсутствии `lastHash` или ошибке делает полную перезагрузку `loadData()`.

### Формат ответа
- [ ] Связи возвращаются с полями `source`, `target`, `weight`, `link_type` (не `source_note_id`/`target_note_id`).
- [ ] На фронте fallback-нормализация `source_note_id`/`target_note_id` в `+page.svelte` не используется.

---

## 4. Fog / performance presets (если реализовано в ветке)

- [ ] Открыть 3D-граф. Туман плавно развеивается от `fogDensityInitial` к `fogDensityFinal` по мере стабилизации симуляции.
- [ ] Если реализованы пресеты (`birth`/`nebula`/`deep-space`): переключение пресета меняет `scene.fog.density` и цвет.
- [ ] `performance-monitor` замеряет FPS в `Graph3DEngine.frame()`.
- [ ] При падении FPS ниже порога (из `knowledge-graph.config.json` → `frontend.graph.performance.fps_threshold`) эффекты снижаются:
  - уменьшается плотность тумана;
  - отключаются анимации;
  - или снижается `starfield` count.
- [ ] На слабом устройстве / эмуляции CPU throttling граф остаётся интерактивным.

---

## 5. Publish / Unpublish (если реализовано в ветке)

- [ ] В `NoteEditor` или `NoteCard` есть тумблер/кнопка "Публичный доступ".
- [ ] `POST /api/v1/notes/<id>/publish` → заметка становится `is_public = true`.
- [ ] `POST /api/v1/notes/<id>/unpublish` → `is_public = false`.
- [ ] Пользователь может публиковать только свои заметки (проверка `creator_id`).
- [ ] После publish/unpublish в публичном графе (инкогнито) заметка появляется/исчезает (с учётом кэша / событий).
- [ ] Приватные заметки не видны неаутентифицированным пользователям ни в графе, ни в списке.

---

## 6. Graph-service: авторизация и события (если реализовано в ветке)

### Авторизация
- [ ] Аутентифицированный пользователь видит только свои заметки в graph-service (запросы фильтруются по `creator_id`).
- [ ] `user_id` не передаётся в query string к graph-service.
- [ ] Запрос без валидного токена к приватному endpoint graph-service возвращает 401.
- [ ] Публичный endpoint graph-service (`/api/v1/graph/public` или аналог) отдаёт только `is_public = true` без auth.

### Событийная инвалидация
- [ ] Создать заметку в UI. В логах graph-service появляется событие `NoteCreated` и `Cache invalidated`.
- [ ] Удалить связь. Событие `LinkDeleted` инвалидирует `graph-service:full:*` и `graph-service:note:*`.
- [ ] Следующий запрос графа не берёт устаревший кэш.

---

## 7. Fallback и регрессия

- [ ] Остановить `graph-service` контейнер (`docker stop kg-test-graph-service`).
- [ ] Обновить страницу `/graph/3d` — frontend/backend fallback отрабатывает (или показывает понятную ошибку, а не пустой canvas).
- [ ] Вернуть `graph-service` (`docker start kg-test-graph-service`).
- [ ] 2D-граф (`/`) продолжает работать: hover, drag-and-drop связи, ghost node, удаление, поиск.
- [ ] List view: фильтрация, сортировка, поиск, batch delete не сломаны после 3D-переключений.
- [ ] Note CRUD: создание/редактирование/удаление заметки обновляет и список, и 2D-граф, и 3D-граф.

---

## 8. Мониторинг и логи

- [ ] В Console нет `401`-циклов при переключении видов.
- [ ] В Network нет повторных запросов `getFullGraphData` при переключении 2D ↔ 3D (если данные не изменились).
- [ ] При удалении заметки `+page.svelte` вызывает `loadNotes()` и `loadGraphData()` корректно (нет race condition).
- [ ] Redis не содержит висячих ключей `graph-service:*` после `scripts/testing/stop-test.ps1`.

---

## Как сообщить о дефекте

Для каждого упавшего пункта:

1. Окружение: браузер, ОС, стек (dev/personal/test).
2. Скриншот или короткая запись экрана.
3. Логи Console / Network / backend / graph-service (`docker logs kg-test-graph-service`).
4. Точные шаги воспроизведения.
5. Ожидаемый vs фактический результат.
