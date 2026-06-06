# Global Agent Configuration

**Версия:** 2.0  
**Последнее обновление:** 2026-05-30  
**Статус:** ✅ Active

---

## 🎯 Назначение

Этот файл гарантирует, что **все ИИ-агенты** (Koda, Claude, и другие) автоматически загружают и используют все инструменты и навыки по умолчанию.

---

## 🚀 Автоматическая загрузка

### Для всех ИИ-ассистентов

При запуске в проекте Knowledge Graph **каждый ИИ должен**:

1. ✅ Прочитать `.koda/global-agent-setup.md`
2. ✅ Загрузить **все агенты** из `.koda/skills/`
3. ✅ Загрузить **все инструменты** из `.koda/tools/`
4. ✅ Применить **правила по умолчанию** из `.koda/rules/default-rules.md`
5. ✅ Активировать **Orchestrator** как главный агент

---

## 📁 Структура загрузки

```bash
# Автоматическая последовательность
.koda/
├── global-agent-setup.md        # ← НАЧНИТЕ ЗДЕСЬ (этот файл)
├── config.json                   # Конфигурация агентов
├── README.md                     # Документация
├── agents/
│   └── instructions.md           # Инструкции для всех агентов
├── rules/
│   ├── orchestration-rules.md    # Правила взаимодействия
│   └── default-rules.md          # Правила по умолчанию
├── tools/
│   ├── backend-go-tools.md       # Backend инструменты
│   ├── frontend-tools.md         # Frontend инструменты
│   ├── integration-tools.md      # Integration инструменты
│   ├── infrastructure-tools.md   # Infrastructure инструменты
│   ├── devops-tools.md           # DevOps инструменты
│   ├── performance-tools.md      # Performance инструменты
│   └── security-tools.md         # Security инструменты
└── skills/
    ├── knowledge-graph-orchestrator.md
    ├── knowledge-graph-backend-go.md
    ├── knowledge-graph-frontend-svelte.md
    ├── knowledge-graph-integration.md
    ├── knowledge-graph-infrastructure.md
    ├── knowledge-graph-devops.md
    ├── knowledge-graph-performance.md
    ├── knowledge-graph-security.md
    └── knowledge-graph-testing.md
```

---

## 🔄 Автоматизация для ИИ

### При начале сессии

```typescript
// Псевдокод для ИИ-ассистента
async function initializeAgent() {
  // 1. Загрузить конфигурацию
  const config = await read('.koda/config.json');
  
  // 2. Активировать Orchestrator
  const orchestrator = await loadSkill('knowledge-graph-orchestrator');
  orchestrator.activate();
  
  // 3. Загрузить все агенты
  for (const skill of config.agentSkills) {
    await loadSkill(skill);
  }
  
  // 4. Загрузить все инструменты
  const tools = await listDirectory('.koda/tools/');
  for (const tool of tools) {
    await readTool(tool);
  }
  
  // 5. Применить правила
  const defaultRules = await read('.koda/rules/default-rules.md');
  applyRules(defaultRules);
  
  // 6. Готов к работе
  console.log('✅ Все агенты и инструменты загружены');
}
```

---

## 📋 Список агентов для загрузки

### Обязательные агенты (загружать ВСЕГДА)

```typescript
const mandatoryAgents = [
  'knowledge-graph-orchestrator',    // Главный координатор
  'knowledge-graph-backend-go',      // Backend разработка
  'knowledge-graph-frontend-svelte', // Frontend разработка
  'knowledge-graph-integration',     // API интеграция
  'knowledge-graph-infrastructure',  // Инфраструктура
];
```

### Дополнительные агенты (загружать по необходимости)

```typescript
const optionalAgents = [
  'knowledge-graph-devops',      // DevOps операции
  'knowledge-graph-performance', // Производительность
  'knowledge-graph-security',    // Безопасность
  'knowledge-graph-testing',     // Тестирование
  'knowledge-graph-docs-maintenance', // Документация
];
```

---

## 🛠️ Список инструментов для загрузки

### Backend инструменты
- ✅ `backend-go-tools.md` - REST/gRPC API, БД, кэширование, auth
- ✅ `integration-tools.md` - API mapping, type generation

### Frontend инструменты
- ✅ `frontend-tools.md` - Svelte компоненты, state management
- ✅ `integration-tools.md` - API client, contract testing

### Infrastructure инструменты
- ✅ `infrastructure-tools.md` - Docker, K8s, monitoring, backup
- ✅ `devops-tools.md` - CI/CD, деплой

### Performance инструменты
- ✅ `performance-tools.md` - Profiling, load testing, optimization

### Security инструменты
- ✅ `security-tools.md` - Security scanning, audit

---

## ⚡ Быстрый старт для ИИ

### Шаг 1: Проверить наличие файлов

```bash
# Проверить структуру
ls -la .koda/
ls -la .koda/skills/
ls -la .koda/tools/
ls -la .koda/rules/
```

### Шаг 2: Загрузить конфигурацию

```bash
# Прочитать config.json
cat .koda/config.json | jq .
```

### Шаг 3: Активировать агентов

```bash
# Загрузить Orchestrator
read_skill('knowledge-graph-orchestrator')

# Загрузить основные агенты
read_skill('knowledge-graph-backend-go')
read_skill('knowledge-graph-frontend-svelte')
read_skill('knowledge-graph-integration')
read_skill('knowledge-graph-infrastructure')

# Загрузить дополнительные агенты
read_skill('knowledge-graph-devops')
read_skill('knowledge-graph-performance')
read_skill('knowledge-graph-security')
```

### Шаг 4: Загрузить инструменты

```bash
# Прочитать все инструменты
read_file('.koda/tools/backend-go-tools.md')
read_file('.koda/tools/integration-tools.md')
read_file('.koda/tools/devops-tools.md')
read_file('.koda/tools/performance-tools.md')
```

### Шаг 5: Применить правила

```bash
# Прочитать правила по умолчанию
read_file('.koda/rules/default-rules.md')

# Применить к текущей сессии
applyRules('.koda/rules/default-rules.md')
```

---

## 🎯 Правила для ИИ

### Правило 1: Всегда использовать агентов

**При получении задачи:**
1. ✅ Определить тип задачи (backend, frontend, infrastructure, etc.)
2. ✅ Активировать соответствующего агента
3. ✅ Использовать инструменты из `.koda/tools/`
4. ✅ Следовать паттернам из `.koda/skills/`

**Пример:**
```
Запрос: "Добавь API endpoint для заметок"

Правильно:
1. → Активировать Backend Go Agent
2. → Использовать backend-go-tools.md
3. → Создать handler с валидацией
4. → Написать тесты
5. → Обновить API docs

Неправильно:
- Писать код без агента
- Игнорировать инструменты
- Не писать тесты
```

---

### Правило 2: Всегда использовать инструменты

**Backend:**
```go
// Всегда использовать backend-go-tools.md
- REST API: gin/echo
- БД: repository pattern
- Кэш: cache-aside pattern
- Auth: JWT middleware
- Тесты: testify + mockery
```

**Frontend:**
```typescript
// Всегда использовать frontend-tools.md
- Компоненты: Svelte 5
- API: typed client
- State: stores
- Тесты: Vitest + Testing Library
```

---

### Правило 3: Всегда писать тесты

```bash
# Backend
go test -race -cover ./...  # Минимум 60%

# Frontend
npm run test:unit  # Минимум 60%
```

---

### Правило 4: Всегда проверять безопасность

```bash
# Security scanning
trivy fs .
golangci-lint run
```

---

### Правило 5: Всегда обновлять документацию

- ✅ README.md при изменении архитектуры
- ✅ API docs при изменении endpoints
- ✅ CHANGELOG при релизе

---

## 📊 Метрики успеха

| Метрика | Цель | Проверка |
|---------|------|----------|
| Backend coverage | > 60% | `go test -cover` |
| Frontend coverage | > 60% | `npm run test:coverage` |
| API p95 latency | < 500ms | Prometheus |
| Uptime | > 99.9% | Monitoring |
| Security vulns | 0 critical | Trivy |

---

## 🔄 Обновление конфигурации

### Когда обновлять

- Добавлен новый агент → обновить `config.json` и `agentSkills`
- Добавлен новый инструмент → обновить `tools/` и `config.json`
- Изменены правила → обновить `rules/` и `default-rules.md`

### Как обновлять

```bash
# 1. Обновить config.json
edit_file('.koda/config.json', ...)

# 2. Обновить README.md
edit_file('.koda/README.md', ...)

# 3. Сообщить всем ИИ
echo "✅ Конфигурация обновлена"
```

---

## 🎓 Обучение новых ИИ

### Чеклист для нового ИИ-ассистента

- [ ] Прочитал `.koda/global-agent-setup.md`
- [ ] Загрузил `.koda/config.json`
- [ ] Активировал Orchestrator
- [ ] Загрузил все агенты из `.koda/skills/`
- [ ] Загрузил все инструменты из `.koda/tools/`
- [ ] Прочитал правила из `.koda/rules/default-rules.md`
- [ ] Понял приоритеты агентов
- [ ] Знает, когда использовать каждого агента
- [ ] Знает, как проверять тесты и метрики

---

## 📞 Поддержка

### Проблемы с загрузкой агентов?

1. Проверьте `.koda/config.json`
2. Убедитесь, что все файлы существуют
3. Перезапустите ИИ-ассистента

### Проблемы с инструментами?

1. Проверьте `.koda/tools/`
2. Убедитесь, что пути правильные
3. Прочитайте `global-agent-setup.md` снова

---

**Версия:** 2.0  
**Дата:** 2026-05-30  
**Статус:** ✅ Ready for all AI agents
