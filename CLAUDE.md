# Knowledge Graph — точка входа для Claude Code

Указатель, а не свод правил. Правила живут в перечисленных ниже файлах и дублировать их здесь нельзя: копия разойдётся с оригиналом.

## Читать перед любой работой

1. [`.windsurfrules`](.windsurfrules) — единственный нормативный источник: архитектура, безопасность, тестирование, язык, Docker, сохранность данных. Имя файла историческое, правила распространяются на всех агентов.
2. [`docs/AI_AGENT_PROTOCOL.md`](docs/AI_AGENT_PROTOCOL.md) — распределение ролей и порядок передачи работы.
3. [`docs/AI_HANDOFF.md`](docs/AI_HANDOFF.md) — что сейчас в работе и чья очередь.
4. [`docs/PROJECT_REVIEW_AI_AGENTS.md`](docs/PROJECT_REVIEW_AI_AGENTS.md) — состояние проекта, недавние правки, известные риски.

По теме задачи дополнительно: [`docs/ARCHITECTURE_SUMMARY.md`](docs/ARCHITECTURE_SUMMARY.md), [`docs/TESTING.md`](docs/TESTING.md), [`docs/REGRESSION_TEST_PLAN.md`](docs/REGRESSION_TEST_PLAN.md), [`docs/BACKUP.md`](docs/BACKUP.md), [`docs/ARGOS.md`](docs/ARGOS.md).

Актуальный аудит обвязки и цепочки верификации — [`docs/AI_PROCESS_AUDIT.md`](docs/AI_PROCESS_AUDIT.md).

## Роль Claude Code

Постановки задач, документация, конфигурация агентов, верификация на живом стеке и ревью чужих изменений. Реализацию кода ведёт Devin. Подробности и исключения — в протоколе.

## Обязательное в этом репозитории

- **Коммиты** — с авторством `--author="Claude Opus 5 <noreply@anthropic.com>"`. Чужая работа коммитится с сохранением настоящего автора.
- **Ветки** — одна ветка, один агент. Перед началом работы `git status`: правящиеся файлы чужие.
- **Свидетельства** — при ручной проверке заполняется поле «Screenshot / Logs» в [`docs/MANUAL_TEST_FEEDBACK.md`](docs/MANUAL_TEST_FEEDBACK.md).
- **Personal-стек** — не запускается без явной просьбы. Команды, способные уничтожить его тома, блокируются хуком [`scripts/devops/guard-personal-data.py`](scripts/devops/guard-personal-data.py), пока нет свежего бэкапа.
- **Изменил норму — обнови производные.** `.windsurfrules` → `.devin/skills/knowledge-graph/SKILL.md` и оба мастер-промпта в `.devin/prompts/`.

## Слэш-команды проекта

- `/kg-work` — прочитать [`docs/AI_HANDOFF.md`](docs/AI_HANDOFF.md) и сделать свою часть. Режим выбирается автоматически: есть непроверенная работа Devin — ревью, нет — своя очередь. Замечания пишутся в [`docs/tasks/`](docs/tasks/), а не в чат.

Префикс `kg-` обязателен для новых команд: без него имя сталкивается со встроенными.

## Проектные скиллы

Хранятся в одном экземпляре в [`.devin/skills/`](.devin/skills/), в `.claude/skills/` — указатели на них.

- `kg-graph-3d` — [подсистема 3D и её ловушки](.devin/skills/kg-graph-3d/SKILL.md)
- `kg-regression` — [тест-стек и полный цикл](.devin/skills/kg-regression/SKILL.md)
- `kg-backup` — [сохранность данных Personal-стека](.devin/skills/kg-backup/SKILL.md)
- `kg-layers` — [границы слоёв](.devin/skills/kg-layers/SKILL.md)

Правила написания новых — в [протоколе](docs/AI_AGENT_PROTOCOL.md), раздел «Как писать скиллы». Коротко: скилл пишется задним числом, обязан ссылаться на реальные артефакты и не дублирует норму.

## Команды

Канонический список — [`COMMANDS.md`](COMMANDS.md) и раздел Common Commands в [`.windsurfrules`](.windsurfrules). Разрешения на прогон тестов уже выданы в [`.claude/settings.json`](.claude/settings.json).
