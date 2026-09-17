# Ревью: Проектные скиллы (e95eba8)

**Что проверялось.**

- Коммит `e95eba8` — `.devin/skills/kg-backup`, `kg-graph-3d`, `kg-layers`, `kg-regression`, а также `.claude/skills/*` указатели, `.devin/skills/knowledge-graph/SKILL.md`, `.windsurfrules`, `CLAUDE.md`, `docs/AI_AGENT_PROTOCOL.md`.
- Проверка якорей: упомянутые файлы, команды, строки/участки кода, порты, скрипты — существуют.
- Проверка внутренних ссылок по всем markdown-файлам проекта — обрывающихся ссылок нет.
- Быстрые пробы команд из `kg-layers` (`grep` по `backend/internal/domain/` и `frontend/src/shared/`, `frontend/src/entities/`) — ноль нарушений, как заявлено.

**Находка.**

- `.devin/skills/knowledge-graph/SKILL.md` — потерялся пункт `Roadmap` в раздел `## Project Skills`.
  
  Сейчас там:
  ```markdown
  ## Project Skills

  Stored once under `.devin/skills/`; `.claude/skills/` holds pointers to the same text.

  - `kg-graph-3d` — 3D subsystem map and its traps
  - `kg-regression` — test stack and the full regression cycle
  - `kg-backup` — Personal stack data safety
  - `kg-layers` — layer boundaries and how to check them

  How to write new ones: `docs/AI_AGENT_PROTOCOL.md`, section «Как писать скиллы». A skill is written after an incident, must cite real artifacts, and never duplicates the norm.
  - Roadmap: `ROADMAP.md`; детальные планы: `docs/BACKLOG.md`

  ## Devin Workflow
  ```

  `Roadmap` — это навигационный пункт, а не project skill. Ранее он входил в `## Primary Navigation`. Его смещение под `How to write new ones:` делает раздел `## Project Skills` визуально незавершённым и сбивает структуру.

**Как исправить.**

1. Убрать строку `- Roadmap: ROADMAP.md; детальные планы: docs/BACKLOG.md` из конца раздела `## Project Skills`.
2. Вернуть её в `## Primary Navigation` (после `Current audit: docs/AI_PROCESS_AUDIT.md`) — так было до коммита.

**Всё остальное в порядке.**

- Новые скиллы содержат якоря на реальные файлы и команды; источники указаны в шапке.
- `.claude/skills/*` корректно указывают на `.devin/skills/*` через `../../../`.
- Чек-лист безопасности влит в `.windsurfrules` без нарушения соседней разметки.
- `docs/AI_AGENT_PROTOCOL.md` и `CLAUDE.md` корректно описывают склад и правила написания скиллов.
- Проектная ссылочная проверка после коммита — чисто.
