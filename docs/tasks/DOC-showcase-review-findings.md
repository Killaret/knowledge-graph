# Ревью: Витрина документации (62a963d)

**Что проверялось.**

- Изменения коммита `62a963d` — README, ROADMAP, CHANGELOG, BACKLOG, IDEAS, docs/README.md, ARCHITECTURE_ROADMAP.md, CRITICAL_FIXES.md, PROJECT_REVIEW_AI_AGENTS.md, а также .devin/prompts/* и .devin/skills/knowledge-graph/SKILL.md.
- Ссылки во всех markdown-файлах проекта (кроме `node_modules`, `.git`, `.kilo`) скриптом `scripts/tmp_link_check_project2.py`.
- Существование и корректность путей в `docs/README.md`.

**Результат.**

Найдена одна действительно обрывающаяся внутренняя ссылка:

- `docs/PROJECT_REVIEW_AI_AGENTS.md:615`
  ```markdown
  - Базовый и «туманный» снимки: [`docs/assets/a1-3d-visual-regression/`](../assets/a1-3d-visual-regression/)
  ```
  Файл `docs/PROJECT_REVIEW_AI_AGENTS.md` находится в `docs/`, поэтому `../assets/...` ведёт в `D:\knowledge-graph\assets\`, которого нет. Реальная директория — `D:\knowledge-graph\docs\assets\a1-3d-visual-regression\`. Правильный относительный путь: `assets/a1-3d-visual-regression/`.

**Почему это отклоняет приёмку.**

В сообщении коммита и в строке доски прямо заявлено, что «no dangling links remain» / «ROADMAP.ru.md и IDEAS_EN.md удалены как дубликаты» и ссылки перенаправлены. Проверка исполнением показывает, что висячая ссылка осталась в изменённом же файле `docs/PROJECT_REVIEW_AI_AGENTS.md`.

**Ложноположительное срабатывание.**

`docs/MASS_IMPORT_TEST_PLAN.md:9` содержит inline-код `` `[title](url)` ``, который скрипт без учёта обратных кавычек интерпретировал как ссылку. После фильтрации inline-кода это срабатывание исчезает.

**Что нужно исправить.**

1. В `docs/PROJECT_REVIEW_AI_AGENTS.md:615` заменить `](../assets/a1-3d-visual-regression/)` на `](assets/a1-3d-visual-regression/)`.
2. Перезапустить `python scripts/tmp_link_check_project2.py` (или эквивалент) и убедиться, что `BROKEN_LINKS` пуст.

**Всё остальное в порядке.**

- `README.md`, `ROADMAP.md`, `CHANGELOG.md`, `docs/BACKLOG.md`, `docs/IDEAS.md`, `docs/README.md` — структура и связи между файлами корректны.
- `ROADMAP.ru.md` и `docs/IDEAS_EN.md` удалены; ссылки в промптах, скилле и трёх документах перенаправлены на `ROADMAP.md` и `docs/BACKLOG.md`.
- `docs/README.md` индексирует 60+ файлов, помечает архивные, все ссылки из индекса валидны.
- `docs/ARCHITECTURE_ROADMAP.md` помечен как архивный в собственном заголовке.
