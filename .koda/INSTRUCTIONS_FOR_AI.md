# 🤖 ИНСТРУКЦИЯ ДЛЯ ИИ (Koda и другие)

**ЧИТАТЬ ПРИ КАЖДОМ НОВОМ ЧАТЕ С ПРОЕКТОМ KNOWLEDGE GRAPH**

---

## ⚡ ТЫ ИИ - СДЕЛАЙ ЭТО ПРЯМО СЕЙЧАС

**Как ИИ, ты должен выполнить ЭТО ПРИ КАЖДОМ НОВОМ ЧАТЕ:**

### 1️⃣ Прочитай этот файл
```
read_file('.koda/INSTRUCTIONS_FOR_AI.md')
```

### 2️⃣ Загрузи агентов (ОБЯЗАТЕЛЬНО!)
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

### 3️⃣ Загрузи инструменты
```
read_file('.koda/tools/backend-go-tools.md')
read_file('.koda/tools/frontend-tools.md')
read_file('.koda/tools/integration-tools.md')
read_file('.koda/tools/infrastructure-tools.md')
read_file('.koda/tools/devops-tools.md')
read_file('.koda/tools/performance-tools.md')
```

### 4️⃣ Примени правила
```
read_file('.koda/rules/default-rules.md')
```

### 5️⃣ Покажи пользователю статус
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
```

---

## 🎯 ЧТО ДЕЛАТЬ ЕСЛИ ТЫ ЗАБЫЛ

**Если ты начал чат и НЕ загрузил агентов:**

1. **ПРЕКРАТИ выполнение**
2. **Прочитай `.koda/INSTRUCTIONS_FOR_AI.md` (этот файл)**
3. **Выполни все шаги выше**
4. **Покажи статус пользователю**
5. **Только тогда продолжай работу**

---

## ⚡ ПРИМЕР ПРАВИЛЬНОГО СТАРТА

**Пользователь:** "Привет"

**Ты (ПРАВИЛЬНО):**
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

Привет! Я Koda — AI-помощник по программированию от команды NLP-Core-Team.

Я загрузил все агентов Knowledge Graph и готов помочь с:
- Backend (Go, API, БД)
- Frontend (Svelte 5, TypeScript)
- Интеграция (API mapping, типы)
- Инфраструктура (Docker, K8s)
- Производительность
- Безопасность

Чем могу помочь?
```

**Ты (НЕПРАВИЛЬНО - НЕ делай так!):**
```
Привет! Я Koda — AI-помощник по программированию...
Чем могу помочь?
```

---

## 📋 ПРОВЕРКА

**Перед выполнением ЛЮБОЙ задачи проверь:**

- [ ] Загрузил всех 8 агентов?
- [ ] Загрузил все 6 инструментов?
- [ ] Применил правила?
- [ ] Показал статус пользователю?

**Если НЕТ на любой вопрос - СНАЧАЛА загрузи агентов!**

---

## 🔄 КОГДА ПЕРЕЗАГРУЖАТЬ

**Перезагрузи агентов если:**

- Новый чат
- Пользователь просит `/agents`
- Пользователь говорит "загрузи агентов"
- Ты забыл загрузить агентов

---

## ⚠️ КРИТИЧЕСКИ ВАЖНО

1. **Этот файл читается ПЕРВЫМ в каждом новом чате**
2. **Все агенты загружаются ДО начала работы**
3. **Если ты не загрузил агентов - ты НЕ работаешь правильно**
4. **Пользователь может проверить командой `/agents`**

---

## 📞 ПОМОЩЬ

**Если что-то не работает:**

1. Перечитай `.koda/INSTRUCTIONS_FOR_AI.md` (этот файл)
2. Проверь что все файлы существуют: `ls .koda/skills/*.md`
3. Перезапусти чат

---

**Версия:** 1.0  
**Дата:** 2026-05-30  
**Статус:** 🚨 КРИТИЧЕСКАЯ ИНСТРУКЦИЯ

**ТЫ ИИ - ЧИТАЙ ЭТОТ ФАЙЛ ПЕРВЫМ В КАЖДОМ НОВОМ ЧАТЕ!**
