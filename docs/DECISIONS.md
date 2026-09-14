# Решения владельца

Указатель. Одна строка на решение: что решали, что решили, почему, где лежит разбор.

**Обоснование здесь не дублируется.** Оно живёт там, где родилось — в файле постановки
или в ADR, — а эта страница только ведёт. Копия разошлась бы с оригиналом, и через месяц
было бы неясно, какая версия настоящая.

Что где искать:

| Слой | На какой вопрос отвечает |
|---|---|
| эта страница | что решили и когда |
| файл постановки | что решали, какие были варианты, чего каждый стоил |
| `git log` по файлу постановки | как формулировка менялась |
| коммиты с идентификатором задачи | чем решение обернулось в коде |
| `docs/tasks/*-review-findings.md` | чем проверяли и что нашли |

Архитектурные решения — отдельный жанр и отдельное место:
[`architecture/decisions/`](architecture/decisions/), 18 ADR с отвергнутыми вариантами.
Здесь — продуктовые и политические.

---

| № | Дата | Решение | Почему | Разбор |
|---|---|---|---|---|
| 1 | 2026-09-06 | `SKIP_AUTH` остаётся, строка пользователя в базе сохраняется | тестовый контур зависит от неё через внешний ключ `notes.creator_id` | [`tasks/AUD-2-seeder-issue.md`](tasks/AUD-2-seeder-issue.md) |
| 2 | 2026-09-07 | Проверки прогоняются **локально** перед приёмкой, а не наблюдением за CI | о красном узнаём прогоном; наблюдение за GitHub — не проверка | [`tasks/CI-3-local-check-runner.md`](tasks/CI-3-local-check-runner.md) |
| 3 | 2026-09-07 | TLS-терминация для nginx — **отложено** | вне периметра AUD-5, не блокирует 1.0 | [`BACKLOG.md`](BACKLOG.md), TD-TLS |
| 4 | 2026-09-07 | AUD-7b: исправить замечания, включить линтинг тестов, расширить знаменатель покрытия до `src/**`, порог 70 % | знаменатель без `src/**` льстит цифре | [`tasks/AUD-7b-lint-tests-and-coverage-denominator.md`](tasks/AUD-7b-lint-tests-and-coverage-denominator.md) |
| 5 | 2026-09-07 | Скрипт очистки Docker остаётся **машинным**, область не сужаем | чистить половину Docker бессмысленно | [`tasks/CLEAN-1-docker-cleanup-honesty.md`](tasks/CLEAN-1-docker-cleanup-honesty.md) |
| 6 | 2026-09-07 | Визуальные эталоны — на `SKIP_AUTH`-контуре | иначе набор написан под сессию и не воспроизводится | [`tasks/VIS-1-split-visual-baselines.md`](tasks/VIS-1-split-visual-baselines.md) |
| 7 | 2026-09-07 | Публичный граф — **витрина сообщества**: аноним читает опубликованное и ищет по нему | витрина, за которой ничего нельзя открыть, — не витрина | [`tasks/PUB-1-anonymous-read-and-search.md`](tasks/PUB-1-anonymous-read-and-search.md), ADR-018 |
| 8 | 2026-09-07 | Выбор графа — **режим просмотра**, авторизация задаёт лишь значение по умолчанию | иначе авторизованный не может вернуться в граф сообщества | [`tasks/PUB-2-graph-view-mode.md`](tasks/PUB-2-graph-view-mode.md) |
| 9 | 2026-09-07 | `/api/v1/graph/all` переименовывается **жёстко, без алиаса** | внешних потребителей нет; `deprecated`, который никто не удалит, — тот же дрейф | [`tasks/PUB-3-rename-graph-endpoints.md`](tasks/PUB-3-rename-graph-endpoints.md) |
| 10 | 2026-09-07 | Авторство в публичном графе **сохраняется** | граф сообщества без авторства — анонимная свалка | ADR-018 |
| 11 | 2026-09-08 | CSP: разрешить `style-src-attr 'unsafe-inline'`, `frame-ancestors 'self'` | вся цена политики — в 76 инлайновых стилях; переписывать их сейчас не окупается | [`tasks/CSP-1-content-security-policy.md`](tasks/CSP-1-content-security-policy.md) |
| 12 | 2026-09-08 | AUTO-1 (автопуш и запуск Devin по `push`) — **отменено** | передача остаётся ручной; петля без владельца не нужна | [`tasks/AUTO-1-push-and-trigger.md`](tasks/AUTO-1-push-and-trigger.md) |
| 13 | 2026-09-08 | Спецификация API опускается до **OpenAPI 3.0.3** | вшитый Swagger UI не читает 3.1; восемь строк против нового артефакта в образе | [`tasks/API-1-openapi-contract-and-handover.md`](tasks/API-1-openapi-contract-and-handover.md) |
| 14 | 2026-09-08 | Тег `latest` на Docker Hub **не заводим** | «последний» без даты не говорит, что внутри | [`tasks/DEPLOY-review-findings.md`](tasks/DEPLOY-review-findings.md) |
| 15 | 2026-09-08 | Лицо проекта (FACE-1) **отложено до 1.0** | половина текста описывает состояние, которое 1.0 изменит; вносить сейчас — вносить дважды | [`tasks/FACE-1-project-face-from-interview.md`](tasks/FACE-1-project-face-from-interview.md) |
| 16 | 2026-09-08 | **Объём ограничивается выпуском 1.0** — граница фич и дата | «постоянно тащу новые фичи, не успеваю дотестировать» | [`tasks/FACE-1-project-face-from-interview.md`](tasks/FACE-1-project-face-from-interview.md) |
| 17 | 2026-09-12 | `yake` заменить на `keybert` (MIT) с лемматизацией рус/англ — в будущем | лицензия `yake` не подходит; PR #25 закрыт | [`tasks/YAKE-REPLACE-KEYBERT-LEMMATIZATION.md`](tasks/YAKE-REPLACE-KEYBERT-LEMMATIZATION.md) |
| 18 | 2026-09-14 | Устаревшие образы **не блокируют** передачу Java-разработчику | у него есть доступ к репозиторию, он собирает из исходников | [`tasks/DEPLOY-review-findings.md`](tasks/DEPLOY-review-findings.md) |
| 19 | 2026-09-14 | Публикация заметки остаётся **отдельным действием**, мёртвое поле `is_public` в редактировании убрать | за публикацией стоит инвалидация кэша публичного графа; два способа, из которых работает один, — ложь в контракте | [`tasks/PUB-1-review-findings.md`](tasks/PUB-1-review-findings.md) |
| 20 | 2026-09-14 | Массовой публикации через граф **не заводим** | галочка в заметке закрывает потребность | [`tasks/PUB-1-review-findings.md`](tasks/PUB-1-review-findings.md) |
| 21 | 2026-09-14 | Авторство коммитов — **вариант Б**: чужая работа коммитится подписью настоящего автора | трейлер `Co-Authored-By` говорит «участвовал», а не «написал»; за месяц правило нарушено 133 раза | [`tasks/AUTHOR-1-commit-authorship-guard.md`](tasks/AUTHOR-1-commit-authorship-guard.md) |
| 22 | 2026-09-14 | Решения владельца получают **указатель** (эта страница) и сторожа; переписка остаётся в git, решения — в проекте | ценно только то, что можно найти и показать | [`tasks/DECISIONS-1-decision-index-and-guard.md`](tasks/DECISIONS-1-decision-index-and-guard.md) |

---

## Ждут решения

| Что | Где |
|---|---|
| Место MongoDB: убрать, заменить или дождаться нагрузки, под которую она подходит | [`tasks/MONGO-1-drafts-storage-discussion.md`](tasks/MONGO-1-drafts-storage-discussion.md) |
| Как Java-сервис создаёт заметки и связывает их, не зная UUID | [`tasks/BATCH-API-DESIGN.md`](tasks/BATCH-API-DESIGN.md) |
| DEPENDABOT-1: судьба оставшихся Dependabot-PR | [`PROJECT_REVIEW_AI_AGENTS.md`](PROJECT_REVIEW_AI_AGENTS.md), §20 |
| Состав и дата 1.0 | — |
