# ➡️ Документ перемещён в архив

**Этот отчёт устарел (дата: 13 апреля 2026)**

Все исправления, описанные в этом отчёте, уже внедрены в версии 1.0.0.

---

## 📚 Актуальная документация

| Документ | Описание |
|----------|----------|
| **`TESTING.md`** | Руководство по тестированию |
| **`CHANGELOG.md`** | История изменений v1.0.0 |

---

## 📦 Архив

Этот файл сохранён в `docs/archive/` для истории.
См. `archive/README.md` для полного списка архивированных документов.

---

## Устаревший отчёт об исправлениях (13 апреля 2026):

## 1. Исправления API ✅

### 1.1 Исправлено передачу type в CreateNoteModal
**Файл:** `src/lib/components/CreateNoteModal.svelte`

```typescript
// Было:
metadata: { type }

// Стало:
type: type,
metadata: {}
```

### 1.2 Добавлены name атрибуты для тестирования
**Файл:** `src/lib/components/CreateNoteModal.svelte`

```html
<input id="note-title" name="title" ... />
<textarea id="note-content" name="content" ...></textarea>
```

### 1.3 Улучшен API клиент
**Файлы:** `src/lib/api/notes.ts`, `src/lib/api/graph.ts`

```typescript
const api = ky.create({ 
  prefixUrl: '/api',
  timeout: 30000,
  retry: {
    limit: 2,
    methods: ['get', 'post', 'put', 'delete'],
    statusCodes: [408, 413, 429, 500, 502, 503, 504]
  }
});
```

## 2. Исправления тестов ✅

### 2.1 Добавлены ожидания модальных окон
```typescript
await page.waitForSelector('.modal, [role="dialog"]', { timeout: 5000 });
```

### 2.2 Добавлены ожидания side panel
```typescript
await page.waitForSelector('.side-panel, .note-side-panel', { timeout: 5000 });
```

### 2.3 Исправлен strict mode violation
```typescript
// Было:
await expect(page.locator('.note-card:has-text("...")')).toBeVisible();

// Стало:
await expect(page.locator('.note-card:has-text("...")').first()).toBeVisible();
```

### 2.4 Увеличены таймауты
```typescript
await page.waitForTimeout(1500); // вместо 1000
await expect(...).toBeVisible({ timeout: 10000 });
```

### 2.5 Исправлены graph тесты
- Canvas проверяется только на странице `/graph/[id]`
- Добавлены first() для избежания strict mode violation

## 3. Текущий статус тестов

### ✅ Успешно проходят (5 тестов):
1. `should show back button on note detail page`
2. `should use browser back when history exists`
3. `should handle back button navigation from graph page`
4. `should display performance mode indicator`
5. `should handle graph page with no nodes gracefully`

### ❌ Требуют отладки (7 тестов):

| Тест | Проблема | Возможная причина |
|------|----------|-------------------|
| `should create a new note` | Не находит `input[name="title"]` | Модальное окно не открывается или сервер не обновился |
| `should edit a note` | Таймаут на `input[name="title"]` | Та же проблема с модальным окном |
| `should delete a note` | Не находит note card | Создание не работает, поэтому и удалять нечего |
| `should render graph page with canvas visible` | Canvas не найден | Graph3D не инициализируется |
| `should show graph container with correct styling` | Canvas не найден | Graph3D не инициализируется |
| `should open graph for a note with links` | Canvas не найден | Graph3D не инициализируется |
| `should search for notes` | Таймаут на `input[name="title"]` | Проблема с созданием |

## 4. Диагностика

### 4.1 Проблема с модальным окном создания
**Симптом:** Playwright не находит `input[name="title"]` после клика на `.create-btn`

**Возможные причины:**
1. Vite dev server кэширует старые файлы
2. Playwright кликает до полной загрузки страницы
3. Модальное окно рендерится с задержкой

**Рекомендации для отладки:**
```bash
# Остановить dev сервер и перезапустить
taskkill /F /IM node.exe
cd frontend && npm run dev

# Запустить тест в headed режиме для визуальной отладки
npm run test:headed
```

### 4.2 Проблема с Graph3D
**Симптом:** Canvas не отображается на странице `/graph/[id]`

**Возможные причины:**
1. Three.js не загружается (динамический импорт)
2. WebGL не поддерживается в headless режиме
3. Ошибка инициализации: `Error initializing 3D scene`

**Лог ошибки из теста:**
```
Error initializing 3D scene
This note has no connections
```

## 5. Рекомендации

### 5.1 Для исправления модального окна:
1. Убедиться что dev сервер перезапущен после изменений
2. Добавить `await page.waitForLoadState('networkidle')` перед кликом
3. Проверить видимость кнопки перед кликом:
   ```typescript
   await expect(page.locator('.create-btn')).toBeVisible();
   await page.click('.create-btn');
   ```

### 5.2 Для Graph3D тестов:
1. Использовать 2D fallback для тестирования
2. Добавить `?force3d=1` параметр для принудительного 3D
3. Проверить работу WebGL в CI environment

## 6. Итог

**Исправлено:**
- ✅ API типы
- ✅ Retry и timeout
- ✅ Селекторы strict mode
- ✅ Ожидания модальных окон
- ✅ name атрибуты для input

**Осталось:**
- ⚠️ Проблема с dev server кэшем
- ⚠️ Graph3D инициализация
- ⚠️ Визуальная отладка модального окна

**Сборка:** ✅ `npm run build` проходит успешно

---

## Следующие шаги:
1. Перезапустить dev сервер вручную
2. Запустить `npm run test:headed` для визуальной отладки
3. Проверить console.log в браузере
4. При необходимости обновить playwright selectors
