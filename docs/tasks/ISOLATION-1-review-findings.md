# ISOLATION-1 — разбор реализации (Devin, 2026-09-30)

Дефект из [`RECO-1-review-findings.md`](RECO-1-review-findings.md), пункт 3:
`FindSimilarNotes` и `FindSimilarNotesBatch` сравнивали вектор со всеми
заметками без условия на владельца.

## Что сделано

`backend/internal/infrastructure/db/postgres/embedding_repo.go`: оба запроса
присоединяют `notes n1` к заметке-источнику и требуют
`n1.creator_id IS NOT DISTINCT FROM n2.creator_id`.

**Решение по `creator_id IS NULL`** (пункт 2 постановки — «решить явно»):
`IS NOT DISTINCT FROM` считает два NULL равными — безвладельческие заметки
(legacy до авторизации, тестовый пользователь миграции 019 с `uuid.Nil`)
продолжают видеть друг друга, но не заметки с владельцем и наоборот.
Строгий `=` заставил бы весь legacy-корпус потерять подсказки — отклонено.

## Красный → зелёный

`TestEmbeddingRepository_OwnerIsolation` (`embedding_repo_test.go`):
два пользователя, вектор чужой заметки — ближайший сосед.
На старом запросе — красный (`FindSimilarNotes leaked another user's note`),
на новом — зелёный; `FindSimilarNotesBatch` проверен в обе стороны.

## Покрытие по путям (критерий 2)

| Путь | Тест | Статус |
|---|---|---|
| semantic-подсказки (`GET /notes/:id/suggestions`, fallback «semantic») | `TestGetSuggestions_SemanticFallback_OwnerIsolation` | зелёный: возвращает только свою заметку |
| автосвязи (`GammaLinkGenerator` с реальными `EmbeddingRepository` и `LinkRepository`) | `TestGammaLinkGenerator_OwnerIsolation` (`isolation_gamma_integration_test.go`) | зелёный: связь к чужой заметке не создаётся и не сохраняется |
| загрузчик графа (`embedding_loader`) | тот же SQL через общий интерфейс; репозиторный тест покрывает оба запроса | зелёный |

## Уже созданные чужие автосвязи (пункт 4)

Новые чужие автосвязи создаваться не могут (запрос не отдаёт чужих кандидатов).
Разовая чистка существующих — SQL для отчёта «было N, стало 0»:

```sql
-- подсчёт
SELECT count(*) FROM links l
JOIN notes s ON s.id = l.source_note_id
JOIN notes t ON t.id = l.target_note_id
WHERE l.deleted_at IS NULL
  AND s.creator_id IS DISTINCT FROM t.creator_id;
-- чистка (soft-delete)
UPDATE links SET deleted_at = now()
WHERE id IN (
  SELECT l.id FROM links l
  JOIN notes s ON s.id = l.source_note_id
  JOIN notes t ON t.id = l.target_note_id
  WHERE l.deleted_at IS NULL
    AND s.creator_id IS DISTINCT FROM t.creator_id);
```

Прогон на Personal-стеке — только с согласия владельца после бэкапа;
в этой сессии не выполнялся (стек не поднимался).

## Живая проверка (пункт 5) — выполнена

Изолированный тест-стек, `SKIP_AUTH=false`, два реальных пользователя
(`iso_a`, `iso_b`). У обоих по заметке с идентичным текстом («про аниме») +
вторая заметка у `iso_a`; векторы посчитаны NLP-конвейером.

- Подсказки `iso_a`: только своя «Note A2» — чужая, семантически ближайшая,
  не показана.
- Подсказки `iso_b`: пусто — заметки `iso_a` не просачиваются.
- `cross_owner_links = 0` в базе.

Запись — в `docs/agents/MANUAL_TEST_FEEDBACK.md` («ISOLATION-1»).
