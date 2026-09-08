# Команды проекта Knowledge Graph

Полный справочник команд для быстрого доступа ко всем операциям проекта.

## Структура скриптов

Скрипты проекта организованы по семантическим категориям в директории `scripts/`:

```
scripts/
├── backfill/         # Разовая заливка данных
├── ci/               # Проверки и воркфлоу CI
├── cleanup/          # Скрипты очистки
├── database/         # Скрипты для работы с базой данных
├── devops/           # Бэкапы, стражи, проверки политик
└── testing/          # Тест-стек, регрессионные прогоны
```

**Cleanup скрипты:**
- `cleanup-docker.ps1` / `cleanup-docker.sh` - Очистка Docker + опциональное сжатие VHD через diskpart (`-WslOptimize`, требует прав администратора; без Hyper-V)
- `cleanup-logs.ps1` / `cleanup-logs.sh` - Очистка логов и старых снапшотов тестов
- `cleanup-test-artifacts.py` - Очистка артефактов тестовых прогонов

## 🚀 Команды запуска и разработки

### Основные команды (корень проекта)
```bash
npm run prepare                    # Установка husky git hooks
npm run lint                       # Линтинг frontend кода
npm run lint:backend               # Линтинг backend Go кода
npm run format                     # Форматирование frontend кода
npm run build-config               # Сборка конфигурации
npm run test                       # Запуск unit тестов
npm run clean:logs                 # Очистка логов и старых снапшотов
npm run clean:logs:dry             # То же в режиме предпросмотра
npm run clean:test-artifacts       # Очистка артефактов тестов
npm run clean:test-artifacts:dry   # То же в режиме предпросмотра
```

### Makefile команды
```bash
make clean-logs                    # Очистка логов и старых снапшотов
make clean-logs-dry                # То же в режиме предпросмотра
```

### Docker cleanup и оптимизация диска
```bash
.\scripts\cleanup\cleanup-docker.ps1              # Базовая очистка; тома сохраняются
.\scripts\cleanup\cleanup-docker.ps1 -DryRun      # Предпросмотр: что уйдёт, ничего не удаляется
.\scripts\cleanup\cleanup-docker.ps1 -Full        # Удалить ВСЕ неиспользуемые образы; тома сохраняются
.\scripts\cleanup\cleanup-docker.ps1 -WslOptimize # + сжатие VHD WSL2 — требует запуск от администратора
bash scripts/cleanup/cleanup-docker.sh            # То же на Linux/macOS: флаги -f, -o, -n|--dry-run
```

Скрипт завершается ненулевым кодом, если хоть один шаг не удался — `[PASS]` печатается только при реальном успехе. Шаги, способные тронуть тома (`-RemoveVolumes`, `-Full`), отказываются без свежего непустого бэкапа Personal-стека.

## 🎨 Frontend команды

### Запуск и сборка
```bash
cd frontend
npm run dev                        # Запуск dev сервера (Vite)
npm run build                      # Production сборка
npm run preview                    # Предпросмотр production сборки
```

### Единая локальная проверка Core Checks

```powershell
.\scripts\testing\check-all.ps1          # Все локально доступные проверки
.\scripts\testing\check-all.ps1 -Quick   # Интеграционные явно помечаются [SKIP]
```

```bash
./scripts/testing/check-all.sh
./scripts/testing/check-all.sh --quick
```

Команда повторяет пять джоб `_core-checks.yml`. Недоступные инструменты не скрываются: каждая такая фаза получает `[SKIP]` с причиной, а итог помечается `COMPLETE WITH SKIPS`.

### Проверка кода
```bash
cd frontend
npm run check                      # Проверка типов SvelteKit
npm run check:watch                # Проверка типов в watch режиме
npm run lint                       # ESLint проверка (без изменения файлов)
npm run lint:fix                   # ESLint с авто-фиксом
npm run format                     # Prettier форматирование
npm run format:check               # Проверка форматирования
```

### Unit тесты (Vitest)
```bash
cd frontend
npm run test:unit                  # Запуск unit тестов
npm run test:unit:watch            # Unit тесты в watch режиме
npm run test:coverage              # Unit тесты с coverage
```

### E2E тесты (Playwright)
```bash
cd frontend
npm run test                       # Запуск E2E тестов
npm run test:headed                # E2E тесты с видимым браузером
npm run test:debug                 # E2E тесты в debug режиме
npm run test:smoke                 # Smoke тесты только
npm run test:visual                # Visual регрессионные тесты
npm run test:lighthouse            # Lighthouse CI тесты
```

### BDD тесты (Cucumber)
```bash
cd frontend
npm run test:cucumber               # Запуск Cucumber BDD тестов
npm run test:bdd                   # Алиас для test:cucumber
```

### Комплексные тесты
```bash
cd frontend
npm run test:all                   # Все тесты (unit + E2E + BDD)
npm run test:ci:smoke              # CI smoke тесты
npm run test:ci:full               # Полные CI тесты
```

## 🔧 Backend команды (Go)

### Запуск и разработка
```bash
cd backend
go run ./cmd/server                # Запуск сервера
go run ./cmd/worker                # Запуск worker
go run ./cmd/seed                  # Запуск seed скрипта
```

### Тестирование
```bash
cd backend
go test ./... -v                   # Все unit тесты
go test ./internal/domain/... -v   # Domain слой тесты
go test -race ./...                # С race detection
go test -tags=integration ./...    # Интеграционные тесты
go test -coverprofile=coverage.out ./...  # С coverage report
go tool cover -html=coverage.out   # HTML coverage report
```

### Сборка и linting
```bash
cd backend
go build -o server ./cmd/server    # Сборка binary
golangci-lint run ./...            # Линтинг Go кода
golangci-lint run --new-from-rev=HEAD~1  # Линтинг новых изменений
```

### Миграции БД
```bash
cd backend
go run ./cmd/checkmigrations      # Проверка миграций
```

## 🧠 NLP Service команды (Python)

### Запуск
```bash
cd nlp-service
uvicorn app.main:app --reload      # Dev сервер с auto-reload
uvicorn app.main:app               # Production сервер
python -m app.main                 # Альтернативный запуск
```

### Тестирование
```bash
cd nlp-service
pytest tests/ -v                   # Все тесты
pytest tests/test_api.py -v        # API тесты
pytest tests/test_nlp_utils.py -v  # NLP utils тесты
pytest --cov=.                    # С coverage
```

### Зависимости
```bash
cd nlp-service
pip install -r requirements.txt    # Установка зависимостей
pip freeze > requirements.txt      # Обновление requirements
```

## 🐳 Docker команды

### Полный стек
```bash
docker compose up                   # Запуск всех сервисов
docker compose up --build           # Пересборка и запуск
docker compose down                 # Остановка всех сервисов
docker compose down -v              # Остановка с удалением volumes
```

### Отдельные сервисы
```bash
docker compose up postgres redis   # Только БД и кэш
docker compose up backend           # Только backend
docker compose up frontend          # Только frontend
docker compose up nlp-service      # Только NLP сервис
```

### Управление
```bash
docker compose ps                   # Статус сервисов
docker compose logs -f backend      # Логи backend
docker compose restart backend     # Перезапуск backend
docker compose exec backend bash   # Shell в backend контейнере
```

## 🧪 Тестирование по уровням

### Unit тесты
```bash
# Go backend
cd backend && go test ./... -v

# Frontend TypeScript
cd frontend && npm run test:unit

# Python NLP
cd nlp-service && pytest tests/ -v
```

### Интеграционные тесты
```bash
# Go backend
cd backend && go test -tags=integration ./...

# Frontend E2E
cd frontend && npm run test
```

### BDD тесты
```bash
# Cucumber сценарии
cd frontend && npm run test:bdd
```

### Все тесты
```bash
# Полный набор frontend
cd frontend && npm run test:all

# Комплексная проверка
npm run test && cd backend && go test ./... && cd ../nlp-service && pytest tests/
```

## AI agents and project context

For AI agent rules, architecture overview and project context, see:
- `.windsurfrules` — single source of truth for Windsurf/Cascade.
- `docs/PROJECT_REVIEW_AI_AGENTS.md` — full knowledge-transfer artifact.
- `docs/AGENTS.md` / `docs/AGENTS_EN.md` — agent capability map.


### Логи
```bash
# Backend
docker compose logs -f backend
docker compose logs backend | grep ERROR

# Frontend
# Логи доступны в браузере консоли
```

### Health checks
```bash
# Backend health endpoint
curl http://localhost:8080/health

# NLP service health
curl http://localhost:8000/health
```

### Database
```bash
# Подключение к PostgreSQL
docker compose exec postgres psql -U kb_user -d knowledge_base

# Резервное копирование
./scripts/devops/backup-personal.sh      # Linux/Mac
./scripts/devops/backup-personal.ps1    # Windows
```

## 🔨 Инфраструктура и CI/CD

### Pre-commit hooks
```bash
npm run prepare                   # Установка husky hooks
```

### Линтинг и форматирование
```bash
# Full stack lint
npm run lint && npm run lint:backend

# Full stack format
npm run format && cd backend && golangci-lint run --fix
```

### Cleanup
```bash
# Очистка логов и старых снапшотов тестов
npm run clean:logs                 # или make clean-logs
npm run clean:logs:dry             # предпросмотр без удаления

# Очистка артефактов тестовых прогонов
npm run clean:test-artifacts
npm run clean:test-artifacts:dry   # предпросмотр без удаления
```

### Docker cleanup и сжатие VHDX WSL2

Скрипт `cleanup-docker` — единственный поддерживаемый способ. Он честно рапортует об отказах: ненулевой код выхода, если хоть один шаг не удался.

**Флаги `cleanup-docker.ps1`** (у `.sh` аналоги: `-f`, `-o`, `--remove-volumes`, `-n|--dry-run`):

- без флагов — остановка контейнеров, dangling-образы, остановленные контейнеры, сети, кэш сборки; **тома не трогаются**;
- `-DryRun` — предпросмотр: печатает, что ушло бы, и ничего не меняет;
- `-Full` — дополнительно `docker system prune -af`: **все** образы без исключения (после шагов 1–3 контейнеров не остаётся, поэтому неиспользуемыми считаются все); тома по-прежнему не трогаются;
- `-RemoveVolumes` — удаляет только анонимные (64-hex) dangling-тома; именованные, `*personal*` и помеченные `protected` сохраняются всегда;
- `-WslOptimize` — сжатие `docker_data.vhdx` через diskpart. **Требует запуск от администратора** — без повышения шаг помечается `[SKIP]` с причиной; процессы WSL/Docker при этом не убиваются: скрипт ждёт освобождения файла и отказывается с сообщением, если он остался заблокированным.

**Защита томов.** Шаги, способные тронуть тома (`-RemoveVolumes`, `-Full`), отказываются без свежего непустого бэкапа Personal-стека (`backups/backup-personal-*`, порог — `scripts/devops/backup-policy.env`). Политика та же, что у хука `guard-personal-data.py`.

**Порядок и периодичность:**

1. Обычный прогон раз в неделю или когда `docker system df` показывает большой reclaimable-кэш сборки:
   ```powershell
   .\scripts\cleanup\cleanup-docker.ps1 -DryRun   # посмотреть цену
   .\scripts\cleanup\cleanup-docker.ps1           # выполнить
   ```
2. Компактизация VHD — редко и только **после** обычной очистки: diskpart возвращает Windows лишь то, что уже освобождено внутри диска. Признак «пора»: файл `docker_data.vhdx` заметно больше суммы из `docker system df`.
   ```powershell
   # из терминала, запущенного от администратора
   .\scripts\cleanup\cleanup-docker.ps1 -WslOptimize
   ```
   Сжимается только диск Docker (`%LOCALAPPDATA%\Docker\wsl`); диски других дистрибутивов WSL не выбираются.

**Чего не делать:**
- не убивать `vmmem*`/`wsl*`/`docker*` через `Stop-Process -Force` — в этой VM лежат тома Personal-стека, форсированное завершение оставляет ФС грязной;
- не запускать `docker system prune --volumes` и `docker volume prune` — `docker system df` помечает тома Personal-стека как «reclaimable», потому что их контейнеры остановлены.

## 📝 Документация

### Генерация документации
```bash
# Swagger UI уже доступен на /swagger
# OpenAPI spec: openAPI.yaml
```

### Работа с документацией
```bash
# Обновление README/docs с помощью knowledge-graph-docs-maintenance агента
```

## 🎯 Быстрые сценарии

### Быстрый старт разработки
```bash
# 1. Установка зависимостей
cd backend && go mod download
cd ../frontend && npm install
cd ../nlp-service && pip install -r requirements.txt

# 2. Запуск стек
docker compose up

# 3. Отдельные сервисы при необходимости
cd backend && go run ./cmd/server
cd frontend && npm run dev
cd nlp-service && uvicorn app.main:app --reload
```

### Полный цикл тестирования
```bash
# Frontend все тесты
cd frontend && npm run test:all

# Backend все тесты
cd backend && go test -tags=integration ./...

# NLP тесты
cd nlp-service && pytest tests/ -v
```

### Подготовка к коммиту
```bash
# Линтинг
npm run lint && npm run lint:backend

# Форматирование
npm run format

# Тесты
npm run test
cd backend && go test ./...
cd ../nlp-service && pytest tests/
```

## 🔍 Поиск проблем

### Проверка зависимостей
```bash
cd backend && go mod verify
cd frontend && npm audit
cd nlp-service && pip check
```

### Проверка конфигурации
```bash
cd backend && go run ./cmd/checkconfig
```

### Database статус
```bash
docker compose ps postgres
docker compose exec postgres pg_isready
```

---

**Примечание:** Все команды предполагают выполнение из корневой директории проекта `d:\knowledge-graph`, если не указано иное с `cd <directory>`.