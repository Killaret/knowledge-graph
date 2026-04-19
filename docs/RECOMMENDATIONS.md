# Асинхронное предвычисление рекомендаций

## Обзор архитектуры

Рекомендательная система Knowledge Graph переведена на **асинхронное событийно-ориентированное предвычисление**. Вместо синхронного расчёта рекомендаций при каждом запросе, рекомендации вычисляются в фоне и сохраняются в таблице `note_recommendations`.

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Создание/      │────▶│  Asynq Queue     │────▶│  Worker         │
│  Обновление     │     │  (Redis)         │     │  (RefreshSvc)   │
│  заметки/связи  │     └──────────────────┘     └────────┬────────┘
└─────────────────┘                                      │
                                                         ▼
                                              ┌────────────────────┐
                                              │  note_recommendations│
                                              │  (PostgreSQL)        │
                                              └──────────┬─────────┘
                                                         │
                                                         ▼
                                              ┌────────────────────┐
                                              │  GET /suggestions   │
                                              │  (быстрое чтение)   │
                                              └────────────────────┘
```

## Компоненты системы

### 1. Таблица предвычисленных рекомендаций

**Миграция:** `backend/migrations/013_create_note_recommendations.up.sql`

```sql
CREATE TABLE note_recommendations (
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    recommended_note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    score REAL NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (note_id, recommended_note_id)
);
```

### 2. Сервис обновления (RefreshService)

**Файл:** `backend/internal/application/recommendation/refresh_service.go`

Атомарно обновляет рекомендации в транзакции:
- Получает рекомендации через `TraversalService.GetSuggestions`
- Сохраняет через `SaveBatch` (UPSERT)
- Удаляет устаревшие через `DeleteNotInBatch`

### 3. Asynq задачи

**Файл:** `backend/internal/infrastructure/queue/tasks/recommendation.go`

```go
const TypeRefreshRecommendations = "recommendation:refresh"

// Опции задачи:
// - MaxRetry(3)                    // 3 попытки
// - Timeout(30s)                   // Таймаут
// - ProcessIn(delay)              // Задержка (dedup)
// - UniqueKey("rec:{note_id}")    // Дедупликация
```

### 4. Событийная логика

**Файлы:**
- `backend/internal/interfaces/api/notehandler/note_handler.go`
- `backend/internal/interfaces/api/linkhandler/link_handler.go`

При создании/обновлении заметки или связи автоматически ставятся задачи на обновление рекомендаций.

### 5. Определение затронутых заметок

**Файл:** `backend/internal/application/recommendation/affected_notes.go`

```go
const reverseCascadeDepth = 1  // Ограничение каскада

func GetAffectedNotes(targetNoteID) []uuid.UUID {
    // 1. Сама заметка
    // 2. Заметки, которые её рекомендуют (только прямые)
}
```

## Конфигурация

### Переменные окружения

```bash
# Количество рекомендаций для сохранения
RECOMMENDATION_TOP_N=20

# Задержка перед выполнением задачи (dedup в секундах)
RECOMMENDATION_TASK_DELAY_SECONDS=5

# Fallback на Redis
RECOMMENDATION_FALLBACK_ENABLED=true
RECOMMENDATION_FALLBACK_TTL_SECONDS=3600

# Fallback на семантических соседей
RECOMMENDATION_FALLBACK_SEMANTIC_ENABLED=true

# Asynq настройки
ASYNQ_CONCURRENCY=10
ASYNQ_QUEUE_DEFAULT=1
ASYNQ_QUEUE_MAX_LEN=10000
```

## API Эндпоинт

### GET /api/v1/notes/{id}/suggestions

**Логика работы:**

1. **Предвычисленные рекомендации** — быстрое чтение из `note_recommendations`
2. **Проверка актуальности** — сравнение `updated_at` рекомендаций и заметки
3. **Fallback на семантических соседей** — если предвычисленных нет
4. **Fallback на Redis** — если семантический отключен
5. **HTTP 202 Accepted** — если ничего нет, запускается фоновый расчёт

**Заголовки ответа:**

```http
X-Recommendations-Stale: true           # Данные устарели
X-Recommendations-Source: semantic-fallback
X-Recommendations-Source: redis-fallback
```

## Первоначальное заполнение

### CLI команда

**Файл:** `backend/cmd/cli/main.go`

```bash
# Сборка
cd backend
go build -o bin/recommendation-cli ./cmd/cli

# Запуск (dry-run для проверки)
./bin/recommendation-cli --dry-run

# Реальный запуск
./bin/recommendation-cli

# Для больших баз (>1000 заметок) увеличивается задержка
./bin/recommendation-cli --batch-delay=60
```

## Мониторинг

### Логи

Воркер логирует время выполнения каждой задачи:

```
[RefreshService] Starting refresh recommendations for note a000...0001
[RefreshService] Finished refresh for note a000...0001 in 45.2ms
```

### Метрики для отслеживания

1. **Длина очереди Asynq:**
   ```bash
   redis-cli LLEN asynq:{default}
   ```

2. **Количество задач в статусе pending:**
   ```bash
   redis-cli ZCARD asynq:scheduled
   ```

3. **Ошибки в логах воркера:**
   ```bash
   grep "failed to refresh" /var/log/worker.log
   ```

### Веб-интерфейс asynqmon

Рекомендуется развернуть `asynqmon` для визуального мониторинга:

```bash
docker run -p 8080:8080 hibiken/asynqmon --redis-addr=localhost:6379
```

## Производительность

### Сравнение подходов

| Подход | Latency | Нагрузка на БД | Масштабируемость |
|--------|---------|----------------|------------------|
| Синхронный BFS | 50-200ms | Высокая | Ограничена |
| Асинхронный (текущий) | 5-10ms | Низкая | Отличная |

### Оптимизации

1. **Batch-загрузка соседей** — `GetNeighborsBatch` уменьшает количество SQL-запросов
2. **Дедупликация задач** — `UniqueKey` предотвращает дублирование
3. **Ограничение каскада** — `reverseCascadeDepth = 1` предотвращает взрыв очереди
4. **Транзакционность** — атомарное обновление через `SaveBatch` + `DeleteNotInBatch`

## Устранение неполадок

### Проблема: Рекомендации не обновляются

**Проверки:**
1. Работает ли воркер? Проверить логи
2. Есть ли задачи в очереди? `redis-cli LLEN asynq:{default}`
3. Применены ли миграции? `\dt note_recommendations`

### Проблема: Очередь переполнена

**Решения:**
1. Увеличить `ASYNQ_QUEUE_MAX_LEN`
2. Добавить rate limiting в CLI
3. Увеличить `RecommendationTaskDelaySeconds`

### Проблема: Устаревшие рекомендации

**Причины:**
1. Воркер перегружен — увеличить `ASYNQ_CONCURRENCY`
2. Задачи падают с ошибками — проверить логи
3. Redis недоступен — проверить подключение

## Миграция с существующей системы

1. Применить миграцию:
   ```bash
   psql -d knowledge_base -f backend/migrations/013_create_note_recommendations.up.sql
   ```

2. Запустить CLI для начального заполнения:
   ```bash
   ./bin/recommendation-cli
   ```

3. Дождаться завершения очереди (мониторинг через asynqmon)

4. Переключить API на новый handler (уже сделано в коде)

5. Удалить старый Redis-кэш при необходимости:
   ```bash
   redis-cli --scan --pattern "recommendations:*" | xargs redis-cli del
   ```
