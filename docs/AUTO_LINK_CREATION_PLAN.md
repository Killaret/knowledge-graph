# План автоматического создания связей на основе NLP

**Создано:** 29 июля 2026 г.  
**Статус:** ⏳ Запланировано

## Обзор

Документ описывает план реализации автоматического создания связей между заметками на основе NLP-анализа (семантическая похожесть, ключевые слова, графовая структура).

## Текущее состояние

### ✅ Уже реализовано

**Расчёт веса связей по трём компонентам:**
- **Alpha (α)** — вес графового компонента (связи между заметками)
- **Beta (β)** — вес семантического компонента (эмбеддинги/похожесть контента)
- **Gamma (γ)** — вес keyword компонента (сходство ключевых слов)

**Файлы реализации:**
- `backend/internal/domain/graph/aggregation.go` — агрегация весов
- `backend/internal/domain/graph/traversal_service.go` — traversal с тремя компонентами
- `backend/internal/domain/graph/suggestion_result.go` — результаты с компонентами
- `backend/internal/domain/link/entity.go` — `NewGammaLink` для рекомендаций
- `backend/internal/application/recommendation/keyword_similarity.go` — стратегий keyword похожести

**Формула агрегации:**
```go
total = (alpha*graphWeight + beta*semanticWeight + gamma*keywordWeight) / (alpha + beta + gamma)
```

**Конфигурация:**
```json
{
  "recommendation_alpha": 0.5,
  "recommendation_beta": 0.3,
  "recommendation_gamma": 0.2
}
```

### ❌ Не реализовано

**Автоматическое создание связей:**
- Нет задачи для автоматического создания связей на основе рекомендаций
- Связи создаются только вручную через UI
- Worker не создаёт связи даже при высоких score рекомендаций

## Требуемая функциональность

### 1. Автоматическое создание связей при создании заметки

**Сценарий:** При создании новой заметки автоматически создавать связи с похожими заметками если score выше порога.

**Реализация:**
```go
// backend/internal/infrastructure/queue/tasks/auto_link_creation.go

const TypeAutoCreateLinks = "auto:link:creation"

type AutoCreateLinksPayload struct {
    NoteID uuid.UUID `json:"note_id"`
    Threshold float64 `json:"threshold"` // порог score для создания связи
    MaxLinks int `json:"max_links"` // максимальное количество связей
}

func Handler(ctx context.Context, t *asynq.Task) error {
    // 1. Получить рекомендации для заметки
    // 2. Отфильтровать по threshold
    // 3. Создать связи для top MaxLinks
    // 4. SourceType = "gamma" (автоматически созданные)
}
```

### 2. Пороговые значения для автоматического создания

**Конфигурация:**
```json
{
  "auto_link_creation_enabled": true,
  "auto_link_threshold": 0.7, // создавать связи если score >= 0.7
  "auto_link_max_per_note": 5, // максимум 5 связей на заметку
  "auto_link_types": ["related", "reference"] // какие типы связей создавать
}
```

### 3. Типы связей для автоматического создания

**Правила выбора типа связи:**
- **Reference** — если семантическая похожесть очень высокая (score >= 0.9)
- **Related** — если графовая структура или ключевые слова похожи (score >= 0.7)
- **Dependency** — если есть явная иерархия (одна заметка содержит термин из другой)

### 4. Веса для автоматически созданных связей

**Расчёт веса на основе recommendation score:**
```go
weight = recommendation_score * 0.9 // умножаем на 0.9 чтобы не было 1.0
```

### 5. Enqueue задачи при создании заметки

**Модификация note_handler:**
```go
// backend/internal/interfaces/api/notehandler/note_handler.go

func (h *Handler) CreateNote(c *gin.Context) {
    // ... создание заметки ...
    
    // Enqueue auto link creation
    if h.autoLinkCreationEnabled {
        _ = h.taskQueue.EnqueueAutoCreateLinks(
            c.Request.Context(),
            newNote.ID(),
            h.autoLinkThreshold,
            h.autoLinkMaxPerNote,
        )
    }
}
```

### 6. UI для управления автоматическим созданием связей

**Настройки пользователя:**
- Включить/выключить автоматическое создание связей
- Настроить порог score
- Настроить максимальное количество связей
- Выбрать предпочтительные типы связей

## План реализации

### Phase 1: Backend (Worker + Tasks)

1. **Создать задачу AutoCreateLinks**
   - Файл: `backend/internal/infrastructure/queue/tasks/auto_link_creation.go`
   - Handler: анализ рекомендаций и создание связей
   - Enqueue метод в AsynqClient

2. **Добавить конфигурацию**
   - Файл: `backend/internal/config/config.go`
   - Параметры: enabled, threshold, max_links, link_types

3. **Интеграция с note_handler**
   - Enqueue задачи при создании заметки
   - Проверка флага enabled

### Phase 2: Frontend (UI)

1. **Настройки в профиле пользователя**
   - Форма для настройки автоматического создания связей
   - Слайдеры для threshold и max_links
   - Чекбоксы для типов связей

2. **Индикация автоматически созданных связей**
   - Разный цвет/стиль для gamma-связей
   - Tooltip: "Автоматически созданная связь на основе похожести"

### Phase 3: Testing

1. **Unit тесты**
   - Тесты задачи AutoCreateLinks
   - Тесты расчёта весов
   - Тесты выбора типа связи

2. **Integration тесты**
   - Создание заметки → проверка автоматических связей
   - Проверка пороговых значений
   - Проверка типов связей

3. **E2E тесты**
   - Создание заметки через UI
   - Проверка что связи появились на графе
   - Проверка что связи имеют правильный тип и вес

## Технические детали

### Источник данных для автоматического создания

**Использовать существующую систему рекомендаций:**
- `note_recommendations` таблица с precomputed scores
- TraversalService.GetSuggestions
- Компоненты: alpha (graph), beta (semantic), gamma (keywords)

### Предотвращение дубликатов

**Уникальность связей:**
- Уникальный constraint: `(source_note_id, target_note_id, link_type)`
- Проверка существования связи перед созданием
- Обновление веса если связь уже существует

### Производительность

**Оптимизации:**
- Batch создание связей (INSERT INTO ... ON CONFLICT UPDATE)
- Limit на количество связей (max_links)
- Background обработка через Asynq
- Deduplication задач (TaskID)

## Риски и проблемы

### 1. Слишком много автоматических связей

**Решение:**
- Консервативный порог (threshold = 0.7)
- Limit на количество связей (max_links = 5)
- Только для новых заметок, не для всех

### 2. Неверные типы связей

**Решение:**
- Консервативный выбор (только "related" по умолчанию)
- Ручная корректировка пользователем
- Возможность отключить автоматическое создание

### 3. Производительность

**Решение:**
- Already async через Asynq
- Batch операции
- Не для всех заметок (new notes only)

## Критерии успеха

- [ ] Автоматическое создание связей работает для новых заметок
- [ ] Связи создаются только если score >= threshold
- [ ] Типы связей выбираются корректно
- [ ] Веса связей рассчитываются на основе recommendation score
- [ ] SourceType = "gamma" для автоматических связей
- [ ] Пользователь может управлять настройками через UI
- [ ] Нет дубликатов связей
- [ ] Производительность приемлема (задача выполняется < 5 сек)
- [ ] E2E тесты покрывают основные сценарии

## Связанные задачи

- Документация: `docs/RECOMMENDATION_ARCHITECTURE.md` — система рекомендаций
- Документация: `docs/LINK_TYPES.md` — типы связей
- Код: `backend/internal/domain/graph/aggregation.go` — агрегация весов
- Код: `backend/internal/domain/graph/traversal_service.go` — traversal с компонентами