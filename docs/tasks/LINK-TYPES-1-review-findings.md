# LINK-TYPES-1 — разбор ревью

Ревью — Claude Code, 2026-09-28. Проверены `c8fda5b` и `86c5bef` на тест-стенде из
`D:\knowledge-graph-review`, `SKIP_AUTH=false`. Постановка —
[`LINK-TYPES-1-link-types-and-visuals.md`](LINK-TYPES-1-link-types-and-visuals.md).

**Вердикт: отклонено.** Миграция 037 портит данные в шести случаях, которые её тест не покрывает.
Остальное принято. 3D в приёмку больше не входит (решение 82), рекомендации пунктиром — тоже: их
делает Claude Code в GRAPH-LIGHT-1 (решение 83).

## Блокер — миграция 037 на пограничных случаях

Воспроизведение — скрипт ниже на базе тест-стенда. Всё идёт в одной транзакции и откатывается.
Файл миграции предварительно копируется в контейнер: `docker cp
backend/migrations/037_link_types_merge_related.up.sql kg-test-postgres:/tmp/037.sql`.

| Случай | До миграции | После | Что не так |
|---|---|---|---|
| A | у пары автосвязь `related` (`gamma`, 0,62) и ручная `reference` (0,9) | живая — автосвязь `gamma` с весом 0,9; ручная удалена | ручная связь владельца стала автосвязью. Решение 53: ручная поверх автоматической — повышение до `user`, вес модели уходит в `metadata.gamma` |
| B | `related` 0,3, `reference` 0,5 и `dependency` 0,95 | у `related` вес 0,95 | вес взят у связи другого типа |
| C | `related` 0,2, `custom` 0,4 и удалённая `parent` 1,0 | у `related` вес 1,0 | вес взят у удалённой связи |
| D | отвергнутая автосвязь (удалена, `gamma`) и живая ручная `reference` | ожила отвергнутая автосвязь, ручная удалена | то же, что A, и вдобавок возвращается связь, которую человек отверг |
| E | `related` и удалённая вместе с заметкой `reference` (`deleted_via_note_id`) | `reference` осталась старого типа; после «Восстановить» у пары две живые связи | старый тип возвращается в живые данные, пара двоится |
| F | `related` A→B и `reference` B→A | две живые `related`: A→B и B→A | `SaveUserLink` и решение 53: одна связь на пару в любом направлении |

Причины — в `037_link_types_merge_related.up.sql`:

- вес выжившей строки — `max(weight)` по всем строкам пары любого типа и в любом состоянии;
- выжившая выбирается по `(deleted_at IS NULL) DESC, weight DESC` без учёта `source_type`;
- цикл идёт только по живым строкам старых типов;
- пара считается направленной — группировка по `(source_note_id, target_note_id)`.

Миграция выполнится один раз — на данных владельца при переходе на 1.0, поэтому её нужно довести до
правильного поведения во всех случаях, даже редких.

**Что сделать:**

1. Пара — неупорядоченная, как в `SaveUserLink`.
2. Вес выжившей строки — максимум только по сливаемой группе: живые `related`, `reference`, `custom`
   этой пары.
3. Ручная поверх автоматической: если в группе есть строка `user`, выжившая становится `user`, вес
   модели — в `metadata.gamma`; это то же повышение, что в `SaveUserLink`. Отвергнутая (удалённая)
   автосвязь не оживает автосвязью.
4. Удалённые строки старых типов тоже переводятся — хотя бы те, что с `deleted_via_note_id`, — чтобы
   «Восстановить» не возвращало старые типы и не двоило пары.
5. Печатать число слитых пар, как в постановке.
6. Тест на каждый случай A–F на фикстуре; каждый красный на нынешней миграции.

```sql
\set ON_ERROR_STOP on
BEGIN;
CREATE TEMP TABLE n AS
  SELECT id, row_number() OVER (ORDER BY created_at, id) AS k
  FROM notes WHERE deleted_at IS NULL LIMIT 12;
DELETE FROM links WHERE source_note_id IN (SELECT id FROM n) OR target_note_id IN (SELECT id FROM n);
CREATE TEMP TABLE cases(label text, s int, t int, lt text, st text, w float8, del bool, via int);
INSERT INTO cases VALUES
 ('A', 1, 2, 'related', 'gamma', 0.62, false, NULL), ('A', 1, 2, 'reference', 'user', 0.90, false, NULL),
 ('B', 3, 4, 'related', 'user', 0.30, false, NULL), ('B', 3, 4, 'reference', 'user', 0.50, false, NULL),
 ('B', 3, 4, 'dependency', 'user', 0.95, false, NULL),
 ('C', 5, 6, 'related', 'user', 0.20, false, NULL), ('C', 5, 6, 'custom', 'user', 0.40, false, NULL),
 ('C', 5, 6, 'parent', 'user', 1.00, true, NULL),
 ('D', 7, 8, 'related', 'gamma', 0.70, true, NULL), ('D', 7, 8, 'reference', 'user', 0.80, false, NULL),
 ('E', 9, 10, 'related', 'user', 0.50, false, NULL), ('E', 9, 10, 'reference', 'user', 0.60, true, 9),
 ('F', 11, 12, 'related', 'user', 0.50, false, NULL), ('F', 12, 11, 'reference', 'user', 0.60, false, NULL);
INSERT INTO links (source_note_id, target_note_id, link_type, weight, source_type, deleted_at, deleted_via_note_id)
SELECT ns.id, nt.id, c.lt, c.w, c.st, CASE WHEN c.del THEN now() END, nv.id
FROM cases c JOIN n ns ON ns.k = c.s JOIN n nt ON nt.k = c.t LEFT JOIN n nv ON nv.k = c.via;
\i /tmp/037.sql
SELECT c.label, l.link_type, l.source_type, l.weight, (l.deleted_at IS NULL) AS live
FROM links l JOIN n ns ON ns.id = l.source_note_id JOIN n nt ON nt.id = l.target_note_id
JOIN (SELECT DISTINCT label, s, t FROM cases) c ON c.s = ns.k AND c.t = nt.k
ORDER BY c.label, ns.k, l.link_type;
UPDATE links SET deleted_at = NULL, deleted_via_note_id = NULL WHERE deleted_via_note_id = (SELECT id FROM n WHERE k = 9);
SELECT link_type, (deleted_at IS NULL) AS live FROM links
WHERE source_note_id = (SELECT id FROM n WHERE k = 9) AND target_note_id = (SELECT id FROM n WHERE k = 10);
ROLLBACK;
```

## Мелочь

- API без `link_type` отвечает 400. Постановка: по умолчанию `related` и в API — сделать поле
  необязательным со значением `related`.
- «Faithful down» в сообщении коммита — не совсем: у выжившей строки down не возвращает прежний вес
  и прежнее «удалена».

## Принято без замечаний

- **API.** `reference` при создании и `custom` при правке сохраняются как `related` — живьём на
  стенде (201 и 200, тип `related`); интеграционные тесты обработчика связей зелёные.
- **Цвет автосвязи.** Тест есть; мутация «автосвязь цветом своего типа» краснеет (1 тест).
- **Цепочка `dependency`.** Фикстура с ветвлением и циклом; мутация «глубина без предела» краснеет
  (3 теста). Живьём: цикл 2→3→4→2 при наведении красный; в панели заметки «Requires» и «Needed for»
  совпадают с созданными связями, предупреждение о цикле есть.
- **Легенда** без `reference` и `custom`, со строкой «Auto link (model)» — живьём.
- **Панель заметки** снова показывает связи: конверт `{incoming, outgoing}` разобран, на стенде
  «Links (7)».
- **Тесты.** Бэкенд — сьют миграции 037, интеграция обработчика связей, домен связей — зелёные на
  базе стенда. Фронтенд — 84 из 84 по шести файлам.

## Снято: «пустой холст у маленького графа»

В первой версии этого разбора здесь был дефект: у пользователя с шестью заметками холст пустой при
открытии и после открытия панели. Он не подтвердился. Проверка шла во встроенном браузере, вкладка
которого была скрыта: `document.visibilityState = "hidden"`, за секунду ни одного вызова
`requestAnimationFrame`, а граф рисуется только в этом цикле. В видимом Chromium (Playwright,
окно 1400×900) тот же пользователь: граф виден через 3 секунды после загрузки, масштаб 1, шесть узлов;
после открытия панели холст сужается и граф остаётся на месте. UI-PANELS-1 не затронут.

Остаётся мелочь: `resize.ts` при нулевом размере родителя берёт высоту `window.innerHeight - 80`, в
скрытом окне это −80, и вписывание даёт отрицательный масштаб. В видимой вкладке не проявляется;
учтено в GRAPH-LIGHT-1.

## Доработка Devin, 2026-09-29 — миграция 037 переписана

Все шесть воспроизведённых случаев закрыты в `037_link_types_merge_related.up.sql`,
каждый под отдельным тестом в `migration_037_integration_test.go`. На старой версии
миграции все шесть тестов красные, на новой — зелёные (12/12 сьют).

- **A — ручная против gamma.** Пара (gamma `related` 0.62 + user `reference` 0.9):
  survivor — строка `user` с повышением в `related`, вес модели уходит в
  `metadata.gamma` (`score` + `generated_at`) — как `SaveUserLink`/`PromoteToUser`.
  Тест `TestManualReferenceBeatsGammaRelated`.
- **B — вес `dependency` не смешивается.** Максимум считается только по живым
  сливаемым строкам (`related`/`reference`/`custom`); `dependency` не входит в группу.
  Тест `TestDependencyWeightDoesNotContaminate`.
- **C — удалённый `parent` не даёт вес.** Негeneric-типы и удалённые строки исключены
  из группы и из максимума. Тест `TestDeletedParentWeightDoesNotContaminate`.
- **D — отклонённая gamma не воскресает.** Порядок survivor: живые → `user` → вес;
  удалённая строка выбирается только когда живых в группе нет, а живые удалённые
  `related` без маркера в поглощение не входят — остаются удалёнными.
  Тест `TestRejectedGammaStaysDeleted`.
- **E — восстановление заметки.** Поглощённые строки, удалённые через заметку,
  отцепляются от `deleted_via_note_id` (значение сохранено в
  `metadata.deleted_via_note_id_before`): `note_repo.Restore` больше не может
  вернуть вторую живую legacy-связь. Тест `TestNoteDeletedRowStaysDetachedOnRestore`
  прогоняет сам `UPDATE` восстановления.
- **F — встречное направление.** Группировка по `LEAST/GREATEST(source,target)` —
  `related` A→B и `reference` B→A схлопываются в одну строку, направление survivor'а
  сохраняется. Тест `TestReverseDirectionPairCollapses`.

UNIQUE `(source,target,link_type)` покрывает и удалённые строки, поэтому поглощённые
`related` переводятся в сентиментальный тип `absorbed_related` (освобождает ключ до
конвертации survivor'а); настоящий тип — в `metadata.link_type_before`, down его
возвращает.

### API: `link_type` необязателен

`createLinkRequest` и `importBatchLinkItem`: `binding:"required"` → `omitempty`,
пустое значение нормализуется в `related`. Тест `TestCreateLinkDefaultsToRelated`
(201 и `related` при отсутствии поля). `openAPI.yaml`: поле убрано из `required`,
добавлено `default: related`.

### Down-миграция

`037...down.sql` переписан по `was_deleted`: поглощённые-живые возвращают тип и
воскресают, поглощённые-удалённые возвращают тип и `deleted_via_note_id`, оставаясь
удалёнными; survivor возвращает свой тип. Вес survivor'а — объединённый максимум,
исходные веса по строкам невосстановимы — это зафиксировано в комментарии файла.
Тест `TestDownMigrationRestoresTypes` зелёный на новой паре up/down.

## Ревью доработки — Claude Code, 2026-09-30

Доработка: Devin, `f48aa77`. **Вердикт: принято; хвост — LINK-TYPES-1-TAIL.** Все шесть случаев первого раунда
закрыты — живьём на базе стенда и тестами. Но два правила выбора выжившей связи тестами не закреплены, и нашлись
два новых пограничных случая. Всё это нужно закрыть до того, как 037 дойдёт до данных владельца.

### Живьём — стенд из `1584906`: сценарий первого раунда плюс A′ и E′, в транзакции с откатом

| Случай | После 037 | Вердикт |
|---|---|---|
| A | живая `related`, `user`, 0,9; вес модели 0,62 — в `metadata.gamma` | исправлено |
| B | `related` 0,5; `dependency` 0,95 не тронута | исправлено |
| C | `related` 0,4; удалённая `parent` не тронута | исправлено |
| D | живая `related`, `user`, 0,8; отклонённая автосвязь осталась удалённой | исправлено |
| E | после «Восстановить» у пары одна живая `related` | исправлено |
| F | одна живая `related` на пару | исправлено |
| A′: автосвязь 0,95 и ручная `reference` 0,5 | живая `related`, `user`, **вес 0,95** | тип верен; вес — модели, а `SaveUserLink` при повышении оставляет вес пользователя |
| E′: у пары только `reference`, удалённая вместе с заметкой | после «Восстановить» у пары **нет живой связи** | второй проход отцепил связь от заметки |

### Мутации

| Мутация | Результат |
|---|---|
| правило «user важнее gamma» убрано из выбора выжившей | **зелёная**: в фикстуре A ручная тяжелее автосвязи, и выживает та же строка |
| максимум веса по всем строкам пары | красная: `TestDependencyWeightDoesNotContaminate`, `TestDeletedParentWeightDoesNotContaminate`, `TestRejectedGammaStaysDeleted` |
| удалённая строка может выжить раньше живой | **зелёная**: в фикстуре D та же строка выживает и без этого правила |
| второй проход сохраняет маркер восстановления | красная: `TestNoteDeletedRowStaysDetachedOnRestore` |
| пара снова направленная | красная: шесть тестов, среди них `TestReverseDirectionPairCollapses` |
| `POST /links` без `link_type` — без умолчания | красная: `TestCreateLinkDefaultsToRelated` |
| пакетный импорт без `link_type` — без умолчания | **зелёная**: теста нет |

### Хвост — LINK-TYPES-1-TAIL (1.0, Devin, маленькая; до перехода данных владельца на 1.0)

1. Фикстура A′ — автосвязь тяжелее ручной: выживает `user`, вес — как в `SaveUserLink`, то есть вес пользователя,
   а вес модели — в `metadata.gamma`. Мутация «user важнее gamma» — красная.
2. Фикстура D′ — удалённая строка тяжелее живой: выживает живая. Мутация «живая важнее удалённой» — красная.
3. Тест умолчания `related` в пакетном импорте.
4. E′: если у пары нет живой связи, удалённая вместе с заметкой старая связь переводится в `related` и сохраняет
   маркер, чтобы «Восстановить» вернуло связь. На данных владельца этого случая не будет — мягкое удаление и 037
   приходят одним выпуском, — но на стендах и копиях он встречается.

Постановка первого раунда противоречила себе в случае A′: пункт 2 требовал максимум по группе, пункт 3 — повышение
как в `SaveUserLink`. Верен пункт 3.

### Принято без замечаний

- Пара неупорядоченная; вес — только по живым сливаемым строкам; ручная связь не становится автосвязью;
  отклонённая не оживает; поглощённые строки помечены для отката.
- Down-миграция возвращает типы; невосстановимый вес выжившей записан в самом файле.
- `link_type` необязателен в `POST /links` и в пакетном импорте, `openAPI.yaml` обновлён.

`check-all.ps1` на `1584906`: 31 из 33 зелёные, `golangci-lint` пропущен (локально не установлен). Две красные: юнит-тесты бэкенда упали с кодом 1 без единой строки FAIL — повтор той же команды `go test -v -coverprofile=cover.out ./...` зелёный, сбой окружения (память); сторож индекса — дрейф из коммита `1584906` (TASKS-INDEX-3), починен пересборкой индекса в коммите ревью.

## Доработка Devin — хвост LINK-TYPES-1-TAIL, 2026-10-01

Все четыре пункта хвоста закрыты; миграция и тесты в
`037_link_types_merge_related.up.sql` / `migration_037_integration_test.go`
(17/17 зелёные, `-tags=integration`).

- **A′ — автосвязь тяжелее ручной.** Дефект был в миграции: выжившей `user`
  присваивался `max_w` — вес gamma. Теперь при `source_type='user'` и наличии
  живой gamma выжившая хранит **вес пользователя**, а модельный уходит в
  `metadata.gamma` (`score`, `generated_at`) — то же повышение, что
  `SaveUserLink`. Тест `TestA2UserWeightSurvivesOverHeavierGamma` (gamma
  `related` 0,95 + user `reference` 0,5 → живая `related`/`user`/0,5, 0,95 в
  `metadata.gamma`). Мутация «вес выжившей = max_w» — красная.
- **D′ — удалённая тяжелее живой.** Тест `TestD2LiveRowBeatsHeavierDeleted`
  (живая `reference` 0,3 + удалённая `reference` 0,9 встречным направлением —
  живая обязана быть legacy-типа, иначе пара не входит в цикл). Выживает живая,
  удалённая не воскресает. Мутация «убрать `(deleted_at IS NULL)` из
  ORDER BY выбора выжившей» — красная (мёртвая строка оживает).
- **Импорт без `link_type`.** Тест `TestImportBatch_LinkWithoutTypeDefaultsToRelated`:
  пакетный импорт связи без `link_type` → `related`. Мутация «убрать дефолт
  `related`» — красная (`NewLinkType("")` отвергает пустое).
- **E′ — единственная связь пары удалена вместе с заметкой.** Дефект был во
  втором проходе: маркер `deleted_via_note_id` отцеплялся, и «Восстановить»
  возвращало пару без связи. Теперь второй проход раздвоён: via-note строка
  на паре без живых конвертируется в `related` с сохранением маркера —
  `TestE2ViaNoteDeletedLegacyReturnsOnRestore` прогоняет и сам `UPDATE`
  восстановления (оживает одна `related`, 0,7). Коллизия — живая или удалённая
  `related` уже есть на паре — отдаётся в часть 2 (детач маркера):
  `TestE2CollisionDetachesInsteadOfDuplicating` (после восстановления одна
  живая связь, не две).

### Найдено доработкой (адресат — миграция, не постановка)

- **UNIQUE-конфликт одного направления.** `(source,target,link_type)` покрывает
  удалённые строки, а `reference` и `custom` одного направления — разные ключи:
  две via-note строки A→B обоих типов при конвертации обеих в `related` давали
  бы дубликат ключа и обрыв миграции. Часть 1 выбирает одного (тяжелейшего,
  тай-брейк по id) кандидата на направленный ключ — тест
  `TestE2SameDirectionDeletedLegacyNoKeyClash`.
