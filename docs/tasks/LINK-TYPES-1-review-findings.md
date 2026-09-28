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
