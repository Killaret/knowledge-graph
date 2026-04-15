# ➡️ Документ удалён

**Этот индекс устарел (2026-04-03).**

---

## 📚 Актуальная документация

| Документ | Описание |
|----------|----------|
| **`START_HERE.md`** | Обновлённое главное меню |
| **`../DEPLOYMENT.md`** | Развёртывание |
| **`../TESTING.md`** | Тестирование |
| **`../CHANGELOG.md`** | История изменений |

---

## 📦 Архив

Этот файл будет удалён.
См. `../archive/README.md` для списка архивированных документов.

---

## Устаревший индекс (2026-04-03):

---

## 🔧 Скрипты для запуска

### 1. rebuild_and_test.ps1 (⭐ Основной)
**Описание:** Полная пересборка и тестирование async tasks

**Использование:**
```powershell
cd D:\knowledge-graph
.\rebuild_and_test.ps1
```

**Что делает:**
- ✅ Останавливает старые контейнеры
- ✅ Пересобирает Docker образы
- ✅ Запускает все сервисы
- ✅ Ждет, пока все будут healthy
- ✅ Создает тестовую заметку
- ✅ Проверяет обработку async tasks
- ✅ Выводит результаты

**Время выполнения:** 5-10 мин

### 2. test_async_tasks.ps1
**Описание:** Только тестирование (без пересборки)

**Использование:**
```powershell
.\test_async_tasks.ps1
```

**Что делает:**
- ✅ Создает тестовую заметку
- ✅ Отправляет на обработку
- ✅ Проверяет логи worker'а
- ✅ Показывает результаты

**Время выполнения:** ~30 сек

### 3. rebuild_and_test.bat
**Описание:** Batch скрипт для Windows cmd.exe

**Использование:**
```cmd
rebuild_and_test.bat
```

---

## 📚 Документация

### 1. README_ASYNC_SETUP.md (⭐ Полное руководство)
**Содержание:**
- Быстрый старт (3 варианта)
- Архитектура async system
- Структура проекта
- Типы async tasks (ExtractKeywords, ComputeEmbedding)
- Отладка и диагностика
- Частые проблемы и решения
- Все полезные команды

**Объем:** ~10 KB, 5-10 мин чтения

**Когда читать:** Для глубокого понимания системы

### 2. REBUILD_CHECKLIST.md (✅ Пошаговый чек-лист)
**Содержание:**
- 7-шаговый процесс пересборки
- Ожидаемые результаты на каждом шаге
- Проверочные команды
- Troubleshooting для каждого шага
- Диагностика проблем

**Объем:** ~5 KB

**Когда читать:** Когда следуете процессу вручную

### 3. ASYNC_TESTING_GUIDE.md (📘 Руководство тестирования)
**Содержание:**
- Пошаговое тестирование
- Примеры запросов (curl, PowerShell)
- Ожидаемые результаты
- Интерпретация логов
- Всё про инструменты тестирования

**Объем:** ~3 KB

**Когда читать:** При тестировании async tasks

### 4. ARCHITECTURE_DIAGRAMS.md (🎨 Визуальные диаграммы)
**Содержание:**
- Диаграмма потока создания заметки
- Архитектура Docker контейнеров
- Жизненный цикл async task
- Параллельная обработка
- Состояния обработки в логах

**Объем:** ~4 KB

**Когда читать:** Для понимания архитектуры

### 5. SUMMARY.md (📋 Подробный отчет)
**Содержание:**
- Что было подготовлено
- Как использовать
- Архитектура системы
- Что проверить после пересборки
- Troubleshooting
- Следующие шаги

**Объем:** ~7 KB

**Когда читать:** Для полного обзора

### 6. QUICK_START.md (⚡ Очень быстрый старт)
**Содержание:**
- 30-секундный старт
- За 3 команды
- Ожидаемый результат
- Быстрые решения

**Объем:** ~2 KB

**Когда читать:** Если спешите

---

## 📂 Структура файлов в проекте

```
D:\knowledge-graph\
│
├── 📄 START_HERE.md                ← ВЫ ЗДЕСЬ
├── 📄 QUICK_START.md               ← Быстро прочитать
│
├── 🔧 СКРИПТЫ:
├── 📜 rebuild_and_test.ps1         ← Запустить это
├── 📜 test_async_tasks.ps1         ← Для тестирования
├── 📜 rebuild_and_test.bat         ← Windows cmd версия
│
├── 📚 ДОКУМЕНТАЦИЯ:
├── 📘 README_ASYNC_SETUP.md        ← Полное руководство
├── 📋 REBUILD_CHECKLIST.md         ← Пошаговый чек-лист
├── 📖 ASYNC_TESTING_GUIDE.md       ← Тестирование
├── 🎨 ARCHITECTURE_DIAGRAMS.md     ← Диаграммы
├── 📄 SUMMARY.md                   ← Итоговый отчет
│
├── 🐳 DOCKER:
├── docker-compose.yml              ← Конфиг контейнеров
├── .env                            ← Переменные окружения
│
├── 💻 BACKEND (Go):
├── backend/
│   ├── cmd/
│   │   ├── server/main.go          ← REST API
│   │   └── worker/main.go          ← Async worker
│   ├── internal/
│   │   ├── infrastructure/queue/
│   │   │   ├── asynq_client.go     ← Отправка задач
│   │   │   ├── worker.go           ← Обработка задач
│   │   │   └── tasks.go            ← Типы задач
│   │   └── ...
│   └── Dockerfile
│
├── 🐍 NLP SERVICE (Python):
├── nlp-service/
│   ├── app/main.py
│   ├── requirements.txt
│   └── Dockerfile
│
├── 💾 БД:
├── init-db/                        ← SQL инициализация
│   └── init.sql
│
└── 📖 ДРУГОЕ:
    └── migrations/                 ← Миграции БД
```

---

## 🗺️ Карта документации

### Если вы хотите:

**Быстро запустить:**
1. Откройте QUICK_START.md
2. Выполните `.\rebuild_and_test.ps1`
3. Готово! ✓

**Подробно разобраться:**
1. Читайте README_ASYNC_SETUP.md (архитектура)
2. Смотрите ARCHITECTURE_DIAGRAMS.md (визуально)
3. Следуйте REBUILD_CHECKLIST.md (пошагово)

**Найти решение проблемы:**
1. Откройте REBUILD_CHECKLIST.md
2. Найдите вашу проблему
3. Следуйте решению

**Протестировать систему:**
1. Используйте ASYNC_TESTING_GUIDE.md
2. Запустите `.\test_async_tasks.ps1`
3. Проверьте логи: `docker-compose logs -f kg-worker`

**Понять архитектуру:**
1. Смотрите ARCHITECTURE_DIAGRAMS.md (визуально)
2. Читайте README_ASYNC_SETUP.md (теория)
3. Изучайте код в backend/cmd/worker/main.go

---

## ⏱️ Рекомендуемый порядок чтения

### Вариант 1: Спешу (5 мин)
1. START_HERE.md (2 мин)
2. QUICK_START.md (2 мин)
3. Запустить скрипт (1 мин)

### Вариант 2: Нормальный темп (15 мин)
1. START_HERE.md (2 мин)
2. README_ASYNC_SETUP.md (10 мин)
3. Запустить скрипт (3 мин)

### Вариант 3: Детально (30+ мин)
1. START_HERE.md (2 мин)
2. QUICK_START.md (2 мин)
3. ARCHITECTURE_DIAGRAMS.md (5 мин)
4. README_ASYNC_SETUP.md (10 мин)
5. REBUILD_CHECKLIST.md (5 мин)
6. ASYNC_TESTING_GUIDE.md (5 мин)
7. Запустить и изучить код

---

## 🔍 Быстрая справка

### Основные команды:

```bash
# Пересборка и тест (всё в одну команду)
.\rebuild_and_test.ps1

# Только тест
.\test_async_tasks.ps1

# Смотреть логи в реальном времени
docker-compose logs -f kg-worker

# Проверить статус
docker-compose ps

# Остановить
docker-compose down

# Полная очистка
docker-compose down -v
```

---

## 📊 Статистика подготовленных материалов

- **Скрипты:** 3 файла
- **Документация:** 6 файлов
- **Общий объем документации:** ~40 KB
- **Время на прочтение:** 5-30 мин (в зависимости от выбора)
- **Время на запуск:** 5-10 мин
- **Итого:** 10-40 мин от чтения до полного тестирования

---

## ✅ Чек-лист использования

- [ ] Прочитал START_HERE.md
- [ ] Прочитал QUICK_START.md
- [ ] Запустил `.\rebuild_and_test.ps1`
- [ ] Дождался завершения (5-10 мин)
- [ ] Увидел "✅ Rebuild and testing complete!"
- [ ] При проблемах прочитал REBUILD_CHECKLIST.md
- [ ] Протестировал `.\test_async_tasks.ps1`
- [ ] Проверил логи: `docker-compose logs -f kg-worker`
- [ ] Система работает! 🎉

---

## 🎓 После пересборки

### Дополнительно можно:

1. **Изучить код:**
   - backend/cmd/worker/main.go
   - backend/internal/infrastructure/queue/worker.go
   - backend/internal/infrastructure/queue/asynq_client.go

2. **Протестировать API:**
   - Создать несколько заметок
   - Проверить keyword extraction
   - Проверить embedding computation

3. **Мониторить обработку:**
   - `docker-compose logs -f kg-worker`
   - `docker-compose logs -f kg-backend`
   - `docker-compose logs -f kg-redis`

4. **Проверить БД:**
   ```bash
   docker exec -it kg-postgres psql -U kb_user -d knowledge_base
   SELECT * FROM notes;
   SELECT * FROM note_keywords;
   SELECT * FROM embeddings;
   ```

---

## 🎯 Итог

Вы получили:

✅ Полный набор скриптов для пересборки
✅ Пошаговые инструкции
✅ Подробную документацию
✅ Диаграммы архитектуры
✅ Solutions для проблем
✅ All комманды для отладки

**Начните:**
```powershell
.\rebuild_and_test.ps1
```

**Удачи!** 🚀

---

**Остались вопросы?** → Откройте REBUILD_CHECKLIST.md
