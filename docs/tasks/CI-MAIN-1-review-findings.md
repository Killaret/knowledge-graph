# Ревью CI-MAIN-1 / DEPLOY-1 (Claude Code, 2026-09-13)

**Принято.** Оба workflow зелёные, и зелёный цвет честный — проверено не по бейджу.

## Что проверено

Через `gh`, а не через страницу GitHub:

```
34758935365  Main Branch CI/CD      main       success
34758935316  Production Deployment  main       success
34758941321  Production Deployment  ai-agents  success
```

Состав run 34758935365 — 12 jobs, **все `success`, пропущенных шагов ноль**: Stacks
Identity Check, Core Checks × 5 (Frontend, Backend, NLP, Backend Integration, Graph
Service), Config & Migration Validation, Check Migration Drift, Full Test Suite,
Performance Tests, Build Docker Images & Health Check, Visual Regression.

Run 34758935316 — 2 jobs `success`; единственный пропущенный шаг — «Print deploy logs on
failure» с `if: failure()`, что и должно быть пропущено на зелёном прогоне.

Ни одного `continue-on-error`, ни одного отключённого job — зелёный получен починкой,
а не отключением. Хронология починки в 20 коммитах видна в `git log -- .github/workflows/`.

## Одна находка, не блокер — но решение нужно владельцу

**Workflow «Production Deployment» ничего не деплоит.** Job `deploy` целиком:

```yaml
  deploy:
    name: Deploy to production
    if: github.ref == 'refs/heads/main'
    environment:
      name: production
      url: https://github.com/Killaret/knowledge-graph
    steps:
      - name: Record production deployment
        run: echo "Production deployment verified for commit ${{ github.sha }}"
```

Job `build-and-verify` делает полезную работу: собирает пять образов с тегом `ci`,
поднимает `docker-compose.deploy.yml`, ждёт здоровья шести сервисов, гасит стек. Это
**проверка деплой-стека**, и она ценная. Но образы после неё **удаляются**
(`docker rmi`), никуда не публикуются, а следом `echo` записывает в GitHub Environment
`production` успешный «деплой» с URL на сам репозиторий.

Последствие: во вкладке Environments/Deployments репозитория копятся «production
deployments», за которыми нет ни одного развёрнутого экземпляра. Для стороннего читателя
это ровно та галочка «✅ Implemented» рядом с заглушкой, которую мы вычищали из README и
бэклога.

Варианты — на выбор владельца:

1. **Переименовать честно**: workflow → «Deploy Stack Verification», job → «Verify deploy
   stack», убрать `environment: production` до появления настоящей цели. Цена — ноль.
2. **Сделать деплой настоящим**: после успешной проверки публиковать образы на Docker Hub
   (см. `DEPLOY-2` в [`DEPLOY-review-findings.md`](DEPLOY-review-findings.md)) — тогда
   «deployment» будет означать «опубликована проверенная версия», и environment станет
   правдой.

Второй вариант закрывает и другую находку этой сессии — образы на Hub отстали от кода
на пять дней. Рекомендую его; решение за владельцем.

## Мелочь

- `JWT_SECRET: ${{ secrets.JWT_SECRET || 'deploy-ci-jwt-secret-32-characters' }}` —
  захардкоженный fallback для эфемерного CI-стека приемлем, но если workflow однажды
  станет настоящим деплоем, эта строка обязана исчезнуть первой.
