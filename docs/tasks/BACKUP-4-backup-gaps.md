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
