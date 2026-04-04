# QUICK START - Пересборка и тестирование

## За 30 секунд:

```powershell
cd D:\knowledge-graph
.\rebuild_and_test.ps1
```

Скрипт автоматически:
✅ Остановит старые контейнеры
✅ Пересоберет образы
✅ Запустит все сервисы
✅ Создаст тестовую заметку
✅ Проверит обработку async tasks

---

## Что произойдет:

1. **Docker Compose пересобирается** (2-5 мин)
   - Backend (Go) - компилируется
   - Worker (Go) - компилируется
   - NLP сервис (Python) - загружается модель

2. **Контейнеры запускаются** (30-60 сек)
   - PostgreSQL инициализируется
   - Redis запускается
   - NLP сервис загружает модель ML

3. **Тестируется async обработка**
   - Создается тестовая заметка
   - Worker обрабатывает keywords extraction
   - Worker обрабатывает embedding computation
   - Результаты сохраняются в БД

---

## Ожидаемый результат в логах:

```
[OK] All services healthy

[OK] Note created successfully
  Note ID: 123e4567-e89b-12d3-a456-426614174000

[OK] Keyword extraction tasks processed
  Found 1 successful processing(s)

[OK] Embedding computation tasks processed
  Found 1 successful processing(s)

✅ Rebuild and testing complete!
```

---

## Если что-то пошло не так:

### Worker не обрабатывает tasks?
```bash
docker-compose logs -f kg-worker
```

### NLP сервис не готов?
```bash
docker-compose logs -f kg-nlp
```

### Полная пересборка:
```bash
cd D:\knowledge-graph
docker-compose down -v
docker-compose up --build -d
```

---

## Дополнительные скрипты:

```bash
# Только тестирование (без пересборки)
.\test_async_tasks.ps1

# Вручную создать заметку
curl -X POST http://localhost:8080/notes \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","content":"Test content"}'

# Смотреть логи worker в реальном времени
docker-compose logs -f kg-worker

# Остановить все
docker-compose down
```

---

## Файлы документации:

- 📄 **README_ASYNC_SETUP.md** - Полное руководство
- 📋 **REBUILD_CHECKLIST.md** - Чек-лист диагностики
- 📘 **ASYNC_TESTING_GUIDE.md** - Детальное руководство тестирования
- 🔧 **rebuild_and_test.ps1** - Полный скрипт пересборки
- ✅ **test_async_tasks.ps1** - Скрипт тестирования

Готово! 🚀
