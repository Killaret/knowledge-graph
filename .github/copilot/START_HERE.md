# 🚀 Knowledge Graph - Документация

**Статус:** ✅ Production Ready v1.0.0

---

## 📚 Актуальная документация

### Основное (начните здесь)
| Документ | Описание | Время чтения |
|----------|----------|--------------|
| **`../DEPLOYMENT.md`** | Развёртывание (Docker, K8s) | 10 мин |
| **`../TESTING.md`** | Тестирование (Go, Playwright, Cucumber) | 10 мин |
| **`../CONFIGURATION.md`** | Env переменные, формулы рекомендаций | 5 мин |
| **`../CHANGELOG.md`** | История изменений v1.0.0 | 5 мин |

### Архитектура
| Документ | Описание |
|----------|----------|
| **`../architecture/README.md`** | C4 Model, UML диаграммы |
| **`../architecture/adr.md`** | 14 архитектурных решений (ADR) |
| **`../FRONTEND_ARCHITECTURE.md`** | Three.js, Progressive Rendering |

### API
| Документ | Описание |
|----------|----------|
| **`../../backend/openAPI.yaml`** | OpenAPI 3.1 спецификация |
| **`../API_ERRORS.md`** | Коды ошибок, примеры обработки |

### Вспомогательное
| Документ | Описание |
|----------|----------|
| **`../TASKS.md`** | Бэклог задач (TODO) |
| **`../IDEAS.md`** | Идеи для развития |
| **`../UX_GUIDELINES.md`** | UX принципы (Шнейдерман, Нильсен, Рамс) |

---

## 🚀 Быстрый старт: 3 шага

### Шаг 1: Откройте PowerShell
```powershell
cd D:\knowledge-graph
```

### Шаг 2: Запустите скрипт пересборки
```powershell
.\rebuild_and_test.ps1
```

### Шаг 3: Ждите результатов ✓
- Пересборка: ~3-5 мин
- Запуск сервисов: ~30-60 сек
- Тестирование: ~15 сек
- **Итого: ~5-7 мин**

---

## Что произойдет:

```
✅ Docker контейнеры остановлены
✅ Образы пересобраны:
   - backend (Go server + worker)
   - nlp-service (Python ML model)
   - PostgreSQL
   - Redis
✅ Все сервисы запущены
✅ Проверена готовность к обработке
✅ Создана тестовая заметка
✅ Worker обработал async tasks
✅ Результаты сохранены в БД
✅ Выведены результаты
```

---

## Ожидаемый результат:

```
✓ All services healthy
✓ Note created successfully
✓ Keyword extraction tasks processed
✓ Embedding computation tasks processed
✅ Rebuild and testing complete!
```

---

## Если что-то пошло не так:

Проблема? Читайте: **REBUILD_CHECKLIST.md**
- Пошаговая диагностика
- Решение для каждой проблемы
- Полезные команды

---

## Что проверяется:

✅ Пересборка Docker образов
✅ Запуск всех контейнеров (PostgreSQL, Redis, Backend, Worker, NLP)
✅ Здоровье всех сервисов
✅ Backend подключен к Asynq/Redis
✅ Worker слушает очередь задач
✅ NLP сервис готов
✅ Создание заметки работает
✅ Async task extraction:keywords обрабатывается
✅ Async task compute:embedding обрабатывается
✅ Результаты сохраняются в БД

---

## Проверка логов вручную:

```bash
# Все в реальном времени:
docker-compose logs -f

# Только worker:
docker-compose logs -f kg-worker

# Только backend:
docker-compose logs -f kg-backend

# Посмотреть последние 50 строк:
docker-compose logs --tail=50 kg-worker
```

---

## Структура проекта:

```
D:\knowledge-graph\
├── 📁 backend/                  # Go: server + worker + tests
│   ├── cmd/server, cmd/worker
│   ├── internal/ (domain, application, infrastructure, interfaces)
│   ├── migrations/
│   └── openAPI.yaml            # API спецификация
├── 📁 frontend/                 # Svelte 5 + TypeScript
│   ├── src/lib/
│   │   ├── api/                # HTTP клиенты
│   │   ├── three/              # Three.js модули (3D граф)
│   │   └── components/
│   └── tests/                  # E2E тесты (Playwright)
├── 📁 nlp-service/              # Python: embeddings, keywords
├── 📁 docs/                     # 📚 Документация
│   ├── DEPLOYMENT.md           # ← Развёртывание
│   ├── TESTING.md              # ← Тестирование
│   ├── CONFIGURATION.md        # ← Конфигурация
│   ├── CHANGELOG.md            # ← История версий
│   ├── API_ERRORS.md           # ← Ошибки API
│   ├── FRONTEND_ARCHITECTURE.md
│   ├── architecture/           # C4, ADR, UML
│   └── COPILOT_DOCS/           # AI-ассистенты
├── docker-compose.yml
├── .env
└── rebuild_and_test.ps1        # ← Скрипт пересборки
```

---

## Что дальше:

1. **Запустите развёртывание:**
   ```powershell
   # Полная инструкция в DEPLOYMENT.md
   .\rebuild_and_test.ps1
   ```

2. **Проверьте результаты:**
   - Должно быть "All services healthy ✅"
   - API доступен: http://localhost:8080/health
   - Frontend: http://localhost:5173

3. **При проблемах:**
   - Читайте `../DEPLOYMENT.md` → раздел Troubleshooting
   - Проверьте логи: `docker-compose logs -f`

4. **Тестирование:**
   ```powershell
   # Backend unit tests
   cd backend && go test ./...
   
   # Frontend E2E
   cd frontend && npm run test
   
   # Полная инструкция в TESTING.md
   ```

5. **Для слежения в реальном времени:**
   ```bash
   docker-compose logs -f kg-worker
   ```

---

## Ключевые компоненты системы:

| Компонент | Порт | Роль | Статус |
|-----------|------|------|--------|
| Backend | 8080 | REST API, отправка tasks | ✓ |
| Worker | - | Обработка tasks | ✓ |
| Redis | 6379 | Очередь задач | ✓ |
| PostgreSQL | 5432 | База данных | ✓ |
| NLP Service | 5000 | ML модель (keywords + embeddings) | ✓ |

---

## Система готова! 🎉

**Начните отсюда:**
```powershell
cd D:\knowledge-graph
docker-compose up -d
```

**Документация:**
- Развёртывание → `../DEPLOYMENT.md`
- Тестирование → `../TESTING.md`
- Конфигурация → `../CONFIGURATION.md`
- История изменений → `../CHANGELOG.md`

**Удачи!** 🚀
