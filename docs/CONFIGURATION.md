# Configuration Guide

This document describes the configuration system for Knowledge Graph project.

## Overview

The project uses a hierarchical configuration system with the following priority:

```
Environment Variables > knowledge-graph.config.json > Hard-coded Defaults
```

## Configuration Files

### 1. `config/` source files

Editable configuration sections are stored in `config/*.json` files in the project root. These files are merged into the runtime config file `knowledge-graph.config.json`.

Current source files:
- `config/backend.json`
- `config/frontend.json`
- `config/backup.json`
- `config/nlp.json`
- `config/mongodb.json`
- `config/ci_cd.json`

To regenerate the runtime config after edits, run:
```bash
npm run build-config
```

### 2. `knowledge-graph.config.json`

Generated runtime JSON configuration file located in the project root. Used by both backend and frontend.

**Structure:**
```json
{
  "backend": {
    "server": { "rate_limit": {...} },
    "database": {...},
    "search": {...},
    "recommendation": {...},
    "pagination": {...},
    "graph": {...},
    "embedding": {...},
    "asynq": {...}
  },
  "frontend": {
    "test": {...},
    "graph": { "2d": {...}, "3d": {...} },
    "api": {...}
  },
  "ci_cd": {...},
  "nlp": {...},
  "backup": {...}
}
```

### 3. `.env` / `.env.example`

Environment variables for sensitive data and deployment-specific settings.

**Required Variables:**
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `NLP_SERVICE_URL` - NLP service endpoint

**Optional Variables (override JSON config):**
- `SERVER_PORT` - Backend port (default: 8080)
- `SERVER_RATE_LIMIT_ENABLED` - Enable rate limiting
- `SERVER_RATE_LIMIT_REQUESTS` - Rate limit requests
- `RECOMMENDATION_DEPTH` - BFS depth for recommendations
- `RECOMMENDATION_ALPHA` - Weight for link component
- `RECOMMENDATION_BETA` - Weight for semantic component
- `RECOMMENDATION_GAMMA` - Weight for keyword component (default: 0.2)
- `RECOMMENDATION_KEYWORD_SIMILARITY_METHOD` - Keyword similarity method: `jaccard`, `overlap`, `tversky`, `weighted_jaccard`, `cosine` (default: `jaccard`)
- `RECOMMENDATION_KEYWORD_TVERSKY_ALPHA` - Alpha parameter for Tversky index (default: 0.5)
- `RECOMMENDATION_KEYWORD_TVERSKY_BETA` - Beta parameter for Tversky index (default: 0.5)
- And many more...

See `.env.example` for complete list.

### Keyword Similarity Methods

The recommendation system supports multiple similarity measures for keyword-based matching:

| Method | Formula | Use Case |
|--------|---------|----------|
| `jaccard` | \|A ∩ B\| / \|A ∪ B\| | Classic set similarity, ignores weights |
| `overlap` | \|A ∩ B\| / min(\|A\|, \|B\|) | Good for asymmetric similarity (subset detection) |
| `tversky` | \|A ∩ B\| / (\|A ∩ B\| + α\|A\\B\| + β\|B\\A\|) | Tunable with α/β parameters. α=β=1 → Jaccard, α=β=0.5 → Dice |
| `weighted_jaccard` | Σmin(w₁,w₂) / Σmax(w₁,w₂) | Uses keyword weights from NLP extraction |
| `cosine` | (A·B) / (\|A\|\|B\|) | Vector space model similarity |

**Configuration Example:**
```json
{
  "backend": {
    "recommendation": {
      "gamma": 0.2,
      "keyword_similarity_method": "tversky",
      "keyword_tversky_alpha": 0.5,
      "keyword_tversky_beta": 0.5
    }
  }
}
```

Or via environment variables:
```bash
RECOMMENDATION_GAMMA=0.2
RECOMMENDATION_KEYWORD_SIMILARITY_METHOD=tversky
RECOMMENDATION_KEYWORD_TVERSKY_ALPHA=0.5
RECOMMENDATION_KEYWORD_TVERSKY_BETA=0.5
```

## Backend Configuration (Go)

### Loading Order

1. Load `knowledge-graph.config.json` (generated from `config/*.json` when present)
2. Check environment variables (override JSON values)
3. Use hard-coded defaults (if neither set)

If the root config file is absent, backend configuration code can fall back to reading `config/*.json` directly.

### Usage in Code

```go
import "knowledge-graph/internal/config"

cfg := config.Load()

// Access configuration
depth := cfg.RecommendationDepth
ttl := cfg.RecommendationCacheTTL
```

### Adding New Configuration Parameter

1. Add field to `JSONConfig` struct in `backend/internal/config/config.go`:
```go
type JSONConfig struct {
    Backend struct {
        // ... existing fields ...
        NewSection struct {
            NewParam int `json:"new_param"`
        } `json:"new_section"`
    }
}
```

2. Add field to `Config` struct:
```go
type Config struct {
    // ... existing fields ...
    NewParam int
}
```

3. Add loading logic in `Load()`:
```go
cfg.NewParam = getIntEnv("NEW_PARAM", getJSONIntOrDefault(jsonCfg, 
    func(j *JSONConfig) int { return j.Backend.NewSection.NewParam }, 
    42)) // default value
```

4. Add to the appropriate source file under `config/`, for example `config/backend.json`:
```json
{
  "backend": {
    "new_section": {
      "new_param": 42
    }
  }
}
```

5. Regenerate the runtime config file:
```bash
npm run build-config
```

6. Document in `.env.example`:
```bash
# New Section
NEW_PARAM=42
```

## Frontend Configuration (TypeScript)

### Loading

Frontend imports the generated runtime config file at build time:

```typescript
import config from '../../../knowledge-graph.config.json';

export const apiConfig = config.frontend.api;
export const graphConfig = config.frontend.graph;
```

### Limitations

- Frontend only uses a subset of configuration
- Changes require rebuild
- No runtime environment variable support (use build args for different environments)

## Docker Compose Configuration

### Development Stack (docker-compose.yml)

- Port: 3000
- Full stack with all services
- Shared volumes for development

### Personal Instance (docker-compose.personal.yml)

- Port: 3001
- Isolated database and Redis
- Persistent volumes

## Validation & CI/CD

### Local Validation

```bash
# Backend
cd backend
make check-config        # Validate JSON config
cd backend
make check-migrations    # Check migration drift

# Manual checks
# Regenerate config from source files
npm run build-config

# Validate JSON syntax
cat knowledge-graph.config.json | jq . > /dev/null

# Check environment variables
source .env && go run backend/cmd/server/main.go
```

### CI/CD Checks

The following checks run automatically:

1. **Config Validation** (`.github/workflows/ci-config-validation.yml`)
   - JSON syntax validation
   - Config loads successfully
   - All env vars documented in `.env.example`

2. **Migration Drift Check**
   - Applies migrations to test database
   - Verifies GORM AutoMigrate compatibility

3. **Docker Compose Validation**
   - Syntax validation
   - Environment variable consistency

## Troubleshooting

### Config Not Loading

1. Check file exists in project root: `knowledge-graph.config.json`
2. If it is missing, regenerate it from `config/*.json`:
   ```bash
   npm run build-config
   ```
3. Validate JSON syntax: `cat knowledge-graph.config.json | jq .`
4. Check backend logs for "[Config] Loading JSON config from..."

### Environment Variables Not Applied

1. Verify variable name matches exactly (case-sensitive)
2. Check `.env.example` for correct naming
3. Ensure variable is exported: `export VAR_NAME=value`

### Migration Errors

1. Check database connection: `pg_isready -h localhost -p 5432`
2. Verify migrations are applied: `migrate -path backend/migrations -database "$DATABASE_URL" version`
3. Check for drift: `make check-migrations`

## Security Considerations

- Never commit `.env` files with real credentials
- Use `.env.example` as template with dummy values
- Store sensitive data in environment variables only
- Rotate secrets regularly in production

---

## Graph Service (gRPC Microservice)

`graph-service` — отдельный gRPC-микросервис для вычисления раскладок графа и потоковой передачи данных. Имеет собственный раздел конфигурации в `knowledge-graph.config.json`.

### JSON Конфигурация (`graph_service`)

```json
{
  "graph_service": {
    "grpc_port": "9090",
    "http_port": "9091",
    "full_limit": 1000,
    "default_depth": 2,
    "event_channel": "graph:events",
    "cache": {
      "note_layout_ttl_seconds": 300,
      "full_layout_ttl_seconds": 300,
      "delta_ttl_seconds": 60
    },
    "layout": {
      "2d_radius": 100.0,
      "3d_radius": 120.0,
      "3d_z_step": 5.0,
      "default_node_size": 1.0
    },
    "stream_chunk_size": 100,
    "event_tracking_ttl_hours": 24,
    "unprocessed_event_check_interval_minutes": 5
  }
}
```

### Переменные Окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `GRPC_PORT` | Порт gRPC сервера | `9090` |
| `HTTP_PORT` | Порт HTTP fallback сервера | `9091` |
| `POSTGRES_URL` | Строка подключения к PostgreSQL | `postgresql://postgres:postgres@postgres:5432/knowledge_base?sslmode=disable` |
| `REDIS_URL` | Адрес Redis | `redis:6379` |
| `EVENT_CHANNEL` | Канал Redis Pub/Sub | `graph:events` |
| `GRAPH_FULL_LIMIT` | Лимит для полного графа | `1000` |
| `GRAPH_DEFAULT_DEPTH` | Глубина для раскладки заметки | `2` |
| `GRAPH_STREAM_CHUNK_SIZE` | Количество узлов в чанке потока | `100` |
| `CACHE_NOTE_TTL_SECONDS` | TTL кэша раскладки заметки | `300` (5 мин) |
| `CACHE_FULL_TTL_SECONDS` | TTL кэша полного графа | `300` (5 мин) |
| `CACHE_DELTA_TTL_SECONDS` | TTL кэша дельты | `60` (1 мин) |

### Параметры Движка Раскладки

| Параметр | Описание | По умолчанию |
|----------|----------|--------------|
| `2d_radius` | Радиус круга для 2D раскладки | `100.0` |
| `3d_radius` | Радиус спирали для 3D раскладки | `120.0` |
| `3d_z_step` | Вертикальный шаг между узлами в 3D | `5.0` |
| `default_node_size` | Базовый размер узла | `1.0` |

### Поведение Кэширования

- **Note Layout Cache** (`CACHE_NOTE_TTL_SECONDS`): Кэш раскладок по заметкам (по умолчанию: 5 мин)
- **Full Layout Cache** (`CACHE_FULL_TTL_SECONDS`): Кэш полных графов (по умолчанию: 5 мин)
- **Delta Cache** (`CACHE_DELTA_TTL_SECONDS`): Кэш дельт (по умолчанию: 1 мин)

Инвалидация кэша происходит автоматически через Redis Pub/Sub при обновлении заметок или ссылок.

### Параметры Отслеживания Событий

| Параметр | Описание | По умолчанию |
|----------|----------|--------------|
| `event_tracking_ttl_hours` | Время хранения записей подтверждения событий | `24` |
| `unprocessed_event_check_interval_minutes` | Интервал работы retry worker | `5` |

### gRPC Эндпоинты

| Эндпоинт | Метод | Описание |
|----------|-------|----------|
| `GetNoteLayout` | Unary | Получить раскладку для конкретной заметки |
| `GetFullLayout` | Server Stream | Поток полного графа чанками |
| `GetDelta` | Unary | Получить изменения с последнего хэша |

### HTTP Fallback Эндпоинты

| Эндпоинт | Метод | Описание |
|----------|-------|----------|
| `/health` | GET | Проверка здоровья |
| `/api/v1/graph/note/:id` | GET | Раскладка заметки (JSON) |
| `/api/v1/graph/full` | GET | Полный граф (JSON) |
| `/api/v1/graph/delta` | GET | Дельта (JSON) |

### Использование в docker-compose.yml

```yaml
graph-service:
  build:
    context: .
    dockerfile: ./services/graph-service/Dockerfile
  environment:
    GRPC_PORT: 9090
    HTTP_PORT: 9091
    POSTGRES_URL: postgresql://user:pass@postgres:5432/knowledge_base
    REDIS_URL: redis:6379
    GRAPH_FULL_LIMIT: 1000
    CACHE_NOTE_TTL_SECONDS: 300
```

### Применение Изменений

#### После изменения переменных в `docker-compose.yml`, перезапустите graph-service:

```bash
docker-compose restart graph-service
```

#### Проверьте логи — должно появиться сообщение о загрузке конфигурации:

```bash
docker logs kg-graph-service --tail 30
```

### Связанная Документация

| Документ | Описание |
|----------|----------|
| [`architecture/decisions/013-graph-service-isolation.md`](architecture/decisions/013-graph-service-isolation.md) | Архитектурное решение по выделению Graph Service |
| [`architecture/decisions/014-event-driven-cache-invalidation.md`](architecture/decisions/014-event-driven-cache-invalidation.md) | Инвалидация кэша через события |
