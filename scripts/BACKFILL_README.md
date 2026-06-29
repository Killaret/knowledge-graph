# Backfill Recommendations

Массовый пересчёт рекомендаций для всех заметок в базе.

## Зачем нужно?

Worker автоматически обновляет рекомендации только для **новых/изменённых** заметок. Для существующих заметок нужно запустить CLI вручную.

## Быстрый запуск

```powershell
# Personal stack
docker compose -f docker-compose.personal.yml run --rm cli_personal ./cli

# Dev stack
docker compose run --rm cli ./cli
```

## Dry Run (проверка без создания задач)

```powershell
docker compose -f docker-compose.personal.yml run --rm cli_personal ./cli --dry-run
```

## Мониторинг прогресса

### 1. Логи worker

```powershell
docker compose -f docker-compose.personal.yml logs -f kg-worker-personal
```

Искать строки:
```
[RefreshService] Starting refresh recommendations for note {id}
[RefreshService] Successfully refreshed N recommendations for note {id}
```

### 2. Проверка в БД

```bash
docker compose -f docker-compose.personal.yml exec postgres_personal psql -U personal -d knowledge_personal -c "SELECT note_id, COUNT(*) as rec_count FROM note_recommendations GROUP BY note_id;"
```

### 3. Redis queue (asynqmon)

```bash
# Установить asynqmon
go install github.com/hibiken/asynqmon@latest

# Запустить
asynqmon --redis-addr localhost:6380
```

## Конфигурация

Параметры из `knowledge-graph.config.json`:

```json
{
  "backend": {
    "recommendation": {
      "alpha": 0.5,           // Вес графа
      "beta": 0.5,            // Вес семантики
      "gamma": 0.2,           // Вес ключевых слов
      "depth": 3,             // Глубина обхода графа
      "decay": 0.5,           // Затухание веса
      "top_n": 50,            // Максимум рекомендаций
      "keyword_similarity_method": "jaccard"
    }
  }
}
```

## Формула расчёта

```
Final Score = α×Graph + β×Semantic + γ×Keyword
```

### Компоненты:

1. **Graph Score** (α=0.5)
   - BFS traversal depth=3
   - Decay factor=0.5
   - Link types: reference, dependency, related

2. **Semantic Score** (β=0.5)
   - Cosine similarity векторов (384 dim)
   - Модель: all-MiniLM-L6-V2
   - Из `note_embeddings` таблицы

3. **Keyword Score** (γ=0.2)
   - Jaccard coefficient
   - Ключевые слова из NLP service
   - Из `note_keywords` таблицы

## Troubleshooting

### Worker не обрабатывает задачи

1. Проверить очередь:
   ```bash
   docker compose -f docker-compose.personal.yml exec redis_personal redis-cli LLEN asynq:{default}
   ```

2. Проверить логи worker:
   ```bash
   docker logs kg-worker-personal --tail 100
   ```

3. Перезапустить worker:
   ```bash
   docker compose -f docker-compose.personal.yml restart worker_personal
   ```

### Ошибки подключения к БД

Проверить DATABASE_URL:
```bash
docker compose -f docker-compose.personal.yml exec backend_personal env | grep DATABASE_URL
```

### Пустые рекомендации

Причины:
- Нет связей между заметками → создать через UI
- Нет embeddings → проверить NLP service
- Нет keywords → проверить NLP service

## Автоматизация

Добавить в crontab для периодического пересчёта:

```bash
# Раз в неделю в 3:00
0 3 * * 0 docker compose -f /path/to/docker-compose.personal.yml run --rm cli_personal ./cli
```

## Скрипт PowerShell

Автоматический скрипт `scripts/backfill-recommendations.ps1`:

```powershell
# Dry run
.\scripts\backfill-recommendations.ps1 -DryRun

# Реальный запуск
.\scripts\backfill-recommendations.ps1
```
