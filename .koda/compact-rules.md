# Knowledge Graph Project - Compact AI Rules

## Проект
Knowledge Graph - система управления заметками с графовыми связями и NLP-анализом.

## Стек
- **Backend:** Go 1.23+, PostgreSQL, Redis, gRPC
- **Frontend:** Svelte 5, Vitest, Playwright
- **NLP:** Python FastAPI, sentence-transformers, HuggingFace
- **Infrastructure:** Docker, Docker Compose

## Ключевые правила для AI

### 1. Тесты обязательны
- Всегда запускайте `go test ./...` после изменений в backend
- Фронтенд: `npm run test:unit` для unit тестов
- E2E: `npm run test` для Playwright
- Coverage должен быть >60%

### 2. Конвенции Go
- Clean Architecture: domain/application/infrastructure/interfaces слои
- Private fields в entities + factory functions
- Repository Pattern для доступа к данным
- DDD: Entities, Value Objects, Aggregates
- Не использовать глобальные переменные - явная передача зависимостей

### 3. Конвенции Frontend (Svelte 5)
- Компоненты: атомы → молекулы → организмы
- Сtores для состояния, stores/lib/services для бизнес-логики
- TypeScript для всех типов
- SvelteKit для роутинга

### 4. Docker & Infrastructure
- Multi-stage builds для оптимизации
- Volumes для персистентности данных (postgres_data, redis_data, huggingface_cache)
- Health checks для всех сервисов
- Graceful shutdown

### 5. Безопасность
- Никогда не коммитить секреты (.env, токены)
- Использовать переменные окружения
- Rate limiting для write operations
- CORS настроен правильно

### 6. NLP Service
- Модели сохраняются в huggingface_cache volume
- Используйте HF_HOME для кэша HuggingFace
- Sentiment-transformers кэшируются в /root/.cache/huggingface/hub/

### 7. Конфигурация
- knowledge-graph.config.json - основная конфигурация
- .env - секреты (не коммитить)
- ENV переменные переопределяют JSON config

## Частые команды

### Backend
```bash
cd backend
go test ./... -v                    # Unit тесты
go build ./cmd/server               # Сборка сервера
go run ./cmd/server                 # Запуск сервера
```

### Frontend
```bash
cd frontend
npm run test:unit                   # Unit тесты
npm run test                         # E2E тесты
npm run build                         # Сборка
```

### Docker
```bash
docker compose up -d                 # Запуск всех сервисов
docker compose -f docker-compose.personal.yml up -d  # Personal instance
docker compose logs -f backend         # Логи backend
```

## Структура проекта
```
backend/
├── cmd/                 # Entry points
├── internal/
│   ├── domain/         # Business logic (entities, value objects)
│   ├── application/    # Use cases
│   ├── infrastructure/  # DB, Redis, external services
│   └── interfaces/     # HTTP handlers, middleware
frontend/
├── src/
│   ├── lib/            # Business logic
│   ├── components/     # UI components
│   └── routes/          # Pages
nlp-service/            # Python FastAPI NLP
services/
docker-compose.yml      # Основной стек
docker-compose.personal.yml  # Personal instance
```

## AI-специфичные подсказки

Для backend задач:
- Используйте Clean Architecture
- Репозитории возвращают domain entities
- Use case слои в application
- Не смешивайте слои

Для frontend задач:
- Компоненты должны быть реактивными
- Stores только для cross-component state
- Компоненты должны получать данные через props/stores

Для инфраструктуры:
- Docker multi-stage builds для оптимизации
- Health checks обязательны
- Используйте volumes для персистентности