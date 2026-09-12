# SECURITY-1: настройки GitHub, Dependabot и порядок чинить security-алерты

## Цель
После включения Dependency graph, Dependabot alerts, CodeQL и Secret Protection появился пул задач по безопасности. Этот файл — порядок настройки и фикса, чтобы Claude и Devin не теряли контекст.

## Что человек должен настроить в UI GitHub

### 1. Branch protection for `main`
- `Settings → Branches → Add rule`
- Branch name pattern: `main`
- Включить:
  - `Require a pull request before merging`
  - `Require approvals: 1`
  - `Dismiss stale PR approvals when new commits are pushed`
  - `Require status checks to pass before merging` (выбрать `Core Checks`, `Smoke Tests`, `Frontend Tests`, `Security Audit`)
  - `Require conversation resolution before merging`
  - `Include administrators` (если хотим строго)
  - `Do not allow bypass the above settings`
  - `Restrict pushes that create files larger than 100MB` / `Block force pushes`

### 2. Dependabot rules (Security → Dependabot → Rules)
- `Dismiss low-impact alerts for development-scoped dependencies` — оставить включённым.
- Опционально: включить `Dismiss package malware alerts`.

### 3. Version updates configure (`Dependabot → Version updates`)
- Пока управляется `.github/dependabot.yml`, но в UI можно включить/отключить `Grouped security updates`.
- Рекомендуется включить `Grouped security updates` — будет меньше security-PR.

### 4. Repository visibility / access
- Подтвердить, что `private/public` соответствует планам.
- `Settings → Actions → General` — убедиться, что `GITHUB_TOKEN` имеет разрешения `Read repository contents and packages` по умолчанию.

## Что добавить в `.github/dependabot.yml` (Devin/Claude)

```yaml
  # Graph Service dependencies
  - package-ecosystem: "gomod"
    directory: "/services/graph-service"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
    open-pull-requests-limit: 5
    reviewers:
      - "knowledge-graph-maintainers"
    assignees:
      - "knowledge-graph-maintainers"
    labels:
      - "dependencies"
      - "graph-service"
    # Разрешить major, но ревьюить вручную
    # ignore:
    #   - dependency-name: "*"
    #     update-types: ["version-update:semver-major"]

  # Root package.json / package-lock.json
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
    open-pull-requests-limit: 5
    reviewers:
      - "knowledge-graph-maintainers"
    assignees:
      - "knowledge-graph-maintainers"
    labels:
      - "dependencies"
      - "root"
```

Также стоит убрать или ослабить `ignore` major-обновлений в `/frontend`, `/backend`, `/nlp-service`, чтобы security-патчи не прятались.

## Порядок чинить security-алерты и CodeQL

### Этап 1: настроить репозиторий (человек + Devin)
1. PR #42 — review GitHub repo security settings.
2. PR #43 — обновить `.github/dependabot.yml` (graph-service + root npm).
3. PR #54 — добавить `permissions:` в `.github/workflows/_core-checks.yml` и `frontend-tests.yml`.

### Этап 2: обновить зависимости по убыванию риска
1. **#46** — root `package-lock.json` (16 critical, `brace-expansion`, `js-yaml`, `undici`).
2. **#48** — `nlp-service/requirements.txt` (`nltk` 3.8.1 → 3.10.x, 33 high).
3. **#44** — `backend/go.mod` (`golang.org/x/crypto`, `grpc`, `docker/docker`, `otel/sdk`, 13 high).
4. **#45** — `services/graph-service/go.mod` (`containerd`, `grpc`, `pgx`, 9 high).
5. **#47** — `frontend/package-lock.json` (Svelte, SvelteKit, sharp, brace-expansion).
6. **#41** — approve/merge open Dependabot PRs #21–#40 in order.

### Этап 3: CodeQL fixes
1. **#49** — cookie `Secure` in `auth/handler.go`.
2. **#50** — SSRF in `import_fetcher.go`.
3. **#51** — weak hashing in `apikey.go` / `user/handler.go`.
4. **#52** — `extract-urls.ts` sanitization.
5. **#53** — `check-core-workflow-sync.mjs` escaping.

### Этап 4: верификация
- После каждого PR: `go test ./...`, `npm run test:unit`, `npm run lint`, `npm audit`, `go list -m -u all`.
- Перед мёрджем: `gh pr checks --watch` и `check-all` локально.

## Ссылки
- `docs/PROJECT_REVIEW_AI_AGENTS.md` §11 — общая сводка.
- Issues #41–#54 — отдельные задачи.
