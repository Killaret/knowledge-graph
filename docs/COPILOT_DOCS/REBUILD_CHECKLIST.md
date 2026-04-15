# ➡️ Документ удалён

**Этот файл устарел.**  
Диагностика теперь в `DEPLOYMENT.md` → раздел Troubleshooting.

---

## 📚 Актуальная документация

| Документ | Описание |
|----------|----------|
| **`../DEPLOYMENT.md`** | Развёртывание и диагностика |
| **`../TESTING.md`** | Проверка работоспособности |

- `Connected to PostgreSQL` ✓
- `Asynq client created successfully` ✓
- Никаких ошибок подключения ✓

---

## ✅ Шаг 3: Проверить логи worker'а

```bash
docker-compose logs kg-worker --tail=30
```

**Ищите:**
- `Worker: Connected to PostgreSQL` ✓
- `Starting worker, connecting to Redis` ✓
- `Worker started, listening for tasks...` ✓

---

## ✅ Шаг 4: Тестировать создание заметки

**Способ 1: PowerShell скрипт (рекомендуется)**
```bash
.\test_async_tasks.ps1
```

**Способ 2: curl**
```bash
curl -X POST http://localhost:8080/notes ^
  -H "Content-Type: application/json" ^
  -d "{\"title\":\"Test\",\"content\":\"Test content for async processing\"}"
```

**Способ 3: Postman/Insomnia**
- Метод: POST
- URL: http://localhost:8080/notes
- Body (JSON):
```json
{
  "title": "My Test Note",
  "content": "This is test content that will trigger async keyword extraction and embedding computation."
}
```

---

## ✅ Шаг 5: Проверить обработку async tasks

После создания заметки подождите **10 секунд** и проверьте логи worker'а:

```bash
docker-compose logs -f kg-worker
```

**Ищите эти логи:**

### Для извлечения ключевых слов:
```
HandleExtractKeywords: received task {"note_id":"...","top_n":10}
HandleExtractKeywords: extracted 5 keywords for note ...
HandleExtractKeywords: successfully processed note ... with 5 keywords
```

### Для вычисления эмбеддинга:
```
HandleComputeEmbedding: received task {"note_id":"..."}
HandleComputeEmbedding: found note ..., processing...
HandleComputeEmbedding: computed embedding for note ... (size=384)
HandleComputeEmbedding: successfully processed note ...
```

---

## ❌ Если что-то не работает:

### Проблема: Backend не подключается к Redis
```bash
docker-compose logs kg-backend | findstr "WARNING.*asynq"
```
**Решение:** Убедитесь, что `kg-redis` running и healthy

### Проблема: Worker не обрабатывает tasks
```bash
docker-compose logs kg-worker | findstr "error\|ERROR"
docker-compose logs kg-redis
```
**Решение:** Проверьте Redis и сетевое соединение

### Проблема: NLP сервис недоступен
```bash
docker-compose logs kg-nlp | tail -20
curl http://localhost:5000/health
```
**Решение:** Дождитесь, пока NLP сервис загрузит модель (может занять несколько минут)

### Проблема: PostgreSQL не подключается
```bash
docker-compose logs kg-postgres | grep -i error
```
**Решение:** Удалите volume и пересоздайте: `docker-compose down -v && docker-compose up --build -d`

---

## 📊 Статус проверка

```bash
# Все в одной команде
docker-compose ps && echo "" && docker-compose logs --tail=5 kg-worker
```

---

## 📝 Что было исправлено

- ✅ Backend правильно инициализирует Asynq клиент
- ✅ Worker корректно обрабатывает `TypeExtractKeywords` задачи
- ✅ Worker корректно обрабатывает `TypeComputeEmbedding` задачи
- ✅ Redis используется как хранилище очереди
- ✅ Логирование добавлено для отслеживания обработки
