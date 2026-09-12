# SECURITY-1: настройки GitHub, Dependabot и порядок чинить security-алерты

## Статус (2026-09-12)

| Этап | Статус | Примечание |
|---|---|---|
| Branch ruleset `main` | ✅ активен | Ruleset ID `23024343`, target `main`, 1 approval, status checks, block force-push/deletion |
| `.github/dependabot.yml` | ✅ обновлён | Убраны blanket-major-игноры, добавлены `services/graph-service` и root `npm`, группировка |
| Workflow permissions | ✅ добавлен `permissions:` | `_core-checks.yml`, `frontend-tests.yml`, `ci.yml`, `security.yml` ограничены `contents: read` (+ `actions: write` где нужно) |
| #50 SSRF | ✅ принят риск | Dismissed в CodeQL как `won't fix` по решению владельца |
| #51 weak hashing | 🔄 на ревью | API-ключи перешли на Argon2id; токен `id:secret`, хранится Argon2-хеш; тесты проходят; ожидает мёрджа и повторного скана CodeQL |
| #49 cookie Secure | ⏳ в очереди |  |
| #52/#53 front-end sanitization | ⏳ в очереди |  |
| #54 workflow permissions | 🔄 на ревью | Добавлены `permissions:`, но CodeQL скан ещё не перезапущен на `main` |

## Цель
После включения Dependency graph, Dependabot alerts, CodeQL и Secret Protection появился пул задач по безопасности. Этот файл — порядок настройки и фикса, чтобы Claude и Devin не теряли контекст.

## Что человек должен настроить в UI GitHub

### 1. Branch ruleset for `main` — статус
- Ruleset `main` создан и активен.
- **Но `Target branches` пустой**: нужно нажать `Add target → Include by pattern: main`, иначе правило не применяется.
- Проверить, что включены:
  - `Require a pull request before merging` (1 approval)
  - `Dismiss stale PR approvals when new commits are pushed`
  - `Require status checks to pass before merging`
  - `Block force pushes`
  - `Restrict deletions`

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

## Решения владельца

- **Ignore major version updates** в `.github/dependabot.yml` — убираем. Major-обновления разрешены, но мёрджатся вручную.
- **#51 weak hashing** — переходим на **Argon2**.
- **#50 SSRF в `import_fetcher.go`** — принимаем риск, ограничивать URL не нужно. Добавить комментарий/ADR: фича предназначена для загрузки произвольных публичных URL, current threat model не требует ограничений.
- **Grouped security updates** — включить.

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
2. **#50** — SSRF in `import_fetcher.go` — ✅ dismissed как `won't fix` по решению владельца.
3. **#51** — weak hashing in `apikey.go` / `user/handler.go` — 🔄 реализовано на ветке `security/findings`:
   - `user/handler.go:CreateAPIKey` теперь хеширует секрет API-ключа через `auth.HashPassword` (Argon2id).
   - `middleware/apikey.go` принимает токен формата `<id>:<secret>`, ищет ключ по `id`, проверяет `secret` через `auth.VerifyPassword`.
   - `APIKeyRepository.FindActiveByHash` заменён на `FindActiveByID`; репозиторий и интерфейс обновлены.
   - Добавлены unit-тесты на валидный, невалидный и malformed токен.
   - ⚠️ Это **breaking change** для существующих API-ключей: старые SHA-256-хеши не проверятся, ключи нужно пересоздать. Формат токена изменился с `uuid` на `uuid:secret`.
4. **#52** — `extract-urls.ts` sanitization.
5. **#53** — `check-core-workflow-sync.mjs` escaping.

### Этап 4: верификация
- После каждого PR: `go test ./...`, `npm run test:unit`, `npm run lint`, `npm audit`, `go list -m -u all`.
- Перед мёрджем: `gh pr checks --watch` и `check-all` локально.

## Ссылки
- `docs/PROJECT_REVIEW_AI_AGENTS.md` §11 — общая сводка.
- Issues #41–#54 — отдельные задачи.
