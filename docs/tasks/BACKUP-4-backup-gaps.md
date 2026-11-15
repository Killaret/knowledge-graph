# BACKUP-4. Три дыры в бэкапе Personal-стека: пустой дамп проходит проверку, MongoDB без бэкапа, стек из чужого каталога

Постановка: Claude Code. Находки сделаны 2026-10-01, тогда же владелец уехал, и в репозиторий они не
попали. Записаны и перепроверены 2026-11-14. Исполнитель: Devin. Продолжение
[BACKUP-3](BACKUP-3-automatic-backups.md).

## Находки

**1. Пустой дамп проходит проверку свежести.** `Desktop\my items\backup-personal-auto-2026-09-30-1441.dump`
весит 913 байт: таблиц в нём нет. Сторож `scripts/devops/guard-personal-data.py` отвергает только файл
нулевого размера (`st_size == 0`, строка 201). Поэтому пустой дамп считается свежим бэкапом. На
следующий день после такого дампа деструктивная команда прошла бы. Перепроверено 2026-11-14: файл на
месте, условие в сторожe то же. `check-personal-backup.ps1` проверить на тот же случай.

**2. MongoDB не бэкапится ничем.** В `docs/operations/BACKUP.md` и в `scripts/devops/backup-personal.*`
MongoDB не упоминается. При этом в ней живут `nlp_artifacts` (NLP-4) и `quality_log` (NOTE-QUALITY-1),
см. решение 99. Перепроверено 2026-11-14 поиском по слову `mongo`. Единственная копия — сырой архив
тома `personal-volumes-raw-2026-10-01-085733-mongodbdata_personal.tar.gz`, его владелец снял вручную.

**3. Personal-стек поднимался из каталога Devin.** 2026-09-30 с 13:47 до 16:38 МСК Personal-стек
работал из `D:\knowledge-graph-ai-agents`. Он создал тома `knowledge-graph-ai-agents_{pg,redis}data_personal`.
Записи в журнале об этом нет. Имя тома по умолчанию берётся из имени каталога. Значит, стек,
запущенный не из `D:\knowledge-graph`, работает на пустых томах, а не на данных владельца. Его бэкапы
перезаписывают настоящие более свежими пустыми. Это вероятная причина находки 1. По
[CLAUDE.md](../../CLAUDE.md) Personal-стек поднимается только по явной просьбе.

## Что сделать

1. Сторож и `check-personal-backup.*`: бэкап считается, только если в нём есть данные. Для `.sql.gz` —
   есть `COPY` или `INSERT` хотя бы одной таблицы заметок. Для `.dump` — `pg_restore --list` показывает
   таблицы. Тест — на копии 913-байтного файла: она должна отвергаться.
2. MongoDB: дамп `mongodump --archive --gzip` теми же производителями, что Postgres. Имя —
   `backup-personal-<mode>-<ts>.mongo.gz` рядом с SQL. Проверка свежести — на оба файла.
3. Запуск Personal-стека не из `D:\knowledge-graph` (имя проекта Compose ≠ `knowledge-graph`) —
   отказ скрипта старта с понятным текстом.
4. `docs/operations/BACKUP.md`: восстановление из сырых архивов томов. Сейчас этот порядок есть только
   в `README.txt` папки переноса `Desktop\knowledge-graph`. Пишет Claude Code.

## Не находка

После 2026-10-01 в `Desktop\my items` нет новых бэкапов. Это не дыра. Бэкап пишет воркер при изменении
заметки, а по `README.txt` переноса данные после 2026-09-21 не менялись. Personal-стек с тех пор не
поднимался (`restart: no`, `d92ba43`).

## Исполнение Devin (2026-11-15)

**П. 1 — бэкап только с данными.** `backup_has_data()` в `guard-personal-data.py` и зеркала в `check-personal-backup.{ps1,sh}`: `.sql[.gz]` — строка `COPY`/`INSERT` таблицы `notes`; `.dump` — `pg_restore --list` с `TABLE ... notes`, отсутствие `pg_restore` → файл непроверяем → не бэкап; `.mongo.gz` — пространство `knowledge_personal.` в сырых байтах или gzip-потоке; `personal-volumes-raw-*` — непустой, по имени тома (`pgdata`/`mongodbdata`). Фикстура — копия 913-байтного файла: `scripts/devops/testdata/backup-personal-empty-envelope.dump`, отвергается.

**П. 2 — MongoDB.** `mongodump --archive --gzip` → `backup-personal-<mode>-<ts>.mongo.gz` рядом с SQL у всех производителей: Go Runner (`WithMongo`, `mongodb-tools` в образе worker), `backup-personal.{ps1,sh}` (локальный mongodump или `docker exec kg-mongo-personal`), cron `backup_scheduler` (`mongodb-tools` + `PERSONAL_MONGO_*` в env). Свежесть — по обоим хранилищам: нет свежего mongo-файла → отказ, даже при свежем pg. Облако и ретенция — на оба файла.

**П. 3 — отказ старта.** `start-personal.{ps1,sh}` вычисляют имя проекта Compose (`COMPOSE_PROJECT_NAME`, иначе имя каталога) и отказывают, если ≠ `knowledge-graph`, с объяснением про префикс томов.

**Проверки:** python-сьют сторожа 50/50 (фикстура-конверт, `notes_archive`, mongo без пространства, pg-only → deny); `go test ./internal/infrastructure/backup` — 5 новых кейсов mongo; ручной прогон `.sh`/`.ps1` на фикстурах — свежая пара pass, конверт и «только pg» → отказ; `start-personal.*` из `knowledge-graph-ai-agents` → отказ до docker. Мутация `backup_has_data → True` — 4 красных включая конверт. Реальный каталог `Desktop\my items`: .ps1 отверг .dump и нашёл устаревший сырой архив → отказ.

**Не трогал:** п. 4 (восстановление из сырых архивов в `docs/operations/BACKUP.md`) — за Claude Code.

## Исполнение Claude Code (2026-11-15) — п. 4

В `docs/operations/BACKUP.md`, раздел «Backup Restoration», добавлены два подраздела:

- «Restoring from raw volume archives» — четыре условия до начала (каталог `knowledge-graph`, томов
  ещё нет, те же мажорные версии, архивы в папке бэкапов), команды для PowerShell и Bash, проверка
  после старта и как снять сырой архив;
- «Restoring MongoDB from a `.mongo.gz` dump» — путь восстановления для дампов из п. 2. Без него
  производители пишут файл, который нечем вернуть.

Указатель — в скилле `kg-backup`.

**Проверено исполнением** на временных томах `kgrt_*` (без `_personal` в имени) с архивами
2026-10-01, Bash-вариант:

- Postgres открылся без ошибок восстановления: 113 заметок и 34 связи. В
  `backup-personal-auto-2026-09-21-214357.sql.gz` тоже 113 строк `notes`.
- Redis — 61 ключ.
- Том без меток Compose даёт предупреждение «already exists but was not created by Docker Compose»,
  том с метками — нет.
- `mongorestore` через `docker cp` вернул тестовый дамп из 3 документов и убрал лишний документ
  (`--drop`).
- Команда снятия архива даёт записи от `./` с владельцем 999, как у архивов 2026-10-01.
- Personal-тома не трогались. Временные тома и контейнеры удалены.

PowerShell-вариант взят из `README.txt` папки переноса и 2026-11-15 не запускался: PowerShell в этой
сессии недоступен.

**Находка к п. 2.** В сыром архиве MongoDB от 2026-10-01 есть только `admin`, `config` и `local`.
Базы `knowledge_personal` на момент снятия не было. Значит, `nlp_artifacts` и `quality_log` в
Personal-стеке тогда не писались, и на 2026-10-01 терять в MongoDB было нечего. Выводы п. 2 это не
отменяет: после включения конвейера (MODEL-2) данные там появятся.

## Ревью пп. 1–3 — Claude Code, 2026-11-15 (ревью №17)

Работа: Devin, `1509c13`. **Вердикт: пп. 1–3 приняты, три мелочи.** П. 4 сделал Claude Code (`441a3ef`), его проверяет
Devin.

### Мутации `guard-personal-data.py` — `python scripts/devops/test_guard_personal_data.py`, 50 проверок

| Мутация | Результат |
|---|---|
| `backup_has_data` всегда `True` | красная: конверт 913 байт, `notes_archive`, не-gzip `.sql.gz` |
| свежесть только по Postgres, MongoDB не требуется | красная: «fresh pg without mongo: expected deny, got allow» |
| нет `pg_restore` — `.dump` засчитывается | красная: «the 913-byte envelope dump was counted as a backup» |
| `.sql.gz` засчитывается без `COPY`/`INSERT` в `notes` | красная: `notes_archive`, не-gzip `.sql.gz` |

### Исполнением

| Что | Как | Результат |
|---|---|---|
| п. 1–2, `check-personal-backup.ps1` и `.sh` | `KG_BACKUP_DIR` на три папки во временном каталоге: конверт + mongo; только `.sql.gz` с `COPY public.notes`; пара `.sql.gz` + `.mongo.gz` с `knowledge_personal.` | оба скрипта: конверт — отказ «Postgres», только pg — отказ «MongoDB», пара — PASS |
| п. 2, производители | `backup-personal.{ps1,sh}`, `runner.go:78` | пишут `.sql.gz` (`--format=plain`) и `.mongo.gz`; ретенция — на оба |
| п. 3, `start-personal.ps1` | подставной `docker.cmd` первым в `PATH`, запуск из `D:\kg-build-cache\rv-eab2` | отказ с объяснением про префикс `rv-eab2_`, `docker` не вызван, код 1. С `COMPOSE_PROJECT_NAME=knowledge-graph` проверка пропускает и доходит до `docker compose … up` |
| раннер Go | `go test ./internal/infrastructure/backup/` | зелёный |

### Мелочь

- **Сырые архивы томов засчитываются по непустоте** (`backup_has_data`, ветка `personal-volumes-raw-`). Архив пустого
  тома `pgdata` — ровно случай находки 3 — весит десятки мегабайт файлов кластера и проходит. Постановка про сырые
  архивы не говорила, и комментарий в коде это ограничение честно называет. Если когда-нибудь их будет снимать
  автоматика, а не владелец руками, — проверять содержимое.
- `_mongo_dump_has_namespace` сначала читает файл целиком (`read_bytes()`). На больших дампах сторож, который
  срабатывает перед каждой подходящей командой, будет тратить память. Достаточно первых мегабайт, как в ветке gzip.
- Образ `worker` с `apk add mongodb-tools` здесь не собирался. Если пакета нет в `alpine:3.19`, сборка упадёт на
  первом деплое.

Последствие для владельца — так и задумано: пока в каталоге бэкапов нет ни одного `.mongo.gz` или сырого архива
`mongodbdata`, сторож отказывает любой деструктивной команде по Personal-стеку и пишет, что сделать.

`check-all.ps1` без `-Quick` на `a2285dd` (включает `1509c13`): фаза «Personal data guard tests» — PASS; всего 34 PASS,
2 FAIL вне задачи (`home-page.svelte.test.ts` — FREEZE-3D-1, порог скорости NLP), 1 SKIP (`golangci-lint`).
