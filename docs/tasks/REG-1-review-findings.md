# Ревью REG-1 (Claude Code, 2026-09-13)

**Принято.** Интеграционный тест прогнан на живой pgvector-БД, обе мутации краснеют.

## Как проверялось

Тест под `//go:build integration`, нужен pgvector. Вместо testcontainers (ловушка с
registry-DNS, ронявшая интеграционные фазы) подключил его к живому тест-Postgres через
`TEST_DATABASE_URL=postgres://kb_user:kb_password@127.0.0.1:15434/knowledge_test`.
`kb_user` имеет `rolcreatedb`, тест сам создаёт и роняет изолированную БД.

```
--- PASS: TestNoteHandlerSemanticIntegrationSuite/TestGetSuggestions_SemanticFallback_PopulatesTitleAndScore
ok  knowledge-graph/internal/interfaces/api/notehandler  2.652s
```

## Мутации — обе стороны бага

1. **SQL-алиас.** Вернул `as score` → `as similarity` в `FindSimilarNotes`
   (`embedding_repo.go:55`) — тест падает. Значит `SimilarNote.Score` действительно
   зависит от совпадения имени колонки с GORM-полем, и `ORDER BY score` без него мёртв.
   Осторожно: `as score` в файле **дважды** (single + batch, строки 55 и 90); мутировать
   надо строку 55 — путь `GetSuggestions` идёт через `FindSimilarNotes`, не batch.
2. **Подгрузка `title`.** Дописал `&& false` в условие
   `if noteEntity, err := h.repo.FindByID(...)` (`note_handler.go:1615`) — тест падает
   на пустом заголовке. Значит именно этот `FindByID` наполняет карточку, без него UI
   рисует пустоту.

## Замечаний нет

Оба критерия приёмки закрыты и проверяемы. `batch`-вариант (`FindSimilarNotesBatch`,
строка 90) тем же тестом не покрыт — но REG-1 про single-путь, это вне её объёма;
кандидат в отдельную проверку, если batch-рекомендации где-то используются в UI.
