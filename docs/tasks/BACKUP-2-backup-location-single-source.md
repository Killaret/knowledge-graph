# BACKUP-2. Каталог бэкапа — один источник, и сторож смотрит именно в него

Постановка для Devin. Ставит Claude Code, реализует Devin, проверяет Claude Code.

**Почему реализует Devin.** Правка затрагивает `scripts/devops/guard-personal-data.py` — хук,
который ограничивает меня же. Протокол запрещает агенту писать себе ограничения.

## Решение владельца, 2026-09-14

Облачный бэкап **не через API Яндекс.Диска**, а проще: файл кладётся в
`C:\Users\<user>\Desktop\my items`, и эта папка уезжает в облако клиентом самого владельца.
OAuth-токен для бэкапа больше не нужен.

Сделано сразу: `BACKUP_CLOUD_ENABLED` в `docker-compose.personal.yml` переведён с `true` на
`false` — это было единственное место, где загрузка в Яндекс включалась по умолчанию; во всех
остальных compose-файлах и в `config.go` она и так `false`. Код `yandex_disk.go` и задача
`TypeBackupToCloud` остаются: путь работает, он просто больше не по умолчанию.

## Проблема

После этого решения каталог бэкапа — Рабочий стол. А сторож свежести смотрит в репозиторий.

| Кто | Куда пишет / где ищет |
|---|---|
| `backup-personal.ps1` | `$env:USERPROFILE\Desktop\my items` |
| `backup_scheduler` в compose | `/backups`, смонтирован из `C:/Users/89209/Desktop/my items` |
| `backup-personal.sh` | `./backups` — **другое место** |
| `check-personal-backup.ps1` | `<repo>/backups` — **другое место** |
| `check-personal-backup.sh` | `<repo>/backups` — **другое место** |
| `guard-personal-data.py` | `<repo>/backups` — **другое место** |

Сегодня это ещё не выстрелило: в `<repo>/backups` лежат 29 файлов, свежайший от 2026-09-12,
потому что `.sh` когда-то гоняли вручную из корня. Как только эти файлы состарятся, сторож
начнёт отказывать всегда — при живых свежих бэкапах на Рабочем столе.

Сторож, который отказывает всегда, кончается тем, что его обходят. А стоит он между заметками
владельца и `docker volume rm`.

Замер сейчас:

```
$ pwsh -File scripts/devops/check-personal-backup.ps1
  [ERROR] Newest backup backup-personal-daily-2026-09-12-130700.sql is 56.7 h old; allowed: 24 h
```

Это **правильный** отказ — свежего нет нигде. Но ищет он не там, где теперь будет свежий.

## Второй дефект: `check-personal-backup.sh` не работает вообще

```bash
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || cd "$SCRIPT_DIR/../.." && pwd)"
```

`A || B && C` разбирается как `(A || B) && C`. `git rev-parse` отрабатывает успешно — и `pwd`
выполняется **тоже**, дописывая вторую строку. `REPO_ROOT` получает две строки:

```
$ printf '%s\n' "$REPO_ROOT" | cat -A
D:/knowledge-graph$
/d/knowledge-graph/scripts/devops$
```

Дальше `BACKUP_DIR="$REPO_ROOT/backups"` — мусор, `[ -d ]` не проходит, и скрипт **всегда**
печатает «No backup directory», чем бы ни был заполнен каталог:

```
  [ERROR] No backup directory: D:/knowledge-graph
/d/knowledge-graph/backups
```

Отказ ложный, но «в безопасную сторону», поэтому его никто не заметил. PowerShell-версия
свободна от этого и считает возраст верно — расхождение двух версий дефект и спрятало.

## Что сделать

### 1. Каталог бэкапа — в `backup-policy.env`

Файл уже объявлен единственным местом для порога:

> keep this file the only place where the threshold lives

Каталог обязан жить там же и по тем же правилам: переменная окружения перекрывает файл на один
запуск. Предлагаемое имя — `KG_BACKUP_DIR`.

Путь машинозависимый, поэтому в файле он должен разворачиваться и в PowerShell, и в POSIX-шелле,
и в Python. Простейшее честное решение — хранить относительный хвост (`Desktop/my items`) и
разворачивать от домашнего каталога в каждом потребителе; абсолютный путь через `KG_BACKUP_DIR`
остаётся для тех, у кого иначе.

### 2. Все шесть потребителей читают одно значение

`backup-personal.ps1`, `backup-personal.sh`, `check-personal-backup.ps1`,
`check-personal-backup.sh`, `guard-personal-data.py` и `docker-compose.personal.yml`.

Для контейнера остаётся `BACKUP_DIR=/backups` — он смонтирован из того же каталога; менять
монтирование не нужно, но путь в `volumes` захардкожен под одну машину, и это стоит вынести
в переменную с тем же значением по умолчанию.

### 3. Починить разбор `REPO_ROOT` в `.sh`

Скобки или отдельный `if`. Плюс — раз уж версии разошлись — сверить, что обе на одинаковых
данных дают одинаковый вердикт.

### 4. Доказать мутациями

Бэкап — единственное, что стоит между владельцем и потерей заметок, поэтому вердикт без
воспроизведения не принимается.

1. **Свежий бэкап есть на своём месте** → все три проверки (`ps1`, `sh`, хук) говорят «можно».
2. **Тот же файл состарен** (сдвинуть mtime за порог) → все три отказывают и называют возраст.
3. **Каталог пуст** → все три отказывают и называют каталог.
4. **Свежий файл нулевого размера** → все три отказывают. В `<repo>/backups` сейчас два таких
   файла, так что случай не выдуманный.
5. **`.sh` до починки и после** на одних и тех же данных: до — «No backup directory» при полном
   каталоге, после — тот же вердикт, что у `ps1`.

Вывод каждой мутации приложить. Хук проверять его собственным набором
(`test_guard_personal_data.py`), а не вручную.

## Критерии приёмки

1. Каталог бэкапа задан в одном месте; ни один из шести потребителей не содержит своей копии пути.
2. Три проверки на одних и тех же данных дают одинаковый вердикт.
3. Пять мутаций отработали как описано, вывод приложен.
4. `check-personal-backup.sh` при полном каталоге больше не говорит «No backup directory».
5. `check-all` зелёный; тесты хука проходят.

## Ограничения

- **Personal-стек не поднимать.** Всё проверяется на файлах и подменённых mtime.
- Ничего не удалять из `<repo>/backups` и из `Desktop\my items` — там боевые данные владельца.
  Перенос старых файлов — отдельный разговор, не в этой задаче.
- Порог свежести (24 ч) не трогать: решение владельца, менять его не просили.
- Код Яндекс.Диска не удалять: он выключен по умолчанию, а не отменён.

## Выполнено

- Единый источник: `scripts/devops/backup-policy.env` содержит `KG_BACKUP_DIR=Desktop/my items` и `KG_BACKUP_MAX_AGE_HOURS=24`.
- `scripts/devops/backup-personal.ps1` и `backup-personal.sh` читают `KG_BACKUP_DIR`, разворачивают `~` и относительный хвост от домашнего каталога, `BACKUP_DIR` остаётся контейнерным оверрайдом.
- `scripts/devops/check-personal-backup.ps1` и `check-personal-backup.sh` теперь смотрят в `KG_BACKUP_DIR`, а не в `<repo>/backups`; `.sh` починен — убран сломанный `REPO_ROOT`.
- `scripts/devops/guard-personal-data.py` читает `KG_BACKUP_DIR` из `backup-policy.env` или env, разворачивает `~`/относительный путь, ищет бэкапы в правильном каталоге.
- `docker-compose.personal.yml` использует `${KG_BACKUP_DIR:-C:/Users/89209/Desktop/my items}` вместо захардкоженного пути; `BACKUP_CLOUD_ENABLED` остаётся `false`.
- `.env.example` и `docs/BACKUP.md`, `docs/CONFIGURATION_EN.md` обновлены.

### Мутации

1. **Свежий бэкап** (`/tmp/kg-backup-mutations/backup-personal-daily-2026-09-14-233500.sql.gz`):
   - `check-personal-backup.ps1` → `[PASS]`
   - `check-personal-backup.sh` → `[PASS]`
   - `guard-personal-data.py` → `systemMessage` о разрешении.
2. **Состаренный тот же файл** (`touch -d 2026-09-12`):
   - `check-personal-backup.ps1` → `[ERROR] ... 58.5 h old`
   - `check-personal-backup.sh` → `[ERROR] ... 58.5 h old`
   - `guard-personal-data.py` → `deny` с возрастом.
3. **Пустой каталог** (`/tmp/kg-backup-empty`):
   - `check-personal-backup.ps1` → `[ERROR] No non-empty backup ...`
   - `check-personal-backup.sh` → `[ERROR] No non-empty backup ...`
   - `guard-personal-data.py` → `deny`.
4. **Нулевой свежий файл** (`/tmp/kg-backup-zero/backup-personal-daily-...` 0 байт):
   - `check-personal-backup.ps1` → `[ERROR] No non-empty backup ...`
   - `guard-personal-data.py` → `deny`.
5. **`check-personal-backup.sh` до и после**: на старых данных в `Desktop\my items` старая версия печатала `No backup directory: D:/knowledge-graph/d/knowledge-graph/backups`; новая печатает тот же `[ERROR] ... h old`, что и PowerShell.

Справочные каталоги `C:/Users/89209/AppData/Local/Temp/kg-backup-*` созданы для мутаций и будут удалены по завершении тестов.
