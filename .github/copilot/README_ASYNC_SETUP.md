# ➡️ Документ удалён

**Этот файл устарел.**  
Полная информация теперь в `DEPLOYMENT.md` и `TESTING.md`.

---

## 📚 Актуальная документация

| Документ | Описание |
|----------|----------|
| **`../DEPLOYMENT.md`** | Полное руководство по развёртыванию (Docker, K8s) |
| **`../TESTING.md`** | Тестирование async tasks, мониторинг |
| **`../CONFIGURATION.md`** | Env переменные, настройка очередей |

---

## 📦 Архив

Этот файл будет удалён.
См. `../DEPLOYMENT.md` для полного руководства по async tasks.
```bash
cd D:\knowledge-graph

# 1. Остановить старые контейнеры
docker-compose down

# 2. Пересобрать и запустить
docker-compose up --build -d

# 3. Подождать, пока сервисы запустятся (30-60 сек)
# Можно проверить: docker-compose ps

# 4. Протестировать async tasks
.\test_async_tasks.ps1
```

## Архитектура Async System

```
┌─────────────────────────────────────────────────────┐
│                   Client (Postman/curl)             │
└────────────────────────┬────────────────────────────┘
                         │ POST /notes
                         ▼
┌─────────────────────────────────────────────────────┐
│            Backend (kg-backend)                      │
│  - REST API                                          │
│  - Сохраняет заметку в PostgreSQL                   │
│  - Отправляет async tasks в Redis (via Asynq)       │
└────────────────────────┬────────────────────────────┘
                         │ 
                    ┌────┴─────┐
                    │           │
        ┌───────────▼──┐  ┌───────────▼──┐
        │ Task: extract │  │ Task: embed  │
        │  keywords    │  │  ding        │
        └───────────────┘  └───────────────┘
                    │           │
                    └───────┬───┘
                            ▼
                   ┌──────────────────┐
                   │  Redis Queue     │
                   │ (kg-redis)       │
                   └────────┬─────────┘
                            │
                            ▼
                   ┌──────────────────┐
                   │  Worker          │
                   │ (kg-worker)      │
                   │ - Обрабатывает   │
                   │   tasks          │
                   │ - Вызывает NLP   │
                   │ - Сохраняет      │
                   │   результаты     │
                   └────────┬─────────┘
                            │
                    ┌───────┴────────┐
                    │                │
          ┌─────────▼───┐  ┌────────▼────┐
          │ PostgreSQL  │  │  NLP Service │
          │ (kg-postgres)│ │  (kg-nlp)    │
          └─────────────┘  └──────────────┘
```

## Структура проекта

```
knowledge-graph/
├── backend/
│   ├── cmd/
│   │   ├── server/      # REST API server
│   │   └── worker/      # Async task worker
│   ├── internal/
│   │   ├── infrastructure/
│   │   │   ├── queue/
│   │   │   │   ├── asynq_client.go     # Клиент для отправки задач
│   │   │   │   ├── worker.go           # Обработчик задач
│   │   │   │   └── tasks.go            # Типы задач
│   │   │   └── ...
│   │   ├── interfaces/
│   │   │   └── api/
│   │   │       └── notehandler/
│   │   │           └── note_handler.go # API endpoints
│   │   └── ...
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── nlp-service/        # Python NLP сервис
├── docker-compose.yml
└── ...
```

## Типы Async Tasks

### 1. ExtractKeywords (`extract:keywords`)

**Триггер:** Создание или обновление заметки

**Обработчик:** `Worker.HandleExtractKeywords()`

**Процесс:**
1. Получить заметку из БД
2. Извлечь ключевые слова через NLP сервис
3. Сохранить ключевые слова в БД (таблица `note_keywords`)

**Логи:**
```
HandleExtractKeywords: received task {"note_id":"...","top_n":10}
HandleExtractKeywords: extracted 5 keywords for note ...
HandleExtractKeywords: successfully processed note ... with 5 keywords
```

### 2. ComputeEmbedding (`compute:embedding`)

**Триггер:** Создание или обновление заметки

**Обработчик:** `Worker.HandleComputeEmbedding()`

**Процесс:**
1. Получить заметку из БД
2. Вычислить embedding через NLP сервис
3. Сохранить embedding в БД (таблица `embeddings` с pgvector)

**Логи:**
```
HandleComputeEmbedding: received task {"note_id":"..."}
HandleComputeEmbedding: found note ..., processing...
HandleComputeEmbedding: computed embedding for note ... (size=384)
HandleComputeEmbedding: successfully processed note ...
```

## Отладка

### Проверить статус контейнеров

```bash
docker-compose ps
```

Ожидаемый результат:
```
NAME          STATUS              
kg-postgres   Up (healthy)        
kg-redis      Up (healthy)        
kg-nlp        Up (healthy)        
kg-backend    Up (healthy)        
kg-worker     Up (healthy)        
```

### Посмотреть логи worker'а

```bash
# Последние 50 строк
docker-compose logs --tail=50 kg-worker

# Follow (слежение за логами в реальном времени)
docker-compose logs -f kg-worker

# С временными метками
docker-compose logs --timestamps kg-worker
```

### Проверить Redis

```bash
# Посмотреть logи Redis
docker-compose logs kg-redis

# Подключиться к Redis и проверить очередь
docker exec kg-redis redis-cli KEYS "*asynq*"
docker exec kg-redis redis-cli LLEN "asynq:queues:default"
```

### Проверить PostgreSQL

```bash
# Логи базы
docker-compose logs kg-postgres

# Подключиться к БД
docker exec -it kg-postgres psql -U kb_user -d knowledge_base

# Проверить таблицы
\dt
SELECT COUNT(*) FROM notes;
SELECT COUNT(*) FROM note_keywords;
SELECT COUNT(*) FROM embeddings;
```

### Проверить NLP сервис

```bash
# Логи NLP
docker-compose logs -f kg-nlp

# Health check
curl http://localhost:5000/health

# Проверить доступность сервиса
docker exec kg-nlp curl http://localhost:5000/health
```

## Частые проблемы

### ❌ Worker не обрабатывает tasks

**Симптомы:**
- Задачи отправляются (видно в логах backend: "Task enqueued")
- Но в worker логах нет "HandleExtractKeywords" или "HandleComputeEmbedding"

**Диагностика:**
```bash
# Проверить Redis
docker-compose logs kg-redis | tail -20

# Проверить connection в worker
docker-compose logs kg-worker | grep -i "redis\|error"

# Проверить очередь в Redis
docker exec kg-redis redis-cli LLEN "asynq:queues:default"
```

**Решение:**
1. Убедитесь, что Redis запущен и здоров: `docker-compose ps`
2. Перезапустите worker: `docker-compose restart kg-worker`
3. Проверьте переменные окружения в docker-compose.yml (REDIS_URL)

### ❌ Backend не может подключиться к Redis

**Логи:**
```
WARNING: failed to create asynq client: connection refused
```

**Решение:**
```bash
# Проверить Redis
docker-compose logs kg-redis

# Убедитесь, что redis в docker-compose.yml имеет правильный адрес
# и что он запущен: docker-compose up -d kg-redis
```

### ❌ NLP сервис не готов

**Симптомы:**
- Worker обрабатывает tasks
- Но есть ошибки: "failed to extract keywords" или "failed to compute embedding"

**Логи:**
```
HandleExtractKeywords: failed to extract keywords: connection refused
```

**Решение:**
```bash
# Дождитесь загрузки модели (может занять 5-10 минут)
docker-compose logs -f kg-nlp

# Проверьте health endpoint
curl http://localhost:5000/health

# Если не работает, пересоберите NLP сервис
docker-compose build --no-cache kg-nlp
docker-compose up -d kg-nlp
```

### ❌ Ошибки базы данных

**Симптомы:**
- Backend или worker падают с ошибками БД

**Решение:**
```bash
# Полная пересборка с удалением data
docker-compose down -v
docker-compose up --build -d

# Это удалит все данные, но часто решает проблемы с БД
```

## Команды для тестирования

### Создать заметку через curl

```bash
curl -X POST http://localhost:8080/notes \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Note",
    "content": "This is test content for keyword extraction and embedding computation",
    "metadata": {"test": true}
  }'
```

### Создать заметку через PowerShell

```powershell
$body = @{
    title = "Test Note"
    content = "This is test content"
    metadata = @{ test = $true }
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/notes" `
  -Method POST `
  -ContentType "application/json" `
  -Body $body
```

### Получить заметку

```bash
curl http://localhost:8080/notes/{note-id}
```

### Проверить, что заметка обработана

```powershell
# Подключиться к БД и проверить наличие keywords и embeddings

docker exec -it kg-postgres psql -U kb_user -d knowledge_base -c "
SELECT 
  n.id,
  n.title,
  COUNT(k.id) as keyword_count,
  COUNT(e.id) as embedding_count
FROM notes n
LEFT JOIN note_keywords k ON n.id = k.note_id
LEFT JOIN embeddings e ON n.id = e.note_id
GROUP BY n.id, n.title
ORDER BY n.created_at DESC
LIMIT 10;
"
```

## Ожидаемое поведение

### При создании заметки (POST /notes):

1. **Backend** (logs):
   ```
   Enqueuing tasks for note {id}
   EnqueueExtractKeywords called for note {id}
   Task enqueued: {...}
   EnqueueComputeEmbedding called for note {id}
   Task enqueued: {...}
   ```

2. **Redis** (очередь заполняется задачами)

3. **Worker** (начинает обрабатывать через 1-2 секунды):
   ```
   HandleExtractKeywords: received task {...}
   HandleExtractKeywords: extracted 5 keywords
   HandleExtractKeywords: successfully processed note {id}
   
   HandleComputeEmbedding: received task {...}
   HandleComputeEmbedding: computed embedding (size=384)
   HandleComputeEmbedding: successfully processed note {id}
   ```

4. **PostgreSQL** (данные сохраняются):
   - Запись в таблице `note_keywords`
   - Запись в таблице `embeddings`

## Полезные команды

```bash
# Полная пересборка
docker-compose down -v
docker-compose up --build -d

# Слежение за логами всех сервисов
docker-compose logs -f

# Перезапуск worker
docker-compose restart kg-worker

# Остановить все
docker-compose down

# Удалить все, включая volumes
docker-compose down -v

# Выполнить команду в контейнере
docker exec -it kg-backend sh
docker exec -it kg-worker sh
docker exec -it kg-postgres bash
```

## Документация

- **REBUILD_CHECKLIST.md** - Чек-лист пересборки и диагностики
- **ASYNC_TESTING_GUIDE.md** - Подробное руководство по тестированию
- **test_async_tasks.ps1** - Скрипт для автоматического тестирования
- **rebuild_and_test.ps1** - Полный скрипт пересборки и тестирования

## Вопросы?

Если что-то не работает:
1. Проверьте логи: `docker-compose logs -f kg-worker`
2. Читайте раздел "Частые проблемы" выше
3. Убедитесь, что все контейнеры running и healthy: `docker-compose ps`
4. Попробуйте полную пересборку: `docker-compose down -v && docker-compose up --build -d`
