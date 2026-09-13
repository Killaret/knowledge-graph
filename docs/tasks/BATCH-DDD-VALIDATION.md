# BATCH-1 follow-up: DDD / Clean Architecture — перенос валидации типа заметки в домен

## Статус

**2026-09-13.** Devin обнаружил, что валидация celestial-body типа заметки (`star`, `planet`, `technical` и т.д.) сейчас дублируется в handler (`note_handler.go`) и не инкапсулирована в домене. Это нарушает ADR-005 «Validation Strategy» и идею rich domain model. Нужна постановка/рефакторинг.

## Что не так

- `note.NewNote` и `note.NewNoteWithCreator` принимают `noteType string` и не валидируют его, не возвращают `error`.
- `note.Note.SetType` также не валидирует значение.
- Валидация выполняется в interface layer: `validateResolvedNoteType` в `note_handler.go`.
- В `POST /notes`, `POST /notes/batch/create`, `POST /import/batch` есть дублирующая логика проверки `metadata.type` + `ValidCelestialBodyTypes`.

Это значит, что любой другой вход в домен (другой handler, use-case, worker, CLI) может создать `Note` с некорректным типом, обходя проверку.

## Почему это важно по Clean Architecture / DDD

- **Domain layer** — это внутренний круг. Он не должен зависеть от `interfaces` и не может полагаться на то, что кто-то "снаружи" уже проверил данные.
- **Business invariants** (допустимые типы, пустой title, длина content) по ADR-005 должны жить в domain.
- **Syntax validation** (JSON, max length в байтах, oneof в binding) — это задача `interfaces`, но она не должна заменять доменные инварианты.

## Что нужно сделать

1. Превратить `noteType` в полноценный value object или хотя бы защитить конструктор/сеттер:
   - Вариант A: новый `note.NewNote(...)` возвращает `(*Note, error)`. Это ломает много тестов, но чисто по архитектуре.
   - Вариант B: отдельный `note.NewType(value string) (NoteType, error)` и `Note.SetType(noteType NoteType)`. Меньше ломает существующий код.
2. Убрать дублирующую `validateResolvedNoteType` из `note_handler.go` после того, как домен начнёт отклонять невалидные типы.
3. Обновить тесты: `entity_test.go`, `value_objects_test.go`, `handler_unit_test.go`, `note_handler_test.go`, `note_handler_import_test.go`.
4. Проверить `Update` handler: `existing.SetType(req.Type)` должен обрабатывать `error`.

## Рекомендация

Вариант A — правильный, но радикальный. Лучше сделать в рамках отдельного PR/ветки, потому что затронет ~80 тестов и несколько handler'ов. Если нужно быстрое решение здесь и сейчас — вариант B плюс пометка TODO о миграции на A.

## Связанное

- ADR-005 `docs/architecture/decisions/005-validation-strategy.md`
- `backend/internal/domain/note/entity.go`
- `backend/internal/domain/note/value_objects.go`
- `backend/internal/interfaces/api/notehandler/note_handler.go`
- `docs/tasks/BATCH-API-DESIGN.md`

## Следующее действие

Claude Code / владелец решает: A или B, Devin выполняет рефакторинг после принятой постановки.
