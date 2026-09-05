# Передача работы между агентами

Короткий инбокс: что в работе, что ждёт ревью, чья очередь. Читается в начале каждой сессии, обновляется при передаче задачи. Долгая история состояния — в [`PROJECT_REVIEW_AI_AGENTS.md`](PROJECT_REVIEW_AI_AGENTS.md), порядок работы — в [`AI_AGENT_PROTOCOL.md`](AI_AGENT_PROTOCOL.md).

Обновлено: 2026-09-05, ветка `feat/2d-adaptive-fog`.

## Сейчас в работе

| Задача | Кто делает | Статус |
|---|---|---|
| A-1, A-2 — сигнал готовности 3D для визуальных тестов | Devin | постановка: [tasks/A-1-3d-readiness-signal.md](tasks/A-1-3d-readiness-signal.md), брать после A-3. Верификация за Claude Code |
| A-3 — честный результат `run-full-test-cycle.ps1`, снос автокоммита | Devin | ревью пройдено с блокером: агрегация `-contains 1` пропускает коды выхода ≠ 1. Нужна правка |
| B-1 — барьер на удаление данных Personal-стека | Claude Code | сделано, `d302e63`, ревью Devin получено, ждёт правок/утверждения |
| Протокол, `CLAUDE.md`, `.claude/` | Claude Code | сделано, `3cd9282`, ревью Devin получено, ждёт правок/утверждения |

## Ждёт ревью

- ~~`d302e63`~~ — ревью получено, замечания отработаны в `b4c1b88`: закрыт обход через `bash -c`, набор тестов расширен с 17 до 36 проверок, добавлены `ask`-правила и `.devin/config.json`.
- Очищенный `.devin/config.local.json` — из allowlist убраны `Exec(rm)`, `Exec(rmdir)`, `Exec(del)`, добавлен deny на удаление томов. Проверить, не сломало ли легитимные сценарии Devin.

## Решения человека, которых ждём

- `.github/CODEOWNERS` ссылается на несуществующую команду `@knowledge-graph-maintainers` — удалить файл или исправить.
- MCP-коннекторы отключаются в настройках claude.ai.

## Следующее в очереди

После A-1: пересборка baseline Argos и смена `ARGOS_REFERENCE_BRANCH` на `main`, затем пороги покрытия в CI, визуальная регрессия на PR, диагностика WebGL, сквозное покрытие 3D под тегом `@3d`, гигиена эталонов.

Полный список с приоритетами — часть E в [`AI_PROCESS_AUDIT.md`](AI_PROCESS_AUDIT.md).

## Примечания Devin → Claude (2026-09-05)

Краткий обмен по результатам ревью d302e63 и 3cd9282:

- **A-3:** `run-full-test-cycle.ps1` и `.sh` переписаны. Каждая фаза регистрируется, `[Final Summary]` показывает фактический статус, автокоммит удалён, `grep` по `add -A`/`successful regression cycle` чист. Полноценный прогон с подставной упавшей фазой не сделан из-за длительной сборки тест-стека.
- **d302e63:** guard не блокирует изолированный тест-стек (`docker compose -f docker-compose.test.yml down -v` проходит, тесты 17/17). Нужны дополнительные тесты: `--volumes`, multi-volume, `-p knowledge-graph-personal`, цепочки, `newest_backup`/возраст бэкапа. `deny` в `.claude/settings.json` стоит расширить.
- **3cd9282:** протокол и `CLAUDE.md` рабочие. В `AI_HANDOFF.md` п. A-1/A-2 был назначен не той роли — исправлено. `SKILL.md` и `.devin/config.yml` обновлены: Devin теперь читает `AI_HANDOFF.md` и `AI_AGENT_PROTOCOL.md` на старте.
- **.devin/config.local.json:** очищен безопасно, но allowlist очень короткий; рекомендуется создать `.devin/config.json` project-level.
