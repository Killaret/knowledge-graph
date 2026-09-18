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
- [`docs/gordon/gordon_training_guide.md`](../gordon/gordon_training_guide.md) — пошаговый обучающий гайд для команды.

## Что решить

1. Какие рекомендации уже покрыты доской / `DECISIONS.md` / `DEPLOYMENT_EN.md` / `docker-compose.deploy.yml` / `deploy.yml`.
2. Что противоречит принятым решениям (Docker Compose-деплой, изолированные стеки, бэкап-политика) или требует решения владельца — прежде всего Kubernetes-трек против текущего compose-деплоя.
3. Что полезно встроить в проектную документацию: production-чек-лист, нагрузочное тестирование, мониторинг.
4. Связать с открытыми задачами деплоя: DEPLOY-2 (публикация образов), CI `deploy.yml`.
5. Итоговый вердикт по `docs/gordon/` совместно с вопросом 4 из GORDON-1: оставить как архив внешних анализов или влить выводы и удалить.

## Статус

Ждёт человека — обзор и вердикт. Собрано Devin 2026-09-18 с `D:\gordon_*.md` в `docs/gordon/`.
