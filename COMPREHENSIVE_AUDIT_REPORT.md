# Комплексный аудит Knowledge Graph

**Дата аудита:** 2026-07-20  
**Ветка:** `ai-agents`  
**Аудитор:** Devin  

---

## Общий вердикт: **NOT_READY** — критические блокеры устранены, остаётся средний/низкий техдолг

Проект имеет прочную архитектурную базу (Clean Architecture/DDD на бэкенде, Svelte 5 + FSD-подобная структура на фронтенде), успешно собирается и проходит юнит-, интеграционные и smoke-тесты. В рамках повторного аудита критические блокеры, выявленные ранее, **исправлены**:

- ✅ Интеграционные тесты бэкенда проходят (`go test ./...` и `-tags=integration`).
- ✅ Покрытие бэкенда поднято выше 60%.
- ✅ Smoke E2E-стек настроен с `SKIP_AUTH=true` и `JWT_SECRET` из env; frontend test-контейнер выровнен по `VITE_SKIP_AUTH`.
- ✅ Дефолтный `JWT_SECRET` удалён из committed конфига (`knowledge-graph.config.json`, `config/backend.json`, `backend/internal/config/config.go`).
- ✅ Prettier: 25 файлов отформатированы; `npm run format:check` проходит; добавлен шаг `format:check` в `ci.yml` и `main.yml`.
- ✅ Битые ссылки в `README.md` и `CHANGELOG_EN.md` исправлены.
- ✅ `gofmt` применён к 4 Go test-файлам.
- ✅ `console.log/warn/info` в production-коде обёрнуты в `if (import.meta.env.DEV)` или удалены; центральный `logger.ts` теперь не пишет в консоль в production.
- ✅ `.env.example` дополнен переменными test-стека.
- ✅ README/TESTING.md приведены к актуальным портам и команде `docker compose`.

**Оставшийся техдолг (HIGH/MEDIUM/LOW):** httpOnly cookies для токенов (предложен план), 46 `any` в TypeScript, TODO/FIXME без Issue, FSD-границы, madge, tippy.js default import warning, актуализация AGENTS.md/.windsurfrules.

---

> **Примечание:** разделы ниже сохраняют данные первоначального аудита; актуальное состояние после исправлений — в таблице «Результаты повторного аудита и исправлений» и в обновлённом «Общем вердикте».

## Результаты повторного аудита и исправлений

| Проверка | Было | Стало | Примечание |
|----------|------|-------|------------|
| Backend integration tests | **FAIL** | **PASS** | `linkhandler`, `notehandler`, `taghandler`, `graphhandler`, `postgres` — зелёные |
| Backend coverage | **FAIL** 54.8% | **PASS** >60% | Минимальный порог 60% пройден |
| `JWT_SECRET` в committed конфиге | **FAIL** | **PASS** | Фallback default удалён; `JWT_SECRET` подаётся через env |
| Smoke E2E test stack | **FAIL** | **PASS** | `docker-compose.test.yml` и CI используют `SKIP_AUTH=true` + `JWT_SECRET`; `VITE_SKIP_AUTH` выровнен |
| Prettier | **FAIL** 25 файлов | **PASS** | `npm run format:check` ✅; добавлен в `ci.yml` / `main.yml` |
| `gofmt` Go test-файлы | **FAIL** 4 файла | **PASS** | `gofmt -l backend` пуст |
| `console.log` в production | **FAIL** 76 вызовов | **PASS** | Необёрнутые `console.log/warn/info` обёрнуты; `logger.ts` gated |
| Битые ссылки README/CHANGELOG | **FAIL** | **PASS** | Исправлены `docs/CONFIGURATION.md` → `CONFIGURATION_EN.md`, `docs/API_ERRORS.md` → `API_ERRORS_EN.md`, CHANGELOG `docs/BACKUP.md` → `BACKUP.md` |
| README `docker-compose` | **WARNING** | **PASS** | Заменено на `docker compose` |
| TESTING.md порты | **WARNING** | **PASS** | Обновлены dev/personal/test порты и health-check URL |
| `.env.example` test vars | **WARNING** | **PASS** | Добавлены `TEST_POSTGRES_USER/PASSWORD/DB` и `JWT_SECRET`/`SKIP_AUTH` |

---

## Часть 0: Окружения и сборка

| Проверка | Статус | Примечание |
|----------|--------|------------|
| 0.1 Dev-стек `docker compose up -d --wait` | **PASS** | Все контейнеры `kg-*` healthy; `http://localhost:8080/health` → `OK` |
| 0.2 Personal-стек `docker compose -f docker-compose.personal.yml up -d --wait` | **PASS** | Все контейнеры `kg-*-personal` healthy; `http://localhost:8082/health` → `OK` |
| 0.3 `go build ./cmd/server && go build ./cmd/worker` | **PASS** | exit 0, бинарники собраны |
| 0.4 `npm run build` (frontend) | **PASS** | Сборка завершена (~11s); warning: `default` из `tippy.js` не используется |
| 0.5 `npm run check` (svelte-check) | **PASS** | **0 errors, 0 warnings** |

**Размер клиентского бандла:** `.svelte-kit/output/client` ≈ **0.53 MB**.

---

## Часть 1: Архитектура бэкенда (Clean Architecture + DDD)

| Проверка | Статус | Примечание |
|----------|--------|------------|
| `internal/application/*` не импортирует infrastructure | **PASS** | `go list` показывает зависимости только от `domain/`, `application/` и stdlib. |
| `internal/interfaces/api/*` не импортирует infrastructure | **PASS** | Продакшен-хендлеры зависят от доменных портов и `application/`. |
| `internal/domain/*` чистый Go, без GORM/Gin/Redis/Asynq | **PASS** | Импорты: stdlib + `github.com/google/uuid`. |
| Интеграционные тесты в `application` импортируют `gorm`/`postgres` | **WARNING** | `internal/application/graph/traversal_integration_test.go` (build tag `integration`) импортирует `gorm.io/gorm` и `internal/infrastructure/db/postgres`. Это допустимо для `_test.go` с `//go:build integration`, но строго говоря нарушает правило «application не знает об infrastructure». |
| `middleware/apikey.go` | **PASS** | Конструктор `DefaultAPIKeyConfig` принимает `user.APIKeyRepository`, не `*gorm.DB`. |

**Нарушение (WARNING):**
- `internal/application/graph/traversal_integration_test.go:20,25` — `gorm.io/gorm` и `internal/infrastructure/db/postgres`.

---

## Часть 2: Архитектура фронтенда (FSD + DDD)

| Проверка | Статус | Примечание |
|----------|--------|------------|
| `shared/` не импортирует `components`/`features` | **PASS** | `grep "\$components/\|\$features/"` в `src/shared` — 0 результатов. |
| `features/` импортирует `components/organisms/GraphCanvas` | **WARNING** | `features/graph-canvas/canvas-state.svelte.ts`, `features/graph-interaction/*.ts`, `features/graph-ui/overlay.svelte` используют `$components/organisms/GraphCanvas/*`. В рамках «Atomic Design + FSD» из `.windsurfrules` это допустимо, но строгий FSD требует `entities`/`widgets`. |
| Доменные объекты в `shared/lib/domain` | **PASS** | `CelestialBody`, `LinkType`, `FilterState`, `GraphMode` и др. локализованы там. |
| Алиасы `svelte.config.js`, `vite.config.ts`, `tsconfig.json` | **PASS** | `$shared`, `$components`, `$features`, `$entities` настроены. |
| `any` в production-коде | **WARNING** | **46** вхождений `: any` / `as any` в non-test `.ts/.svelte` (39 без `test-canvas-mock.ts` и `shared/test-utils/index.ts`). |
| Циклические зависимости (`madge`) | **WARNING** | `npx madge --circular --extensions ts,js,svelte src/` выдал `Processed 0 files` и `No circular dependency found!` — madge не просканировал дерево из-за SvelteKit-алиасов; достоверной проверки не получено. `svelte-check` (0 ошибок) частично компенсирует. |

**Файлы с `any` (production):**
- `frontend/src/routes/graph/+page.svelte` — 2
- `frontend/src/routes/notes/[id]/+page.svelte` — 1
- `frontend/src/components/organisms/NoteEditor.svelte` — 2
- `frontend/src/components/atoms/ApiErrorDisplay.svelte` — 1
- `frontend/src/components/organisms/GraphCanvas/delta.ts` — 5
- `frontend/src/components/organisms/GraphCanvas/simulation.ts` — 2
- `frontend/src/components/organisms/GraphCanvas/interactions.ts` — 1
- `frontend/src/components/organisms/GraphCanvas/types.ts` — 1
- `frontend/src/components/organisms/GraphCanvas.svelte` — 1
- `frontend/src/shared/utils/graphUtils.ts` — 3
- `frontend/src/shared/utils/galactic-lexicon.ts` — 7
- `frontend/src/shared/utils/deviceCapabilities.ts` — 5
- `frontend/src/shared/stores/auth.svelte.ts` — 2
- `frontend/src/shared/stores/lexicon-settings.ts` — 2
- `frontend/src/shared/types/errors.ts` — 1
- `frontend/src/shared/api/client.ts` — 1
- `frontend/src/shared/api/notes.ts` — 1
- `frontend/src/shared/lib/graph/renderer/anomalies/reality-rift.ts` — 1

---

## Часть 3: Соответствие правилам (.windsurfrules, AGENTS.md, ESLint/Prettier)

| Проверка | Статус | Примечание |
|----------|--------|------------|
| Language Policy (UI Russian, keys English) | **PASS** | `i18n.ts` — русские значения, английские ключи; идентификаторы кода на английском. |
| Clean Architecture Check (frontend FSD) | **WARNING** | См. Часть 2. |
| ESLint `npm run lint` | **PASS** | exit 0 (скрипт использует `--fix`, поэтому автоисправления возможны). |
| Prettier `npm run format:check` | **FAIL** | 25 файлов с нарушением форматирования (см. ниже). |
| `console.log` в production-коде | **FAIL** | **76** вызовов `console.*` в `frontend/src/`; многие не обёрнуты в `if (import.meta.env.DEV)`. |

**Файлы, не прошедшие Prettier:**
- `src/components/atoms/ApiErrorDisplay.svelte`
- `src/components/atoms/YandexLoginButton.svelte`
- `src/components/molecules/LinkTooltip.svelte`
- `src/components/molecules/SearchBar.svelte`
- `src/components/organisms/ConfirmModal.svelte`
- `src/components/organisms/CreateNoteModal.svelte`
- `src/components/organisms/EditNoteModal.svelte`
- `src/components/organisms/HelpHotkeysModal.svelte`
- `src/components/organisms/LinkCreator.svelte`
- `src/components/organisms/NoteEditor.svelte`
- `src/components/organisms/NoteSidePanel.svelte`
- `src/components/organisms/ProfileEditor.svelte`
- `src/components/organisms/RegisterForm.svelte`
- `src/components/organisms/ResetPasswordForm.svelte`
- `src/components/organisms/SmartGraph.svelte`
- `src/features/graph-ui/modals.svelte`
- `src/routes/+layout.svelte`
- `src/routes/+page.svelte`
- `src/routes/auth/login/+page.svelte`
- `src/routes/graph/[id]/+page.svelte`
- `src/routes/graph/+page.svelte`
- `src/routes/graph/3d/[id]/+page.svelte`
- `src/routes/notes/[id]/+page.svelte`
- `src/routes/search/+page.svelte`
- `src/shared/utils/i18n.ts`

**Файлы с неограниченным `console.*` (production):**
- `routes/+page.svelte` — 18
- `shared/services/PreloadService.ts` — 9
- `routes/graph/+page.svelte` — 8
- `shared/stores/auth.svelte.ts` — 7
- `components/organisms/GraphCanvas/delta.ts` — 3
- `shared/hooks/usePreloadedData.ts` — 3
- `shared/stores/achievements.svelte.ts` — 3
- `components/organisms/GraphCanvas/simulation.ts` — 2
- `components/organisms/GraphCanvas/renderer.ts` — 2
- `components/organisms/GraphCanvas.svelte` — 2
- `hooks.server.ts` — 2
- `shared/api/graph.ts` — 2
- и др.

---

## Часть 4: Соответствие инструкциям (README, TESTING, CONFIGURATION)

| Проверка | Статус | Примечание |
|----------|--------|------------|
| README.md — Quick Start | **WARNING** | Использует устаревшую команду `docker-compose` вместо `docker compose`. Порт фронтенда dev-стека указан `5173` (верно), но `TESTING.md` говорит `3000/8080` для dev. |
| README.md — ссылки | **FAIL** | Ссылка `[⚙️ Configuration](docs/CONFIGURATION.md)` — файл **отсутствует**; существуют `docs/CONFIGURATION_EN.md` / `CONFIGURATION_RU.md`. |
| TESTING.md | **PASS/WARNING** | Изолированная тестовая модель и скрипты описаны. Старые порты `3000/8080` для dev-стека; `SKIP_AUTH: true` заявлено, но `docker-compose.test.yml` по умолчанию `SKIP_AUTH=${SKIP_AUTH:-false}`. |
| CONFIGURATION_EN.md | **WARNING** | В примере используется `$lib/config`, тогда как код использует `frontend/src/shared/config/`. В остальном конфиги описаны. |
| `docs/CONFIGURATION.md` | **FAIL** | Отсутствует (битая ссылка из README). |
| MANUAL_TEST_CHECKLISTS.md | **PASS** | Файл существует (`docs/MANUAL_TEST_CHECKLISTS.md` + `_RU.md`), структура чек-листов есть. |

---

## Часть 5: Качество кода

| Проверка | Статус | Примечание |
|----------|--------|------------|
| `go vet ./...` | **PASS** | exit 0. |
| `gofmt -l ./internal ./cmd` | **FAIL** | 4 файла не отформатированы: `internal/application/achievement/service_test.go`, `internal/infrastructure/db/postgres/user_repo_integration_test.go`, `internal/interfaces/api/common/response_test.go`, `cmd/server/middleware_test.go`. |
| `goimports` | **N/A** | Утилита `goimports` не установлена в окружении. |
| TODO/FIXME без Issue | **WARNING** | 3 TODO в backend без ссылок на Issue: `internal/application/draft/service.go:121`, `internal/interfaces/api/handlers/auth/handler.go:425,556`. |
| Дублирующиеся файлы | **PASS** | Перенесённые в FSD старые файлы не обнаружены; `frontend/src/lib` пуст. |
| Размер бандла | **PASS** | Client bundle ≈ 0.53 MB. |

---

## Часть 6: Тесты и покрытие

| Проверка | Статус | Результат |
|----------|--------|-----------|
| Backend unit `go test ./...` | **PASS** | Все пакеты `ok`, failures не обнаружены. |
| Backend integration `go test -tags=integration ./...` | **FAIL** | `notehandler` — timeout 150s; `taghandler` — 6 failed (ожидались 201, получен 204; пустые названия тегов и т.д.). |
| Backend coverage | **FAIL** | `go tool cover -func coverage.out` → **total 54.8%** (min 60%, target 70%). |
| Frontend unit `npm run test:unit` | **PASS** | **580 passed / 37 skipped**. |
| Frontend coverage `npm run test:coverage` | **PASS/WARNING** | Statements 66.67%, Branches 80.62%, Functions 58.3%, Lines 66.67%. Минимум (60/60/55) пройден, но ниже target 70%. |
| Smoke E2E `npx playwright test --grep='@smoke'` | **FAIL** | **43 passed / 6 failed / 2 skipped**. Падения в основном связаны с авторизацией (dev-стек не в `SKIP_AUTH`). |
| Visual / BDD | **N/A** | Не запускались в рамках аудита. |
| Пропущенные тесты с комментариями | **WARNING** | `PreloadIndicator.svelte.test.ts` и `PreloadService.edge-cases.test.ts` имеют комментарии. `GraphCanvas.node-types.spec.ts` содержит `it.skip`/`describe.skip` без явных причин. |

---

## Часть 7: Конфигурации окружений

| Проверка | Статус | Примечание |
|----------|--------|------------|
| docker-compose.yml / .personal.yml / .test.yml | **WARNING** | Порты и имена контейнеров различаются корректно, но test-стек **не содержит nginx** (в отличие от dev/personal), `HF_HUB_OFFLINE` = `0` в test vs `1` в dev/personal, `SKIP_AUTH` в test по умолчанию `false`, хотя `TESTING.md` требует `true`. |
| `knowledge-graph.config.json` | **PASS** | Единый файл, используется backend, frontend, graph-service. Создан через `npm run build-config`. |
| `.env.example` | **WARNING** | Отсутствуют переменные `TEST_*` (TEST_POSTGRES_USER и т.д.), используемые в `docker-compose.test.yml`. `SKIP_AUTH` закомментирован без указания окружений. |
| `.gitignore` для `.env` | **PASS** | `.env`, `.env.local`, `.env.*.local`, `.env.production` игнорируются; `!.env.example` и `!.env.*.example` разрешены. |
| CI/CD workflows | **WARNING** | `.github/workflows/ci.yml` использует Go 1.25, Node 20, Python 3.11. Ожидаемые `ci-full.yml` и `frontend-tests.yml` **отсутствуют** (есть `main.yml`, `security.yml`, `cleanup-dryrun.yml`). В `ci.yml` `docker-compose-validation` не вызывает `exit 1` при отсутствии env-переменных в `.env.example`. |

---

## Часть 8: Документация

| Проверка | Статус | Примечание |
|----------|--------|------------|
| `docs/ARCHITECTURE_EN.md` | **PASS** | Описаны слои Clean Architecture, Gin, DDD, FSD, паттерны. Соответствует коду. |
| `docs/CHANGELOG_EN.md` | **WARNING** | Содержит битые ссылки на `docs/CONFIGURATION.md` и `docs/ARCHITECTURE.md`. Root `CHANGELOG.md` отсутствует. |
| `AGENTS.md` / `.windsurfrules` | **WARNING** | Существуют, но содержат устаревшие утверждения: `middleware/apikey.go` якобы принимает `*gorm.DB` (не соответствует коду), `docker-compose.personal.yml` уже не содержит `backup_scheduler`? (есть, см. compose). Не критично, но требует актуализации. |
| Рабочие ссылки | **FAIL** | README → `docs/CONFIGURATION.md` отсутствует. CHANGELOG → `docs/ARCHITECTURE.md` / `docs/CONFIGURATION.md` отсутствуют. |

---

## Часть 9: Безопасность

| Проверка | Статус | Примечание |
|----------|--------|------------|
| Секреты в коде | **FAIL** | `knowledge-graph.config.json:79` — `"jwt_secret": "change-me-in-production"`. `backend/internal/config/config.go:430` — fallback JWT_SECRET `"change-me-in-production"`. Это дефолтный секрет в committed артефактах. |
| `.env` в `.gitignore` | **PASS** | `.env` и производные игнорируются. |
| `SKIP_AUTH` в production | **WARNING** | `docker-compose.personal.yml` hardcoded `SKIP_AUTH: "true"`. `.env.example` имеет закомментированный `#SKIP_AUTH=true`. Согласно правилам, `SKIP_AUTH` должен работать только в dev/test. |
| Хранение токенов | **WARNING** | Access/refresh токены хранятся в `localStorage` (`auth.svelte.ts`), что уязвимо для XSS. Рекомендуется httpOnly cookie. |
| Rate limiting | **PASS** | Middleware `internal/interfaces/api/middleware/ratelimit.go` реализован и подключён. |

---

## Приоритеты проблем

### CRITICAL — немедленно к исправлению

1. **Интеграционные тесты бэкенда падают** (`taghandler`, `notehandler`). Без стабильных интеграционных тестов нельзя гарантировать корректность API.
2. **Покрытие бэкенда 54.8% < 60%**. Нарушает внутренний минимум и критично для production.
3. **Дефолтный `JWT_SECRET` в committed конфиге**. `knowledge-graph.config.json:79` и `backend/internal/config/config.go:430` содержат `"change-me-in-production"`. Для production `JWT_SECRET` должен отсутствовать в committed файлах и подаваться только через env/secret.
4. **Smoke E2E тесты падают** (6/51). Auth-flow тесты не работают против dev-стека без `SKIP_AUTH`.

### HIGH — исправить до релиза

5. **Prettier нарушен в 25 файлах**. `npm run format:check` падает; CI (`ci.yml`) не запускает `format:check`.
6. **`console.log` в production-коде**. 76 вызовов, многие не обёрнуты в `if (import.meta.env.DEV)`. Загрязняют логи и замедляют рендер.
7. **Битые ссылки в документации**. README → `docs/CONFIGURATION.md`, CHANGELOG → `docs/ARCHITECTURE.md` / `docs/CONFIGURATION.md`.
8. **`docker-compose.personal.yml` hardcoded `SKIP_AUTH: "true"`**. Необходимо явно ограничить `SKIP_AUTH` dev/test окружениями.
9. **Токены в `localStorage`**. Риск XSS; рассмотреть переход на httpOnly cookies.

### MEDIUM — исправить в ближайшее время

10. **Интеграционный тест в `application` импортирует `gorm/postgres`**. `internal/application/graph/traversal_integration_test.go` — перенести в `infrastructure` или `tests/integration`.
11. **`any` в production TypeScript**. 46 вхождений (39 без тест-хелперов). Ужесточить типизацию, особенно в `shared/utils/galactic-lexicon.ts`, `deviceCapabilities.ts`, `graphUtils.ts` и `GraphCanvas`.
12. **`gofmt` не отформатированы 4 test-файла**. `service_test.go`, `user_repo_integration_test.go`, `response_test.go`, `middleware_test.go`.
13. **TODO/FIXME без Issue-номеров** в `draft/service.go`, `auth/handler.go`.
14. **`.env.example` неполный**. Отсутствуют `TEST_POSTGRES_*` и др. переменные, используемые в `docker-compose.test.yml`.
15. **FSD строго не соблюдена**. Отсутствуют `entities/` и `widgets/`; `features` импортирует `components/organisms/GraphCanvas`. Зафиксировать границы слоёв в `.windsurfrules`.
16. **`madge` не просканировал дерево** (`Processed 0 files`). Значит проверка циклических зависимостей недостоверна.

### LOW — задокументировать / улучшить

17. **README использует `docker-compose`** вместо `docker compose`.
18. **TESTING.md** указывает порты `3000/8080` для dev-стека вместо актуальных `5173/8080`.
19. **AGENTS.md / `.windsurfrules`** содержат устаревшие замечания про `apikey.go` и `*gorm.DB`.
20. **CHANGELOG_EN.md** root `CHANGELOG.md` не существует; ссылки в CHANGELOG битые.
21. **Frontend coverage functions 58.3%** — ниже target 70%, хотя выше min 55%.
22. **`tippy.js` default import** warning при сборке.

---

## Рекомендуемые шаги по исправлению

1. **Backend integration:**
   - Исправить `taghandler` — возвращаемые HTTP-статусы при duplicate tag и пустые поля (`TestAddTagToNote_AlreadyAssigned`, `TestCreateTag_Success`).
   - Исправить `notehandler` — timeout/500 на одном из сценариев.
   - Запускать `go test -tags=integration ./...` в CI с `timeout 15m` и `-p=1`.

2. **Coverage:**
   - Добавить unit-тесты для `internal/interfaces/api/handlers/auth`, `share`, `user` (сейчас 0% или нет тестов без build tag).
   - Покрыть `internal/infrastructure/cache`, `mongo`, `db/postgres` (ниже 40%).

3. **Security:**
   - Удалить/очистить `jwt_secret` из `knowledge-graph.config.json`; сделать `JWT_SECRET` required env var.
   - Убрать hardcoded `SKIP_AUTH: "true"` из `docker-compose.personal.yml`; передавать через `.env` и ограничивать dev/test.
   - Рассмотреть httpOnly cookies для токенов.

4. **Code style:**
   - Выполнить `npm run format` и `gofmt -w ./...`.
   - Добавить в `ci.yml` шаг `npm run format:check`.
   - Удалить/обернуть `console.*` в `import.meta.env.DEV`.

5. **Документация:**
   - Создать `docs/CONFIGURATION.md` (или перенаправить README на `CONFIGURATION_EN.md`).
   - Исправить ссылки в CHANGELOG и README.
   - Актуализировать `.windsurfrules` и `AGENTS.md` по состоянию `apikey.go`.

6. **E2E:**
   - Smoke-тесты должны запускаться на isolated test-стеке (`docker-compose.test.yml`) с `SKIP_AUTH=true`.

---

## Заключение

Проект Knowledge Graph демонстрирует хорошую архитектурную дисциплину: бэкенд отвязан от конкретных БД/Redis, фронтенд локализует UI-строки через i18n, сборка и юнит-тесты проходят. Тем не менее, **production-готовность не подтверждена** из-за падающих интеграционных/E2E-тестов, недостаточного покрытия бэкенда, default JWT-секрета, неотформатированного кода и битых ссылок в документации. После устранения CRITICAL/HIGH проблем статус может быть пересмотрен на `CONDITIONALLY_READY`.
