# BATCH-1 follow-up: DDD / Clean Architecture — перенос валидации типа заметки в домен

## Статус

**2026-09-14.** Реализовано Devin в ветке `devin/batch-api-37`; передано на ревью Claude Code.
- Выбран вариант B: `note.NewType(value string) (NoteType, error)` и `Note.SetType(noteType NoteType)`.
- `NoteType` value object живёт в `backend/internal/domain/note/type.go`.
- Конструкторы и реконструкторы `Note` принимают `NoteType`.
- Обработчики преобразуют строку в `NoteType` через `NewType` и отдают 400 при невалидном значении.
- `validateResolvedNoteType` удалён; `IsValidCelestialBodyType` делегирует домену.
- Тесты `note/type_test.go`, `entity_test.go`, `value_objects_test.go`, `notehandler/*` обновлены.

## Что было не так

- `note.NewNote` и `note.NewNoteWithCreator` принимают `noteType string` и не валидируют его, не возвращают `error`.
- `note.Note.SetType` также не валидирует значение.
- Валидация выполняется в interface layer: `validateResolvedNoteType` в `note_handler.go`.
- В `POST /notes`, `POST /notes/batch/create`, `POST /import/batch` есть дублирующая логика проверки `metadata.type` + `ValidCelestialBodyTypes`.

Это значит, что любой другой вход в домен (другой handler, use-case, worker, CLI) может создать `Note` с некорректным типом, обходя проверку.

## Почему это важно по Clean Architecture / DDD

- **Domain layer** — это внутренний круг. Он не должен зависеть от `interfaces` и не может полагаться на то, что кто-то "снаружи" уже проверил данные.
- **Business invariants** (допустимые типы, пустой title, длина content) по ADR-005 должны жить в domain.
- **Syntax validation** (JSON, max length в байтах, oneof в binding) — это задача `interfaces`, но она не должна заменять доменные инварианты.

## Что сделано

1. `note.NewType(value string) (NoteType, error)` — value object с валидацией, `scaleRank`, `IsUserSelectable`.
2. `Note` хранит `NoteType`; конструкторы и реконструкторы принимают `NoteType`.
3. `note.SetType(noteType NoteType)` игнорирует/не применяет zero value.
4. Дублирующая `validateResolvedNoteType` удалена из `note_handler.go`; `resolveNoteType` использует `note.NewType`.
5. `POST /notes`, `POST /notes/batch/create`, `POST /import/batch`, bookmarklet, batch update — все через `note.NewType`.
6. `IsValidCelestialBodyType` делегирует `note.NewType`.
7. Тесты обновлены и дополнены.

## Рекомендация по дальнейшему улучшению

Вариант A (`NewNote(...) (*Note, error)`) остаётся более чистым с точки зрения DDD, но требует отдельного рефакторинга ~80 тестов. Вариант B закрывает текущий риск обхода валидации.

## Связанное

- ADR-005 `docs/architecture/decisions/005-validation-strategy.md`
- `docs/tasks/NOTE-TYPE-TAXONOMY.md` — таксономия и порядок типов, от которой зависит доменный `NoteType`
- `backend/internal/domain/note/entity.go`
- `backend/internal/domain/note/value_objects.go`
- `backend/internal/interfaces/api/notehandler/note_handler.go`
- `docs/tasks/BATCH-API-DESIGN.md`

## Следующее действие

Ревью Claude Code: `backend/internal/domain/note/type.go`, `entity.go`, `note_handler.go`, `validators.go`; приём решения о миграции на вариант A.
