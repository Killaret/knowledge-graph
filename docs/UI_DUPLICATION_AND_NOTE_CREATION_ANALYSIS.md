# Анализ дублирования в UI и точек создания заметки

**Дата:** 2026-07-21  
**Статус:** действующая заметка, не план рефакторинга  
**Контекст:** ручное тестирование публичного графа показало, что на странице одновременно присутствуют несколько похожих элементов управления и как минимум два разных способа создать заметку. Этот документ фиксирует находки и даёт рекомендации. Сами способы создания заметки пока оставляем как есть.

## 1. Что видит пользователь

На скриншоте/выборе элементов видны:

- Верхняя плавающая панель `floating-controls`:
  - переключатель вида (`2D / 3D / List`);
  - фильтры по типам (`All`, `Stars`, `Planets`, ...);
  - поле поиска (`search-input`);
  - кнопка входа (для анонима);
  - меню (`import / export`);
  - переключатель языка;
  - **кнопка создания заметки** (`create-note-button`, `+`).
- Левый overlay графа `graph-controls`:
  - `graph-controls-reset`;
  - `graph-controls-search`;
  - `graph-controls-mode`;
  - `graph-controls-focus`.
- Canvas графа (`graph-canvas`).
- Открытая призрачная форма заметки (`ghost-note-form`) с кнопкой закрытия `ghost-note-close`.
- Кнопка сворачивания боковой панели (`toggle-button`).

Пользователь отметил, что видит дублирование и **два способа создать заметку**. Мы их оставляем, но разбираем.

## 2. Потоки создания заметки

### 2.1. Поток A — плавающая кнопка `+` / модальное окно

| Атрибут | Значение |
|---------|----------|
| UI-элемент | `data-testid="create-note-button"` в `FloatingControls.svelte` |
| Компонент | `frontend/src/components/organisms/FloatingControls.svelte` |
| Обработка в роуте | `frontend/src/routes/+page.svelte` (`onCreate={() => (showCreateModal = true)}`) |
| Модалка | `frontend/src/components/organisms/CreateNoteModal.svelte` |
| Тип по умолчанию | `CelestialBody.PLANET.type` → `"planet"` |
| Создание | `CreateNoteModal` сам вызывает `createNote(...)` из `$shared/api/notes` |
| После успеха | `onSuccess` → `handleNoteCreated(note)` в `+page.svelte` → `showCreateModal = false`, `graphStore.selectedNodeId = note.id`, `refreshAfterMutation()` |
| Empty-state fallback | В `+page.svelte` есть `<button class="new-note-button" onclick={() => (showCreateModal = true)}>` — та же модалка |

### 2.2. Поток B — горячая клавиша `N` / призрачная форма

| Атрибут | Значение |
|---------|----------|
| Триггер | Клавиша `N` (без модификаторов) или кнопка в canvas-логике |
| Хоткей | `frontend/src/features/graph-interaction/hotkeys.ts:104` |
| Позиционирование | `frontend/src/features/graph-interaction/event-bridge.ts:532` (центр canvas) |
| Форма | `frontend/src/features/graph-ui/modals.svelte` (`data-testid="ghost-note-form"`) |
| Закрытие | `data-testid="ghost-note-close"` |
| Состояние | `frontend/src/features/graph-forms/note-form.ts` |
| Тип по умолчанию | `"planet"` |
| Создание | `createNote(noteFormState, { onNoteCreate, onFormClose })` → `+page.svelte` передаёт `handleNoteCreate(data)` |
| После успеха | `handleNoteCreate` → `createNote(data)` → `refreshAfterMutation()` |

### 2.3. Различия и риски

- **Разный дефолтный тип:** исправлено — оба потока теперь используют `planet` по умолчанию.
- **Разный UI:** исправлено — обе формы теперь используют компонент `TypeSelector` (призрачная форма в `features/graph-ui/modals.svelte` рендерит `NoteForm`).
- **Обработчик ошибок:** унифицирован — обе формы показывают ошибку через `ApiErrorDisplay` внутри `NoteForm`.
- **Разный уровень состояния:** модалка владеет собственным `title/content/type` и вызывает API сама; призрачная форма использует `noteFormState` и прокидывает данные в родителя.

### 2.4. Почему оба остаются

Поток A удобен для мыши и для новичков, поток B — для горячих клавиш и «power users». Это распространённый UX-паттерн. Пока продукт не решит иначе, оба остаются.

## 3. Дублирование поиска

### 3.1. Поиск в верхней панели

- Элемент: `data-testid="search-input"` (`FloatingControls.svelte`).
- Логика: `FloatingControls.handleSearch()` → `onSearch?.(q.value)` → `+page.svelte` обновляет `filterState.searchQuery`.
- Результат: фильтрует **и** список заметок, **и** узлы на графе (`filteredNotes`, `filteredGraphData`).

### 3.2. Поиск в левом overlay

- Элемент: `data-testid="graph-controls-search"` (`features/graph-ui/controls.svelte`).
- Логика: `onSearch` → открывает `hotkeysState.showSearchBox` → `features/graph-ui/overlay.svelte` (`data-testid="search-box"`).
- Результат: ищет узел по заголовку, **не фильтрует** граф, а перемещает вид на следующее совпадение (`focusNextSearchMatch` в `features/graph-interaction/hotkeys.ts`).

### 3.3. Проблема

Обе кнопки имели одинаковую иконку лупы и обычно переводятся как «поиск», но поведение разнится. Это может путать пользователя.

### 3.4. Решение

- Левый поиск (graph controls) переименован в «Найти на графе» (`Find on graph (F)`), иконка изменена с `🔍` на `🎯`.
- Верхний поиск в `FloatingControls` остаётся фильтром/поиском по заметкам.

## 4. Дублирование на уровне кода

### 4.1. Продублированные обработчики

| Обработчик | `frontend/src/routes/+page.svelte` | `frontend/src/routes/graph/+page.svelte` | Отличия |
|------------|------------------------------------|------------------------------------------|---------|
| `handleNoteCreate` | строка 419 | строка 202 | `refreshAfterMutation()` vs `loadGraphData({ nocache: true })`; `alert()` vs `console.error()` |
| `handleNoteDelete` | ~строка 375 | ~строка 170 | аналогично |
| `handleNoteRestore` | ~строка 390 | ~строка 185 | аналогично |
| `handleLinkCreate` | ~строка 395 | ~строка 213 | `createLink(...)` и разные функции обновления графа |

Это повторяющийся CRUD-код. В долгосрочной перспективе его стоит вынести в shared-хелпер (`$shared/services/graph-mutations.ts` или `$features/graph-canvas/graph-mutations.ts`).

### 4.2. Разные функции загрузки графа

- `+page.svelte`: `refreshAfterMutation()` использует `getGraphWithPreload({ force: true })` и `PreloadService`.
- `graph/+page.svelte`: `loadGraphData({ nocache: true })` использует `getFullGraphData({ skipCache: true })`.

Результат похожий, но реализации разные. Это усложняет поддержку и тестирование.

### 4.3. FSD-нарушение

`frontend/src/components/organisms/GraphCanvas.svelte` импортирует из `features/*`:

```ts
import { GraphCanvasOverlay, GraphCanvasModals, GraphCanvasControls } from "$features/graph-ui";
import { createDragDropState, ... } from "$features/graph-interaction/drag-and-drop";
import { createHotkeysState, ... } from "$features/graph-interaction/hotkeys";
import { createZoomPanState, ... } from "$features/graph-interaction/zoom-pan";
import { attachEvents, ... } from "$features/graph-interaction/event-bridge";
import { createGraphCanvasState, ... } from "$features/graph-canvas/canvas-state.svelte";
import { createNoteFormState, createNote, closeNoteForm } from "$features/graph-forms/note-form";
import { createLinkFormState, createLink, closeLinkForm } from "$features/graph-forms/link-form";
import { createLayoutProvider, toRuntimeConfig } from "$features/graph-3d";
```

Согласно правилам проекта, `components/organisms/` может импортировать только `components/atoms/`, `components/molecules/` и `shared/`. `GraphCanvas.svelte` фактически является виджетом и должен жить в `widgets/graph-canvas/` или в `features/graph-canvas/`, чтобы иметь право импортировать другие `features`.

Это не причина дублирования напрямую, но усложняет рефакторинг и делает дублирование менее очевидным.

## 5. Дублирование контролов

### 5.1. Верхняя панель `FloatingControls.svelte`

- `view-toggle` — `2D / 3D / List`.
- `layout-provider-toggle` — только в 3D-режиме.
- `type-filters` — фильтр по типам.
- `search-input` — фильтрация.
- `login-btn` — редирект на `/auth/login`.
- `menu-button` — import/export.
- `LangSwitcher` — смена языка.
- `create-note-button` — создание заметки.

### 5.2. Левый overlay `features/graph-ui/controls.svelte`

- `graph-controls-reset` — сброс вида.
- `graph-controls-search` — поиск/центрирование.
- `graph-controls-mode` — режим взаимодействия (`GraphMode`).
- `graph-controls-focus` — фокус-режим.

### 5.3. Пересечения

- **Поиск** — уже разобран выше.
- **Фокус** — `graph-controls-focus` переключает фокус-режим; в `overlay.svelte` есть индикатор `focus-mode-indicator`. Это один и тот же режим, но управление и отображение разнесены.
- **View vs. Mode** — `view-toggle` переключает между 2D/3D/List; `graph-controls-mode` переключает режимы внутри canvas (например, обычный и drag-link). Имена похожи, но функции разные.

## 6. Рекомендации

### 6.1. Не трогать сейчас

- Оба способа создания заметки (`+` и `N`) оставить.
- Два поиска оставить, но задокументировать разницу.
- `GraphCanvas.svelte` не переносить в другой слой FSD без отдельной задачи.

### 6.2. Сделать в ближайшем спринте

1. **Унифицировать дефолтный тип заметки** — решить, должен ли он быть `star` или `planet` во всех потоках. Это либо быстрый багфикс, либо намеренный UX-выбор.
2. **Вынести `handleNoteCreate/Delete/Restore/LinkCreate` в shared-хелпер** — убрать дублирование между `+page.svelte` и `graph/+page.svelte`. Нужно согласовать стратегию обновления графа (`refreshAfterMutation` vs `loadGraphData`) и стратегию ошибок (`alert` vs `console.error`).
3. **Добавить Playwright-регрессию** на оба потока создания заметки, чтобы дальнейшие изменения не сломали один из них.
4. **Переименовать/подписать поисковые элементы** так, чтобы пользователю было понятно: один фильтрует, другой находит.
5. **Вынести `GraphCanvas.svelte` в `widgets/`** при следующем рефакторинг-спринте graph-UI.

### 6.3. Технический долг

- Документировать в `docs/CRITICAL_FIXES.md` или `docs/REGRESSION_TEST_PLAN.md`, что `+page.svelte` и `graph/+page.svelte` содержат дублирующий CRUD-код.
- Отслеживать FSD-нарушение `GraphCanvas.svelte` в `docs/CRITICAL_FIXES.md`.

## 7. Покрытие тестами

Чтобы закрепить анализ и не допустить регрессии, рекомендуется добавить:

- `frontend/tests/note-creation-flows.spec.ts`:
  - `test('creates a note via floating + button')` — `/`, real-auth, клик `create-note-button`, заполнить `CreateNoteModal`, сохранить, проверить появление заметки.
  - `test('creates a note via N hotkey ghost form')` — `/` и `/graph`, real-auth, нажать `N`, заполнить `ghost-note-form`, сохранить, проверить появление заметки.
  - `test('both flows use consistent defaults')` — проверить, что после создания выбранный тип соответствует ожидаемому (сейчас может падать, пока баг с дефолтным типом не исправлен).

## 8. Связанные файлы

- `frontend/src/components/organisms/FloatingControls.svelte`
- `frontend/src/components/organisms/CreateNoteModal.svelte`
- `frontend/src/components/organisms/GraphCanvas.svelte`
- `frontend/src/features/graph-ui/controls.svelte`
- `frontend/src/features/graph-ui/overlay.svelte`
- `frontend/src/features/graph-ui/modals.svelte`
- `frontend/src/features/graph-interaction/hotkeys.ts`
- `frontend/src/features/graph-interaction/event-bridge.ts`
- `frontend/src/features/graph-forms/note-form.ts`
- `frontend/src/routes/+page.svelte`
- `frontend/src/routes/graph/+page.svelte`
