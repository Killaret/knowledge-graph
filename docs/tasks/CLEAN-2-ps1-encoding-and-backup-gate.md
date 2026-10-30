# CLEAN-2. Кодировка .ps1 под Windows PowerShell 5.1 + изоляция гейта бэкапа

Продолжение CLEAN-1. Реализовано Devin 2026-09-19 в основном чекауте, статус — **на ревью у Claude Code**.

## Дефект: em-dash в .ps1 ломает парсер PowerShell 5.1

PS 5.1 читает .ps1 без BOM как ANSI (cp1251). Байты `—` (UTF-8 `E2 80 94`)
становятся `â€"`, где `0x94` — это `”` (умная кавычка). PowerShell считает `”`
**закрывающей кавычкой строки**, поэтому em-dash внутри `"..."` обрывает строку,
а остаток строки становится кодом → `ParserError` или искажение строки.
В комментариях (`#`) — безвредно, там это просто текст.

Аудит `scripts/**/*.ps1` (не-ASCII без BOM):

| Файл | Где `—` | Эффект под PS 5.1 |
|---|---|---|
| `cleanup/cleanup-docker.ps1` | строки + комментарии | `ParserError` на строке 282 — скрипт не запускался |
| `testing/lib/phase-tracking.ps1` | строки | dot-source падал → все `Register-Phase`/`Write-FinalSummary` «не найдены» |
| `testing/test-cleanup-honesty.ps1` | строки | та же ловушка (не запускался) |
| `devops/check-personal-backup.ps1` | только комментарии | работал, но заменено для консистентности |
| `testing/run-full-test-cycle.ps1` | есть BOM | **не тронут** — BOM заставляет PS 5.1 читать UTF-8 корректно |

Заменено `—` → `-`/`--` в четырёх файлах. В чекауте `ai-agents` эти файлы уже
ASCII — баг жил только в `main`.

## Улучшение: `KG_BACKUP_DIR` для изоляции теста

`test-cleanup-honesty.ps1` Case 3 предполагал «нет свежего бэкапа», но после
реального бэкапа 19.09 гейт честно пропускал — тест краснел от окружения.
Добавлена переменная `KG_BACKUP_DIR` (дефолт `<repo>/backups` без изменений):

- `devops/check-personal-backup.ps1` — `$backupDir` из env при наличии;
- `devops/check-personal-backup.sh` — `${KG_BACKUP_DIR:-$REPO_ROOT/backups}`;
- `devops/guard-personal-data.py` — `newest_backup` и текст deny учитывают env;
- тест Case 3 пишет фейковый бэкап в `%TEMP%` и выставляет env — «нет бэкапа» теперь детерминировано.

## Проверка

`powershell.exe` (5.1) `test-cleanup-honesty.ps1` → **ALL PASSED** (кейсы 1–5,
включая sh-вариант и гейт бэкапа). Ранее файл вообще не парсился.

## Ограничение фикса

- `pwsh` (PS 7) читал UTF-8 корректно всегда — регрессии нет.
- `run-full-test-cycle.ps1` работает под 5.1 благодаря BOM; эмодзи в его выводе — по назначению.
- Воркфлоу CI/локальные прогоны используют pwsh — поведение не менялось.

## Замечание для Claude Code

Правило проекта для будущих .ps1: либо чистый ASCII, либо UTF-8 **с BOM** —
иначе Windows PowerShell 5.1 читает файл как ANSI и ломается на Unicode-символах
в строках. Возможно, стоит добавить в `lint-scripts.py` проверку «не-ASCII без BOM».
