# Knowledge Graph AI Agents - Universal Setup

**Поддержка: Cursor + GitHub Copilot + Windsurf**

---

## 🎯 Overview

Эта конфигурация обеспечивает **автоматическую активацию оркестратора и 8 агентов** во всех AI-помощниках:

- ✅ **Cursor** — полная поддержка через `.cursor/rules/`
- ✅ **GitHub Copilot** — контекст через `.github/copilot-instructions.md`
- ✅ **Windsurf** — контекст через `.windsurf/rules.md`

---

## 📁 Структура файлов

```
knowledge-graph/
├── .cursor/
│   └── rules/
│       ├── knowledge-graph-orchestrator.md     # ⭐ Always active meta-agent
│       ├── knowledge-graph-performance.md      # Performance optimization
│       ├── knowledge-graph-security.md         # Security audit
│       └── knowledge-graph-devops.md           # DevOps/infrastructure
│
├── .github/
│   └── copilot-instructions.md                 # Copilot context
│
├── .windsurf/
│   └── rules.md                                # Windsurf rules
│
├── .koda/
│   ├── config.json                             # Koda configuration
│   └── skills/
│       └── knowledge-graph-*.md                # Koda agent definitions
│
└── docs/
    └── AGENTS_IMPLEMENTATION_COMPLETE.md       # Implementation report
```

---

## 🚀 Как использовать

### Cursor (Рекомендуется)

**Настройка:** Автоматическая (файлы уже созданы)

**Как работает:**
```
1. Открой проект в Cursor
2. Начни чат с AI
3. Оркестратор активируется автоматически (alwaysApply: true)
4. Просто опиши задачу — AI сам выберет агента
```

**Пример:**
```
User: "Оптимизируй загрузку графа"
Cursor AI: → Routes to performance agent → Returns optimization guide
```

**Файлы:**
- `.cursor/rules/knowledge-graph-orchestrator.md` — оркестратор
- `.cursor/rules/knowledge-graph-performance.md` — производительность
- `.cursor/rules/knowledge-graph-security.md` — безопасность
- `.cursor/rules/knowledge-graph-devops.md` — DevOps

---

### GitHub Copilot

**Настройка:** Автоматическая (файл уже создан)

**Как работает:**
```
1. Открой проект в VS Code с Copilot
2. Начни чат с Copilot
3. AI использует контекст из .github/copilot-instructions.md
4. Опиши задачу — AI учтёт правила агентов
```

**Пример:**
```
User: "Добавь API endpoint"
Copilot: → Uses backend agent context → Generates Go code
```

**Файл:**
- `.github/copilot-instructions.md` — контекст для всех агентов

**Ограничение:** Нет делегирования между агентами, только единый контекст

---

### Windsurf

**Настройка:** Автоматическая (файл уже создан)

**Как работает:**
```
1. Открой проект в Windsurf
2. Начни чат с AI
3. AI использует контекст из .windsurf/rules.md
4. Опиши задачу — AI учтёт правила агентов
```

**Пример:**
```
User: "Создай Kubernetes deployment"
Windsurf AI: → Uses devops agent context → Generates manifests
```

**Файл:**
- `.windsurf/rules.md` — правила для Windsurf

**Ограничение:** Нет делегирования между агентами, только единый контекст

---

## 🤖 Все агенты (8 total)

| # | Агент | Назначение | Поддержка |
|---|-------|------------|-----------|
| 0 | **knowledge-graph-orchestrator** | Мета-агент координатор | ✅ Cursor (полная) |
| 1 | knowledge-graph-frontend-svelte | Frontend (Svelte 5) | ✅ Все |
| 2 | knowledge-graph-backend-go | Backend (Go) | ✅ Все |
| 3 | knowledge-graph-docs-maintenance | Документация | ✅ Все |
| 4 | knowledge-graph-testing | Тестирование | ✅ Все |
| 5 | knowledge-graph-integration | API интеграция | ✅ Все |
| 6 | **knowledge-graph-performance** | **Производительность** | ✅ Все |
| 7 | **knowledge-graph-security** | **Безопасность** | ✅ Все |
| 8 | **knowledge-graph-devops** | **DevOps** | ✅ Все |

---

## 📊 Сравнение поддержки

| Инструмент | Делегирование | Автоактивация | Контекст | Рекомендация |
|------------|---------------|---------------|----------|--------------|
| **Cursor** | ✅ Полное | ✅ Да | ✅ Отдельные файлы | ⭐ **Лучший выбор** |
| **Copilot** | ⚠️ Частичное | ⚠️ Через контекст | ✅ Единый файл | Хорошо для VS Code |
| **Windsurf** | ⚠️ Частичное | ⚠️ Через контекст | ✅ Единый файл | Хорошо для VS Code |

---

## 🎯 Примеры использования

### Пример 1: Оптимизация

```
Запрос: "Оптимизируй bundle size на 50%"

Cursor:
  → Оркестратор анализирует → performance task
  → Делегирует: knowledge-graph-performance
  → Результат: Code splitting + lazy loading guide

Copilot/Windsurf:
  → AI использует контекст performance agent
  → Результат: Аналогичный совет
```

### Пример 2: Безопасность

```
Запрос: "Проведи аудит безопасности auth"

Cursor:
  → Оркестратор анализирует → security task
  → Делегирует: knowledge-graph-security
  → Результат: Полный security audit report

Copilot/Windsurf:
  → AI использует контекст security agent
  → Результат: Security checklist
```

### Пример 3: DevOps

```
Запрос: "Создай Kubernetes deployment"

Cursor:
  → Оркестратор анализирует → devops task
  → Делегирует: knowledge-graph-devops
  → Результат: Полные K8s manifests

Copilot/Windsurf:
  → AI использует контекст devops agent
  → Результат: Deployment manifests
```

---

## 🔧 Настройка (если нужно)

### Cursor

```bash
# Файлы уже созданы в .cursor/rules/
# Просто открой проект в Cursor

# Проверка:
ls -la .cursor/rules/
```

### Copilot

```bash
# Файл уже создан в .github/copilot-instructions.md
# Просто открой проект в VS Code с Copilot

# Проверка:
ls -la .github/copilot-instructions.md
```

### Windsurf

```bash
# Файл уже создан в .windsurf/rules.md
# Просто открой проект в Windsurf

# Проверка:
ls -la .windsurf/rules.md
```

---

## ✅ Проверка

### Cursor

```bash
# 1. Открой проект в Cursor
# 2. Нажми Cmd+K (или Ctrl+K)
# 3. Введи: "Оптимизируй загрузку графа"
# 4. Ожидай: AI использует performance agent
```

### Copilot

```bash
# 1. Открой проект в VS Code
# 2. Нажми Cmd+Shift+P → "Copilot: New Chat"
# 3. Введи: "Оптимизируй загрузку графа"
# 4. Ожидай: AI учтёт performance контекст
```

### Windsurf

```bash
# 1. Открой проект в Windsurf
# 2. Начни чат с AI
# 3. Введи: "Оптимизируй загрузку графа"
# 4. Ожидай: AI учтёт performance контекст
```

---

## 📚 Документация

### Основные документы

- **Cursor Rules:** `.cursor/rules/knowledge-graph-orchestrator.md`
- **Copilot Instructions:** `.github/copilot-instructions.md`
- **Windsurf Rules:** `.windsurf/rules.md`

### Дополнительная документация

- **Main Guide:** `.koda/README.md`
- **Implementation Report:** `docs/AGENTS_IMPLEMENTATION_COMPLETE.md`
- **Commands:** `COMMANDS.md`

---

## 🎯 Рекомендации

### Для максимальной эффективности

**Используй Cursor** — он поддерживает:
- ✅ Отдельные файлы для каждого агента
- ✅ Автоматическое делегирование
- ✅ Полную автоактивацию оркестратора

### Для VS Code

**Используй Copilot или Windsurf** — они поддерживают:
- ✅ Единый контекст всех агентов
- ✅ Интеграцию в VS Code
- ⚠️ Ограниченное делегирование (через контекст)

---

## 🚀 Быстрый старт

### 1. Выбери AI-помощник

- **Cursor** → Лучший опыт (рекомендуется)
- **Copilot** → Хороший опыт в VS Code
- **Windsurf** → Хороший опыт в VS Code

### 2. Открой проект

```bash
# Cursor
cursor .

# VS Code (с Copilot)
code .

# Windsurf
windsurf .
```

### 3. Начни чат

```
Просто опиши задачу:
"Оптимизируй загрузку графа"
"Создай API endpoint"
"Проведи аудит безопасности"
```

### 4. Получи результат

AI автоматически использует нужного агента!

---

## 📞 Поддержка

**Вопросы по агентам?**
- См. `.koda/README.md`
- Или `docs/AGENTS_IMPLEMENTATION_COMPLETE.md`

**Проблемы с настройкой?**
- Проверь, что все файлы созданы
- Перезапусти AI-помощник
- Убедись, что проект открыт в правильной директории

---

## 🎉 Summary

**Создано:**
- ✅ 4 файла для Cursor (оркестратор + 3 агента)
- ✅ 1 файл для Copilot (контекст всех агентов)
- ✅ 1 файл для Windsurf (контекст всех агентов)
- ✅ Полная документация

**Результат:**
- ✅ Оркестратор всегда активен во всех инструментах
- ✅ 8 агентов доступны в любом AI-помощнике
- ✅ Автоматическое делегирование (Cursor)
- ✅ Единый контекст (Copilot/Windsurf)

**Готово к использованию!** 🚀

---

**Last Updated:** 2026-05-27  
**Version:** 1.0  
**Status:** ✅ Complete
