# Конфигурация Knowledge Graph

Параметры системы можно настроить через:
1. **Файлы `config/*.json`** — редактируемые секции, которые объединяются в `knowledge-graph.config.json`
2. **`knowledge-graph.config.json`** — итоговый файл конфигурации (рекомендуется)
3. **Переменные окружения** — переопределение конкретных значений
4. **Файл `.env`** — для локальной разработки

В продакшне передавайте переменные через `environment` в `docker-compose.yml` или через `ConfigMap`/`Secrets` в Kubernetes.

🌐 **Язык:** Русский | [English version](CONFIGURATION_EN.md)

---

## Единый файл конфигурации

Файл `knowledge-graph.config.json` в корне проекта — это **итоговый артефакт сборки** и единственный источник истины для всех структурных параметров. Он создаётся из редактируемых файлов в `config/*.json` и содержит настройки бэкенда, фронтенда, NLP-сервиса, бэкапов и CI/CD.

Для пересоздания конфига после редактирования исходных файлов выполните:
```bash
npm run build-config
```

### Приоритет (от высшего к низшему)

1. **Переменные окружения** — переопределяют любое значение из JSON
2. **`knowledge-graph.config.json`** — итоговый конфиг, общий для всех компонентов
3. **Жёстко заданные умолчания** — fallback в коде Go/TypeScript

Примечание: редактируйте файлы в `config/*.json` и пересоздавайте `knowledge-graph.config.json` командой `npm run build-config`.

### Структура файла

```json
{
  "backend": {
    "server": { "rate_limit": { ... } },
    "database": { ... },
    "search": { ... },
    "recommendation": { ... },
    "pagination": { ... },
    "graph": { ... },
    "embedding": { ... },
    "asynq": { ... }
  },
  "frontend": {
    "test": { ... },
    "graph": { "2d": { ... }, "3d": { ... } },
    "api": { ... },
    "achievements": { ... }
  },
  "ci_cd": {
    "integration_test": { ... }
  },
  "nlp": { ... },
  "backup": { ... }
}
```

### Использование во фронтенде (TypeScript)

```typescript
import { graphConfig2D, apiConfig, testConfig, ACHIEVEMENT_POLL_INTERVAL_MS } from '$shared/config/config.ts';

// Используем централизованный конфиг
const enableShadows = nodes.length < graphConfig2D.shadows_threshold;
const limit = apiConfig.default_limit;
const pollInterval = ACHIEVEMENT_POLL_INTERVAL_MS;
```

---

## Параметры 2D-графа (`frontend.graph.2d`)

Все параметры читаются через `$shared/config/config.ts.ts → graphConfig2D`. Жёсткое задание этих значений в исходном коде **запрещено**.

```json
{
  "frontend": {
    "graph": {
      "2d": {
        "max_nodes": 500,
        "shadows_threshold": 100,
        "animated_links_threshold": 50,
        "gravity_nodes_threshold": 100,
        "gravity_max_distance": 300,
        "ghost_node_radius": 30
      }
    }
  }
}
```

| Параметр | Тип | Умолчание | Файл | Описание |
|----------|-----|-----------|------|----------|
| `max_nodes` | integer | `500` | `renderer.ts` | Максимальное количество узлов в 2D-графе |
| `shadows_threshold` | integer | `100` | `renderer.ts` | Количество узлов, **ниже** которого рисуются CSS-тени. Выше порога тени отключаются для производительности |
| `animated_links_threshold` | integer | `50` | `renderer.ts` | Количество связей, **выше** которого анимированные линии заменяются статическими. Предотвращает лаги на плотных графах |
| `gravity_nodes_threshold` | integer | `100` | `gravity-system.ts` | Количество узлов, **выше** которого система гравитации отключается полностью. Гравитация O(n²), поэтому пропускается на больших графах |
| `gravity_max_distance` | integer | `300` | `gravity-system.ts` | Максимальный радиус притяжения между узлами в мировых координатах |
| `ghost_node_radius` | integer | `30` | `ghost-node.ts` | Радиус кнопки создания заметки («+») в экранных пикселях (рисуется в экранных координатах, всегда видна) |

### Дерево решений по производительности

```
nodes.length < shadows_threshold (100)       → включить CSS-тени
nodes.length < gravity_nodes_threshold (100) → включить гравитационную симуляцию
links.length > animated_links_threshold (50) → использовать статические линии вместо анимированных
```

### Использование в бэкенде (Go)

```go
import "knowledge-graph/internal/config"

cfg := config.Load()
// cfg.ServerRateLimitEnabled
// cfg.RecommendationDepth
// cfg.GraphDefaultLimit
```

---

## Галактический лексикон и достижения

### Настройки пользователя (база данных)

Система галактического лексикона и достижений использует пользовательские настройки из таблицы `user_settings`:

| Ключ | Тип | Умолчание | Описание |
|------|-----|-----------|----------|
| `galactic_mode` | boolean | `false` | Включить галактический (космический) стиль сообщений |
| `show_achievement_notifications` | boolean | `true` | Показывать уведомления о полученных достижениях |
| `preferred_language` | string | `ru` | Предпочитаемый язык пользователя (ru/en) |

Настройки обновляются через API пользователя и учитываются фронтенд-компонентами.

### Конфигурация фронтенда (`frontend.achievements`)

```json
{
  "frontend": {
    "achievements": {
      "poll_interval_ms": 7000
    }
  }
}
```

| Параметр | Тип | Умолчание | Описание |
|----------|-----|-----------|----------|
| `poll_interval_ms` | integer | `7000` | Интервал опроса новых достижений (миллисекунды) |

### Переопределение через переменные окружения

| Переменная | Описание | Умолчание |
|-----------|----------|-----------|
| `FRONTEND_ACHIEVEMENTS_POLL_INTERVAL_MS` | Интервал опроса достижений | `7000` |

### Типы условий достижений

Достижения используют JSON-условия в поле `condition_json`:

**Условие по количеству (count):**
```json
{
  "type": "count",
  "entity": "note",
  "action": "create",
  "filter": { "type": "star" },
  "threshold": 10
}
```

**Условие по серии дней (streak):**
```json
{
  "type": "streak",
  "days": 7
}
```

### API-эндпоинты

| Эндпоинт | Метод | Описание |
|----------|-------|----------|
| `/api/v1/achievements` | GET | Список всех доступных достижений |
| `/api/v1/users/me/achievements` | GET | Достижения текущего пользователя |
| `/api/v1/users/me/achievements/:id/mark-seen` | POST | Отметить уведомление о достижении как просмотренное |

---

## Конфигурация резервного копирования

### JSON-конфигурация (`backup`)

```json
{
  "backup": {
    "local_path": "./backups",
    "cloud": {
      "enabled": false,
      "provider": "yandex",
      "yandex": {
        "oauth_token": ""
      }
    },
    "schedule": "0 23 * * 0",
    "retention_days": 14,
    "draft_ttl_hours": 168
  }
}
```

### Переопределение через переменные окружения

| Переменная | Описание | Умолчание |
|-----------|----------|-----------|
| `BACKUP_LOCAL_PATH` | Локальная директория для бэкапов | `./backups` |
| `BACKUP_CLOUD_ENABLED` | Включить облачное резервирование | `false` |
| `BACKUP_CLOUD_PROVIDER` | Провайдер облака (`yandex`) | `yandex` |
| `BACKUP_YANDEX_OAUTH_TOKEN` | OAuth-токен Яндекс.Диска | — |
| `BACKUP_YANDEX_FOLDER` | Папка на Яндекс.Диске | `/KnowledgeGraphBackups` |
| `BACKUP_YANDEX_MAX_BACKUPS` | Максимальное количество хранимых бэкапов | `10` |
| `BACKUP_SCHEDULE` | Расписание cron | `0 23 * * 0` |
| `BACKUP_RETENTION_DAYS` | Срок хранения бэкапов (дни) | `14` |
| `BACKUP_DRAFT_TTL_HOURS` | TTL черновиков в MongoDB | `168` |

### Скрипты резервного копирования

**Linux/Mac:** `scripts/utility/backup-personal.sh`
- Выполняет `pg_dump` в `backups/backup-personal-YYYY-MM-DD.sql.gz`
- Загружает на Яндекс.Диск через WebDAV если облачный бэкап включён
- Удаляет старые бэкапы по сроку хранения

**Windows:** `scripts/utility/backup-personal.ps1`
- Аналогичная функциональность для Windows

---

**Сервис резервирования (Docker):** `backup_scheduler` в `docker-compose.personal.yml`
- Автоматически запускает скрипты каждые 24 часа
- Поддерживает локальный и облачный (Яндекс.Диск) бэкап

**Go-сервис:** `backend/internal/infrastructure/cloud/yandex_backup.go`
- `YandexBackupService` для интеграции с Яндекс.Диском через WebDAV
- Методы: `UploadBackup`, `DownloadBackup`, `ListBackups`, `DeleteBackup`
- Логика повтора (3 попытки) при сбоях загрузки
- Автоматическая очистка старых облачных бэкапов (`max_backups`, по умолчанию: 10)

**Подробное руководство по настройке:** [`docs/BACKUP.md`](BACKUP.md)

---

## MongoDB

### Переменные окружения

| Переменная | Описание | Умолчание |
|-----------|----------|-----------|
| `MONGO_URL` | Строка подключения к MongoDB | `mongodb://localhost:27017` |
| `MONGO_DATABASE` | Имя базы данных MongoDB | `knowledge_graph` |

### Использование MongoDB

MongoDB используется для хранения черновиков заметок:
- **Паттерн состояний**: Active → Publishing → Published / Conflict
- **TTL-индекс**: автоматическая очистка просроченных черновиков
- **Синхронизация черновиков**: асинхронная синхронизация с PostgreSQL

---

## Обязательные переменные окружения

Должны быть заданы исключительно через переменные окружения (не в JSON):

| Переменная | Компонент | Описание | Пример |
|-----------|-----------|----------|--------|
| `DATABASE_URL` | backend, worker | Строка подключения к PostgreSQL | `postgresql://kb_user:kb_pass@postgres:5432/knowledge_base?sslmode=disable` |

Все остальные параметры можно настроить через `knowledge-graph.config.json` или переопределить переменными окружения.

---

## Инфраструктурные параметры

| Переменная | Компонент | Описание | Умолчание |
|-----------|-----------|----------|-----------|
| `SERVER_PORT` | backend | Порт HTTP-сервера (Gin) | `8080` |
| `REDIS_URL` | backend, worker | Адрес Redis для очередей asynq и кеша рекомендаций | `localhost:6379` |
| `NLP_SERVICE_URL` | backend, worker | URL Python NLP-сервиса | `http://localhost:5000` |

### Детали компонентов

**SERVER_PORT** — порт, на котором слушает бэкенд:
- Внутри Docker Compose: `8080`
- Снаружи: `8080:8080`

**REDIS_URL** — хранилище для:
- Очередей задач asynq (воркер читает отсюда)
- Кеша рекомендаций (TTL настраивается отдельно)

**NLP_SERVICE_URL** — адрес для HTTP-вызовов:
- `/extract_keywords` — извлечение ключевых слов
- `/embeddings` — генерация векторных эмбеддингов
- `/health` — проверка работоспособности сервиса

---

## Сервер и ограничение запросов (Rate Limiting)

### JSON-конфигурация (`backend.server`)

```json
{
  "backend": {
    "server": {
      "rate_limit": {
        "enabled": true,
        "requests": 100,
        "window_seconds": 60,
        "endpoints": {
          "notes_create": 30,
          "links_create": 50,
          "notes_update": 20
        },
        "fallback_ports": ["8081", "8082"]
      }
    }
  }
}
```

### Переопределение через переменные окружения

| Переменная | Описание | Умолчание |
|-----------|----------|-----------|
| `SERVER_RATE_LIMIT_ENABLED` | Включить rate limiting | `true` |
| `SERVER_RATE_LIMIT_REQUESTS` | Общий лимит запросов | `100` |
| `SERVER_RATE_LIMIT_WINDOW_SECONDS` | Временное окно | `60` |
| `SERVER_PORT` | Порт HTTP-сервера | `8080` |
| `SERVER_FALLBACK_PORTS` | Резервные порты (через запятую) | `8081,8082` |

### Поведение rate limiting

- **Общие запросы**: все GET-запросы и неуказанные эндпоинты
- **Операции записи**: более строгие лимиты для POST/PUT/DELETE
- **По IP**: лимиты применяются на клиентский IP-адрес
- **Ответ при превышении**: `429 Too Many Requests`

---

## Алгоритм рекомендаций

### JSON-конфигурация (`backend.recommendation`)

```json
{
  "backend": {
    "recommendation": {
      "depth": 3,
      "decay": 0.5,
      "top_n": 20,
      "alpha": 0.5,
      "beta": 0.5,
      "gamma": 0.2,
      "cache_ttl_seconds": 300,
      "task_delay_seconds": 5,
      "batch_rate_limit": 10,
      "fallback_enabled": true,
      "fallback_ttl_seconds": 3600,
      "fallback_semantic_enabled": true,
      "keyword_enabled": true,
      "bfs_aggregation": "max",
      "bfs_normalize": true
    }
  }
}
```

### Переопределение через переменные окружения

| Переменная | Описание | Умолчание | Диапазон |
|-----------|----------|-----------|---------|
| `RECOMMENDATION_ALPHA` | Вес явных связей | `0.5` | 0.0 – 1.0 |
| `RECOMMENDATION_BETA` | Вес семантической схожести | `0.5` | 0.0 – 1.0 |
| `RECOMMENDATION_DEPTH` | Глубина обхода BFS | `3` | 1 – 5 |
| `RECOMMENDATION_DECAY` | Затухание веса для косвенных связей | `0.5` | 0.0 – 1.0 |
| `RECOMMENDATION_CACHE_TTL_SECONDS` | TTL кеша | `300` | 60 – 3600 |
| `EMBEDDING_SIMILARITY_LIMIT` | Лимит кандидатов pgvector | `30` | 10 – 100 |
| `RECOMMENDATION_FALLBACK_SEMANTIC_ENABLED` | Включить семантический fallback | `true` | — |
| `RECOMMENDATION_KEYWORD_ENABLED` | Включить ключевую составляющую (gamma) | `true` | — |

### Подробное описание

**ALPHA + BETA** — формула комбинирования:
```
score = α × explicit_score + β × semantic_score
```
- `α + β` не обязаны равняться 1, но рекомендуется для единой шкалы
- `α = 1.0, β = 0.0` — только явные связи (семантика игнорируется)
- `α = 0.0, β = 1.0` — только семантика (связи игнорируются)
- `α = 0.5, β = 0.5` — баланс: равный вес связей и семантики (по умолчанию)

**DEPTH** — глубина BFS:
- `1` — только прямые связи (быстро)
- `3` — связи до 3-го уровня (оптимально)
- `5` — глубокий поиск (медленно, много данных)

**DECAY** — затухание веса:
- Применяется к связям начиная со 2-го уровня
- Формула: `weight × decay^(depth-1)`
- `0.5` означает: 2-й уровень = 50%, 3-й = 25%

**CACHE_TTL** — время кеширования:
- Рекомендации кешируются в Redis
- Ключ кеша: `suggestions:{note_id}:{limit}`

**EMBEDDING_SIMILARITY_LIMIT** — лимит семантических кандидатов:
- Сколько заметок извлекать из pgvector по схожести эмбеддингов
- Выше = точнее, но медленнее

**RECOMMENDATION_FALLBACK_ENABLED** — переключатель синхронного fallback:
- `false` (по умолчанию): использовать только предвычисленные рекомендации из таблицы `note_recommendations`. Самый быстрый вариант, но новые заметки временно могут не получать рекомендации.
- `true`: включить синхронный fallback через pgvector и Redis. Медленнее, но всегда возвращает результат.

При `false`:
- API читает только из `note_recommendations` (одиночный индексированный SELECT)
- Без запросов pgvector в пути запроса
- Без обращений к Redis за рекомендациями
- Новые заметки возвращают пустой список до завершения воркера

Рекомендуется: установите `false` в продакшне после проверки надёжности воркера. Используйте `true` только как аварийный откат при проблемах с очередью.

См. также: [RECOMMENDATION_ARCHITECTURE.md](RECOMMENDATION_ARCHITECTURE.md)

---

## Граф: визуализация и API

### JSON-конфигурация (`backend.graph`, `frontend.graph`)

```json
{
  "backend": {
    "graph": {
      "load_depth": 2,
      "max_nodes": 500,
      "default_limit": 100,
      "max_limit": 1000,
      "link_default_limit": 500,
      "link_max_limit": 5000
    }
  },
  "frontend": {
    "graph": {
      "2d": {
        "max_nodes": 500,
        "shadows_threshold": 100,
        "animated_links_threshold": 50,
        "gravity_nodes_threshold": 100,
        "gravity_max_distance": 300,
        "ghost_node_radius": 30
      },
      "3d": { "max_nodes": 500 }
    }
  }
}
```

### Переопределение через переменные окружения

| Переменная | Описание | Умолчание |
|-----------|----------|-----------|
| `GRAPH_LOAD_DEPTH` | Глубина загрузки графа | `2` |
| `GRAPH_MAX_NODES` | Максимальное количество узлов | `500` |
| `GRAPH_DEFAULT_LIMIT` | Лимит узлов по умолчанию | `100` |
| `GRAPH_MAX_LIMIT` | Максимальный лимит узлов | `1000` |
| `GRAPH_LINK_DEFAULT_LIMIT` | Лимит связей по умолчанию | `500` |
| `GRAPH_LINK_MAX_LIMIT` | Максимальный лимит связей | `5000` |

---

## Аутентификация и безопасность

### JSON-конфигурация (`backend.auth`)

```json
{
  "backend": {
    "auth": {
      "jwt_secret": "change-me-in-production",
      "jwt_access_ttl_seconds": 900,
      "jwt_refresh_ttl_seconds": 604800,
      "argon2_time": 3,
      "argon2_memory": 65536,
      "argon2_threads": 4,
      "api_key_enabled": true,
      "static_api_key": "",
      "skip_auth": false,
      "yandex_client_id": "",
      "yandex_client_secret": "",
      "pkce_enabled": true,
      "pkce_code_challenge_length": 128,
      "smtp_host": "",
      "smtp_port": 587,
      "smtp_user": "",
      "smtp_password": "",
      "smtp_from": "noreply@example.com",
      "password_reset_ttl_seconds": 900,
      "password_policy_min_length": 10,
      "password_policy_require_upper": true,
      "password_policy_require_lower": true,
      "password_policy_require_digit": true,
      "password_policy_require_special": true
    }
  }
}
```

### Переопределение через переменные окружения

| Переменная | Описание | Умолчание |
|-----------|----------|-----------|
| `JWT_SECRET` | Секрет для подписи JWT-токенов (**обязательно менять в продакшне**) | `change-me-in-production` |
| `JWT_ACCESS_TTL_SECONDS` | TTL access-токена | `900` (15 мин) |
| `JWT_REFRESH_TTL_SECONDS` | TTL refresh-токена | `604800` (7 дней) |
| `AUTH_SKIP` | Отключить аутентификацию (только для разработки) | `false` |
| `API_KEY_ENABLED` | Включить аутентификацию по API-ключу | `true` |
| `STATIC_API_KEY` | Статический API-ключ (если пустой — генерируется) | — |
| `YANDEX_CLIENT_ID` | OAuth client ID для входа через Яндекс | — |
| `YANDEX_CLIENT_SECRET` | OAuth client secret для Яндекс | — |
| `SMTP_HOST` | SMTP-сервер для сброса пароля | — |
| `SMTP_PORT` | SMTP-порт | `587` |
| `SMTP_USER` | Логин SMTP | — |
| `SMTP_PASSWORD` | Пароль SMTP | — |

### Политика паролей

| Параметр | Умолчание | Описание |
|----------|-----------|----------|
| `password_policy_min_length` | `10` | Минимальная длина пароля |
| `password_policy_require_upper` | `true` | Требовать заглавные буквы |
| `password_policy_require_lower` | `true` | Требовать строчные буквы |
| `password_policy_require_digit` | `true` | Требовать цифры |
| `password_policy_require_special` | `true` | Требовать специальные символы |

> ⚠️ **Никогда не коммитьте секреты** в репозиторий. Всегда используйте переменные окружения или `.env`-файл, исключённый из `.gitignore`.

---

## NLP-сервис

### JSON-конфигурация (`nlp`)

```json
{
  "nlp": {
    "model_name": "all-MiniLM-L6-V2",
    "max_text_length": 10000,
    "hf_home": "/root/.cache/huggingface",
    "hf_hub_disable_telemetry": true,
    "hf_hub_offline": true
  }
}
```

### Переопределение через переменные окружения

| Переменная | Описание | Умолчание |
|-----------|----------|-----------|
| `NLP_MODEL_NAME` | Название модели HuggingFace | `all-MiniLM-L6-V2` |
| `NLP_MAX_TEXT_LENGTH` | Максимальная длина текста для обработки | `10000` |
| `HF_HOME` | Путь к локальному кешу HuggingFace | `/root/.cache/huggingface` |
| `HF_HUB_OFFLINE` | Работа в оффлайн-режиме (без интернета) | `true` |
| `HF_HUB_DISABLE_TELEMETRY` | Отключить телеметрию HuggingFace | `true` |

### Lazy-loading модели (обязательный паттерн)

```python
# nlp_utils.py
_model = None

def get_embedding_model():
    global _model
    if _model is None:
        _model = SentenceTransformer("model-name")
    return _model
```

- Uvicorn стартует за ~1 секунду
- Модель загружается за ~15 секунд из локального кеша
- Первый запрос к `/embed` или `/extract_keywords` инициирует загрузку

---

## Graph Service

### JSON-конфигурация (`graph_service`)

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

| Параметр | Описание |
|----------|----------|
| `grpc_port` | Порт gRPC-сервера |
| `http_port` | Порт HTTP-сервера (health, API) |
| `full_limit` | Максимальное количество узлов в полном графе |
| `default_depth` | Глубина загрузки графа по умолчанию |
| `event_channel` | Redis pub/sub канал для событий графа |
| `cache.note_layout_ttl_seconds` | TTL кеша layout для отдельной заметки |
| `cache.full_layout_ttl_seconds` | TTL кеша полного layout графа |
| `cache.delta_ttl_seconds` | TTL дельта-обновлений |
| `layout.2d_radius` | Начальный радиус размещения узлов в 2D |
| `layout.3d_radius` | Начальный радиус размещения узлов в 3D |
| `stream_chunk_size` | Размер чанка при потоковой передаче узлов |
| `event_tracking_ttl_hours` | TTL отслеживания обработанных событий |
| `unprocessed_event_check_interval_minutes` | Интервал проверки необработанных событий |

---

## Аномалии 2D-графа (`frontend.graph.anomaly`)

Аномалии — особые типы узлов с уникальными визуальными эффектами.

```json
{
  "frontend": {
    "graph": {
      "anomaly": {
        "reality_rift": {
          "core_color": "#0a0a0f",
          "glow_color": "#8b5cf6",
          "crack_count_min": 5,
          "crack_count_max": 8,
          "deform_amount_min": 0.2,
          "deform_amount_max": 0.5
        },
        "chromatic_maw": {
          "tentacle_count_min": 6,
          "tentacle_count_max": 10,
          "hue_shift_base": 280,
          "hue_shift_range": 60
        },
        "void_whisper": {
          "particle_count_min": 20,
          "particle_count_max": 30,
          "hue_shift_base": 220,
          "hue_shift_range": 40,
          "connection_distance_threshold": 0.4
        },
        "cosmic_abomination": {
          "particle_count_min": 12,
          "particle_count_max": 15,
          "tentacle_count_min": 3,
          "tentacle_count_max": 4,
          "crack_count_min": 2,
          "crack_count_max": 3
        }
      }
    }
  }
}
```

| Аномалия | Параметр | Описание |
|----------|----------|----------|
| `reality_rift` | `core_color` | Цвет ядра разлома реальности |
| `reality_rift` | `glow_color` | Цвет свечения |
| `reality_rift` | `crack_count_min/max` | Диапазон количества трещин |
| `reality_rift` | `deform_amount_min/max` | Степень деформации формы |
| `chromatic_maw` | `tentacle_count_min/max` | Диапазон количества щупалец |
| `chromatic_maw` | `hue_shift_base/range` | Базовый сдвиг и разброс оттенка |
| `void_whisper` | `particle_count_min/max` | Диапазон количества частиц |
| `void_whisper` | `connection_distance_threshold` | Порог расстояния для соединений частиц |
| `cosmic_abomination` | все выше | Комбинация частиц, щупалец и трещин |

---

## База данных

### Переменные окружения PostgreSQL

| Переменная | Описание | Умолчание |
|-----------|----------|-----------|
| `DATABASE_URL` | Полная строка подключения | — (обязательно) |
| `DB_RETRY_MAX_ATTEMPTS` | Максимальное количество попыток подключения | `3` |
| `DB_RETRY_DELAY_SECONDS` | Задержка между попытками | `5` |
| `DB_MIGRATIONS_FAIL_ON_ERROR` | Падать при ошибке миграции | `false` |

### JSON-конфигурация (`backend.database`)

```json
{
  "backend": {
    "database": {
      "retry_max_attempts": 3,
      "retry_delay_seconds": 5,
      "migrations_fail_on_error": false
    }
  }
}
```

---

## Полнотекстовый поиск

### JSON-конфигурация (`backend.search`)

```json
{
  "backend": {
    "search": {
      "fulltext_languages": ["russian", "simple"],
      "ranking_weights": {
        "russian": 1,
        "simple": 1
      },
      "fallback_to_ilike": true
    }
  }
}
```

| Параметр | Описание |
|----------|----------|
| `fulltext_languages` | Языки для полнотекстового индекса PostgreSQL |
| `ranking_weights` | Веса конфигураций при ранжировании |
| `fallback_to_ilike` | При нулевых результатах FTS перейти на `ILIKE` |

---

## Асинхронные задачи (asynq)

### JSON-конфигурация (`backend.asynq`)

```json
{
  "backend": {
    "asynq": {
      "concurrency": 10,
      "queue_default": 1,
      "queue_max_len": 10000
    }
  }
}
```

| Параметр | Описание |
|----------|----------|
| `concurrency` | Количество параллельных воркеров обработки задач |
| `queue_default` | Приоритет очереди по умолчанию |
| `queue_max_len` | Максимальный размер очереди |

---

## Параметры портов по окружениям

| Окружение | Компонент | Хост:Порт |
|-----------|-----------|-----------|
| Dev | Nginx gateway | `8080` |
| Dev | Backend | `9000` (внутри контейнера) |
| Dev | Graph Service HTTP | `9091` |
| Personal | Backend | `8085` |
| Personal | API gateway | `8082` |
| Personal | Graph Service | `8092` |
| Test | Frontend | `3002` |
| Test | Backend | `8083` |
| Test | Graph Service | `9095` |
| Test | PostgreSQL | `5434` |
| Test | Redis | `6381` |

---

## CI/CD конфигурация

### JSON-конфигурация (`ci_cd`)

```json
{
  "ci_cd": {
    "integration_test": {
      "migrate_all": true,
      "truncate_list": ["notes", "links", "embeddings", "recommendations"]
    }
  }
}
```

| Параметр | Описание |
|----------|----------|
| `migrate_all` | Применять все миграции перед интеграционными тестами |
| `truncate_list` | Таблицы, которые очищаются перед каждым тестовым запуском |

---

## Ссылки

- [CONFIGURATION_EN.md](CONFIGURATION_EN.md) — Английская версия
- [TESTING_RU.md](TESTING_RU.md) — Тестирование
- [DEPLOYMENT_EN.md](DEPLOYMENT_EN.md) — Деплой
- [BACKUP.md](BACKUP.md) — Резервное копирование
- [RECOMMENDATION_ARCHITECTURE.md](RECOMMENDATION_ARCHITECTURE.md) — Архитектура рекомендаций
- [ANOMALY_TYPES.md](ANOMALY_TYPES.md) — Типы аномалий графа
