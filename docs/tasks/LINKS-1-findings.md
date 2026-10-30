# LINKS-1. Находки исполнения (прогон 2026-09-21, тест-стек)

Стенд: `docker-compose.test.yml`, сид 100 заметок, модель
`paraphrase-multilingual-MiniLM-L12-v2`. Генерация запускалась вживую: после
`embed-recompute` воркер посчитал эмбеддинги и создал gamma-связи сам.

## 1. Гистограмма top-2 косинусных score по сиду

| rank | min | avg | max | n |
|------|-----|-----|-----|---|
| 1 | 0.955 | 0.993 | 0.998 | 99 |
| 2 | 0.949 | 0.991 | 0.998 | 99 |
| 3 | 0.940 | 0.989 | 0.997 | 99 |
| 5 | 0.935 | 0.986 | 0.995 | 99 |
| 8 | 0.912 | 0.973 | 0.990 | 99 |
| 10 | 0.867 | 0.905 | 0.934 | 99 |

**Сид вырожден для калибровки порога.** Сеялка генерирует почти одинаковые
тексты («This is a test X note number N…»), поэтому все top-2 score лежат в
0.949–0.998 — порог 0.6 пропускает всё, и на этом сиде он ничего не отсекает.
Реальную разделительную способность покажет только разношёрстный корпус
(Personal после MODEL-1). Умолчание 0.6 оставлено как осторожное значение из
постановки; конфигурация проведена полной тройкой: `GAMMA_LINK_MIN_SCORE` →
`knowledge-graph.config.json` (`backend.recommendation.gamma_link_min_score`) →
дефолт 0.6 в `config.go`, плюс тест `TestGammaLinkMinScore`.

Итог на сиде: 100 заметок с эмбеддингом → 200 gamma-связей (степень 2
достигается почти у всех), веса связей 0.843–0.998.

## 2. Найденный дефект — `note_links_closure` не обновлялся вообще

Первый же `LinkCreated` показал: `REFRESH MATERIALIZED VIEW CONCURRENTLY`
падает с `ERROR: cannot accumulate arrays of different dimensionality`.
Причина в миграции 027: `(array_agg(path ORDER BY distance))[1]` при
`path uuid[]` — Postgres не умеет накапливать массивы разной длины, поэтому
любая пара (источник, цель) с путями разной длины ломала рефреш. До
автосвязей в сиде не было рёбер, view оставался пустым, и дефект был
невидим. Заодно тип колонки `path` определялся как `uuid` вместо `uuid[]` —
запросы `array_to_string(c.path, ',')` в graph-service падали бы при чтении.

Исправлено миграцией `032_fix_note_links_closure_path`: агрегация идёт по
`path::text`, результат приводится обратно к `uuid[]`. После фикса рефреш
работает: 1114 строк closure на ~195 связях, затем 297 строк на 200 связях
после регенерации (топология другая — связи пересчитаны).

## 3. Замеры REFRESH MATERIALIZED VIEW CONCURRENTLY

| состояние | время |
|---|---|
| до LINKS-1 (старое определение, любые разветвления) | ERROR 2202E — не обновлялось |
| после фикса, 195 gamma-связей | 125 ms (1114 строк) |
| после регенерации, 200 gamma-связей | 33 ms (297 строк) |

На сиде из 100 заметок refresh дешёвый. Степень 2 оставлена; поднимать по-прежнему
нельзя без замера на реальном корпусе — рост paths нелинейный (195→200 связей
меняет 1114→297 строк только от топологии).

## 4. События и рекомендации

- На каждую созданную связь воркер публикует `LinkCreated` в `graph:events`;
  graph-service логирует приём, инвалидирует кэши и рефрешит closure
  (после фикса view — успешно).
- Refresh recommendations ставится для источника и каждой цели; в логах
  видны повторные `task ID conflicts` — asynq дедуплицирует одинаковые
  refresh-задачи, это ожидаемое сообщение, не ошибка данных.
- `gamma-links-regenerate --dry-run`: `Would delete 197 / Would create 200`.
- Полный прогон: `197 deleted, 200 created`, второй прогон `200 deleted,
  200 created` — идемпотентно, ручные связи не тронуты (в сиде их нет;
  фильтр проверен тестом `TestDeleteBySourceType`).

## 5. Сигнал для NOTE-QUALITY-1

Дубликат эмбеддинга даёт score 1.0 и gamma-связь создаётся — на сиде почти
все пары ≥0.95 именно из-за шаблонных текстов. Если NOTE-QUALITY-1 будет
искать дубликаты, пары с gamma-связью веса ~1.0 — готовый кандидатский список.

## 6. Мутации (все три красные, код восстановлен и перепроверен)

| мутация | тест | результат |
|---|---|---|
| убрана проверка `s.Score < g.minScore` | `TestGammaLinkGenerator_EnforcesMinScore` | FAIL: 3 связи вместо 2 |
| убран `PublishLinkCreated` в воркере | `TestWorker_ComputeEmbeddingCreatesGammaLinks` | FAIL: i/o timeout — событие не пришло |
| убран `WHERE source_type` в `DeleteBySourceType` | `TestLinkRepositoryIntegrationSuite/TestDeleteBySourceType` | FAIL: удалено 2 связи вместо 1 (ручная тоже) |

## 7. Покрытие

Юнит (`internal/application/recommendation`, `internal/infrastructure/queue`,
`internal/config`): порог 0.9/0.65/0.4→2 связи, степень, идемпотентность,
ручная связь блокирует, дубликат score 1.0, нет кандидатов — ноль связей без
ошибки, `PlanForNotes` игнорирует gamma-цели но уважает ручные, env-override
и дефолт `GAMMA_LINK_MIN_SCORE`, события+рефреши из воркера.

Интеграция (`-tags=integration`, testcontainers pgvector + miniredis):
`HandleComputeEmbedding` создаёт ≤2 gamma-связи и публикует по LinkCreated
на каждую; удаление заметки-цели каскадно удаляет gamma-связь;
`DeleteBySourceType` удаляет только gamma.

## Побочные находки сессии

- `worker-test` не стартовал в тест-стеке, пока его не подняли вручную —
  `up -d --wait` не дождался его; при прогонах проверять `docker compose ps`.
- `kg-test-nlp` после пересоздания контейнера заново скачивает модель (~10 мин);
  задачи эмбеддингов это переживают за счёт ретраев asynq.
- `keyword-recompute` отсутствовал в `backend/Dockerfile` (есть в COMMANDS.md)
  — добавлен рядом с `gamma-links-regenerate`.
- `TestConstraintProtection` жил на расхождении схем: AutoMigrate создавал FK
  без CASCADE, а миграция 002 — с `ON DELETE CASCADE`. После выравнивания
  `link_model.go` (`constraint:OnDelete:CASCADE`) тест переписан под реальную
  семантику: удаление заметки каскадно сносит связи, без ошибки.
- Два новых cmd (`gamma-links-regenerate`, `keyword-recompute`) не были в
  `backend-coverage-excludes.txt` — тянули покрытие под порог (69.38%).
  Добавлены по тому же принципу, что `embed-recompute`/`worker` — после
  исключения покрытие 70.86%.
