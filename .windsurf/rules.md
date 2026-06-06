# Windsurf Agent Context

Этот файл содержит единый контекст для Windsurf AI и полный перечень агентов.
Не удаляйте.

## Список агентов (9 штук)
1. knowledge-graph-orchestrator — координация и делегирование
2. knowledge-graph-backend-go — Go backend, API, БД, auth
3. knowledge-graph-frontend-svelte — Svelte UI, компоненты и state
4. knowledge-graph-integration — OpenAPI, контракты, webhooks
5. knowledge-graph-infrastructure — Docker, K8s, мониторинг
6. knowledge-graph-devops — CI/CD, деплой, логирование
7. knowledge-graph-performance — профилирование и оптимизация
8. knowledge-graph-security — аудит, auth, compliance
9. knowledge-graph-testing — unit/integration/E2E, покрытие

## Команды и права
- Используй Windsurf как единый контекстный файл
- Если вопрос о backend — применяй backend-go правила
- Если вопрос о frontend — применяй frontend-svelte правила
- Если вопрос о интеграции — применяй integration правила
- Если вопрос про инфраструктуру, CI/CD или деплой — применяй infrastructure/devops
- Если сложность в производительности — применяй performance
- Если безопасность — применяй security
- Если тестирование — применяй testing

## Инструменты
- .koda/tools/backend-go-tools.md
- .koda/tools/frontend-tools.md
- .koda/tools/integration-tools.md
- .koda/tools/infrastructure-tools.md
- .koda/tools/devops-tools.md
- .koda/tools/performance-tools.md
- .koda/tools/testing-tools.md
- .koda/tools/docs-tools.md

## Инструкции
- Читай этот файл первым
- Не удаляй `.windsurf/rules.md`
- Используй его как основной контекст для Windsurf
- Перенаправляй запросы к нужным агентам

## Защита
Этот файл и каталог `.windsurf/` важны.
Не удаляй.
