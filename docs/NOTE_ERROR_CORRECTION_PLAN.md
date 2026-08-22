# План функционала исправления ошибок и дополнения заметок

**Создано:** 29 июля 2026 г.  
**Статус:** ⏳ Запланировано

## Обзор

Документ описывает план реализации функционала исправления ошибок и автодополнения заметок, аналогичного QuickCaptureWidget для dust-заметок, но для полноценных заметок.

## Аналогия с QuickCaptureWidget (Dust)

**QuickCaptureWidget (dust):**
- ✅ Быстрое создание заметок через ✨ кнопку или Ctrl+Shift+N
- ✅ Автообрезание заголовка при превышении лимита
- ✅ Тип заметки автоматически устанавливается в "dust"
- ✅ Минимальный UI (только textarea + кнопка сохранения)
- ✅ Быстрый захват мысли без необходимости подробного заполнения

## Требуемый функционал

### 1. Исправление ошибок в заметках

**Сценарий:** Пользователь заметил ошибку в существующей заметке и хочет быстро её исправить без перехода на страницу редактирования.

**Реализация:**
```svelte
// frontend/src/components/organisms/QuickEditWidget.svelte

- Быстрое редактирование заголовка/контента
- Контекстное меню или быстрая клавиша (Ctrl+E)
- Inline редактирование без перехода на отдельную страницу
- Автосохранение при потере фокуса
- Показывать diff изменений
```

**UI варианты:**
- **Inline редактирование** — клик по заголовку/контенту → прямое редактирование
- **Quick Edit Modal** — лёгкая модалка для быстрого исправления
- **Side Panel Quick Edit** — правая панель с быстрым редактированием

### 2. Автодополнение заметок

**Сценарий:** При создании заметки система автоматически предлагает дополнения на основе контента.

**Реализация:**
```typescript
// frontend/src/features/autocompletion/NoteAutocompletion.ts

- Автосуггестии заголовка на основе контента
- Автосуггестии типа заметки (star, planet, comet и т.д.)
- Автосуггестии тегов/ключевых слов
- Автосуггестии связей с существующими заметками
- Автосуггестии продолжения контента
```

**Интеграция с NLP:**
- Использовать существующий NLP service
- Анализ контента в реальном времени
- Предлагать релевантные дополнения

### 3. Обогащение Dust-заметок (DustInboxPanel)

**Сценарий:** Dust-заметки собираются в специальной панели и могут быть "обогащены" до полноценных заметок.

**Реализация:**
```svelte
// frontend/src/features/dust/DustInboxPanel.svelte

- Панель для dust-заметок (как входящие)
- Группировка dust-заметок по темам
- Автоанализ и предложения типа (star, planet, comet)
- Автосоздание связей между related dust-заметками
- Конвертация dust → полноценная заметка в один клик
```

**Обогащение dust-заметок:**
- Предложение типа на основе контента
- Предложение заголовка (если короткий)
- Предложение тегов
- Автосоздание связей с похожими dust-заметками

### 4. Исправление ошибок NLP

**Сценарий:** NLP service может ошибаться в анализе (ключевые слова, эмбеддинги), пользователь должен иметь возможность исправить.

**Реализация:**
```typescript
// frontend/src/features/nlp/NLPCorrectionPanel.svelte

- Показать ключевые слова, предложенные NLP
- Позволить пользователю редактировать их
- Показать similarity score с другими заметками
- Позволить пользователю вручную создать связи
- Пересчёт эмбеддингов по требованию
```

## План реализации

### Phase 1: Quick Edit Widget

1. **Компонент QuickEditWidget**
   - Файл: `frontend/src/components/organisms/QuickEditWidget.svelte`
   - Быстрое редактирование заголовка/контента
   - Inline режим или модалка
   - Горячая клавиша: Ctrl+E

2. **Интеграция с заметками**
   - Клик по заголовку → inline редактирование
   - Клик по контенту → inline редактирование
   - Автосохранение при blur

3. **API integration**
   - PATCH /api/v1/notes/{id} для быстрого обновления
   - Оптимистическое обновление UI
   - Откат при ошибке

### Phase 2: Autocompletion Engine

1. **NLP autocompletion service**
   - Файл: `nlp-service/autocompletion.py`
   - API: `/autocomplete/title`, `/autocomplete/type`, `/autocomplete/tags`
   - Анализ контента в реальном времени

2. **Frontend autocompletion UI**
   - Файл: `frontend/src/features/autocompletion/NoteAutocompletion.svelte`
   - Dropdown suggestions при вводе
   - Ctrl+Space для триггера
   - Предварительный просмотр дополнений

3. **Интеграция с CreateNoteModal**
   - Добавить autocompletion в форму создания
   - Показывать предложения при вводе
   - One-click принятие предложений

### Phase 3: DustInboxPanel

1. **Компонент DustInboxPanel**
   - Файл: `frontend/src/features/dust/DustInboxPanel.svelte`
   - Список dust-заметок с группировкой
   - Кнопки "Обогатить" и "Конвертировать"
   - Предложения типа и связей

2. **Worker для обогащения dust**
   - Файл: `backend/internal/infrastructure/queue/tasks/dust_enrichment.go`
   - Анализ dust-заметок пачками
   - Предложение типа и связей
   - Автосоздание связей между related dust

3. **API для dust-обогащения**
   - POST /api/v1/dust/enrich
   - GET /api/v1/dust/suggestions
   - PATCH /api/v1/dust/{id}/enrich

### Phase 4: NLP Correction Panel

1. **Компонент NLPCorrectionPanel**
   - Файл: `frontend/src/features/nlp/NLPCorrectionPanel.svelte`
   - Показать NLP результаты (keywords, embeddings)
   - Ручное редактирование ключевых слов
   - Ручное создание связей на основе similarity
   - Пересчёт эмбеддингов

2. **Worker для пересчёта NLP**
   - Файл: `backend/internal/infrastructure/queue/tasks/nlp_recalc.go`
   - Пересчёт embeddings по требованию
   - Пересчёт keywords
   - Обновление рекомендаций

## UI/UX Design

### Quick Edit Widget

**Вариант 1: Inline**
```
┌─────────────────────────────────┐
│ 📝 React Hooks Tutorial        │ ← клик → редактируется
│ Hooks are functions that...     │ ← клик → редактируется
└─────────────────────────────────┘
```

**Вариант 2: Quick Edit Modal**
```
┌─────────────────────────────────┐
│ Quick Edit                    [×]│
├─────────────────────────────────┤
│ Title: [React Hooks Tutorial    ]│
│ Content: [Hooks are functions...]│
│                               [Save]│
└─────────────────────────────────┘
```

### Autocompletion

**Dropdown suggestions:**
```
┌─────────────────────────────────┐
│ Title: [React Hooks           ]  │
│         ↓ suggestions           │
│         • React Hooks Usage    │
│         • Best Practices      │
│         • Common Mistakes    │
└─────────────────────────────────┘
```

### DustInboxPanel

**Layout:**
```
┌─────────────────────────────────┐
│ Dust Inbox                    [×]│
├─────────────────────────────────┤
│ 📁 Programming                │
│   • React Hooks (dust)       [Enrich]│
│   • CSS Tricks (dust)         [Enrich]│
│                                 │
│ 📁 Anime                       │
│   • Shingeki no Kyō (dust)    [Enrich]│
└─────────────────────────────────┘
```

## Приоритет реализации

1. **P0 (Критический):** Quick Edit Widget — быстрые исправления
2. **P1 (Высокий):** Autocompletion Engine — помощь при создании
3. **P2 (Средний):** DustInboxPanel — обогащение dust-заметок
4. **P3 (Низкий):** NLP Correction Panel — расширенное управление

## Технические требования

### Performance

- Autocompletion: <200ms latency
- Quick Edit: оптимистическое обновление UI
- DustInbox: batch обработка, не замедлять UI

### UX Requirements

- Не прерывать рабочий процесс пользователя
- Минимум кликов для действий
- Интуитивные горячие клавиши
- Отмена действий (undo/redo)

### Integration Points

- NLP service для autocompletion
- Worker для background задач
- Existing recommendation system для связей
- Existing keyword/embedding extraction

## Критерии успеха

- [ ] Quick Edit Widget позволяет быстро исправлять ошибки
- [ ] Autocompletion помогает при создании заметок
- [ ] DustInboxPanel позволяет обогащать dust-заметки
- [ ] NLP результаты можно корректировать вручную
- [ ] Все функции работают асинхронно, не блокируя UI
- [ ] E2E тесты покрывают основные сценарии
- [ ] Производительность в норме (<200ms для autocompletion)

## Связанные задачи

- Документация: `docs/LINK_TYPES.md` — типы связей для автосоздания
- Код: `frontend/src/components/organisms/QuickCaptureWidget.svelte` — аналог для dust
- Код: `backend/internal/infrastructure/queue/tasks/` — worker задачи
- Система рекомендаций: `docs/RECOMMENDATION_ARCHITECTURE.md` — для автосоздания связей