# Knowledge Graph AI Agents

**Оркестратор активен всегда в каждой сессии!** 🚀

---

## 🎯 Автоматическая активация

Оркестратор **автоматически** активируется при:
- ✅ Любом новом чате/сессии
- ✅ Любом запросе пользователя
- ✅ Любом контексте проекта

**Без исключений! Без ручной активации!**

---

## 📁 Конфигурация

### Global Config
**Путь:** `~/.agents/config.json` или `~/.koda/config.json`

```json
{
  "defaultAgent": "knowledge-graph-orchestrator",
  "autoActivate": true
}
```

### Project Config
**Путь:** `.koda/config.json`

```json
{
  "orchestrator": {
    "enabled": true,
    "alwaysActive": true,
    "delegateAllTasks": true,
    "autoSelectAgent": true
  },
  "defaultAgent": "knowledge-graph-orchestrator"
}
```

---

## 🤖 Доступные агенты

### Мета-агент (всегда активен)

| Агент | Назначение | Статус |
|-------|------------|--------|
| **knowledge-graph-orchestrator** | Координация всех агентов | ✅ Всегда активен |

### Основные агенты (Высокий приоритет 🟢)

| Агент | Назначение | Родитель |
|-------|------------|----------|
| **knowledge-graph-backend-go** | Backend (Go, API, БД, gRPC) | orchestrator |
| **knowledge-graph-frontend-svelte** | Frontend (Svelte 5, UI/UX) | orchestrator |
| **knowledge-graph-integration** | API интеграция, type generation | orchestrator |
| **knowledge-graph-infrastructure** | Инфраструктура (Docker, K8s) | orchestrator |
| **knowledge-graph-performance** | Производительность и оптимизация | orchestrator |

### Специализированные агенты (Средний приоритет 🟡)

| Агент | Назначение | Родитель |
|-------|------------|----------|
| knowledge-graph-devops | CI/CD, деплой, мониторинг | orchestrator |
| knowledge-graph-security | Безопасность, аудит | orchestrator |

### Вспомогательные агенты (Низкий приоритет 🔵)

| Агент | Назначение | Родитель |
|-------|------------|----------|
| knowledge-graph-testing | Тестирование (unit, E2E) | orchestrator |
| knowledge-graph-docs-maintenance | Документация | orchestrator |

---

## 🎮 Как использовать

### Автоматически (рекомендуется)

Просто опишите задачу - оркестратор сам решит:

```
Пользователь: "Добавь тёмную тему в header"

Оркестратор:
1. Анализирует → frontend задача
2. Делегирует → knowledge-graph-frontend-svelte
3. Получает результат
4. Возвращает ответ
```

### С явным указанием (опционально)

```
"Используя knowledge-graph-performance, оптимизируй загрузку графа"
```

---

## 📊 Примеры работы

### Пример 1: Простой запрос

```
Запрос: "Создай API endpoint для статистики пользователей"

Оркестратор:
└─→ knowledge-graph-backend-go
    ├─ Создаёт handler
    ├─ Создаёт repository
    ├─ Пишет тесты
    └─ Обновляет docs

Результат: Готовый endpoint с тестами и документацией
```

### Пример 2: Сложная фича

```
Запрос: "Реализуй шаринг заметок с email"

Оркестратор:
├─→ knowledge-graph-backend-go (API, email service, БД)
├─→ knowledge-graph-frontend-svelte (ShareModal, формы)
├─→ knowledge-graph-integration (API mapping, типы)
├─→ knowledge-graph-testing (Unit + E2E тесты)
├─→ knowledge-graph-performance (оптимизация email)
└─→ knowledge-graph-docs-maintenance (документация)

Результат: Полностью готовая фича
```

### Пример 3: Производительность

```
Запрос: "Оптимизируй загрузку графа для 1000+ узлов"

Оркестратор:
└─→ knowledge-graph-performance
    ├─ Анализирует bottleneck
    ├─ Оптимизирует запросы БД
    ├─ Добавляет кэширование
    ├─ Оптимизирует 3D рендеринг
    └─ Устанавливает мониторинг

Результат: FPS 30 → 60, загрузка 5s → 1s
```

---

## 🔍 Как это работает

### 1. Запуск сессии

```
AI Assistant стартует
    ↓
Загружает .koda/config.json
    ↓
Активирует knowledge-graph-orchestrator
    ↓
Загружает все специализированные агенты
    ↓
Готов к приёму задач
```

### 2. Обработка запроса

```
Пользователь: "Добавь новую фичу"
    ↓
Orchestrator анализирует запрос
    ↓
Определяет: backend + frontend + tests + docs
    ↓
Делегирует 4 агентам (параллельно)
    ↓
Собирает результаты
    ↓
Возвращает единый ответ
```

### 3. Делегирование

```yaml
Правила маршрутизации:
  frontend:
    keywords: [svelte, component, UI, стиль]
    agent: knowledge-graph-frontend-svelte
    
  backend:
    keywords: [Go, API, endpoint, БД]
    agent: knowledge-graph-backend-go
    
  performance:
    keywords: [оптимизация, скорость, FPS]
    agent: knowledge-graph-performance
    
  security:
    keywords: [безопасность, auth, уязвимость]
    agent: knowledge-graph-security
    
  devops:
    keywords: [deploy, docker, мониторинг]
    agent: knowledge-graph-devops
```

---

## ✅ Проверка активации

### Проверить текущий агент

```bash
# В чате с ИИ
status
# Ожидаем: Default agent: knowledge-graph-orchestrator
```

### Проверить загруженные агенты

```bash
# В чате с ИИ
list-agents
# Ожидаем: orchestrator + 8 специализированных агентов
```

### Протестировать делегирование

```
Запрос: "Оптимизируй bundle size"

Ожидаемый ответ:
"Использую knowledge-graph-performance для оптимизации...
1. Проанализировал bundle
2. Предложил code splitting
3. Добавил lazy loading
4. Экономия: 300KB → 150KB"
```

---

## 📚 Документация

### Основные документы
- **[Orchestrator Instructions](agents/instructions.md)** - Инструкции для всех агентов
- **[Orchestration Rules](rules/orchestration-rules.md)** - Правила взаимодействия
- **[DevOps Tools](tools/devops-tools.md)** - Инструменты инфраструктуры
- **[Performance Tools](tools/performance-tools.md)** - Инструменты оптимизации

### Skills
- **[Orchestrator Guide](.koda/skills/knowledge-graph-orchestrator.md)** - Полная документация
- **[Performance Agent](.koda/skills/knowledge-graph-performance.md)** - Оптимизация
- **[Security Agent](.koda/skills/knowledge-graph-security.md)** - Безопасность
- **[DevOps Agent](.koda/skills/knowledge-graph-devops.md)** - DevOps
- **[Examples](docs/ORCHESTRATOR_EXAMPLES.md)** - Примеры использования
- **[Auto Activation](docs/ORCHESTRATOR_AUTO_ACTIVATION.md)** - Настройка автоактивации
- **[Agent Matrix](docs/AGENT_MATRIX.md)** - Матрица покрытия

---

## 🚀 Быстрый старт

### 1. Проверка конфигурации

```bash
# Проверить .koda/config.json
cat .koda/config.json

# Проверить наличие файлов агентов
ls -la .koda/skills/
```

### 2. Запуск AI Assistant

```bash
# Запустить AI помощник
ai-chat
# Или открыть в IDE/браузере
```

### 3. Отправить запрос

```
"Добавь новую функцию"
# Оркестратор автоматически активируется и обработает запрос
```

---

## 🎯 Преимущества

### Автоматическая активация

- ✅ **Не нужно помнить** о включении оркестратора
- ✅ **Всегда работает** в каждой сессии
- ✅ **Никаких ручных действий** требуется

### Умное делегирование

- ✅ **Автоматический выбор** нужного агента
- ✅ **Параллельное выполнение** независимых задач
- ✅ **Координация** сложных многошаговых задач

### Полное покрытие

- ✅ **Frontend** - UI/UX, компоненты, тесты
- ✅ **Backend** - API, БД, инфраструктура
- ✅ **Тестирование** - все уровни тестов
- ✅ **Документация** - полная актуальность
- ✅ **Интеграция** - API контракты, типы
- ✅ **Производительность** - оптимизация всего
- ✅ **Безопасность** - аудит и hardening
- ✅ **DevOps** - деплой, мониторинг

---

## ⚙️ Конфигурация

### Файлы конфигурации

| Файл | Назначение | Путь |
|------|------------|------|
| Global Config | Глобальные настройки | `~/.agents/config.json` |
| Project Config | Настройки проекта | `.koda/config.json` |
| Startup Hook | Автоматический запуск | `.koda/hooks/on-start.*` |

### Ключевые параметры

```json
{
  "orchestrator": {
    "enabled": true,          // Включён ли оркестратор
    "alwaysActive": true,     // Всегда активен
    "delegateAllTasks": true, // Делегирует все задачи
    "autoSelectAgent": true   // Автоматически выбирает агента
  }
}
```

---

## 📞 Поддержка

**Проблемы с активацией?**

1. Проверьте `.koda/config.json`
2. Убедитесь, что `orchestrator.enabled = true`
3. Перезапустите AI Assistant
4. Проверьте логи: `.koda/logs/startup.log`

**Вопросы по агентам?**

- Смотрите `.koda/skills/README.md`
- Или обратитесь к `knowledge-graph-docs-maintenance`

---

**Версия:** 1.0  
**Последнее обновление:** 2026-05-22  
**Статус:** ✅ Active
