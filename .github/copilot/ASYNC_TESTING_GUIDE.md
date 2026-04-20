# Инструкция по пересборке и тестированию async tasks

## Шаг 1: Пересобрать проект и запустить контейнеры

Выполните один из следующих вариантов:

### Вариант А: Использовать готовый скрипт
```bash
cd D:\knowledge-graph
rebuild_and_test.bat
```

### Вариант Б: Вручную
```bash
cd D:\knowledge-graph

# Остановить старые контейнеры
docker-compose down

# Пересобрать и запустить
docker-compose up --build -d

# Подождать 30 секунд
timeout /t 30

# Проверить статус контейнеров
docker-compose ps
```

## Шаг 2: Проверить логи

### Логи backend'а:
```bash
docker-compose logs -f kg-backend --tail=50
```

### Логи worker'а (самое важное для async tasks):
```bash
docker-compose logs -f kg-worker --tail=50
```

### Логи Redis:
```bash
docker-compose logs -f kg-redis --tail=20
```

### Логи NLP сервиса:
```bash
docker-compose logs -f kg-nlp --tail=20
```

## Шаг 3: Тестировать async tasks

### Способ 1: Использовать скрипт
```bash
test_async_tasks.bat
```

### Способ 2: Вручную через curl/PowerShell

#### Создать заметку:
```powershell
$body = @{
    title = "Test Note"
    content = "This is a test note for keyword extraction and embedding computation."
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/notes" `
  -Method POST `
  -ContentType "application/json" `
  -Body $body
```

#### Проверить логи worker'а на обработку:
```bash
docker-compose logs -f kg-worker
```

## Ожидаемые логи при обработке tasks

При создании заметки вы должны увидеть в логах worker'а:

```
HandleExtractKeywords: received task {"note_id":"...","top_n":10}
HandleExtractKeywords: extracted 5 keywords for note ...
HandleExtractKeywords: successfully processed note ... with 5 keywords

HandleComputeEmbedding: received task {"note_id":"..."}
HandleComputeEmbedding: found note ..., processing...
HandleComputeEmbedding: computed embedding for note ... (size=384)
HandleComputeEmbedding: successfully processed note ...
```

## Troubleshooting

### Если worker не обрабатывает tasks:

1. **Проверить Redis:**
   ```bash
   docker-compose logs kg-redis
   ```

2. **Проверить connection к Redis в worker:**
   ```bash
   docker-compose logs kg-worker | findstr "Worker started" -A 5
   ```

3. **Проверить, что backend правильно отправляет tasks:**
   ```bash
   docker-compose logs kg-backend | findstr "Enqueue"
   ```

4. **Проверить логи backend на ошибки:**
   ```bash
   docker-compose logs kg-backend
   ```

### Если backend не может создать соединение с Redis/DB:

1. Проверьте `.env` файл
2. Убедитесь, что PostgreSQL и Redis здоровы:
   ```bash
   docker-compose ps
   ```

### Если NLP сервис недоступен:

```bash
docker-compose logs kg-nlp
curl http://localhost:5000/health
```

## Основные компоненты системы async:

- **Backend** (`kg-backend`): REST API, отправляет tasks в Redis через asynq
- **Worker** (`kg-worker`): обрабатывает tasks из Redis, вызывает NLP и сохраняет результаты
- **Redis** (`kg-redis`): очередь задач
- **PostgreSQL** (`kg-postgres`): хранилище заметок и результатов
- **NLP Service** (`kg-nlp`): Python сервис для извлечения keywords и embeddings
