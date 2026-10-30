# DISK-1 findings — исполнение (Devin, 2026-09-24)

Постановка: [`DISK-1-everything-on-d.md`](DISK-1-everything-on-d.md) (решение 57).

Свободное место на C:: **7,1 ГБ (замер Claude 23.09) → 9,7 ГБ (перед B) → 20,5 ГБ (после B) →
33,9 ГБ (после D и E, выполненных Devin 24.09)**. Остаётся на C: только `devin\cli` 4,52 ГБ
(раздел C — команда для владельца ниже, Devin CLI нельзя закрыть изнутри самого себя) → ~38,4 ГБ.

## A. Модель не запекается — выполнено

- `nlp-service/Dockerfile`: убраны `RUN snapshot_download` (builder) и `COPY` кэша в итоговый
  образ; `HF_HOME` в builder больше не нужен.
- `nlp-service/entrypoint.sh`: при пустом кэше + `HF_HUB_OFFLINE=1` — немедленный выход с
  сообщением: смонтированный путь, что задать в `HF_CACHE_DIR`, команда одноразовой закачки.
- `.env.example`: `HF_CACHE_DIR=D:/kg-hf-cache` — раскомментированный дефолт.
- `BOOTSTRAP.md`: шаг 3.5 «Диски» + правка формулировки про модель.

Проверка сборкой (`docker build`, кэш пустой — честный прогон):
лог без `snapshot_download` — шаги: apt, venv, pip, nltk, COPY. Мутация — сборка со старым
Dockerfile (`git show HEAD~`): шаг `snapshot_download` присутствует и выполняется
(слой закэширован повторным прогоном).

Проверка запуском:

| Сценарий | Результат |
|---|---|
| `-v D:/kg-hf-cache:...` + `HF_HUB_OFFLINE=1` | `Model found in cache. Skipping download.` — старт продолжается |
| пустой кэш + `HF_HUB_OFFLINE=1` | ERROR с путём, подсказкой `HF_CACHE_DIR=D:/kg-hf-cache` и командой закачки; exit 1 сразу |

## B. Кэши инструментов — на D:

Переключено (все пути подтверждены `check-disk-layout.ps1`):

| Инструмент | Было (C:) | Стало | Способ |
|---|---|---|---|
| Go сборка | `AppData\Local\go-build` 2,53 ГБ | `D:\kg-build-cache\go-build` | `go env -w GOCACHE` |
| Go модули | `go\pkg\mod` 2,98 ГБ | `D:\kg-build-cache\go-mod` | `go env -w GOMODCACHE` + robocopy /E |
| Go temp | — | `D:\kg-build-cache\tmp` | `go env -w GOTMPDIR` |
| npm | `npm-cache` 1,30 ГБ | `D:\kg-build-cache\npm` | `npm config set cache --location=user` |
| pip | `pip\cache` 0,54 ГБ | `D:\kg-build-cache\pip` | `pip config set global.cache-dir` |
| Playwright | `ms-playwright` 0,69 ГБ | `D:\kg-build-cache\ms-playwright` | `setx` + robocopy /E |
| puppeteer | `.cache` | `D:\kg-build-cache\puppeteer` | `setx` |
| torch | `.cache` | `D:\kg-build-cache\torch` | `setx` |
| HF (скрипты) | `.cache` | `D:\kg-hf-cache` | `setx HF_HOME` |

Проверки после переключения: `go test ./...` бэкенда — все пакеты зелёные (testcontainers
тянул образы заново — данные Docker уже на D:); `npm run test:unit` — **142 файла, 1446
тестов зелёных**. После зелёных сборок старые каталоги на C: удалены
(`go\pkg\mod`, `go-build`, `npm-cache`, `pip\cache`, `ms-playwright`, `.cache`) — **−10,8 ГБ**.

`setx` переменные действуют на новые процессы — открытые до переключения шеллы живут со
старыми путями до перезапуска.

## C. Сессии Devin CLI — команды для владельца

Настройки каталога данных в документации Devin CLI нет (просмотрены docs агента) — путь
один: junction при закрытом приложении.

Что в базе: `sessions.db` 4,29 ГБ — 108 сессий, 391 367 `message_nodes`, 61 890
`tool_call_state`. Штатной очистки старых сессий в CLI не нашлось; база растёт от
накопления истории. Если удаление старых сессий из UI доступно — это уменьшит файл.

**Команды (закрыть Devin Desktop полностью, включая трей):**

```cmd
mkdir D:\agents\devin
robocopy "%APPDATA%\devin\cli" "D:\agents\devin\cli" /E /MOVE
mklink /J "%APPDATA%\devin\cli" "D:\agents\devin\cli"
```

Переносится только подкаталог `cli` (вся масса в нём), профиль приложения остаётся на C:.
Проверка после запуска Devin: `dir "%APPDATA%\devin\cli"` — тип `JUNCTION`, сессии на месте,
`sessions.db` физически на D: (`dir D:\agents\devin\cli`).

## D. VM Cowork (Claude Desktop) — команды для владельца

Настройки места у приложения нет (подтверждено разбором Claude 23.09) — junction.

**Выполнено Devin 2026-09-24** (владелец закрыл Claude Desktop; `vmwp` на тот момент — это
VM Docker-десктопа, не Cowork): robocopy перенёс 9,42 ГБ за ~1 мин, исходный каталог удалён,
`Junction -> D:\Claude\vm_bundles` создан и проверен — `claudevm.bundle` читается по старому пути.
Если при обновлении приложение пересоздаст каталог на C: — junction не переживается, вопрос
владельцу: нужен ли Cowork.

## E. Дистрибутив WSL Ubuntu — разведка

- Состояние: `Stopped`, версия WSL2, дистрибутив по умолчанию.
- Внутри **2,4 ГБ**: `/usr` 1,2 ГБ (пакеты), `/var` 1,2 ГБ (из них `/var/log` 858 МБ — логи),
  `/root` 1,2 МБ (только `.docker`-конфиг), `/home` пуст — **пользовательских данных нет**.
- Последнее использование: **2026-05-28** (4 месяца назад).
- Это стоковая Ubuntu с накопленными логами; Docker Desktop работает через отдельный
  `docker-desktop` дистрибутив.

**Выполнено Devin 2026-09-24** по варианту «не нужен»: `wsl --export` в
`D:\WSL\backup\ubuntu-2026-09-24.tar` (2,0 ГБ), затем `wsl --unregister Ubuntu` — vhdx с C:
удалён, `wsl -l -v` показывает только `docker-desktop`. Восстановление при необходимости:
`wsl --import Ubuntu D:\WSL\Ubuntu D:\WSL\backup\ubuntu-2026-09-24.tar`.

## F. Сторож — `scripts/devops/check-disk-layout.ps1`

- Печатает каждый настроенный путь (go env ×3, npm, pip, 4 переменные пользователя из
  реестра, `HF_CACHE_DIR` из `.env`) и размеры известных тяжёлых каталогов на C:.
- Системный диск определяется через `$env:SystemDrive`, а не буквой `C:` (уточнение владельца).
- Падает (exit 1), если настроенный путь ведёт на системный диск — с командой исправления.
- Junction/symlink в списке тяжёлых каталогов показывается как `junction -> target`, без обхода
  (иначе перенесённое продолжало бы считаться лежащим на C:).
- Мутация пройдена: `go env -w GOCACHE=<C:\...>` → `FAIL: go GOCACHE -> C:\...`; возвращено.
- В `check-all` не включён (у раннера своя раскладка). Шаг «Диски» — в `BOOTSTRAP.md`.
- Файл ASCII-only (сторож `check-ps1-ascii.py` зелёный).

## Осталось за владельцем

Только пункт C — junction для `%APPDATA%\devin\cli` (−4,5 ГБ): каталог занят, пока Devin работает,
поэтому выполняется после закрытия Devin CLI/Desktop. Команды (cmd, не PowerShell):

```cmd
mkdir D:\agents\devin
robocopy "%APPDATA%\devin\cli" "D:\agents\devin\cli" /E /MOVE
if exist "%APPDATA%\devin\cli" rmdir "%APPDATA%\devin\cli"
mklink /J "%APPDATA%\devin\cli" "D:\agents\devin\cli"
dir "%APPDATA%\devin"
```

Проверка: `dir` показывает `cli <JUNCTION> [D:\agents\devin\cli]`; после запуска Devin — сессии
на месте, `sessions.db` физически на D:. Если `robocopy` выдал ошибки доступа — Devin не до конца
закрыт (трей, фоновые процессы).

**Выполнено владельцем 2026-09-24:** junction создан, `sessions.db` (4,0 ГБ) физически на D:,
сессия Devin пишет через ссылку. C:: 33,9 → ~39 ГБ свободных.

## Дополнение 2026-09-24: сжатие vhdx Docker

`docker builder prune` освобождает место **внутри** виртуального диска, но файл
`docker_data.vhdx` на хосте сам не уменьшается. Рабочая последовательность (проверена):
`wsl -d docker-desktop -e sh -c "fstrim -av"` (distro должен быть запущен — trim метит
освобождённые блоки) → остановить Docker Desktop → `wsl --shutdown` → `diskpart` со скриптом
`select vdisk` / `attach vdisk readonly` / `compact vdisk` / `detach vdisk`. Без fstrim compact
отчитывается «successfully compacted», но не возвращает ничего. Без остановки Docker Desktop
файл остаётся залоченным. Итог замера: 34 → 18 ГБ vhdx, D:: 2,7 → 19 ГБ свободных.
Готовый bat: `D:\agents\compact-docker-vhdx.bat` (требует админа). `wsl --manage
docker-desktop --set-sparse` отклонён самой WSL (`--allow-unsafe`, риск повреждения томов с БД).
