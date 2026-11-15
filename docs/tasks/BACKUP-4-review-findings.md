# BACKUP-4 — разбор ревью

## Ревью №18 (Devin, 2026-11-15) — п. 4, коммит `441a3ef`: **отклонено**

Один подтверждённый исполнением дефект: команда `mongorestore` из нового раздела
«Restoring MongoDB from a `.mongo.gz` dump» под Git Bash на Windows не выполняется —
MSYS переписывает `/tmp/...` в Windows-путь. Предупреждение про `MSYS_NO_PATHCONV=1`
в документе есть, но только в соседнем разделе «Restoring from raw volume archives».
Всё остальное воспроизведено и подтверждено, включая PowerShell-вариант, который автор
не запускал.

### Дефект 1 — `mongorestore` падает под Git Bash (Windows)

**Воспроизведение** (Git Bash, без `MSYS_NO_PATHCONV`):

```text
$ docker exec kgrt-mongo mongorestore --archive=/tmp/restore.mongo.gz --gzip --drop --nsInclude 'knowledge_personal.*'
Failed: stat C:/Users/89209/AppData/Local/Temp/restore.mongo.gz: no such file or directory
0 document(s) restored successfully. 0 document(s) failed to restore.
$ echo $?
1
```

MSYS path conversion переписывает `/tmp/restore.mongo.gz` внутри `--archive=` в
`%TEMP%` Windows. Файл в контейнере на месте (`docker cp` из блока отрабатывает —
его аргумент `kg-mongo-personal:/tmp/...` конверсии не подвержен), восстановление
не происходит, exit code 1. mongorestore при этом печатает
`0 document(s) restored successfully` — строка выглядит как успех, если читать
только хвост вывода; так же ведёт себя и на заведомо отсутствующем файле.

С `MSYS_NO_PATHCONV=1` та же команда отрабатывает: 4 документа восстановлены,
протухший документ убран `--drop`.

Документ решает проблему ровно этой ловушки, но в другом разделе:
«Bash (Git Bash on Windows needs `MSYS_NO_PATHCONV=1`, or it rewrites `/data`
and `/backup`)» — предупреждение стоит перед блоком восстановления сырых архивов.
Раздел `.mongo.gz` выше его не имеет, а падает в нём именно `/tmp`-путь. Владелец
на Windows: для него блок «bash» — это Git Bash.

**Что исправить:** та же оговорка в раздел `.mongo.gz` (или `sh -c 'mongorestore ...'`,
через `sh -c` аргумент конверсии не подвержен — так уже пишет `backup-personal.ps1`).
Заодно, на усмотрение автора: стоит одной фразой сказать, что хвост
`restored successfully` не доказательство — проверка счётчиками уже есть в блоке
верификации, но при Failed-строке на неё нельзя полагаться молча.

### Наблюдение, не дефект

Блок «Making a raw archive» имеет то же подвержение (`tar czf /backup/...`), но стоит
внутри раздела, где предупреждение уже дано двумя блоками выше, — покрыто, если читать
раздел целиком. Автор решает, повторять ли оговорку.

### Проверено исполнением — что подтверждено

Временные тома `kgrt_*` + архивы `personal-volumes-raw-2026-10-01-085733-*` из
`Desktop\my items`, образы `pgvector/pgvector:pg16`, `mongo:7`, `redis:7-alpine`,
`alpine:3.19` — все заявления автора воспроизведены:

| Заявлено | Прогон №18 |
|---|---|
| Postgres открывается, 113 заметок / 34 связи | `113` / `34` — совпало |
| Redis 61 ключ | `DBSIZE` = `61` — совпало |
| В mongo-архиве нет `knowledge_personal` | `[ 'admin', 'config', 'local' ]` — совпало (находка автора к п. 2 верна) |
| Записи архива от `./`, владелец uid 999 | `tar tzvf`: `./`, `999/999` (pg), `999/1000` (redis `dump.rdb`) — совпало |
| `PG_VERSION` = 16 | `16` — совпало |
| Метки Compose снимают предупреждение | `docker volume inspect` — обе метки на месте; сам warning не воспроизводил (требует поднять Personal-стек) |
| `docker cp` + `mongorestore --drop` | 4 документа восстановлены, протухший убран; `zz_keep` вне дампа пережил `--drop` — заявление «коллекция не из дампа не трогается» подтверждено |
| `mongosh` верификация (`getDBNames`, `getCollectionNames` со счётчиками) | синтаксис рабочий на `mongo:7` (живой прогон на `kg-test-mongo` и `kgrt-mongo`) |
| PowerShell-вариант «не запускался» | **запущен** (`kgrt_ps_redisdata`): `docker volume create --label`, `-v "$($PWD.Path):/backup:ro"`, `tar xzf` — dump.rdb на месте. Работает дословно |

Статически сверено с кодом на `93cd426`:

- `kg-postgres-personal` / `kg-mongo-personal` / `kg-redis-personal`, тома
  `pgdata_personal` / `mongodbdata_personal` / `redisdata_personal`,
  `MONGO_DATABASE=knowledge_personal`, `KG_BACKUP_DIR` → `Desktop\my items` —
  `docker-compose.personal.yml`.
- Производители пишут `mongodump --archive --gzip` → `mongorestore --archive --gzip`
  корректен; имя `backup-personal-<mode>-<ts>.mongo.gz` — `backup-personal.{ps1,sh}`.
- Сторож и `check-personal-backup.*` считают сырой архив бэкапом только своего стора
  (`*pgdata*` → pg, `*mongodbdata*` → mongo) и только непустой —
  `guard-personal-data.py:271`, `check-personal-backup.{ps1,sh}`. Совпадает с фразой
  в «Making a raw archive». Redis-архив не считается бэкапом ни для чего — документ
  этого и не обещает.
- `start-personal.{ps1,sh}` отказывают при имени проекта ≠ `knowledge-graph`
  (п. 3 Devin) — `start-personal.ps1:9-13`. Условие 1 документа опирается на это
  верно.
- «No script produces these archives» — по `personal-volumes-raw` в `scripts/`
  только потребители (сторож, проверки, тесты); производителя нет — верно.
- `docs/tasks/README.md` перегенерирован генератором, строка = строке доски.

### Лимит проверки

Предупреждение Compose «already exists but was not created by Docker Compose»
не воспроизводилось: прогон `docker compose -f docker-compose.personal.yml up`
поднял бы Personal-стек, что без явной просьбы запрещено. Механика меток
подтверждена `docker volume inspect`.

`check-all.ps1` (без `-Quick`) на `93cd426` в worktree ревью, полный прогон после
`npm ci` (root + frontend) и `npm run prepare`: **PASS** — workflow sync, guard-сьют
50/50, backend unit (coverage 70.12%) + integration, graph-service, config, все
сторожа (board, tasks-index, decisions, authorship, hooks), NLP 136/136, license
inventory, frontend circular/lint/format/types/BDD-tsc. **FAIL одна:**
`frontend-unit` — 1683/1684, `home-page.svelte.test.ts` («3d»→'graph') — известный
хвост FREEZE-3D-1 (ревью №8), на доске у Devin; к проверяемому коммиту отношения
не имеет — в нём менялись только документы. SKIP: golangci-lint не установлен.

---

## Ревью №17 (Claude Code, 2026-11-15) — пп. 1–3, исполнение Devin: **принято**

Разбор см. журнал `docs/AI_LOG.md` от 2026-11-15 и секцию «Исполнение Devin
(2026-11-15)» в [`BACKUP-4-backup-gaps.md`](BACKUP-4-backup-gaps.md).
