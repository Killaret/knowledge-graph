# Команды проекта Knowledge Graph

Полный справочник команд для быстрого доступа ко всем операциям проекта.

## Структура скриптов

Скрипты проекта организованы по семантическим категориям в директории `scripts/`:

```
scripts/
├── cleanup/          # Скрипты очистки и сжатия
├── diagnostics/      # Диагностические и проверочные скрипты
├── testing/          # Тестовые скрипты и эксперименты
├── utility/          # Вспомогательные скрипты
├── database/         # Скрипты для работы с базой данных
└── docs/             # Документация скриптов
```

**Cleanup скрипты:**
- `diskpart_compress_admin.ps1` - VHDX сжатие через DiskPart (основной)
- `clean_and_compress_lunix.ps1` - Сжатие lunix образов
- `cleanup_and_compress.ps1` - Полная очистка и сжатие
- `cleanup-docker.ps1` - Очистка Docker + опциональное сжатие VHD через diskpart (без Hyper-V)

**Diagnostics скрипты:**
- `check_all_vhdx.ps1` - Проверка размеров VHDX файлов
- `check_disk_lock.ps1` - Проверка блокировок диска
- `check_file_lock.ps1` - Проверка блокировки файла
- `find_docker.ps1` - Поиск Docker процессов

**Utility скрипты:**
- `stop_docker.ps1` - Остановка Docker
- `force_stop_docker.ps1` - Принудительная остановка Docker
- `fix_vhdx_attributes.ps1` - Исправление атрибутов VHDX

## 🚀 Команды запуска и разработки

### Основные команды (корень проекта)
```bash
npm run prepare                    # Установка husky git hooks
npm run lint                       # Линтинг frontend кода
npm run lint:backend               # Линтинг backend Go кода
npm run format                     # Форматирование frontend кода
npm run build-config               # Сборка конфигурации
npm run test                       # Запуск unit тестов
npm run clean:lunix                # Очистка и компрессия lunix (PowerShell)
npm run clean:lunix:dry            # Очистка lunix (dry run, PowerShell)
npm run clean:lunix:sh             # Очистка и компрессия lunix (bash)
npm run clean:lunix:sh:dry         # Очистка lunix (dry run, bash)
```

### Makefile команды
```bash
make clean-lunix                   # Очистка и компрессия lunix (PowerShell с Compact.exe)
make clean-lunix-sh                # Очистка и компрессия lunix (bash)
make clean-lunix-dry               # Очистка lunix (dry run, PowerShell)
make clean-lunix-sh-dry            # Очистка lunix (dry run, bash)
```

### Docker cleanup и оптимизация диска
```bash
npm run clean:lunix                # Поиск и сжатие всех lunix образов (Compact.exe)
npm run clean:lunix:vhd            # Поиск и сжатие с Optimize-VHD (требует Hyper-V)
npm run clean:lunix:dry            # Dry run для проверки что будет сжато
npm run clean:lunix:sh             # Bash версия для Linux/WSL
npm run clean:lunix:sh:dry         # Bash dry run
```

## 🎨 Frontend команды

### Запуск и сборка
```bash
cd frontend
npm run dev                        # Запуск dev сервера (Vite)
npm run build                      # Production сборка
npm run preview                    # Предпросмотр production сборки
```

### Проверка кода
```bash
cd frontend
npm run check                      # Проверка типов SvelteKit
npm run check:watch                # Проверка типов в watch режиме
npm run lint                       # ESLint linting с авто-фиксом
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
# Очистка временных файлов и docker образов
npm run clean:lunix                # PowerShell с Compact.exe (рекомендуется)
npm run clean:lunix:vhd            # PowerShell с Optimize-VHD (Hyper-V)
npm run clean:lunix:sh            # Bash версия
make clean-lunix                  # Через Makefile
```

### Docker cleanup оптимизация
**Исправленные функции:**
- 🔍 **Поиск нескольких файлов:** Теперь находит все lunix файлы в разных директориях
- 📁 **Поддержка директорий:** Можно указать путь к директории для поиска всех lunix файлов
- 💾 **Compact.exe:** Использует встроенную утилиту Windows для сжатия (без Hyper-V)
- 🗜️ **Optimize-VHD:** Опциональное сжатие VHD (требует Hyper-V)
- ⚡ **Sparse файлы:** Автоматическое включение sparse атрибута для оптимизации диска
- 🔄 **Множественные файлы:** Обрабатывает все найденные файлы, а не только первый
- 💿 **DiskPart VHDX:** Сжатие VHDX файлов Docker WSL2 через diskpart vdisk

**Примеры:**
```bash
# Поиск и сжатие всех lunix образов с Compact.exe
npm run clean:lunix

# Сжатие с Optimize-VHD (если есть Hyper-V)
npm run clean:lunix:vhd

# Проверка что будет сжато без выполнения
npm run clean:lunix:dry

# Сжатие всех файлов в конкретной директории
.\scripts\cleanup\clean_and_compress_lunix.ps1 -ImagePath "D:\images\" -Compress -Force

# Принудительное сжатие без запросов подтверждения
.\scripts\cleanup\clean_and_compress_lunix.ps1 -Search -Compress -UseCompact -Force
```

### Сжатие VHDX Docker WSL2 (рекомендуемый метод)
Для максимальной экономии дискового пространства используйте **DiskPart VHDX сжатие**:

**🎉 Результаты:** Сжатие docker_data.vhdx с 38.98GB до 7.80GB (**31.18GB экономия, 80% reduction**)

**Автоматический скрипт с разблокировкой:**
```bash
# С автоматической остановкой WSL и разблокировкой файла (требует админа)
.\scripts\cleanup\diskpart_compress_admin.ps1

# Или через единый cleanup-скрипт (очистка Docker + сжатие VHD в одном):
.\scripts\cleanup\cleanup-docker.ps1 -Full -WslOptimize
```

**Или через npm:**
```bash
npm run clean:docker:vhdx
```

**Ручной метод (максимальный контроль):**
```powershell
# 1. Остановка WSL
wsl --shutdown

# 2. Принудительная остановка WSL процессов
Get-Process | Where-Object { $_.ProcessName -like "*wsl*" -or $_.ProcessName -like "*vmmem*" } | Stop-Process -Force

# 3. Проверка разблокировки файла
.\scripts\diagnostics\check_file_lock.ps1

# 4. Сжатие VHDX через diskpart (run as admin)
diskpart
# В diskpart:
select vdisk file="C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx"
attach vdisk readonly
compact vdisk
detach vdisk
exit

# 5. Перезапуск WSL
wsl
```

**Оптимальный workflow для максимального эффекта:**
```bash
# Единая команда (требует права админа, Hyper-V НЕ нужен):
.\scripts\cleanup\cleanup-docker.ps1 -Full -WslOptimize
#   - останавливает контейнеры
#   - очищает dangling images, build cache, unused volumes
#   - удаляет ВСЕ неиспользуемые образы (system prune -af --volumes)
#   - останавливает WSL и убивает leftover процессы
#   - находит самый большой .vhdx и сжимает через diskpart
#   - перезапускает WSL

# Или по шагам:
# 1. Запустить Docker (если не запущен)
# 2. Очистка Docker (без volumes)
docker system prune -a --force
# 3. Остановить контейнеры
docker compose down
# 4. Полностью остановить Docker Desktop
# 5. Запустить сжатие
.\scripts\cleanup\diskpart_compress_admin.ps1
# 6. Перезапустить Docker Desktop
```

**Проверка результатов и диагностика:**
```powershell
# Проверка размеров VHDX файлов
.\scripts\diagnostics\check_all_vhdx.ps1

# Проверка что блокирует файл
.\scripts\diagnostics\check_disk_lock.ps1

# Простая проверка блокировки
.\scripts\diagnostics\check_file_lock.ps1
```

**Примечание:**
- Автоматический скрипт обрабатывает остановку WSL и разблокировку файла
- VHDX файл блокируется процессами WSL (vmmemWSL, wsl, wslhost)
- Для максимального эффекта сначала очистите Docker образы
- Требует прав администратора для работы с diskpart

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