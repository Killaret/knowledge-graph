# SKILLS-AUDIT-1 — разбор ревью №19 (Devin, 2026-11-15)

Вердикт: **отклонено**. Проверено исполнением на `ai-agents` (`8c594b9`): якоря, метрики,
побайтовое равенство описаний, `triggers`, списки команд — прогонами и чтением файлов; вызов
`/kg-*` у Devin — инструментом `skill`.

## Блокер: `: ` внутри `description:` ломает YAML-фронтматтер в 8 файлах

Новые описания-триггеры содержат двоеточие с пробелом внутри обычного скаляра —
`description: ... зависимостей: Clean ...`. По спецификации YAML это запрещено, и строгий
парсер отбрасывает весь файл:

```
$ python -c "yaml.safe_load(<front matter>)"
=== ai-agents
FAIL .claude/skills/kg-backup/SKILL.md     | mapping values are not allowed here
FAIL .claude/skills/kg-graph-3d/SKILL.md   | (то же)
FAIL .claude/skills/kg-layers/SKILL.md     | (то же)
FAIL .claude/skills/kg-regression/SKILL.md | (то же)
FAIL .devin/skills/kg-backup/SKILL.md      | (то же)
FAIL .devin/skills/kg-graph-3d/SKILL.md    | (то же)
FAIL .devin/skills/kg-layers/SKILL.md      | (то же)
FAIL .devin/skills/kg-regression/SKILL.md  | (то же)
OK   kg-new-chat (обе копии), kg-work      | парсится
```

Подтверждено на стороне потребителя, а не только парсером:

- `skill search` по дереву `D:\knowledge-graph-ai-agents` вернул **ровно два** файла:
  `kg-new-chat` (фронтматтер валиден) и `knowledge-graph` (фронтматтера нет — листуется без
  описания). Все четыре скилла с `: ` в описании отброшены.
- `skill invoke kg-layers` в сессии Devin → `Skill "kg-layers" not found`; kg-layers
  отсутствует и в списке доступных на старте сессии.
- На `main` та же ошибка была и до коммита — **только у kg-layers** (старое описание тоже
  содержало `: `). Поэтому kg-layers никогда не появлялся в списках Devin — предсуществующий
  дефект, который ревью должен был поймать, но переписанное описание сохранило ловушку.

На стороне Claude Code дефект невидим: его загрузчик терпелив к битому YAML, таблица вызовов
у него настоящая — указатели у него работают. Но для Devin (строгий парсер) и любого CI-инструмента
файлы мёртвы: цель задачи («человек вызывает каждый скилл по имени у обоих агентов») после правок
для четырёх скиллов у Devin не выполняется.

**Что чинить:** экранировать описания — `description: "…"` в кавычках или свёрнутый скаляр
`description: >-` — в обеих копиях (`.devin/skills/` и `.claude/skills/`); либо убрать `: `
из текста. Проверка после правки: `yaml.safe_load` по всем 12 файлам + вызов по имени.

## Что проверено и подтверждено

| Проверка | Способ | Результат |
|---|---|---|
| Строки `engine.ts` 193/226/247/342/98 и `:153-159` | `sed -n` по файлам на `8c594b9` | все шесть точны; ветка инициализации на 153–159 делает `simulateToStable→updateScene→renderOnce→finishLoading` |
| `Graph3DScene.svelte:39-43` | `sed -n 39,43p` | `$derived stableRender`: проп → `process.env.VITEST` → URL-параметр — точно |
| «50 проверок» сторожа | `python scripts/devops/test_guard_personal_data.py` | `All checks passed (50 assertions)` — точно; новый текст kg-backup соответствует BACKUP-4 (пара `.sql.gz`+`.mongo.gz`, `check-personal-backup.ps1`) |
| Описания указателей = канон | сравнение строк `description:` попарно | EQUAL для всех пяти |
| `user` в `triggers` | чтение фронтматтера | есть у всех пяти проектных скиллов + `kg-work` (`user` only); `knowledge-graph` корректно без триггеров |
| Список команд в `CLAUDE.md` | `grep` | строка 34 `/kg-work` + строка 35 — пять скиллов по имени; каталог 43–47 совпадает с `.devin/skills/` |
| Протокол «Как запускается работа» | `grep` | «одна рабочая команда `/kg-work`; проектные скиллы вызываются по имени» — есть |
| Якорь `domain/README.md` | `ls` | `backend/internal/domain/README.md` есть |
| Команды kg-layers | прогон grep | Go-файлы домена → верхние слои: 0 (3 попадания — текст README, не импорты); `shared`→верхние: 0; `entities`→`features/widgets`: 0; `check:circular` — 249 файлов, 0 циклов (ранее в этой сессии) |
| Якоря kg-regression | `ls` | все файлы есть; `testuser.json` создаётся `frontend/tests/setup/auth.setup.ts:12` при прогоне — «файл создаётся прогоном» верно |
| Относительные пути kg-new-chat | чтение | `../kg-work/SKILL.md`, `WORKTREES.md` разрешаются от каталога скилла |
| Вызов `/kg-*` у Devin | `skill invoke` | `kg-backup`, `kg-graph-3d` вызвались и вернули указатель; `kg-layers` — not found (см. блокер); остальные недостижимы из-за того же дефекта |

## Ответ на вопрос постановки про `.devin/config.yml`

**Вызов по имени не зависит от списка в `config.yml`.** Доказательство исполнением:
`D:\knowledge-graph` (main) имеет `config.yml`, идентичный ai-agents (в списке только
`knowledge-graph` + `kg-work`), — при этом сессия обнаружила и вызвала `kg-backup`/`kg-graph-3d`/
`kg-regression`, которых в списке нет. Обнаружение — сканом каталогов `.devin/skills/` и
`.claude/skills/`, а не чтением `skills:` в `config.yml`. Добавлять пять скиллов в
`config.yml` не требуется.

## Наблюдения (не блокеры)

- `knowledge-graph` не имеет фронтматтера вообще — в списках виден с пустым описанием.
  Как `defaultSkill` работает, но записано сюда.
- Фраза «Единственная проектная команда» осталась в `.devin/skills/kg-work/SKILL.md:8` и
  `.claude/commands/kg-work.md:5` — передана Devin постановкой; правка — в сессии работы,
  не в ревью.
- `check-all.ps1` не прогонялся: правки только в документации/скиллах, вердикт — `отклонено`;
  правило требует прогон перед `принято`.

## Что дальше

Claude Code — экранировать/поправить описания в 8 файлах и повторить сдачу. Devin — после
повторной сдачи: фраза «единственная команда» в двух копиях `kg-work`. SKILLS-GUARD-1 — у владельца.
