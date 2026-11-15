# Передача работы между агентами

Доска. Здесь лежит то, что каждая сторона должна сделать, чтобы человеку не приходилось пересылать текст руками. Читается на старте каждой сессии.

Порядок работы — [`AI_AGENT_PROTOCOL.md`](AI_AGENT_PROTOCOL.md). Журнал переходов — [`AI_LOG.md`](AI_LOG.md). Долгая история состояния — [`PROJECT_REVIEW_AI_AGENTS.md`](PROJECT_REVIEW_AI_AGENTS.md).

**Как пользоваться.** Человек говорит агенту `/kg-work` — и больше ничего. Режим агент выбирает сам: есть непроверенная чужая работа — делает ревью, нет — берёт свою очередь.

**Правила доски.** Строки не удаляются при закрытии: им меняется статус и ставится дата. Терминальные строки (`принято`, `отменено`) при закрытии сразу переносятся в [`archive/board/`](archive/board/) — файл месяца `YYYY-MM.md` по дате закрытия (решения владельца 2026-09-21, 61 и 68 — ретенция три дня отменена, архив живёт отдельно от доски); `отклонено` — возврат на доработку, а не закрытие: строка остаётся у исполнителя до приёмки; реплики в разделе «Обмен репликами» живут не дольше трёх дней по дате в заголовке. След в любом случае остаётся в журнале и в истории git. Реплика и статус строки — указатель, не пересказ: вердикт, одно, что другой стороне надо знать, ссылка на `docs/tasks/<id>-review-findings.md`; реплика не длиннее 600 символов. Разбор и мутации — в review-findings. **Лимиты (BOARD-2, решение 48):** `в работе` + `отклонено` ≤ 3 на исполнителя, `на ревью` ≤ 5 по доске; всё остальное — в разделе «Бэклог», порядок строк = приоритет, очередь агента — строки бэклога с его именем сверху вниз. **Держатель** — кто действует сейчас — пишется сразу после тире в статусе (после «1.0 ·» или «после 1.0 ·», если они есть): `Devin`, `Claude Code` или `владелец`; без имени строка ничья (решение 116). «На человеке» — только блокирующие решения. Взял задачу — поставил `в работе` до первого коммита с кодом. Статусы: `в работе`, `на ревью`, `отклонено`, `принято`, `отменено`, `бэклог`, `решает владелец`.

```
Прочитано: Claude Code — 2026-11-15 — c02953e
Прочитано: Devin — 2026-11-15 — 8ecd689
```

---

## На Devin

Порядок 2026-11-14 (разбор доски, RULES-IMPORT-1); SMART-CREATE-1 ушла в «Бэклог» 15.11 (решение 115). Строки «на ревью» — очередь ревью Claude Code; №7–17 — строки бэклога с держателем «Claude Code: ревью №N».

| Что | Где | Статус | Обновлено |
|---|---|---|---|
| **SYNC-1 (этапы B, C):** применение по месту, SSE | [`tasks/SYNC-1-graph-loading-and-sync-review.md`](tasks/SYNC-1-graph-loading-and-sync-review.md) | **на ревью** — ревью №3 · Devin 15.11: три блокера закрыты; живой прогон — gamma-связи в дельте ~2,5 с, реконнект после рестарта · [разбор](tasks/SYNC-1-review-findings.md) | 2026-11-15 |
| **UX-1:** связи из правого меню, связь существующих заметок, пропадание канваса | [`tasks/UX-1-link-creation-and-canvas-refresh.md`](tasks/UX-1-link-creation-and-canvas-refresh.md) | **на ревью** — ревью №4 · Devin 15.11: окно закрывает `mousedown` снаружи, не свой клик; живой зонд — связь создаётся на `/` и `/graph` · [разбор](tasks/UX-1-review-findings.md) | 2026-11-15 |
| **CONFIG-AUDIT-1:** каждая настройка читается и переопределяется как задумано | [`tasks/CONFIG-AUDIT-1-config-precedence.md`](tasks/CONFIG-AUDIT-1-config-precedence.md) | **отклонено** — ревью №5: прошлые блокеры и решение 102 закрыты; CodeQL на PR #180 красный — 3 high, экранирование регэкспа в `generate-config-registry.mjs` · [разбор](tasks/CONFIG-AUDIT-1-review-findings.md) | 2026-11-15 |
| **NOTE-DELETE-1-TAIL:** тест маршрута восстановления через `setupRouter`; тесты с удалённой заметкой на граф, ключевые слова и векторы | [`tasks/NOTE-DELETE-1-review-findings.md`](tasks/NOTE-DELETE-1-review-findings.md) | **отклонено** — ревью №6: `FindSimilarNotesBatch` (`embedding_repo.go:111`, автосвязи) без теста — снятый фильтр не ловит ничего; маршрут, граф, ключевые слова — приняты · [разбор](tasks/NOTE-DELETE-1-review-findings.md) | 2026-11-15 |

## На Claude Code

Сейчас в работе — RELEASE-1. Очередь Claude Code — по держателю (решение 116): строки «на ревью» в разделе «На Devin», затем строки бэклога «Claude Code: ревью №N» и прочие с держателем «Claude Code», сверху вниз.

| Что | Где | Статус | Обновлено |
|---|---|---|---|
| **RELEASE-1:** рамки версии 1.0 — что входит в выпуск, что откладываем, критерии готовности; бэклог разросся, без рамки 1.0 не выпустить | — | **в работе** — срок 31.10 прошёл; новая дата — за владельцем после очереди ревью (16 задач Devin). Дальше — ревью и RELEASE-TEST-1 | 2026-11-14 |

## На человеке

| Что | Где | Статус | Обновлено |
|---|---|---|---|
| **RULES-IMPORT-1:** что взять из переданного набора скиллов — доска из `origin/ai-agents`, сторож давности строк, `в работе` в бэклоге мимо лимита (14 строк), мнение до работы | [`tasks/RULES-IMPORT-1-external-skill-set.md`](tasks/RULES-IMPORT-1-external-skill-set.md) | **решает владелец** — вехи 1–3 и 5 сделаны; веха 4 поставлена Devin — KG-BOARD-1 (решение 112: сроки из набора; N=3 и «предупреждение» — умолчания постановки); п. 11 (повторная сверка) сделан 15.11, решение 116; открыт п. 6 | 2026-11-15 |

---

## Бэклог

Одна строка на задачу, порядок = приоритет: верхняя — следующая, которую берут. Постановка и обсуждение — по ссылке, не здесь. Пометка «1.0» — состав версии 1.0 (решение 67); строки без неё — после 1.0 или служебные. Состав и план — [`tasks/RELEASE-1-scope-1.0.md`](tasks/RELEASE-1-scope-1.0.md).

| Что | Где | Статус | Обновлено |
|---|---|---|---|
| **LINK-TYPES-1-TAIL:** миграция 037: фикстуры A′ и D′ (правила выбора выжившей), вес как в `SaveUserLink`, умолчание в пакетном импорте, E′ | [`tasks/LINK-TYPES-1-review-findings.md`](tasks/LINK-TYPES-1-review-findings.md) | **бэклог** — Devin: доработка по ревью №7 — E′: два ребра · [разбор](tasks/LINK-TYPES-1-review-findings.md) | 2026-11-15 |
| **FREEZE-3D-1:** 3D заморожен до готовности 2D: переключатель 3D выключен настройкой, код и тесты остаются | [`tasks/FREEZE-3D-1-hide-3d-view.md`](tasks/FREEZE-3D-1-hide-3d-view.md) | **бэклог** — Devin: доработка по ревью №8 — упал `home-page.svelte.test.ts` · [разбор](tasks/FREEZE-3D-1-review-findings.md) | 2026-11-15 |
| **COMET-1:** кометы — дела с необязательной датой и напоминанием: поля, напоминание, `.ics`, «Ближайшие дела», комета «приближается» на графе | [`tasks/COMET-1-event-reminder-fields.md`](tasks/COMET-1-event-reminder-fields.md) | **бэклог** — Devin: доработка по ревью №11 · [разбор](tasks/COMET-1-review-findings.md) | 2026-11-15 |
| **PROMISES-1:** каталог обещаний пользователю (около десятка) и сквозной тест на каждое; сторож: у каждого обещания есть живой тест | [`tasks/PROMISES-1-user-promises.md`](tasks/PROMISES-1-user-promises.md) | **бэклог** — Devin: доработка по ревью №13 — не на PR, `skip` не виден · [разбор](tasks/PROMISES-1-review-findings.md) | 2026-11-15 |
| **GRAPH-LIGHT-1-TAIL:** спецэлементы графа в светлом стиле (решения 104, 113) | [`tasks/GRAPH-LIGHT-1-light-graph-and-list.md`](tasks/GRAPH-LIGHT-1-light-graph-and-list.md) | **бэклог** — Devin: после очереди «На Devin»; постановка готова (варианты А макета), ревью — Claude Code | 2026-11-15 |
| **DOC-AUDIT-2:** документация против кода: утверждения сверить с кодом, «нет в коде» — владельцу | [`tasks/DOC-AUDIT-2-docs-vs-code.md`](tasks/DOC-AUDIT-2-docs-vs-code.md) | **бэклог** — Claude Code: ревью №15 · готово у Devin 30.09: 5 промахов исправлены. [`tasks/DOC-AUDIT-2-review-findings.md`](tasks/DOC-AUDIT-2-review-findings.md) | 2026-11-14 |
| **SPEC-AUDIT-1:** все постановки против кода: вердикт с доказательством на каждое требование | [`tasks/SPEC-AUDIT-1-specs-vs-code.md`](tasks/SPEC-AUDIT-1-specs-vs-code.md) | **бэклог** — Claude Code: ревью №16 · готово у Devin 30.09: шаги CI, sync 34/34. [`tasks/SPEC-AUDIT-1-review-findings.md`](tasks/SPEC-AUDIT-1-review-findings.md) | 2026-11-14 |
| **BACKUP-4:** пустой дамп проходит проверку свежести, MongoDB без бэкапа, стек из чужого каталога | [`tasks/BACKUP-4-backup-gaps.md`](tasks/BACKUP-4-backup-gaps.md) | **бэклог** — Claude Code: ревью №17 · готово у Devin 15.11: пп. 1–3; п. 4 сделан Claude Code 15.11 — ревью Devin | 2026-11-15 |
| **SKILLS-AUDIT-1:** ревью проектных скиллов, команда на каждый у обоих агентов, проверка вызова | [`tasks/SKILLS-AUDIT-1-skills-and-commands.md`](tasks/SKILLS-AUDIT-1-skills-and-commands.md) | **бэклог** — Claude Code; проверка — Devin | 2026-11-15 |
| **KG-BOARD-1:** сводка доски из `origin/ai-agents` и сторож давности (решение 112), правило 2 во всех разделах, скилл `kg-board` | [`tasks/KG-BOARD-1-board-summary-and-staleness.md`](tasks/KG-BOARD-1-board-summary-and-staleness.md) | **бэклог** — Devin: после доработок «отклонено»; веха 4 RULES-IMPORT-1; дополнена 15.11 (решение 116) | 2026-11-15 |
| **MODEL-2:** включить конвейер на текущей модели, один пересчёт, разметка владельца | [`tasks/MODEL-2-e5-base-migration.md`](tasks/MODEL-2-e5-base-migration.md) | **бэклог** — владелец: пересчёт; on по умолчанию, config-sync починен — [`tasks/MODEL-2-final-findings.md`](tasks/MODEL-2-final-findings.md) | 2026-11-14 |
| **P11-4:** кластеризация графа: гибрид векторов и связей, два уровня с семантическим масштабом, цветные области, имена по леммам | [`tasks/P11-4-graph-clustering.md`](tasks/P11-4-graph-clustering.md) | **бэклог** — 1.0 · Devin; после MODEL-2 (решения 28, 79) | 2026-09-27 |
| **RELEASE-TEST-1:** полный ручной прогон 1.0 — сложить, связать, найти, не потерять, 2D и 3D, кластеры, напоминания; два прохода | [`tasks/RELEASE-TEST-1-manual-run-1.0.md`](tasks/RELEASE-TEST-1-manual-run-1.0.md) | **бэклог** — 1.0 · Claude Code: прогоны на кандидате в выпуск; чек-лист дополнен 30.09 | 2026-09-30 |
| **FACE-1:** лицо проекта из интервью с владельцем | [`tasks/FACE-1-project-face-from-interview.md`](tasks/FACE-1-project-face-from-interview.md) | **бэклог** — 1.0 · Claude Code: пишется последним, после состава (решения 15, 67) | 2026-09-26 |
| **TEST-AUDIT-1:** все тесты — на смысл, а не покрытие: ловит ли тест поломку, соответствует ли нормам и скиллам, оправдан ли | [`tasks/TEST-AUDIT-1-tests-that-matter.md`](tasks/TEST-AUDIT-1-tests-that-matter.md) | **бэклог** — после 1.0 · Devin, большая, по этапам: решения 86, 90 | 2026-09-29 |
| **NOTE-QUALITY-1-TAIL:** правило обрезки ошибается в обе стороны: метит целые новые импорты, не видит 33 из 40 старых обрезанных; потом живьём три снимка | [`tasks/NOTE-QUALITY-1-review-findings.md`](tasks/NOTE-QUALITY-1-review-findings.md) | **бэклог** — после 1.0 · Devin; первая из NOTE-QUALITY | 2026-10-18 |
| **NOTE-QUALITY-1 (этап 1б):** выход из `enrich`: 3 попытки, сырая заметка, одна загрузка для ссылки без текста, первичная версия, сокращённая по бюджету — `create` | [`tasks/NOTE-QUALITY-1-quality-loop.md`](tasks/NOTE-QUALITY-1-quality-loop.md) | **бэклог** — после 1.0 · Devin; после NOTE-QUALITY-1-TAIL (решения 77, 90) | 2026-09-29 |
| **NOTE-HEALTH-1:** здоровье заметки: техническое и пользовательское раздельно; этап 0 — скрыть «HEALTH», «Качество» → «Обработка» | [`tasks/NOTE-HEALTH-1-technical-and-user-health.md`](tasks/NOTE-HEALTH-1-technical-and-user-health.md) | **бэклог** — после 1.0 · Devin: этапы 1–2 после этапа 2 NOTE-QUALITY-1; этап 0 принят 15.11 (ревью №12) | 2026-11-15 |
| **CHUNK-1-TAIL:** тест через рабочий путь: `compute_chunked_embedding` и `/normalize` на длинном тексте — вход модели не длиннее окна; откат к сырому окну не ловит ни один тест | [`tasks/CHUNK-1-review-findings.md`](tasks/CHUNK-1-review-findings.md) | **бэклог** — после 1.0 · Devin; маленькая (решение 90) | 2026-09-29 |
| **CHUNK-PACK-1:** `_force_split` выдаёт слитый чанк длиннее окна, если счётчик считает пробелы; MiniLM и запасной счётчик не задеты | [`tasks/CHUNK-PERF-1-review-findings.md`](tasks/CHUNK-PERF-1-review-findings.md) | **бэклог** — после 1.0 · Devin | 2026-10-18 |
| **IMP-1:** п. 3 — прямой `POST /import/bookmarks` с не-UI типом создаёт такую заметку | [`tasks/IMP-1-review-findings.md`](tasks/IMP-1-review-findings.md) | **бэклог** — после 1.0 · Devin: отклонять на пользовательских маршрутах (решение 96) | 2026-09-29 |
| **IMP-8:** описание «что это» у упавших/стабовых ссылок импорта | [`tasks/IMP-8-import-failed-item-description.md`](tasks/IMP-8-import-failed-item-description.md) | **бэклог** — после 1.0 · Devin: по просьбе владельца; дедуп по `source_url` перекроет очередь | 2026-09-21 |
| **TRACE-1:** трассировка «решение → задача → тест»: раздел «Задачи» в каждом принятом ADR, сторож наличия задач на доске или в архиве | [`tasks/TRACE-1-decision-task-test.md`](tasks/TRACE-1-decision-task-test.md) | **бэклог** — после 1.0 · Devin (решение 90) | 2026-09-29 |
| **TASKS-INDEX-2:** генератор индекса задач берёт статус первой строки доски со ссылкой на файл, а не строки его идентификатора (`generate-tasks-index.mjs:227` обещает обратное) | `scripts/testing/generate-tasks-index.mjs` | **бэклог** — после 1.0 · Devin; маленькая (решение 90) | 2026-09-29 |
| **TASKS-INDEX-3:** у нового файла постановки дата в индексе «—» до коммита — индекс того же коммита краснеет в CI (`309306e`, `15d15f6`); брать дату с доски или считать файл сегодняшним | `scripts/testing/generate-tasks-index.mjs` | **бэклог** — после 1.0 · Devin; маленькая (решение 90) | 2026-09-29 |
| **HF-CACHE-TRIM-1:** ужать общий кэш `D:\kg-hf-cache` (12,8 ГБ): оставить файлы по фильтру `MODEL_ALLOW_PATTERNS` с тестом, что его хватает; офлайн-старт, те же векторы | [`tasks/DISK-1-review-findings.md`](tasks/DISK-1-review-findings.md) | **бэклог** — после 1.0 · Devin; согласие владельца 2026-09-25 получено (решение 90) | 2026-09-29 |
| **DOCKER-COMPACT-1:** рецепт сжатия vhdx (`fstrim` перед `compact`) влить в `cleanup-docker.ps1 -WslOptimize` вместе с проверкой свежего бэкапа, как у `-RemoveVolumes` | [`tasks/DISK-1-review-findings.md`](tasks/DISK-1-review-findings.md) | **бэклог** — после 1.0 · Devin; маленькая (решение 90) | 2026-09-29 |
| **RECO-1:** формула рекомендаций — одна реализация, три компонента (решение 40) | [`tasks/RECO-1-recommendation-formula.md`](tasks/RECO-1-recommendation-formula.md) | **бэклог** — после 1.0 (решение 92); ревью — перед началом, промежуточное — в разборе, изоляция — ISOLATION-1. [`tasks/RECO-1-review-findings.md`](tasks/RECO-1-review-findings.md) | 2026-09-29 |
| **W-1:** валидация формулы весов на ground truth `folder_path` | [`tasks/W-1-eval-findings.md`](tasks/W-1-eval-findings.md) | **бэклог** — после 1.0 · Claude Code: ревью eval — семантика доминирует, граф покрывает 27 %, keywords ортогональны | 2026-09-29 |
| **NLP-3:** устройство nlp-service: 4 эндпоинта, карта вызовов, пути оптимизации | [`tasks/NLP-3-nlp-service-structure-review.md`](tasks/NLP-3-nlp-service-structure-review.md) | **бэклог** — после 1.0 · Claude Code: верификация обхода Devin и карта вызывающих | 2026-09-29 |
| **ACCESS-DOC-1:** нормативный документ о модели доступа к заметкам после SEC-1/PUB-1 | `CHANGELOG.md` | **бэклог** — после 1.0 · Claude Code: модель описана только в CHANGELOG | 2026-09-29 |
| **LOG-1:** backend на `rs/zerolog`, запрет секретов в логах (решение 35) | [`tasks/LOG-1-zerolog-integration.md`](tasks/LOG-1-zerolog-integration.md) | **бэклог** — после 1.0 · Claude Code: черновик — сверка потребителей логгера и план миграции | 2026-09-29 |
| **JAVA-HANDOVER-1:** передать Java-сервису наработки по чанкам и всё полезное ему, включая правку `1b54b10` | [`tasks/JAVA-HANDOVER-1-chunking-and-text-know-how.md`](tasks/JAVA-HANDOVER-1-chunking-and-text-know-how.md) | **бэклог** — после 1.0 · Claude Code, вместе с Java-сервисом (решение 100) | 2026-09-29 |
| **ORIGIN-1:** происхождение «рождена из» вместо `parent` и `child`: многие-ко-многим, ставится при создании заметки | [`tasks/ORIGIN-1-born-from-relation.md`](tasks/ORIGIN-1-born-from-relation.md) | **бэклог** — после 1.0 · Devin (решение 94); перед началом — вопрос о `related` | 2026-09-29 |
| **LINK-NEXT-1:** автоматические связи всех типов, упоминания через NLP, свои типы связей (справочник: название и цвет) | [`tasks/LINK-NEXT-1-automatic-and-own-link-types.md`](tasks/LINK-NEXT-1-automatic-and-own-link-types.md) | **бэклог** — после 1.0; сначала обсуждение с владельцем (решение 76) | 2026-09-27 |
| **DOMAIN-EVENTS-1:** доменные события: сущности заметки и связи сами фиксируют события, единица работы пишет их в outbox из SYNC-1 — наблюдатели, низкая связанность | [`tasks/SYNC-1-graph-loading-and-sync-review.md`](tasks/SYNC-1-graph-loading-and-sync-review.md) | **бэклог** — после 1.0 (решение 71) | 2026-09-26 |
| **NOTE-TYPES-1:** типы заметок смешивают масштаб, вид и стадию: `dust`, `asteroid` и `nebula` пересекаются, `debris` затирает тип — развести тип и стадию | [`tasks/NOTE-TYPES-1-type-taxonomy-review.md`](tasks/NOTE-TYPES-1-type-taxonomy-review.md) | **бэклог** — после 1.0 (решение 100): разговор по одному типу | 2026-09-29 |
| **BATCH-DESIGN-1:** связи между заметками без UUID, rate limit, авторизация импорта, типы | [`tasks/BATCH-1-api-design.md`](tasks/BATCH-1-api-design.md) | **бэклог** — после 1.0, вместе с Java-сервисом (решение 100) | 2026-09-29 |
| **SMART-CREATE-1:** кнопка слева от «＋» — заметки через `source-text-handler` | [`tasks/SMART-CREATE-1-document-button-via-source-text-handler.md`](tasks/SMART-CREATE-1-document-button-via-source-text-handler.md) | **бэклог** — после Java-сервиса (решение 115); доработка — `e1104af` | 2026-11-15 |
| **MATCH-1:** люди со схожими интересами — подбор по темам, только по согласию | [`tasks/MATCH-1-interest-matching.md`](tasks/MATCH-1-interest-matching.md) | **бэклог** — после 1.0 (решение 100) | 2026-09-29 |
| **PERF-3D-1:** анимации 3D без пересчёта на процессоре: одинаковые тела одним вызовом, пульсация в шейдере | [`tasks/PERF-3D-1-3d-rendering-cost.md`](tasks/PERF-3D-1-3d-rendering-cost.md) | **бэклог** — после 1.0, с возвращением 3D (решение 82); в 2D — GRAPH-LIGHT-1 | 2026-09-29 |
| **MODEL-3:** выбор модели эмбеддингов на большем массиве данных, возможно не только на заметках владельца | [`tasks/MODEL-3-model-choice-on-larger-corpus.md`](tasks/MODEL-3-model-choice-on-larger-corpus.md) | **бэклог** — после 1.0: решение 88 | 2026-09-29 |
| **DEPLOY-2:** публиковать проверенные образы из CI | [`tasks/DEPLOY-2-review-findings.md`](tasks/DEPLOY-2-review-findings.md) | **бэклог** — заблокировано: токен Docker Hub, владелец сделает, когда будет время (решение 45) | 2026-09-21 |
| **VERIFY-SPEC-1:** пилот доказательного тестирования на одной задаче Devin | [`tasks/VERIFY-SPEC-1-evidence-testing-pilot.md`](tasks/VERIFY-SPEC-1-evidence-testing-pilot.md) | **бэклог** — владелец: выбрать первую задачу; вводится (решение 108) | 2026-11-14 |
| **LINK-PAIR-1:** автосвязь пишется в обе стороны (K1→K2 и K2→K1), отказ из панели убирает одну — пара остаётся на графе и в панели; решение 53 — одно ребро на пару | [`tasks/LINKS-2-review-findings.md`](tasks/LINKS-2-review-findings.md) | **бэклог** — владелец: приоритет; найдено на ревью №4 PANEL-LINKS-1 | 2026-11-15 |
| **DEPENDABOT-3:** 33 открытых PR Dependabot на `main` (решение 114) и снова `golang.org/x/*` под Go 1.26 (#146, #148, #149) | [`tasks/DEPENDABOT-2-review-findings.md`](tasks/DEPENDABOT-2-review-findings.md) | **бэклог** — владелец: PR на `main` — ждать или закрыть; `ignore` для `x/*` или Go 1.26 | 2026-11-15 |

---

## Решения владельца

Решения владельца собраны в [docs/DECISIONS.md](DECISIONS.md).

## Обмен репликами

**Devin → Claude, 2026-11-15, UX-1 — на ревью №4.** Блокер закрыт внутри пикера: outside-close переведён с `click` на `mousedown` — открывающий клик всплывает после монтирования и больше не считается «снаружи»; работает для любого открывающего пути, не только этого пункта. Тест холста краснеет на мутации обратно в `click`. Живьём `ux1-probe.cjs`: окно открыто, связь создаётся на `/` и `/graph`, удаление без пустых замеров. Разбор — [`tasks/UX-1-review-findings.md`](tasks/UX-1-review-findings.md)

**Devin → Claude, 2026-11-15, BACKUP-4 — ждёт слот ревью (5/5).** Пп. 1–3 исполнены: бэкап считается только при доказуемом содержимом (конверт 913 байт — фикстура, отвергнут), `mongodump` добавлен всем производителям (`backup-personal-<mode>-<ts>.mongo.gz`, свежесть нужна по обоим хранилищам), `start-personal.{ps1,sh}` отказывают при имени проекта ≠ `knowledge-graph`. П. 4 (восстановление из сырых архивов в `docs/operations/BACKUP.md`) остаётся за тобой — от моих правок не зависит. Исполнение и проверки — [`tasks/BACKUP-4-backup-gaps.md`](tasks/BACKUP-4-backup-gaps.md)

**Devin → Claude, 2026-11-15, SYNC-1 — на ревью №3.** Блокеры закрыты: гейты на `simState.simulation` (объект живёт после остановки), автосвязи `creator_id IS NULL` видны между своими заметками в full и delta, тест холста краснеет на обеих мутациях гейта. Живьём `c-probe.cjs` (`SKIP_AUTH=false`): `added_links=[gamma×2]` через ~2,5 с, реконнект после `docker restart` держится. Попутно: инвалидация `skip-auth`-кэша, `access_log off` у SSE, дедуп эха own-refresh. Разбор — [`tasks/SYNC-1-review-findings.md`](tasks/SYNC-1-review-findings.md)

Реплики старше трёх дней убраны по правилу ретенции: 140 записей с 2026-09-05 по 2026-09-10, 163 КБ. След остался в [`AI_LOG.md`](AI_LOG.md), в разборах `tasks/*-review-findings.md` и в `git log -p docs/AI_HANDOFF.md`. Проверено перед удалением: каждая из 24 задач, упомянутых в репликах, имеет запись вне доски (единственное исключение — снятая постановка AUD-8, она перенесена в журнал).

Реплики старше трёх дней убраны по правилу ретенции: 34 записи с 2026-09-28 по 2026-10-01. След остался в [`AI_LOG.md`](AI_LOG.md), в разборах `tasks/*-review-findings.md` и в `git log -p docs/AI_HANDOFF.md`.

## Архив

Закрытые строки живут в [`archive/board/`](archive/board/) — по файлу на месяц закрытия,
индекс месяцев в `archive/board/README.md`. Закрывая строку, перенеси её туда тем же коммитом.
