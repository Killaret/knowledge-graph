# План автоматического тестирования валидации заголовков и пользовательских сообщений об ошибках

**Создано:** 29 июля 2026 г.  
**Статус:** ⏳ Запланировано

## Обзор

Документ описывает план создания автоматических тестов для проверки валидации длины заголовков заметок и пользовательских сообщений об ошибках.

## Текущее состояние

**Фронтенд-валидация:**
- ✅ Реализована валидация длины заголовка (MAX_TITLE_LENGTH = 200)
- ✅ Реализован индикатор длины в реальном времени
- ✅ Реализована визуальная обратная связь (цвета, отключение кнопки)
- ✅ Реализована локализация сообщений об ошибках (русский/английский)
- ✅ Реализовано автообрезание в QuickCaptureWidget

**Бэкенд-валидация:**
- ✅ Реализована валидация на уровне domain (NewTitle)
- ✅ Реализована валидация на уровне API handler (Gin binding)
- ✅ Реализованы понятные сообщения об ошибках (VALIDATION_ERROR)

## План автоматических тестов

### 1. Unit тесты фронтенда (Vitest)

**Файл:** `frontend/src/components/organisms/CreateNoteModal.spec.ts`

**Тест-кейсы:**
```typescript
describe('CreateNoteModal - Title Validation', () => {
  it('should show error when title exceeds 200 characters', () => {
    // Ввести заголовок длиной 201 символ
    // Проверить что кнопка отключена
    // Проверить что индикатор показывает 201/200 и красный цвет
    // Проверить что появляется сообщение об ошибке
  });

  it('should show warning when title approaches limit (>180 chars)', () => {
    // Ввести заголовок длиной 181 символ
    // Проверить что индикатор показывает желтый цвет
  });

  it('should allow valid title (<200 characters)', () => {
    // Ввести заголовок длиной 50 символов
    // Проверить что кнопка включена
    // Проверить что индикатор показывает 50/200 и зеленый цвет
  });

  it('should prevent API call when validation fails', () => {
    // Ввести заголовок длиной 201 символ
    // Попытаться отправить форму
    // Проверить что API не вызывается (mock createNote)
  });

  it('should clear error when title is corrected', () => {
    // Ввести заголовок длиной 201 символ
    // Исправить до 50 символов
    // Проверить что ошибка исчезла
  });
});
```

**Файл:** `frontend/src/components/organisms/QuickCaptureWidget.spec.ts`

**Тест-кейсы:**
```typescript
describe('QuickCaptureWidget - Title Truncation', () => {
  it('should auto-truncate title when content is too long', () => {
    // Ввести контент длиной 300 символов
    // Создать пылинку
    // Проверить что заголовок обрезан до 197 символов + "..."
  });

  it('should not truncate when content is within limit', () => {
    // Ввести контент длиной 50 символов
    // Создать пылинку
    // Проверить что заголовок равен полному контенту
  });
});
```

### 2. Unit тесты бэкенда (Go)

**Файл:** `backend/internal/domain/note/value_objects_test.go`

**Тест-кейсы:**
```go
func TestNewTitle(t *testing.T) {
    tests := []struct {
        name    string
        value   string
        wantErr bool
        errMsg  string
    }{
        {"valid title", "Test Note", false, ""},
        {"empty title", "", true, "title cannot be empty"},
        {"whitespace only", "   ", true, "title cannot be empty"},
        {"exactly 200 chars", strings.Repeat("a", 200), false, ""},
        {"201 chars", strings.Repeat("a", 201), true, "title too long"},
        {"very long title", strings.Repeat("a", 1000), true, "title too long"},
    }
    // ... test implementation
}
```

**Файл:** `backend/internal/interfaces/api/notehandler/note_handler_test.go`

**Тест-кейсы:**
```go
func TestCreateNote_Validation(t *testing.T) {
    t.Run("title too long", func(t *testing.T) {
        // Создать заметку с заголовком >200 символов
        // Проверить что возвращается 422 Unprocessable Entity
        // Проверить структуру ошибки: VALIDATION_ERROR, field: title
    });

    t.Run("empty title", func(t *testing.T) {
        // Создать заметку с пустым заголовком
        // Проверить что возвращается 422
        // Проверить что есть описание ошибки
    });
}
```

### 3. E2E тесты (Playwright)

**Файл:** `frontend/tests/validation-title-length.spec.ts`

**Тест-кейсы:**
```typescript
test.describe('Title Validation', () => {
  test('should show error for long title in CreateNoteModal', async ({ page }) => {
    // Открыть форму создания заметки
    // Ввести заголовок длиной 201 символ
    // Проверить что кнопка "Создать" отключена
    // Проверить что индикатор показывает "201/200" красным
    // Проверить что появляется сообщение об ошибке
  });

  test('should allow valid title', async ({ page }) => {
    // Открыть форму создания заметки
    // Ввести заголовок длиной 50 символов
    // Проверить что кнопка "Создать" включена
    // Проверить что индикатор показывает "50/200" зеленым
  });

  test('should truncate title in QuickCaptureWidget', async ({ page }) => {
    // Нажать кнопку ✨ или Ctrl+Shift+N
    // Ввести контент длиной 300 символов
    // Создать пылинку
    // Проверить что пылинка создана успешно
    // Проверить что заголовок обрезан
  });

  test('should show localized error messages', async ({ page }) => {
    // Сменить язык на русский
    // Ввести заголовок длиной 201 символ
    // Проверить что сообщение об ошибке на русском
    // Сменить язык на английский
    // Проверить что сообщение об ошибке на английском
  });
});
```

### 4. API тесты

**Файл:** `backend/internal/interfaces/api/notehandler/note_handler_api_test.go`

**Тест-кейсы:**
```go
func TestCreateNoteAPI_Validation(t *testing.T) {
    // Тесты через HTTP запросы к API
    // Проверка структуры JSON ответов
    // Проверка CORS заголовков в ответах об ошибках
}
```

## Приоритет реализации

1. **P0 (Критический):** E2E тесты валидации формы создания заметок
2. **P1 (Высокий):** Unit тесты фронтенда для CreateNoteModal и QuickCaptureWidget
3. **P2 (Средний):** Unit тесты бэкенда (уже частично реализованы)
4. **P3 (Низкий):** API тесты структуры ответов

## Критерии успеха

- [ ] E2E тесты покрывают основные сценарии валидации
- [ ] Unit тесты покрывают логику валидации на фронтенде
- [ ] Unit тесты покрывают логику валидации на бэкенде
- [ ] Все тесты проходят в CI/CD pipeline
- [ ] Покрытие кода валидации > 80%

## Связанные задачи

- Документация: `docs/MANUAL_TEST_CHECKLISTS_RU.md` — добавлен раздел валидации
- Документация: `docs/MANUAL_TEST_CHECKLIST_MINIMAL.md` — добавлены проверки валидации
- Код: `frontend/src/components/organisms/CreateNoteModal.svelte` — реализована валидация
- Код: `frontend/src/components/organisms/QuickCaptureWidget.svelte` — реализовано обрезание