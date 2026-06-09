# Knowledge Graph

База знаний с графовой структурой, перекрёстными ссылками и интеллектуальными рекомендациями. Визуализирует заметки как небесные тела (звёзды, планеты, кометы) в интерактивном 3D-пространстве.

## 🚀 Быстрый старт

### Режимы запуска

| Режим | Порт | Назначение | Команда |
|-------|------|------------|---------|
| **Dev-стек** | 3000 | Разработка с hot-reload | `docker-compose up` |
| **Личный инстанс** | 3001 | Персональное использование | `./start-personal.sh` или `start-personal.ps1` |

Подробнее о развёртывании см. [`docs/architecture/uml/deployment-local.puml`](docs/architecture/uml/deployment-local.puml).

## 🏗️ Архитектура

### Технологический стек

| Компонент | Технология |
|-----------|-----------|
| **Backend** | Go 1.23 + Gin + GORM |
| **Frontend** | Svelte 5 + TypeScript + Three.js |
| **Database** | PostgreSQL 16 + pgvector |
| **Draft Storage** | MongoDB 7 |
| **Cache/Queues** | Redis 7 + asynq |
| **NLP** | Python + FastAPI + sentence-transformers |
| **Backup** | Яндекс.Диск (WebDAV) |
| **Инфраструктура** | Docker Compose |

### Архитектурные паттерны

- **Clean Architecture** — чёткое разделение слоёв (Domain, Application, Infrastructure, Interfaces)
- **Domain-Driven Design (DDD)** — богатая доменная модель с Value Objects и Aggregates
- **CQRS-Lite** — разделение команд и запросов для оптимизации чтения/записи
- **Strategy/Template Method** — гибкие алгоритмы рекомендаций и обработки графов

Подробная документация: [`docs/architecture/README.md`](docs/architecture/README.md)

### Структура проекта

```
knowledge-graph/
├── backend/           # Go backend (REST API, workers)
│   ├── cmd/          # Server и Worker entry points
│   ├── internal/     # DDD слои (domain, application, infrastructure, interfaces)
│   └── migrations/   # SQL миграции
├── frontend/         # SvelteKit frontend (3D визуализация, UI)
├── nlp-service/      # Python NLP сервис (эмбеддинги)
├── docs/             # Архитектура, ADR, UML диаграммы
├── scripts/          # Скрипты по категориям (cleanup, diagnostics, testing, utility)
└── tests/            # BDD тесты (Cucumber + Playwright)
```

## 🛠️ Разработка

### Требования

- Go 1.23+
- Node.js 20+
- Python 3.11+
- Docker & Docker Compose

### Установка зависимостей

```bash
# Backend
cd backend && go mod download

# Frontend
cd frontend && npm install

# NLP Service
cd nlp-service && pip install -r requirements.txt
```

### Запуск для разработки

```bash
# Полный стек через Docker Compose
docker-compose up

# Или отдельно:
# Backend
cd backend && go run ./cmd/server

# Frontend
cd frontend && npm run dev

# NLP Service
cd nlp-service && uvicorn app.main:app --reload
```

### Тестирование

```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && npm run test:unit

# NLP Service
cd nlp-service && pytest
```

Статус тестов: [`TEST_STATUS.md`](TEST_STATUS.md)

## 📚 Документация

- [🎯 Roadmap](docs/ROADMAP.md) — план развития продукта (фазы, задачи, приоритеты)
- [Архитектура](docs/architecture/README.md) — C4 модель, UML диаграммы, ADR
- [Architecture Roadmap](docs/ARCHITECTURE_ROADMAP.md) — план SaaS миграции
- [Резервное копирование](docs/BACKUP.md) — настройка бэкапов для личного инстанса
- Frontend patterns: [frontend/FRONTEND_PATTERNS.md](frontend/FRONTEND_PATTERNS.md)
- [Agents guide](docs/AGENTS.md) — как использовать агенты репозитория
- [Commands reference](COMMANDS.md) — полный справочник команд проекта
- [Scripts reorganization](scripts/docs/SCRIPTS_REORGANIZATION.md) — структура и назначение скриптов
- [API Errors](docs/API_ERRORS.md) — формат ошибок и коды
- [Тесты](TEST_STATUS.md) — статус и покрытие
- [Конфигурация](docs/CONFIGURATION.md) — полное руководство по настройке
- [Конфигурация системы](docs/CONFIGURATION_EN.md) — технические параметры
- [Развертывание](docs/DEPLOYMENT_EN.md) — руководство по развертыванию

## 🔒 Безопасность

Проект использует комплексный подход к безопасности цепочки поставок и CI/CD:

### Защита зависимостей
- ✅ **Только `npm ci`** в CI - никогда не используется `npm install`
- ✅ **minimumReleaseAge=7** в `.npmrc` - блокирует пакеты моложе 7 дней
- ✅ **npm audit** с `--audit-level=high` в CI - билд падает при критических уязвимостях
- ✅ **package-lock.json** закоммичен и контролируется CODEOWNERS
- ✅ **Whitelist для lifecycle scripts** - см. `frontend/NPM_SCRIPTS_WHITELIST.md`

### Автоматический аудит
- ✅ **Dependabot** - еженедельное обновление зависимостей (`.github/dependabot.yml`)
- ✅ **Dependency Review Action** - проверка PR на вредоносные зависимости
- ✅ **Daily security scans** - автоматический аудит уязвимостей

### Защита GitHub и CI/CD
- ✅ **CODEOWNERS** - обязательное review для изменений зависимостей
- ✅ **Minimal permissions** - `contents: read` для GitHub Actions
- ✅ **Branch protection** - требуется approval для main/release веток
- ✅ **Secret scanning** и **push protection** включены

### Для разработчиков
1. Никогда не используйте `npm install` в CI - только `npm ci`
2. Изменения в `package.json`/`package-lock.json` требуют approval maintainers
3. Новые lifecycle scripts должны быть задокументированы в whitelist
4. Всегда проверяйте security отчёты Dependabot PR
## 🤖 AI AGENTS ECOSYSTEM (9 Agents)

**⚠️ CRITICAL FOR AI ASSISTANTS: Read `.koda/STARTUP.md` FIRST in every new chat!**

This project uses a comprehensive AI agent ecosystem for development assistance.

### Agents (9 total)

| Agent | Focus | Tools |
|-------|-------|-------|
| **Orchestrator** | Coordination of all agents | All tools |
| **Backend Go** | Go API, DB, microservices | `backend-go-tools.md` |
| **Frontend Svelte** | Svelte 5, UI/UX, components | `frontend-tools.md` |
| **Python NLP** | Python FastAPI, ML models, NLP | `python-nlp-tools.md` |
| **Integration** | API mapping, DTOs, contracts | `integration-tools.md` |
| **Infrastructure** | Docker, K8s, monitoring | `infrastructure-tools.md` |
| **DevOps** | CI/CD, deployment, backup | `devops-tools.md` |
| **Performance** | Optimization, profiling | `performance-tools.md` |
| **Security** | Security scanning, auth | `security-tools.md` |

### Tools (9 total)

All tools are in `.koda/tools/`:
- `backend-go-tools.md` - REST/gRPC, PostgreSQL, Redis, JWT
- `frontend-tools.md` - Svelte 5, Vitest, Playwright, Testing Library
- `python-nlp-tools.md` - FastAPI, sentence-transformers, YAKE, NLTK
- `integration-tools.md` - OpenAPI, Protocol Buffers, Contract testing
- `infrastructure-tools.md` - Docker multi-stage, Kubernetes, Prometheus
- `devops-tools.md` - CI/CD pipelines, backup scripts
- `performance-tools.md` - pprof, wrk, k6 load testing
- `testing-tools.md` - Unit, integration, E2E, BDD patterns
- `docs-tools.md` - README, ADR, changelog generation

### Rules (2 total)

- `default-rules.md` - Default behavior for all agents
- `orchestration-rules.md` - Agent interaction patterns

### How AI Assistants Should Work

**In every new chat, AI must:**

1. Read `.koda/STARTUP.md` - Critical startup instructions
2. Load all 9 agents via `read_skill()`
3. Load all 9 tools via `read_file()`
4. Apply rules from `.koda/rules/`
5. Show status to user

**Example startup:**
```
🤖 Knowledge Graph Agents - Auto Loaded

✅ Orchestrator: ACTIVE
✅ Backend Go Agent: LOADED
✅ Frontend Svelte Agent: LOADED
✅ Python NLP Agent: LOADED
✅ Integration Agent: LOADED
✅ Infrastructure Agent: LOADED
✅ DevOps Agent: LOADED
✅ Performance Agent: LOADED
✅ Security Agent: LOADED

📊 Loaded agents: 9/9
📊 Loaded tools: 9/9

Ready to work!
```

**Commands for AI:**
- `/agents` - Show loaded agents
- `/tools` - Show loaded tools
- `/rules` - Show applied rules

Эти агенты — не исполняемые команды. Это метаданные, которые помогают выбрать правильный контекст при работе с репозиторием.

## � Devin AI - Development Assistant

**Primary AI Agent:** [Devin](https://cli.devin.ai/docs) by Cognition

Devin is used as the main AI development assistant for this project, providing autonomous coding capabilities with full tool access.

### Devin's Capabilities

- **Full Repository Access:** Read, write, and execute any file in the project
- **Tool Ecosystem:** Access to grep, file operations, shell commands, git, and more
- **Autonomous Problem Solving:** Can debug, implement features, and run tests independently
- **Multi-step Planning:** Uses todo lists to track complex tasks across files
- **CI/CD Integration:** Can run builds, tests, and push changes

### Best Practices for Devin

#### 1. Task Breakdown
- Use `todo_write` tool for complex tasks (3+ steps)
- Mark todos as `in_progress` when starting, `completed` when done
- Only keep ONE todo in progress at a time

#### 2. Code Exploration
- Use `grep` for searching code patterns (not file names)
- Use `find_file_by_name` for finding files by pattern
- Use `read` to understand file contents before editing
- Batch independent file reads for performance

#### 3. Code Changes
- Follow existing code patterns and conventions
- Check dependencies before adding new packages (use `package.json`, `go.mod`)
- Run `npm add` or `go get` instead of editing files directly
- Use existing libraries and utilities
- Follow security best practices (no secrets in code)

#### 4. Testing & Verification
- Run project-specific tests after changes
- Check lint/typecheck/build commands
- Look for verification steps in project rules (AGENTS.md)
- Self-critique changes before marking complete

#### 5. Git Operations
- Run `git status`, `git diff`, `git log` in parallel before committing
- Draft commit messages focusing on "why" not "what"
- Use the standard commit format with Devin attribution:
  ```bash
  git commit -m "type: description

  Generated with [Devin](https://cli.devin.ai/docs)

  Co-Authored-By: Devin <158243242+devin-ai-integration[bot]@users.noreply.github.com>"
  ```
- **NEVER push** unless explicitly requested by user

#### 6. Error Recovery
- Try different approaches when encountering errors
- Search codebase for similar issues/patterns
- Only ask user for help as last resort (except auth/permissions)
- Keep trying reasonable options before giving up

### Devin's Toolbelt

| Tool | Purpose | Example Usage |
|------|---------|--------------|
| `grep` | Search code patterns | Find function usage, type definitions |
| `find_file_by_name` | Find files by pattern | Locate `*.go`, `**/*.svelte` files |
| `read` | Read file contents | Understand implementation before editing |
| `exec` | Execute shell commands | Run tests, builds, git operations |
| `todo_write` | Track tasks | Break down complex features |
| `edit` | Edit files | Make code changes |
| `write` | Write/create files | New components, configuration |
| `git` operations | Version control | Status, diff, commit, push |

### Project-Specific Commands for Devin

**Backend (Go):**
```bash
cd backend && go test ./...           # Run tests
cd backend && go build ./cmd/server   # Build server
cd backend && golangci-lint run       # Lint
```

**Frontend (Svelte):**
```bash
cd frontend && npm run test:unit      # Unit tests
cd frontend && npm run test           # E2E tests (Playwright)
cd frontend && npm run check          # svelte-check
cd frontend && npm run lint           # ESLint
```

**Docker:**
```bash
docker-compose up -d                   # Start services
docker-compose -f docker-compose.test.yml up -d  # CI testing
```

### Configuration Files

- `.devin/config.json` - Devin-specific configuration
- `.devin/skills/knowledge-graph/SKILL.md` - Project-specific Devin skill
- `AGENTS.md` - Project rules and verification commands
- `.cursorrules` - Code style and conventions

### Example Devin Workflow

1. **Receive Task:** "Fix failing tests and add new feature"
2. **Plan:** Create todo list with 5 items
3. **Explore:** Use grep/read to understand codebase
4. **Implement:** Make changes following existing patterns
5. **Verify:** Run tests, lint, typecheck
6. **Commit:** Stage files and commit with standard message
7. **Report:** Update todo list, mark complete

### Recent Devin Contributions

- ✅ Fixed Playwright test errors (_page → page)
- ✅ Resolved 27 svelte-check errors (type fixes, config updates)
- ✅ Updated GraphDelta types (string[] → GraphLink[])
- ✅ Added anomaly configuration types
- ✅ Fixed docker-compose configurations
- ✅ Resolved linting errors (golangci-lint, eslint)
- ✅ Updated npm dependencies for security

## �🧰 Важные команды и утилиты

### Скрипты организованы по категориям в `scripts/`:
- `cleanup/` — скрипты очистки и сжатия
- `diagnostics/` — диагностические скрипты
- `testing/` — тестовые скрипты
- `utility/` — вспомогательные скрипты
- `database/` — скрипты базы данных
- `docs/` — документация скриптов

### Команды очистки:
- `npm run clean:lunix` — запускает Windows-скрипт с Compact.exe (оптимизация диска без Hyper-V).
- `npm run clean:lunix:vhd` — запускает скрипт с Optimize-VHD (требует Hyper-V для VHD сжатия).
- `npm run clean:lunix:dry` — запустить dry-run PowerShell версию.
- `npm run clean:lunix:sh` — запускает bash-версию `scripts/cleanup/clean_and_compress_lunix.sh`.
- `npm run clean:lunix:sh:dry` — запустить dry-run bash версию.
- `make clean-lunix` — alias для `npm run clean:lunix`.
- `make clean-lunix-sh` — alias для `npm run clean:lunix:sh`.
- `npm run clean:docker:vhdx` — VHDX сжатие через DiskPart с авто-разблокировкой (требует админа).
- `make clean-lunix-dry` — dry-run проверки без изменений.

Для периодической очистки:

- `scripts/cleanup/register_cleanup_task.ps1` — регистрирует Scheduled Task в Windows;
- `scripts/cleanup/register_cron.sh` — добавляет запись в WSL/cron.

Диагностические скрипты:
- `scripts/diagnostics/check_all_vhdx.ps1` — проверка размеров VHDX файлов
- `scripts/diagnostics/check_disk_lock.ps1` — проверка блокировок диска
- `scripts/diagnostics/check_file_lock.ps1` — проверка блокировки файла

## ✨ Новые функции

### Резервное копирование (Яндекс.Диск)
- Автоматическое локальное резервное копирование PostgreSQL
- Облачное резервное копирование через Яндекс.Диск (WebDAV)
- Скрипты: `scripts/utility/backup-personal.sh` (Linux/Mac), `scripts/utility/backup-personal.ps1` (Windows)
- Go-сервис: `backend/internal/infrastructure/cloud/yandex_backup.go`
- Docker сервис: `backup_scheduler` в `docker-compose.personal.yml`
- Конфигурация в `knowledge-graph.config.json` секция `backup` с `provider=yandex`
- Подробная документация: [`docs/BACKUP.md`](docs/BACKUP.md)

### Черновики (MongoDB)
- Хранение черновиков заметок в MongoDB
- State pattern для управления состоянием (Active, Publishing, Published, Conflict)
- Автоматическая очистка устаревших черновиков (TTL)
- API endpoints для синхронизации черновиков

### Быстрый захват "Cosmic Dust"
- Плавающий виджет для быстрого создания заметок
- Создание заметок типа `dust` (космическая пыль)
- Фильтр Inbox для быстрых заметок
- Сочетания клавиш: Ctrl+Enter для отправки

### Ускоренная загрузка графа
- `Redis` используется для кэша приватного графа пользователя
- `GET /api/v1/me/graph/cached` возвращает мгновенный кэш
- `GET /api/v1/me/graph/fresh` выдаёт актуальный граф и `delta` для инкрементальных обновлений
- Фронтенд применяет `delta` через `GraphCanvas`, избегая полной перерисовки

### Гибкое сходство ключевых слов в рекомендациях
- **5 стратегий сходства**: Jaccard, Overlap, Tversky, Weighted Jaccard, Cosine
- Конфигурация через `knowledge-graph.config.json`:
  - `keyword_similarity_method` — метод сходства (default: "jaccard")
  - `keyword_tversky_alpha` — параметр alpha для Tversky (default: 0.5)
  - `keyword_tversky_beta` — параметр beta для Tversky (default: 0.5)
- Интеграция через `TraversalService` с использованием `KeywordMatcher`
- Keyword component включается при `gamma > 0`
- Архитектура: `backend/internal/application/recommendation/keyword_similarity.go`
- Подробности: [Рекомендации](docs/RECOMMENDATION_ARCHITECTURE.md)

## 🔧 Pre-commit Hooks

Активировать проверку кода перед коммитом:

```bash
npm run prepare
```

Хуки запускают:
- Frontend: ESLint и Prettier
- Backend: golangci-lint

## 📝 Лицензия

MIT License — см. [LICENSE](LICENSE)

---

**Note:** Скрипты `start-personal.sh` и `start-personal.ps1` запускают личный инстанс на порту 3001 с персистентными данными.

## 🧰 Maintenance scripts

В репозитории есть утилиты для обслуживания локальной среды и оптимизации образов/дисков.

- `scripts/cleanup-docker.ps1` — очистка Docker (dangling images, stopped containers, networks, volumes, builder cache). Пример запуска:

```powershell
.\scripts\cleanup-docker.ps1 -Full -WslOptimize
```

- `scripts/clean_and_compress_lunix.ps1` — поиск и сжатие локального образа/диска `lunix` (Windows). Примеры:

```powershell
# Найти и сжать найденный образ
.\scripts\clean_and_compress_lunix.ps1 -Search -Compress

# Сжать конкретный файл без запроса
.\scripts\clean_and_compress_lunix.ps1 -ImagePath 'D:\images\lunix.vhdx' -Compress -Force
```

- `scripts/clean_and_compress_lunix.sh` — Linux/WSL версия, использует `qemu-img` для сжатия в `qcow2`:

```bash
./scripts/clean_and_compress_lunix.sh -c -p /path/to/lunix.vhdx
```

Также добавлены npm-скрипты для удобного вызова из корня проекта:

```bash
# Windows PowerShell
npm run clean:lunix

# Linux/WSL
npm run clean:lunix:sh
```

Дополнительно поддерживается безопасный режим "dry-run" для оценки действий без внесения изменений:

```powershell
# PowerShell dry-run
.\scripts\clean_and_compress_lunix.ps1 -Search -Compress -DryRun
```

```bash
# Bash dry-run
./scripts/clean_and_compress_lunix.sh --dry-run -c -p /path/to/lunix.vhdx
```

Имеется также удобная `make` цель в корне репозитория:

```bash
make clean-lunix         # запускает npm run clean:lunix
make clean-lunix-sh      # запускает npm run clean:lunix:sh
make clean-lunix-dry     # запускает dry-run (PowerShell)
```

CI: уведомления
----------------
Workflow `.github/workflows/cleanup-dryrun.yml` выполняет dry-run еженедельно и при ручном запуске. При ошибке workflow может отправлять email-уведомления.

Перед включением email-уведомлений добавьте в `Settings -> Secrets` вашего репозитория следующие значения:

- `SMTP_HOST` — адрес SMTP сервера
- `SMTP_PORT` — порт (обычно 465 или 587)
- `SMTP_USERNAME` — логин
- `SMTP_PASSWORD` — пароль
- `NOTIFY_EMAIL_TO` — адрес получателя уведомлений
- `NOTIFY_EMAIL_FROM` — адрес отправителя

Без этих секретов уведомления не будут отправляться, но dry-run продолжит выполняться и загружать логи как артефакты.
