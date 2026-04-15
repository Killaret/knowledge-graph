# Changelog - Knowledge Graph

> Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.0.0/)
> 
> Версионирование следует [Semantic Versioning](https://semver.org/lang/ru/spec/v2.0.0.html)

---

## [Unreleased]

### Added
- **Config**: 5 новых параметров для рекомендаций и Asynq:
  - `RECOMMENDATION_GAMMA` — дополнительный коэффициент (default: 0.2)
  - `BFS_AGGREGATION` — метод агрегации BFS: max/sum/avg (default: max)
  - `BFS_NORMALIZE` — нормализация весов в BFS (default: true)
  - `ASYNQ_CONCURRENCY` — уровень параллелизма воркера (default: 10)
  - `ASYNQ_QUEUE_DEFAULT` — приоритет очереди (default: 1)
- Планируется: Graph export (изображение/видео)
- Планируется: Node grouping в визуализации
- Планируется: Keyboard shortcuts для графа

---

## [1.0.0] - 2026-04-15

### ✨ Highlights

**3D Progressive Rendering с Fog of War**
- Инкрементальная загрузка узлов с анимацией
- Three.js модульная архитектура (core/simulation/rendering/camera)
- Плавные переходы камеры (lerpCamera)
- Авто-зум при загрузке графа

**Полный стек тестирования**
- Backend: 25+ unit тестов (~1100 строк Go)
- Frontend: 48 E2E тестов Playwright
- 3D Graph: тесты WebGL, fog animation, progressive loading
- Cucumber BDD: 5 feature файлов

### 🚀 Added

#### Backend
- **Domain Layer**: Entities (Note, Link), Value Objects (Title, Content), Graph Traversal с MAX стратегией
- **Application Layer**: Composite Neighbor Loader с взвешенным объединением (α=0.7, β=0.3)
- **Infrastructure Layer**: PostgreSQL репозитории, Redis кэширование рекомендаций
- **Interface Layer**: HTTP handlers (notes, links, graph) с Gin
- **Worker**: asynq обработчик задач для NLP эмбеддингов
- **Graph Algorithms**: BFS обход (depth=3, decay=0.5), семантический поиск через pgvector

#### Frontend
- **Note Management**: CRUD операции, поиск, типы заметок (star, planet, comet, galaxy)
- **2D Graph**: D3-force визуализация, Canvas rendering
- **3D Graph**: Three.js + d3-force-3d, progressive rendering, fog of war
- **Progressive Loading**: Партии по 5 узлов, анимация opacity, auto-zoom
- **API Client**: ky-based HTTP клиент с прокси через Vite

#### NLP Service
- Keyword extraction с использованием KeyBERT
- Text embeddings через sentence-transformers (all-MiniLM-L6-v2, 384 dim)
- FastAPI endpoints: `/extract_keywords`, `/embeddings`, `/health`
- Кэширование моделей в `/app/cache`

#### Infrastructure
- PostgreSQL 16 + pgvector расширение
- Redis 7 (asynq очереди + кэш)
- Docker Compose для локальной разработки
- Kubernetes манифесты (production-ready)

#### Documentation
- **Architecture**: C4 Model, UML диаграммы, 14 ADR
- **Configuration**: Полное описание env переменных
- **API**: OpenAPI 3.1 спецификация
- **Frontend Arch**: Three.js модули, Progressive Rendering
- **Deployment**: Полное руководство по развёртыванию
- **Testing**: Стратегия, структура тестов, запуск
- **API Errors**: Справочник ошибок

### 🔧 Changed

- **Note Type Handling**: Исправлена передача `type` в `CreateNoteModal` (API_VERIFICATION_REPORT fix #1)
- **Link Schema**: Поле `description` → `link_type` + добавлены timestamps
- **Routing**: Note-centric navigation вместо graph-first

### ⚡ Performance

- **Recommendation Caching**: Redis TTL 300 секунд для рекомендаций
- **Progressive Graph Loading**: Узлы загружаются партиями, не блокируя UI
- **WebGL Optimization**: Device capability detection, performance mode toggle
- **Database**: IVFFlat индексы для pgvector similarity search

### 🐛 Fixed

- **OpenAPI**: Исправлены несоответствия с реализацией:
  - Link schema (добавлен `link_type`, `created_at`, `updated_at`)
  - Error schema (поля `error`, `message`, `detail`, `code`)
  - Добавлены endpoints: `/notes/search`, `/graph/all` с параметром `limit`
  - Добавлены теги с описаниями
- **Frontend Arch**: Обновлена документация для 3D модулей
- **ADR-014**: Добавлено описание Progressive Rendering решения

### 🧪 Testing

| Компонент | Покрытие | Детали |
|-----------|----------|--------|
| Backend Domain | 85% | 25+ тестов |
| Backend Application | 75% | Composite loader |
| Backend Interface | 70% | HTTP handlers |
| Frontend E2E | 48 тестов | Playwright |
| 3D Graph | Полное | WebGL, fog, animation |

### 📦 Dependencies

```go
// Backend
github.com/gin-gonic/gin v1.9.1
github.com/hibiken/asynq v0.24.0
gorm.io/gorm v1.25.4
github.com/pgvector/pgvector-go v0.1.1

// Frontend
d3-force-3d v3.0.2
three v0.158.0
ky v1.0.0

// NLP
fastapi v0.104.1
sentence-transformers v2.2.2
keybert v0.8.3
```

---

## [0.9.0] - 2026-03-15

### 🚀 Added

- Базовая архитектура DDD (Domain, Application, Infrastructure, Interface)
- CQRS pattern с in-memory Command/Query Bus
- PostgreSQL схема: notes, links, note_embeddings таблицы
- Базовые CRUD операции для заметок
- Базовая связь между заметками

### 🧪 Testing

- Domain unit tests (Note, Link entities)
- Repository integration tests

---

## [0.5.0] - 2026-02-01

### 🚀 Added

- Проект инициализирован
- Базовая структура backend (Go + Gin)
- Frontend skeleton (Svelte 5 + TypeScript)
- Docker Compose setup
- MVP спецификация

---

## История версий

```
1.0.0 - 2026-04-15 (текущая stable)
  ├─ 3D Progressive Rendering
  ├─ Полный стек тестирования
  └─ Production-ready документация

0.9.0 - 2026-03-15 (MVP завершён)
  ├─ DDD архитектура
  ├─ Базовые CRUD операции
  └─ Интеграция NLP

0.5.0 - 2026-02-01 (инициализация)
  └─ Проект структура
```

---

## Сравнение с планом

### ✅ Выполнено из Roadmap

| Функция | План | Факт | Статус |
|---------|------|------|--------|
| DDD архитектура | v1.0 | v1.0 | ✅ |
| BFS рекомендации | v1.0 | v1.0 | ✅ |
| Эмбеддинги (pgvector) | v1.0 | v1.0 | ✅ |
| 3D визуализация | v1.0 | v1.0 | ✅ |
| Progressive Rendering | v1.1 | v1.0 | ✅ Ранний релиз |
| Полный стек тестов | v1.0 | v1.0 | ✅ |

### ⏳ Запланировано на v1.1

- Graph export (PNG/SVG/MP4)
- Node grouping
- Graph filters (по типу связи, весу)
- Keyboard shortcuts
- Offline support (Service Worker)

### 📅 Запланировано на v1.2

- AI-powered inline suggestions
- Mobile app (Capacitor)
- Plugin system
- Collaborative editing

### 🔮 Запланировано на v2.0

- WebRTC real-time collaboration
- Version history
- Voice input
- Advanced analytics

---

## Обратная совместимость

### Миграция с 0.9.0 → 1.0.0

```bash
# 1. Бэкап БД
docker-compose exec postgres pg_dump -U kg_user knowledge_graph > backup_v0.9.sql

# 2. Обновление миграций
docker-compose run --rm backend migrate up

# 3. Пересчёт эмбеддингов (опционально)
docker-compose run --rm worker go run scripts/reprocess_embeddings.go
```

**Breaking Changes**:
- Link schema: поле `type` переименовано в `link_type`
- API: добавлена пагинация к `/notes`
- Frontend: изменена структура роутинга

---

## Контрибьюторы

- Architecture & Backend: Development Team
- Frontend & 3D Visualization: Frontend Team
- NLP & ML: ML Team
- Documentation: AI Assistant (Cascade)

---

## Ссылки

- [Полный код проекта](../PROJECT_CODE.md)
- [Архитектура](./architecture/README.md)
- [API спецификация](../backend/openAPI.yaml)
- [Руководство по развёртыванию](./DEPLOYMENT.md)
- [Тестирование](./TESTING.md)
