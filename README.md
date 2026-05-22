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
| **Backup** | Cloudflare R2 (S3-compatible) |
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

- [Архитектура](docs/architecture/README.md) — C4 модель, UML диаграммы, ADR
- Frontend patterns: [frontend/FRONTEND_PATTERNS.md](frontend/FRONTEND_PATTERNS.md)
- [Agents guide](docs/AGENTS.md) — как использовать агенты репозитория
- [API Errors](docs/API_ERRORS.md) — формат ошибок и коды
- [Тесты](TEST_STATUS.md) — статус и покрытие
- [Конфигурация](docs/CONFIGURATION.md) — полное руководство по настройке
- [Конфигурация системы](docs/CONFIGURATION_EN.md) — технические параметры
- [Развертывание](docs/DEPLOYMENT_EN.md) — руководство по развертыванию

## 🧠 Работа с агентами и командами

Этот репозиторий содержит три специальных агента, которые помогают оформить и поддержать документацию и задачи:

- `knowledge-graph-frontend-svelte` — для frontend задач: анализ UI, Svelte-компонентов, тестовой инфраструктуры и frontend-доков.
- `knowledge-graph-backend-go` — для backend/infra задач: Go-код, БД, Docker, Redis, API и backend-документации.
- `knowledge-graph-docs-maintenance` — для создания, актуализации и оформления документации: `README.md`, `docs/`, ADR, описания изменений и сопроводительных материалов.

Эти агенты — не исполняемые команды. Это метаданные, которые помогают выбрать правильный контекст при работе с репозиторием.

## 🧰 Важные команды и утилиты

- `npm run clean:lunix` — запускает Windows-скрипт `scripts/clean_and_compress_lunix.ps1`.
- `npm run clean:lunix:sh` — запускает bash-версию `scripts/clean_and_compress_lunix.sh`.
- `make clean-lunix` — alias для `npm run clean:lunix`.
- `make clean-lunix-sh` — alias для `npm run clean:lunix:sh`.
- `make clean-lunix-dry` — dry-run проверки без изменений.

Для периодической очистки:

- `scripts/register_cleanup_task.ps1` — регистрирует Scheduled Task в Windows;
- `scripts/register_cron.sh` — добавляет запись в WSL/cron.

## ✨ Новые функции

### Резервное копирование (Cloudflare R2)
- Автоматическое локальное резервное копирование PostgreSQL
- Облачное резервное копирование через Cloudflare R2
- Скрипты: `scripts/backup-personal.sh` (Linux/Mac), `scripts/backup-personal.ps1` (Windows)
- Конфигурация в `knowledge-graph.config.json` секция `backup`

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
