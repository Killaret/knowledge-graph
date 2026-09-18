# GORDON-2. Проанализировать deployment-пакет Gordon

В корне `D:\` найдены 10 документов второго пакета от внешнего AI-ассистента (Gordon) — «Deployment Optimization Package» для Knowledge Graph (~4 700 строк). Файлы собраны в `docs/gordon/` рядом с первым пакетом (см. GORDON-1). Их нужно прочитать, сверить с принятыми решениями и доской, и решить, что встроить в проект.

## Источники

- [`docs/gordon/gordon_index.md`](../gordon/gordon_index.md) — точка входа пакета, карта документов.
- [`docs/gordon/gordon_complete_package.md`](../gordon/gordon_complete_package.md) — свод пакета (89 КБ, 7 файлов).
- [`docs/gordon/gordon_files_summary.md`](../gordon/gordon_files_summary.md) — перечень файлов с назначением.
- [`docs/gordon/gordon_visual_summary.md`](../gordon/gordon_visual_summary.md) — визуальное резюме.
- [`docs/gordon/gordon_deployment_detailed_guide.md`](../gordon/gordon_deployment_detailed_guide.md) — подробный deployment-гайд (736 строк).
- [`docs/gordon/gordon_infrastructure_research.md`](../gordon/gordon_infrastructure_research.md) — инфраструктурное исследование и стратегия (810 строк).
- [`docs/gordon/gordon_infrastructure_discussion_guide.md`](../gordon/gordon_infrastructure_discussion_guide.md) — гайд для обсуждения решений.
- [`docs/gordon/gordon_kubernetes_setup.md`](../gordon/gordon_kubernetes_setup.md) — Kubernetes и мониторинг для production.
- [`docs/gordon/gordon_production_checklist.md`](../gordon/gordon_production_checklist.md) — чек-лист production и нагрузочное тестирование.
- `gordon_training_guide.md` — пошаговый обучающий гайд для команды (удалён решением владельца 2026-09-18).

## Что решить

1. Какие рекомендации уже покрыты доской / `DECISIONS.md` / `DEPLOYMENT_EN.md` / `docker-compose.deploy.yml` / `deploy.yml`.
2. Что противоречит принятым решениям (Docker Compose-деплой, изолированные стеки, бэкап-политика) или требует решения владельца — прежде всего Kubernetes-трек против текущего compose-деплоя.
3. Что полезно встроить в проектную документацию: production-чек-лист, нагрузочное тестирование, мониторинг.
4. Связать с открытыми задачами деплоя: DEPLOY-2 (публикация образов), CI `deploy.yml`.
5. Итоговый вердикт по `docs/gordon/` совместно с вопросом 4 из GORDON-1: оставить как архив внешних анализов или влить выводы и удалить.

## Разбор Devin (2026-09-18)

Каждое утверждение сверено с кодом. Пакет — это дорожная карта «production на Kubernetes» для команды, а не набор багов; полезны точечные оптимизации.

### Принято к реализации

| Предложение | Проверка в репо | Итог |
|---|---|---|
| `start_period: 600s` → 180s у NLP | подтверждено во всех трёх compose (`docker-compose.yml:75`, `test:76`, `personal:67`) | **сделано** — DEPLOY-3: 180s, interval 15s, retries 24 (запас на холодный кэш модели) |
| Запинить `alpine:latest` | `backend/Dockerfile:13` был единственным непиннутым образом | **сделано** — `alpine:3.19`, как у graph-service |
| Параметризовать пул БД | `db.go` — было захардкожено 25/5/5m/1m | **сделано** — секция `database.pool` в `config/backend.json`, env-оверрайды `POSTGRES_MAX_OPEN_CONNS` / `POSTGRES_MAX_IDLE_CONNS` / `POSTGRES_CONN_MAX_LIFETIME_SECONDS` / `POSTGRES_CONN_MAX_IDLE_TIME_SECONDS` (имена из `.env.example`, раньше были задокументированы, но не реализованы); дефолты = прежние значения; `db.ConnectWithPool` + нормализация неположительных значений; применено в `cmd/server` и `cmd/worker` |
| Circuit breaker для NLP client | timeout 10s + exponential backoff уже есть; `gobreaker` — новая зависимость | **отклонено владельцем** — взаимодействие асинхронное (очередь asynq), задачи не теряются при лежачем NLP |
| Метрики пула (`GetPoolStats`) в health/metrics | метрик в проекте нет | **сделано** — `GET /api/v1/metrics/database` за `RequireAdmin()` (DB-POOL-1); периодическое логирование сохранено, интервал настраивается (`stats_interval_seconds`) |

### Решения владельца (2026-09-18)

- **Kubernetes-трек — отложено.** Остаёмся на Docker Compose (`docker-compose.deploy.yml` + `deploy.yml` на Docker Hub). K8s-манифесты и Prometheus-стек оставлены в `docs/gordon/` как справочник на будущее.
- **`gordon_training_guide.md` — удалён**, самостоятельной ценности нет.
- **Production-чек-лист — не нужен отдельным документом:** ранбук деплоя покрыт `DEPLOYMENT_EN.md` + `REGRESSION_TEST_PLAN.md`.

### Неточности пакета

- Модель названа «~300MB» — реально ~4.4 ГБ (torch + safetensors).
- Все «gains» (p95 850→400ms и т.п.) декларированы, не измерены.
- Код из гайдов verbatim не переносить: `contains()` в его `db.go` проверяет только длину строки, не подстроку; readinessProbe ссылается на `/ready`, которого в NLP-сервисе нет.
- `gordon_files_summary.md`, `gordon_index.md`, `gordon_visual_summary.md`, `gordon_complete_package.md` — мета-файлы, дублируют содержимое остальных; оставлены как навигация пакета.

## Статус

Разбор выполнен Devin 2026-09-18; решения владельца записаны выше. Все три открытых вопроса закрыты: пул БД и метрики реализованы (DB-POOL-1), circuit breaker отклонён владельцем (взаимодействие с NLP асинхронное, очередь asynq не теряет задачи).
