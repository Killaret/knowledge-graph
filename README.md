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
- [API Errors](docs/API_ERRORS.md) — формат ошибок и коды
- [Тесты](TEST_STATUS.md) — статус и покрытие
- [Конфигурация](docs/CONFIGURATION.md) — полное руководство по настройке
- [Конфигурация системы](docs/CONFIGURATION_EN.md) — технические параметры
- [Развертывание](docs/DEPLOYMENT_EN.md) — руководство по развертыванию

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
