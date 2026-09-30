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

## Ревью — Claude Code, 2026-09-30

Реализация: Devin, `1584906`. **Вердикт: принято.**

### Мутации

| Мутация | Результат |
|---|---|
| снять условие владельца в `FindSimilarNotes` | красная: `TestEmbeddingRepository_OwnerIsolation`, `TestGammaLinkGenerator_OwnerIsolation`, `TestGetSuggestions_SemanticFallback_OwnerIsolation` |
| снять условие владельца в `FindSimilarNotesBatch` | красная: `TestEmbeddingRepository_OwnerIsolation` |

### Живьём — стенд из `1584906`, `SKIP_AUTH=false`

9/9. Два одноразовых пользователя, у каждого заметка с одним и тем же текстом; косинус их векторов 0,955 — без
исправления чужая заметка была бы ближайшим соседом. В подсказках каждого — только свои заметки. Живых связей
между владельцами на всём стенде, включая сид, — 0. Две заметки одного владельца автосвязь получили, то есть
конвейер отработал. Чужая заметка по прямому адресу — 404.

### Критерии

| # | Критерий | Вердикт | Чем проверен |
|---|---|---|---|
| 1 | тест красный на старом запросе; мутация «снять условие владельца» его краснит | выполнен | обе мутации красные |
| 2 | путь «semantic», автосвязи и загрузчик графа не отдают чужих заметок | выполнен | у «semantic» и автосвязей — свои тесты; загрузчик графа отдаёт результат того же запроса, его держит репозиторный тест |
| 3 | чужих автосвязей на стенде — 0 | выполнен | отчёт Devin; живой прогон ревьюера — 0 по всей базе |
| 4 | живая проверка в `MANUAL_TEST_FEEDBACK.md` | выполнен | записи Devin и ревьюера |

### Вопрос к переходу на 1.0

`IS NOT DISTINCT FROM` объединяет заметки без владельца (`creator_id IS NULL`) в одну группу, отдельную от заметок
с владельцем. Если на Personal-стеке часть заметок владельца без `creator_id` — например, созданы до включения
авторизации, — после 1.0 они перестанут предлагаться рядом с остальными и получать с ними автосвязи. Перед
переходом — посчитать такие заметки на копии данных владельца и, если они есть, привязать их к владельцу.
Personal-стек — только с согласия владельца.

### Принято без замечаний

- Решение по `creator_id IS NULL` принято явно и записано.
- SQL подсчёта и чистки чужих связей — в разборе; прогон на Personal — только с согласия владельца.

`check-all.ps1` на `1584906`: 31 из 33 зелёные, `golangci-lint` пропущен (локально не установлен). Две красные: юнит-тесты бэкенда упали с кодом 1 без единой строки FAIL — повтор той же команды `go test -v -coverprofile=cover.out ./...` зелёный, сбой окружения (память); сторож индекса — дрейф из коммита `1584906` (TASKS-INDEX-3), починен пересборкой индекса в коммите ревью.
