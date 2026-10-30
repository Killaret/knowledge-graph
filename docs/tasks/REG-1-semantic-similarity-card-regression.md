# REG-1: регрессионный тест на семантическое сходство в карточке заметки

**Статус:** на ревью  
**Дата:** 2026-09-10  
**Источник:** ручная проверка карточки `GET /api/v1/notes/:id/suggestions` на Personal-стеке.

## Суть

В карточке заметки не отображались семантически похожие заметки, хотя `note_embeddings` были. Найдено два связанных дефекта:

1. `EmbeddingRepository.FindSimilarNotes` (`backend/internal/infrastructure/db/postgres/embedding_repo.go`) использовал SQL-алиас `similarity`, а GORM-структура `SimilarNote` ожидает поле `score`. Из-за этого `ORDER BY score` падает, а `SimilarNote.Score` остаётся `0`.
2. Семантический fallback в `NoteHandler.GetSuggestions` (`backend/internal/interfaces/api/notehandler/note_handler.go`) возвращал только `note_id` и `score`, не подгружая `title`, поэтому UI рисовал пустые карточки.

Исправления уже в рабочем дереве. Теперь нужен регрессионный интеграционный тест, который ловит откат любого из этих двух исправлений.

## Критерии приёмки

- [x] Интеграционный тест с настоящей `pgvector`-БД создан.
- [x] Тест создаёт две заметки, сохраняет эмбеддинги, вызывает `GET /notes/:id/suggestions` и проверяет:
  - `200 OK`;
  - `X-Recommendations-Source: semantic`;
  - ровно одна подсказка;
  - `note_id` совпадает с похожей заметкой;
  - `title` совпадает с заголовком похожей заметки (не пустой);
  - `score` > 0.
- [x] Тест падает, если SQL-алиас вернуть на `similarity` (проверено мутацией).
- [x] Тест падает, если убрать `FindByID` для `title` (проверено мутацией).
- [x] `go test -tags=integration -run TestNoteHandlerSemanticIntegrationSuite ./internal/interfaces/api/notehandler` проходит.
- [x] `go test ./internal/interfaces/api/notehandler` проходит.

## Связанные файлы

- `backend/internal/infrastructure/db/postgres/embedding_repo.go`
- `backend/internal/interfaces/api/notehandler/note_handler.go`
- `backend/internal/interfaces/api/notehandler/handler_unit_test.go`
- `backend/internal/interfaces/api/notehandler/note_handler_semantic_integration_test.go` (новый)
- `docs/AI_HANDOFF.md`

## Проверяющий

Claude Code — ревью кода и подтверждение, что тест ловит оба дефекта.
