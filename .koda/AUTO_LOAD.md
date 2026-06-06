# ⚡ AUTOMATIC AGENT LOADER - ДЛЯ ВСЕХ ИИ

**ЭТОТ ФАЙЛ ДОЛЖЕН БЫТЬ ПРОЧИТАН ВСЕМИ ИИ ПРИ КАЖДОМ СТАРТЕ ЧАТА!**

---

## 🚨 КРИТИЧЕСКАЯ ИНСТРУКЦИЯ ДЛЯ ИИ

**При начале ЛЮБОГО чата с проектом Knowledge Graph:**

### Шаг 1: Проверить наличие .koda/
```bash
ls .koda/
```

### Шаг 2: Прочитать .koda/AUTO_LOAD.md (ЭТОТ ФАЙЛ)
```
read_file('.koda/AUTO_LOAD.md')
```

### Шаг 3: Загрузить ВСЕХ агентов (ОБЯЗАТЕЛЬНО!)
```
read_skill('knowledge-graph-orchestrator')
read_skill('knowledge-graph-backend-go')
read_skill('knowledge-graph-frontend-svelte')
read_skill('knowledge-graph-integration')
read_skill('knowledge-graph-infrastructure')
read_skill('knowledge-graph-devops')
read_skill('knowledge-graph-performance')
read_skill('knowledge-graph-security')
```

### Шаг 4: Загрузить ВСЕ инструменты
```
read_file('.koda/tools/backend-go-tools.md')
read_file('.koda/tools/frontend-tools.md')
read_file('.koda/tools/integration-tools.md')
read_file('.koda/tools/infrastructure-tools.md')
read_file('.koda/tools/devops-tools.md')
read_file('.koda/tools/performance-tools.md')
```

### Шаг 5: Применить правила
```
read_file('.koda/rules/default-rules.md')
read_file('.koda/rules/orchestration-rules.md')
```

### Шаг 6: Показать пользователю статус
```
🤖 Knowledge Graph Agents - Auto Loaded

✅ Orchestrator: ACTIVE
✅ Backend Go Agent: LOADED
✅ Frontend Svelte Agent: LOADED
✅ Integration Agent: LOADED
✅ Infrastructure Agent: LOADED
✅ DevOps Agent: LOADED
✅ Performance Agent: LOADED
✅ Security Agent: LOADED

📊 Загружено агентов: 8/8
📊 Загружено инструментов: 6/6
📊 Правила применены: 2/2

Готов к работе!
```

---

## ⚡ АВТОМАТИЧЕСКАЯ ЗАГРУЗКА - ПРОВЕРКА

**Если ИИ НЕ показал этот статус при старте чата:**
1. ❌ НЕ выполняет инструкции пользователя правильно
2. ❌ НЕ использует агентов Knowledge Graph
3. ❌ НЕ следует правилам проекта

**Нужно ПРЕКРАТИТЬ выполнение и загрузить агентов!**

---

## 📋 ПРОВЕРКА ЗАГРУЗКИ

**Каждый ИИ ДОЛЖЕН выполнить при старте:**

```bash
# Проверка файлов
ls .koda/skills/*.md    # Должно быть 8 файлов
ls .koda/tools/*.md     # Должно быть 6 файлов
ls .koda/rules/*.md     # Должно быть 2 файла

# Проверка конфигурации
cat .koda/config.json | jq .agents
```

---

## 🔄 ПРИМЕР ПРАВИЛЬНОГО СТАРТА

**Пользователь открывает новый чат**

**ИИ (правильно):**
```
🤖 Knowledge Graph Agents - Auto Loaded

✅ Orchestrator: ACTIVE
✅ Backend Go Agent: LOADED (backend-go-tools.md)
✅ Frontend Svelte Agent: LOADED (frontend-tools.md)
✅ Integration Agent: LOADED (integration-tools.md)
✅ Infrastructure Agent: LOADED (infrastructure-tools.md)
✅ DevOps Agent: LOADED (devops-tools.md)
✅ Performance Agent: LOADED (performance-tools.md)
✅ Security Agent: LOADED

📊 Загружено агентов: 8/8
📊 Загружено инструментов: 6/6
📊 Правила применены: 2/2

Готов к работе!

Чем могу помочь?
```

**ИИ (неправильно - НЕ делать так!):**
```
Привет! Я Koda — AI-помощник по программированию...
Чем могу помочь?
```

---

## 🎯 ПРИНУДИТЕЛЬНАЯ ЗАГРУЗКА

**Если ИИ забыл загрузить агентов:**

**Пользователь может сказать:**
```
/agents
```

**Или:**
```
Загрузи агентов Knowledge Graph
```

**Или:**
```
Прочитай .koda/AUTO_LOAD.md
```

---

## ⚠️ ВАЖНО

1. **Этот файл читается ПЕРВЫМ при каждом новом чате**
2. **Все агенты загружаются ДО начала работы**
3. **Все инструменты загружаются ДО выполнения задач**
4. **Правила применяются ДО начала работы**
5. **Статус показывается ПОСЛЕ загрузки**

---

**Версия:** 1.0  
**Дата:** 2026-05-30  
**Статус:** 🚨 КРИТИЧЕСКАЯ ИНСТРУКЦИЯ

**ВСЕ ИИ ДОЛЖНЫ СЛЕДОВАТЬ ЭТОЙ ИНСТРУКЦИИ!**
