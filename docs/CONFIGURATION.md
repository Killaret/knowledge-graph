# Конфигурация Knowledge Graph

Все параметры системы задаются через переменные окружения. Для локальной разработки их можно определить в файле `.env` в корне проекта. В продакшене переменные передаются через `environment` в `docker-compose.yml` или `ConfigMap`/`Secrets` в Kubernetes.

---

## Обязательные переменные

| Переменная | Компонент | Описание | Пример |
|------------|-----------|----------|--------|
| `DATABASE_URL` | backend, worker | PostgreSQL connection string. Формат: `postgresql://user:password@host:port/dbname?sslmode=disable` | `postgresql://kb_user:kb_pass@postgres:5432/knowledge_base?sslmode=disable` |

**Где используется:**
- `backend` — подключение к БД для API запросов
- `worker` — подключение для сохранения результатов async tasks

---

## Инфраструктурные параметры

| Переменная | Компонент | Описание | По умолчанию |
|------------|-----------|----------|--------------|
| `SERVER_PORT` | backend | Порт HTTP-сервера Gin | `8080` |
| `REDIS_URL` | backend, worker | Арес Redis для asynq очередей и кэша рекомендаций | `localhost:6379` |
| `NLP_SERVICE_URL` | backend, worker | URL Python-сервиса для NLP задач | `http://localhost:5000` |

### Описание компонентов

**SERVER_PORT** — Порт на котором backend слушает HTTP запросы
- В Docker Compose: `8080` (внутри контейнера)
- Пробрасывается наружу: `8080:8080`

**REDIS_URL** — Хранилище для:
- Очереди задач asynq (worker читает отсюда)
- Кэширование рекомендаций (TTL настраивается отдельно)

**NLP_SERVICE_URL** — Адрес для HTTP вызовов:
- `/extract_keywords` — извлечение ключевых слов
- `/embeddings` — генерация векторных представлений
- `/health` — проверка здоровья сервиса

---

## Параметры рекомендаций (граф + эмбеддинги)

| Переменная | Где | Описание | По умолчанию | Диапазон |
|------------|-----|----------|--------------|----------|
| `RECOMMENDATION_ALPHA` | backend | Вес явных связей (BFS) в финальной оценке | `0.5` | 0.0 - 1.0 |
| `RECOMMENDATION_BETA` | backend | Вес семантического сходства (эмбеддинги) | `0.5` | 0.0 - 1.0 |
| `RECOMMENDATION_DEPTH` | backend | Максимальная глубина BFS-обхода графа | `3` | 1 - 5 |
| `RECOMMENDATION_DECAY` | backend | Затухание веса для косвенных связей (depth > 1) | `0.5` | 0.0 - 1.0 |
| `RECOMMENDATION_CACHE_TTL_SECONDS` | backend | TTL кэша рекомендаций в Redis | `300` | 60 - 3600 |
| `EMBEDDING_SIMILARITY_LIMIT` | backend | Лимит кандидатов от pgvector similarity search | `30` | 10 - 100 |

### Детальное описание

**ALPHA + BETA** — Формула комбинирования:
```
score = α × explicit_score + β × semantic_score
```
- `α + β` не обязательно = 1, но рекомендуется для сохранения масштаба
- `α = 1.0, β = 0.0` — только явные связи (игнорировать семантику)
- `α = 0.0, β = 1.0` — только семантика (игнорировать связи)
- `α = 0.7, β = 0.3` — баланс: связи важнее, но семантика учитывается

**DEPTH** — Глубина BFS:
- `1` — только прямые связи (быстро)
- `3` — связи до 3-го уровня (оптимально)
- `5` — глубокий поиск (медленно, много данных)

**DECAY** — Затухание веса:
- Применяется к связям начиная со 2-го уровня
- Формула: `weight × decay^(depth-1)`
- `0.5` означает: 2-й уровень = 50%, 3-й уровень = 25%

**CACHE_TTL** — Время кэширования:
- Рекомендации кэшируются в Redis для скорости
- Ключ кэша: `suggestions:{note_id}:{limit}`

**EMBEDDING_SIMILARITY_LIMIT** — Лимит семантических кандидатов:
- Сколько заметок получить от pgvector по схожести эмбеддингов
- Больше = точнее, но медленнее

---

## Параметры визуализации графа

| Переменная | Где | Описание | По умолчанию | Диапазон |
|------------|-----|----------|--------------|----------|
| `GRAPH_LOAD_DEPTH` | backend, frontend | Глубина загрузки графа для 3D визуализации | `2` | 1 - 4 |

---

## Расширенные параметры рекомендаций (BFS + Asynq)

> ⚠️ Эти параметры объявлены в `config.go`, но **пока не используются** в коде. Они зарезервированы для будущих улучшений алгоритмов.

| Переменная | Где | Описание | По умолчанию | Статус |
|------------|-----|----------|--------------|--------|
| `RECOMMENDATION_GAMMA` | backend | Дополнительный коэффициент для третьего компонента | `0.2` | 🚧 Зарезервировано |
| `BFS_AGGREGATION` | backend | Метод агрегации весов: `max`, `sum`, `avg` | `max` | 🚧 Зарезервировано |
| `BFS_NORMALIZE` | backend | Нормализация весов связей | `true` | 🚧 Зарезервировано |
| `ASYNQ_CONCURRENCY` | worker | Уровень параллелизма Asynq | `10` | 🚧 Зарезервировано |
| `ASYNQ_QUEUE_DEFAULT` | worker | Приоритет дефолтной очереди | `1` | 🚧 Зарезервировано |

### Описание зарезервированных параметров

**RECOMMENDATION_GAMMA** — Коэффициент для третьего компонента:
- Позволяет добавить третий компонент в расчёт (например, временной фактор или популярность)
- `0.2` — небольшой вклад дополнительного фактора

**BFS_AGGREGATION** — Как агрегировать веса при обходе графа:
- `max` — использовать максимальный вес пути (рекомендуется, устойчив к шуму)
- `sum` — суммировать веса всех путей (более агрессивный скоринг)
- `avg` — среднее значение (сглаживает выбросы)

**BFS_NORMALIZE** — Нормализация весов:
- `true` — веса нормализуются к диапазону [0, 1]
- `false` — используются сырые веса

**ASYNQ_CONCURRENCY** — Параллелизм воркера:
- Сколько задач обрабатывается одновременно
- Больше = быстрее, но выше нагрузка на CPU

**ASYNQ_QUEUE_DEFAULT** — Приоритет очереди:
- Приоритет обработки дефолтной очереди задач
- Выше значение = выше приоритет

## Полный пример `.env` файла

```env
# Обязательные
DATABASE_URL=postgresql://kb_user:kb_password@postgres:5432/knowledge_base?sslmode=disable

# Опциональные
SERVER_PORT=8080
REDIS_URL=redis:6379
NLP_SERVICE_URL=http://nlp:5000

# Рекомендации
RECOMMENDATION_ALPHA=0.5
RECOMMENDATION_BETA=0.5
RECOMMENDATION_DEPTH=3
RECOMMENDATION_DECAY=0.5
RECOMMENDATION_CACHE_TTL_SECONDS=300
EMBEDDING_SIMILARITY_LIMIT=30

# Визуализация графа
GRAPH_LOAD_DEPTH=2

# Расширенные параметры рекомендаций
RECOMMENDATION_GAMMA=0.2
BFS_AGGREGATION=max
BFS_NORMALIZE=true

# Asynq worker параметры
ASYNQ_CONCURRENCY=10
ASYNQ_QUEUE_DEFAULT=1

## Примеры различных конфигураций

### 1. Только явные связи (без эмбеддингов)
```env
RECOMMENDATION_ALPHA=1.0
RECOMMENDATION_BETA=0.0
RECOMMENDATION_DEPTH=3
RECOMMENDATION_DECAY=0.5
```

### 2. Только семантическое сходство (игнорировать явные связи)
```env
RECOMMENDATION_ALPHA=0.0
RECOMMENDATION_BETA=1.0
```

### 3. Сбалансированный режим (50/50)
```env
RECOMMENDATION_ALPHA=0.5
RECOMMENDATION_BETA=0.5
```

### 4. Глубокий обход графа (осторожно, может быть медленно)
```env
RECOMMENDATION_DEPTH=5
RECOMMENDATION_DECAY=0.3
```

### 5. Увеличенный кэш рекомендаций (10 минут)
```env
RECOMMENDATION_CACHE_TTL_SECONDS=600
```

### 6. Больше кандидатов от эмбеддингов
```env
EMBEDDING_SIMILARITY_LIMIT=50
```

### 7. Высокая чувствительность к эмбеддингам (α=0.3, β=0.7)
```env
RECOMMENDATION_ALPHA=0.3
RECOMMENDATION_BETA=0.7
```

### 8. Глубокая визуализация графа (больше уровней соседей)
```env
GRAPH_LOAD_DEPTH=4
```

### 9. Минимальная визуализация графа (только прямые связи)
```env
GRAPH_LOAD_DEPTH=1
```

### 10. Агрессивный BFS (sum агрегация, без нормализации)
```env
BFS_AGGREGATION=sum
BFS_NORMALIZE=false
```

### 11. Консервативный BFS (avg агрегация, с нормализацией)
```env
BFS_AGGREGATION=avg
BFS_NORMALIZE=true
```

### 12. Высокопроизводительный воркер
```env
ASYNQ_CONCURRENCY=20
ASYNQ_QUEUE_DEFAULT=2
```

### 13. Экономичный воркер (минимум ресурсов)
```env
ASYNQ_CONCURRENCY=2
ASYNQ_QUEUE_DEFAULT=1
```

### Как применить изменения

#### После редактирования `.env` или переменных в `docker-compose.yml` перезапустите бэкенд:

```bash
docker-compose restart backend
```

#### Проверьте логи – должна появиться строка с загруженной конфигурацией:

```bash
docker logs kg-backend --tail 30 | grep "Config loaded"
```

#### Пример вывода:

```
Config loaded: alpha=0.50, beta=0.50, depth=3, decay=0.50, cacheTTL=5m0s, embeddingLimit=30, graphLoadDepth=2
```

### Формула итогового веса рекомендации

#### Для заметки-источника A и кандидата C:

```
score = α * explicit_weight(A, C) + β * content_sim(A, C)
explicit_weight(A, C) — сумма весов всех явных связей от A к C (с затуханием для косвенных путей, начиная со второго уровня).

content_sim(A, C) — косинусное сходство эмбеддингов, нормализованное в диапазон [0,1].

α + β не обязательно равно 1, но для сохранения масштаба рекомендуется сумма = 1.

### Примечания

#### Эмбеддинги вычисляются асинхронно воркером. Убедитесь, что воркер (kg-worker) запущен и обработал задачи для заметок, иначе content_sim будет равен 0.

#### При beta > 0 требуется, чтобы в таблице note_embeddings были векторы для всех заметок. Если их нет, рекомендации будут полагаться только на явные связи.

#### Изменение RECOMMENDATION_DEPTH больше 3 может существенно увеличить нагрузку на БД и время ответа.

#### Кэш рекомендаций инвалидируется по TTL. Для немедленной очистки можно перезапустить бэкенд или удалить ключи вручную:

```bash
docker exec kg-redis redis-cli KEYS "suggestions:*" | xargs docker exec kg-redis redis-cli DEL
```

---

## Связанная документация

| Документ | Описание |
|----------|----------|
| [`architecture/README.md`](architecture/README.md) | Backend архитектура, DDD слои |
| [`architecture/adr.md`](architecture/adr.md) | Архитектурные решения (ADR) |
| [`DEPLOYMENT.md`](DEPLOYMENT.md) | Развёртывание, проверка конфигурации |
| [`TESTING.md`](TESTING.md) | Тестирование, проверка работы параметров |
| [`backend/openAPI.yaml`](../backend/openAPI.yaml) | API спецификация |