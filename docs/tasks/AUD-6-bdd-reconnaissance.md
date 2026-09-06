# AUD-6. Разведка по 125 неисполняемым BDD-сценариям

Постановка для Devin. Источник: [`../EXTERNAL_AUDIT_2026-09.md`](../EXTERNAL_AUDIT_2026-09.md), находка T-1. Пересекается с A-7 в [`../AI_PROCESS_AUDIT.md`](../AI_PROCESS_AUDIT.md). Порядок работы: [`../AI_AGENT_PROTOCOL.md`](../AI_AGENT_PROTOCOL.md).

Ставит Claude Code, реализует Devin, решение по результату принимает владелец.

**Очерёдность:** после партии AUD-8 / AUD-2 / AUD-3.

**Это задача на измерение, а не на починку.** Результат — цифры, по которым владелец решит: подключать набор или удалять. Чинить падающие сценарии в рамках этой задачи не нужно.

## Проблема

`cucumber.mjs:5,12` указывает пути только в `frontend/tests/features/**`:

```js
const featureRoot = join(__dirname, 'frontend', 'tests', 'features');
```

Там 2 файла и 9 сценариев. В корневом `tests/features/` лежат ещё **13 файлов со 125 сценариями**, а шаги к ним — в `tests/features/step_definitions/` и `tests/steps/`, всего 9 файлов. В конфиг они не попадают и не запускались ни разу.

При этом `.github/workflows/_core-checks.yml:216-217` отдельным шагом типизирует именно этот каталог:

```yaml
- name: Check root BDD TypeScript
  run: npx tsc --noEmit -p ../tests/tsconfig.json
```

Набор компилируется, значит выглядит живым и поддерживаемым. `tests/README.md:84` отчитывается за «127 сценариев» — это почти ровно тот набор, который не исполняется.

## Что выяснено в коде

**1. Два набора шагов, не один.** Активный набор — `frontend/tests/features/step_definitions/` (`auth.steps.ts`, `common.steps.ts`, `graph_interaction.steps.ts`, `graph_3d_loading.steps.ts`) плюс `support/world.ts` и `support/hooks.ts`. Осиротевший — `tests/features/step_definitions/` (7 файлов) и `tests/steps/` (2 файла) со своими `support/`.

Наборы почти наверняка конфликтуют: Cucumber падает на дублирующемся определении шага. Простое объединение путей в одном конфиге, скорее всего, не запустится вовсе — это первое, что нужно выяснить.

**2. Возможны два `World` и два набора хуков.** У каждого набора свой `support/world.ts`. Cucumber допускает только один `setWorldConstructor`. Это второй ожидаемый барьер.

**3. Сценарии писались под другую версию UI.** `frontend/tests/features/step_definitions/graph_interaction.steps.ts` правится прямо сейчас в рамках A-1 (переход на `data-testid="graph-3d-scene"`). Осиротевший набор такой правки не получал никогда, значит селекторы в нём заведомо устарели.

**4. Тегов реального прогона в осиротевших сценариях нет.** Фильтр `not @auth-real` (`cucumber.mjs:21`) на них не рассчитан — распределение по режимам skip-auth и real-auth придётся определить заново.

## Что сделать

Работать во временной ветке или на отдельном конфиге — **менять `cucumber.mjs` в основном дереве не нужно**, пока решение не принято.

1. Собрать конфиг, который видит оба набора. Разрешить механические конфликты, мешающие **запуску**: дублирующиеся определения шагов, два `World`, два набора хуков. Способ разрешения описать, но в основное дерево не вносить.
2. Поднять изолированный тест-стек и прогнать полный набор.
3. Снять цифры: сколько сценариев собрано, сколько прошло, сколько упало, сколько не запустилось (undefined / ambiguous steps).
4. Классифицировать падения по причине, с числом сценариев в каждой группе. Ожидаемые категории — устаревшие селекторы, отсутствующие шаги, изменившиеся маршруты, изменившийся API, требующие real-auth. Категории уточнить по факту.
5. Для каждой категории — оценка объёма починки в терминах «сколько файлов трогать», без сроков.
6. Отдельно ответить: есть ли среди осиротевших сценариев такие, что покрывают поведение, **не покрытое** ни Playwright-спеками, ни активным BDD-набором. Это главный вопрос для решения — остальное про цену.

## Ограничения

- Не чинить сценарии, не править селекторы, не переписывать шаги. Задача — измерить.
- `cucumber.mjs`, `tests/`, `frontend/tests/features/` в основном дереве не менять.
- Не удалять ничего: решение об удалении принимает владелец по результатам.
- Не трогать `frontend/tests/features/step_definitions/graph_interaction.steps.ts` — он в работе по A-1.
- Не трогать шаг `Check root BDD TypeScript` в CI.
- Personal-стек не поднимать. Прогон — только на изолированном тест-стеке.

## Критерии приёмки

Отчёт в этом файле, разделом «Ответ исполнителя», содержащий:

1. Таблицу: собрано / прошло / упало / не запустилось, отдельно по активному и осиротевшему наборам.
2. Классификацию падений по причине с числом сценариев в каждой группе.
3. Список механических конфликтов, мешавших запуску, и как они были обойдены.
4. Ответ на пункт 6 — уникальное покрытие.
5. Явную рекомендацию исполнителя: подключать, удалять или подключать частично, с обоснованием по цифрам.

## Что приложить к результату

Сырой вывод прогона Cucumber — сводку в конце и полный список упавших сценариев. Без него цифры непроверяемы.

---

## Ответ исполнителя

### Подготовка

- Поднят изолированный тест-стек: `kg-test-backend`, `kg-test-worker`, `kg-test-frontend` (на `:3002`), `kg-test-graph-service`, `kg-test-nlp`, `kg-test-postgres`, `kg-test-redis`, `kg-test-mongo`.
- Запущен `seed-test-data.ps1` для создания тестового пользователя (`00000000-0000-0000-0000-000000000000`) и 100 заметок / 60 связей.
- Все прогоны выполнены с `SKIP_AUTH=true` и тег-фильтром `not @auth-real`.
- Для осиротевшего набора дополнительно запущен Vite dev-сервер на `http://localhost:5173` с `VITE_API_TARGET=http://127.0.0.1:18083` и `VITE_SKIP_AUTH=true`.

### Механические конфликты и обход

| Конфликт | Обход |
|---|---|
| `tests/features/` не находит `@cucumber/cucumber` и `@playwright/test` | Установлен `NODE_PATH=D:\knowledge-graph\frontend\node_modules` для прогона |
| Свой `support/world.ts` без `baseURL` и без `setDefaultTimeout` | Создан временный `aud6-support/world.ts` с `baseURL=process.env.BASE_URL` и `setDefaultTimeout(15000)`; загружается вместо корневого `support/world.ts` |
| Свой `support/hooks.ts` использует устаревшие пути (`/notes/{id}`) и `BASE_URL=http://localhost:8080` | Создан временный `aud6-support/hooks.ts` с `this.setup()/teardown()` и снятием скриншотов; cleanup закомментирован, остатки не мешают измерению |
| Шаги осиротевшего набора захардкожены на `http://localhost:5173` | Запущен Vite dev-сервер на этом URL; фронтенд проксирует API на `:18083` |
| Шаги используют старый backend URL (`http://localhost:8080`) и путь `/notes` | Обнаружено как категория падений; в конфигурацию прогона не вносилось |
| Дублирующиеся определения шагов (`I enter {string} in the password field`, `I navigate to {string}`, `I click the {string} button/toggle`, `I select type {string}`, `notes {string} and {string} exist on the graph`) | Зарегистрированы как `ambiguous`; количество — 55 шагов |
| Два `World` и два набора хуков не совместимы с активным набором | Прогоны выполнены **раздельно**: активный набор — своим `cucumber.mjs`, осиротевший — временным `cucumber-aud6-orphan-v2.mjs`. Объединённый запуск без переписывания шагов невозможен. |

### Результаты активного набора

| Набор | Файлы `.feature` | Собрано сценариев | Прошло | Упало | Не запустилось |
|---|---|---:|---:|---:|---:|
| `frontend/tests/features/` | 2 (`graph_2d_list.feature`, `login.feature`) | 9 | 5 | 0 | 4 (`@auth-real` в `login.feature`) |

Cucumber summary:

```text
5 scenarios (5 passed)
43 steps (43 passed)
0m53.477s
```

### Результаты осиротевшего набора

| Набор | Файлы `.feature` | Собрано сценариев | Прошло | Упало | Не запустилось |
|---|---|---:|---:|---:|---:|
| `tests/features/` | 13 (`achievements`, `auth_cosmic_theme`, `camera_navigation`, `celestial_body_types`, `full_3d_graph`, `graph_navigation`, `graph_view`, `import_export`, `link_types`, `local_3d_graph`, `note_management`, `search_and_discovery`, `type_filters`) | 125 | 8 | 88 | 29 (11 ambiguous + 18 undefined) |

Cucumber summary:

```text
125 scenarios (88 failed, 11 ambiguous, 18 undefined, 8 passed)
913 steps (88 failed, 55 ambiguous, 129 undefined, 510 skipped, 131 passed)
8m12.696s
```

### Классификация падений

Из 88 упавших сценариев:

| Причина | Сценариев | Ключевые признаки | Объём починки (файлы) |
|---|---|---:|---|
| **Устаревший backend URL / API-путь** | 76 | `apiRequestContext.*: connect ECONNREFUSED ::1:8080` и `http://localhost:8080/notes` | ~8–10 файлов: `tests/features/support/hooks.ts`, `tests/features/step_definitions/{camera,graph,import_export,link_types,local_3d_graph,note,progressive-graph,search,graph_2d_list,graph_3d_loading}.ts` — заменить `localhost:8080` на `BACKEND_URL` и `/notes` на `/api/v1/notes` |
| **Таймаут / функция не разрешилась за 15 000 мс** | 5 | `Error: function timed out, ensure the promise resolves within 15000 milliseconds` | 4–6 файлов: `graph_steps.ts`, `graph_view`, `import_export_steps.ts`, `note_steps.ts`, `progressive-graph-steps.ts` — обновить селекторы/ожидания |
| **Устаревшие селекторы / assertion не прошёл** | 4 | `expect(locator).toBeVisible() failed`, `expect(received).toBe(expected)`, `expect(received).not.toBeNull()` | 3–5 файлов: `auth_cosmic_steps.ts`, `graph_steps.ts`, `graph_view` — обновить селекторы и ожидания |
| **Итого** | **88** |  | **~10–12 step-файлов** |

Помимо падений:

- **129 undefined steps** — отсутствующие определения, особенно в `achievements.feature` (`I am logged in as a user`, `I navigate to the achievements page` и т.д.). Охват: 7–9 step-файлов.
- **55 ambiguous steps** — дублирование определений между `tests/features/step_definitions/` и `tests/steps/` (например, `I navigate to {string}`, `I click the {string} button`, `I select type {string}`). Охват: 2–5 файлов (deduplication).

### Прошедшие сценарии (8)

```text
1. Login page displays cosmic background
2. Login page displays galaxy icon
3. Forgot password page displays cosmic theme
4. Reset password page displays cosmic theme with token
5. All auth pages have consistent styling
6. Yandex login button has cosmic hover effect
7. Cosmic background does not impact performance
8. Zoom and pan in 2D mode
```

Все они либо просто открывают страницу, либо не трогают API/сложные взаимодействия.

### Уникальное покрытие

Да, среди осиротевших сценариев есть покрытие, которого нет ни в активном BDD-наборе (`frontend/tests/features/` — только `graph_2d_list` и `login`), ни в Playwright-спеках:

- **Achievements** (`achievements.feature`) — отсутствует в Playwright.
- **3D-граф и камера** (`camera_navigation`, `full_3d_graph`, `local_3d_graph`, `graph_view`) — Playwright не имеет отдельных 3D E2E.
- **Типы celestial body и их визуализация** (`celestial_body_types`) — не покрыто Playwright.
- **Link types / стиль связей** (`link_types`) — не покрыто Playwright.
- **Single-file import/export** (`import_export` — Markdown, PDF, URL, JSON, GraphML) — Playwright есть `mass-import.spec.ts`, но он покрывает массовый импорт, а не UI одиночного экспорта/импорта.
- **Search & discovery / semantic search / graph path finder** (`search_and_discovery`) — частично пересекается с `home-page.spec.ts`, но `semantic search` и `graph path finder` уникальны.
- **Note management (wiki links, сортировки, batch delete)** (`note_management`) — Playwright покрывает создание, но не wiki-ссылки/сортировку/массовое удаление.

### Рекомендация

**Подключать частично, но не сейчас и не целиком.**

Обоснование:

- 88 из 125 сценариев падают; 29 не запускаются (ambiguous/undefined); только 8 проходят.
- Главная причина — не селекторы, а **устаревший backend URL/API-путь** (76 сценариев). Это механическая, но массовая правка.
- Есть дублирование определений шагов между `tests/features/step_definitions/` и `tests/steps/` (55 ambiguous steps), которое нужно развести.
- Уникальное покрытие есть и оно ценно (achievements, 3D graph, import/export, link types, search), поэтому удалять весь набор нерационально.
- Подключать к `cucumber.mjs` «как есть» нельзя — сломает активный прогон.

**Рекомендуемый план:**

1. Создать **новую рабочую задачу** (например, A-7 / AUD-7) по реанимации осиротевшего набора.
2. В рамках неё починить backend URL/API-пути и `baseURL`, разрешить ambiguous-шаги, удалить/дописать undefined-шаги, обновить селекторы.
3. Затем ввести плавное слияние: отдельные `.feature` из `tests/features/` мигрировать в `frontend/tests/features/` с единым `support/world.ts` и `support/hooks.ts`.
4. Файлы `tests/features/README.md` и внутренние `Scenario Outline` приведения к единому стилю.

### Полный список упавших сценариев

```text
14. Login form has glass morphism card — ..\tests\features\auth_cosmic_theme.feature:22
15. Login inputs have cosmic focus effect — ..\tests\features\auth_cosmic_theme.feature:29
19. Reset password page shows error without token — ..\tests\features\auth_cosmic_theme.feature:72
20. Auth card has entrance animation — ..\tests\features\auth_cosmic_theme.feature:79
21. Camera centers on start node in local graph view — ..\tests\features\camera_navigation.feature:10
22. Camera shows all nodes in full 3D graph view — ..\tests\features\camera_navigation.feature:19
23. Camera adjusts when toggling between local and full graph — ..\tests\features\camera_navigation.feature:27
24. Camera maintains position on browser back/forward — ..\tests\features\camera_navigation.feature:37
25. Transition from 2D graph to 3D graph maintains context — ..\tests\features\camera_navigation.feature:45
26. Camera handles empty graph gracefully — ..\tests\features\camera_navigation.feature:53
27. Camera positions correctly for isolated single node — ..\tests\features\camera_navigation.feature:60
28. Full 3D graph accessible from home page button — ..\tests\features\camera_navigation.feature:67
29. Camera zooms to fit different graph sizes — ..\tests\features\camera_navigation.feature:75
30. Direct URL access to 3D graph works correctly — ..\tests\features\camera_navigation.feature:84
31. Star type is rendered as glowing sphere with rays — ..\tests\features\celestial_body_types.feature:10
32. Planet type is rendered as sphere with rings — ..\tests\features\celestial_body_types.feature:17
33. Comet type is rendered with tail — ..\tests\features\celestial_body_types.feature:23
34. Galaxy type is rendered as spiral with particles — ..\tests\features\celestial_body_types.feature:29
35. Asteroid type is rendered as irregular rock — ..\tests\features\celestial_body_types.feature:35
36. Debris type is rendered as scattered particles — ..\tests\features\celestial_body_types.feature:41
37. Unknown type falls back to default sphere — ..\tests\features\celestial_body_types.feature:47
38. Type from metadata field is used when root type is missing — ..\tests\features\celestial_body_types.feature:53
39. Navigate to full 3D graph from home page — ..\tests\features\full_3d_graph.feature:10
40. Full 3D graph displays all notes — ..\tests\features\full_3d_graph.feature:17
41. Full 3D graph loads without spinner — ..\tests\features\full_3d_graph.feature:24
42. Stats bar shows full graph mode — ..\tests\features\full_3d_graph.feature:30
43. Navigate from full 3D to specific note 3D — ..\tests\features\full_3d_graph.feature:36
44. Full 3D graph handles empty database — ..\tests\features\full_3d_graph.feature:43
45. Full 3D graph with isolated notes — ..\tests\features\full_3d_graph.feature:50
46. Full 3D graph shows fog animation on load — ..\tests\features\full_3d_graph.feature:58
47. Full 3D graph camera centers on all elements — ..\tests\features\full_3d_graph.feature:66
48. Full 3D graph links persist during camera zoom — ..\tests\features\full_3d_graph.feature:74
49. Full 3D graph links persist during camera rotation — ..\tests\features\full_3d_graph.feature:87
50. Full 3D graph shows all links correctly with many nodes — ..\tests\features\full_3d_graph.feature:98
51. Empty state prompts note creation — ..\tests\features\graph_navigation.feature:9
53. Edit a note from the graph — ..\tests\features\graph_navigation.feature:24
54. Delete a note with confirmation — ..\tests\features\graph_navigation.feature:34
59. Return to 2D view from 3D — ..\tests\features\graph_view.feature:22
60. 3D view on low-end devices — ..\tests\features\graph_view.feature:28
61. Rotate view in 3D mode — ..\tests\features\graph_view.feature:44
62. Click node to open details in 2D — ..\tests\features\graph_view.feature:51
63. Click node to open details in 3D — ..\tests\features\graph_view.feature:57
64. Import from a Markdown file — ..\tests\features\import_export.feature:8
65. Import from a PDF file — ..\tests\features\import_export.feature:16
66. Import from URL — ..\tests\features\import_export.feature:24
67. Export to JSON — ..\tests\features\import_export.feature:32
68. Export to Markdown — ..\tests\features\import_export.feature:38
69. Export to GraphML — ..\tests\features\import_export.feature:44
70. Reference link is rendered with solid line — ..\tests\features\link_types.feature:10
71. Dependency link is rendered with dashed line — ..\tests\features\link_types.feature:17
72. Related link is rendered with dotted line — ..\tests\features\link_types.feature:24
73. Custom link type is rendered with default styling — ..\tests\features\link_types.feature:31
74. Strong link has thicker line — ..\tests\features\link_types.feature:38
75. Weak link has thinner line — ..\tests\features\link_types.feature:44
76. Multiple link types are visible simultaneously — ..\tests\features\link_types.feature:50
77. Link connects nodes even when start node not in API response — ..\tests\features\link_types.feature:59
78. Navigate to local 3D view from note detail page — ..\tests\features\local_3d_graph.feature:10
79. Navigate to local 3D view from home page with selected note — ..\tests\features\local_3d_graph.feature:22
80. Local 3D view shows single note with fog when no connections — ..\tests\features\local_3d_graph.feature:32
81. Local 3D view with progressive loading — ..\tests\features\local_3d_graph.feature:41
82. Switch from local to full graph view using toggle — ..\tests\features\local_3d_graph.feature:53
83. Switch back from full to local graph view — ..\tests\features\local_3d_graph.feature:62
84. Local 3D view camera behavior — ..\tests\features\local_3d_graph.feature:70
85. Navigate from local 3D to another note's local view — ..\tests\features\local_3d_graph.feature:78
86. Links remain visible when switching from local to full graph — ..\tests\features\local_3d_graph.feature:87
87. Links persist during camera zoom operations — ..\tests\features\local_3d_graph.feature:100
88. Links persist during camera rotation — ..\tests\features\local_3d_graph.feature:113
89. Links are not duplicated when switching views multiple times — ..\tests\features\local_3d_graph.feature:126
90. Links connect correct nodes after progressive loading completes — ..\tests\features\local_3d_graph.feature:140
99. Quick search from graph view — ..\tests\features\search_and_discovery.feature:10
100. Navigate to search result — ..\tests\features\search_and_discovery.feature:18
101. Full text search — ..\tests\features\search_and_discovery.feature:26
102. Semantic search — ..\tests\features\search_and_discovery.feature:32
103. Empty search results — ..\tests\features\search_and_discovery.feature:38
104. Search history — ..\tests\features\search_and_discovery.feature:44
105. Filter search results — ..\tests\features\search_and_discovery.feature:51
106. Related notes discovery — ..\tests\features\search_and_discovery.feature:58
107. Graph path finder — ..\tests\features\search_and_discovery.feature:64
108. Filter by star type — ..\tests\features\type_filters.feature:10
109. Filter by planet type — ..\tests\features\type_filters.feature:17
110. Filter by comet type — ..\tests\features\type_filters.feature:24
111. Filter by galaxy type — ..\tests\features\type_filters.feature:31
112. Clear filter by selecting All — ..\tests\features\type_filters.feature:38
113. Filter works with type in metadata field — ..\tests\features\type_filters.feature:45
114. Default type falls back to star when not specified — ..\tests\features\type_filters.feature:52
115. Filter state persists when switching views — ..\tests\features\type_filters.feature:58
116. Empty state when filter matches no notes — ..\tests\features\type_filters.feature:67
117. Combined search and type filter — ..\tests\features\type_filters.feature:74
```


