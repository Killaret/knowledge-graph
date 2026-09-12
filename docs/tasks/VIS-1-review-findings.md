# Findings ревью изменения Argos (`11cab1f`)

Дата: 2026-09-07  
Автор изменения: Claude Code  
Ревьюер: Devin  
Вердикт: **отклонено**

## Что проверено исполнением

1. `git diff 11cab1f^ 11cab1f --ignore-space-at-eol -- .github/workflows/main.yml` показывает единственное содержательное изменение workflow: комментарий и `ARGOS_REFERENCE_BRANCH: ai-agents` → `main`. `SKIP_AUTH: "true"` и запуск только `--project=visual` сохранены осознанно.
2. `docker compose -f docker-compose.test.yml config --quiet` проходит.
3. `FRONTEND_URL=http://127.0.0.1:3002 BACKEND_URL=http://127.0.0.1:18083 npx playwright test --project=visual --list` проходит и перечисляет setup плюс 13 visual-тестов. Локальный upload не запускался: постановка VIS-1 прямо запрещает создавать официальный baseline с машины разработчика.
4. `git diff --numstat 11cab1f^ 11cab1f` показывает 476/473 строки в `main.yml`; с `--ignore-space-at-eol` остаётся только малый содержательный diff — остальное является нормализацией EOL.

## Блокирующие находки

### 1. Из доски потеряна принятая CI-2

Коммит преобразовал разные строки CI-1 и CI-2 в две одинаковые строки CI-1:

```text
CI-1: циклическая зависимость снята
CI-1: циклическая зависимость снята
```

В результате на доске больше нет принятой CI-2 про восстановление `format:check`, хотя журнал и коммит `162c06c` существуют. Воспроизведение:

```powershell
Select-String -Path docs/AI_HANDOFF.md -Pattern 'CI-1: циклическая зависимость снята|CI-2'
```

Ожидание: по одной терминальной строке CI-1 и CI-2. Факт: две CI-1, ноль CI-2 в разделе «На Devin».

### 2. Табличная строка вставлена внутрь «Обмена репликами»

`docs/AI_HANDOFF.md` около строки 392 содержит строку таблицы:

```markdown
| Пересборка baseline Argos: ... | ... | **на ревью** | 2026-09-07 |
```

Она находится не в таблице доски, а между абзацами реплики о setup авторизации. Это повреждает структуру durable handoff и дублирует строку из раздела «На Claude Code».

Ожидание: в «Обмене репликами» обычный prose; строка задачи существует только в таблице владельца.

### 3. Каноническая документация Argos противоречит новой конфигурации

Workflow теперь содержит:

```yaml
ARGOS_REFERENCE_BRANCH: main
```

Но два пользовательских документа по-прежнему утверждают обратное:

- `docs/ARGOS.md:23`: `ai-agents for this work`;
- `docs/ARGOS.md:99`: baseline — `ai-agents`;
- `docs/TESTING.md:467`: baseline branch — `ai-agents for this work`.

Воспроизведение:

```powershell
Select-String -Path docs/ARGOS.md,docs/TESTING.md -Pattern 'ARGOS_REFERENCE_BRANCH|ai-agents'
```

Это блокирует приёмку согласно `.windsurfrules` → Documentation: изменение workflow/configuration требует обновить релевантную документацию. Нужно заменить актуальное значение на `main` и описать, что CI пока сознательно запускает только старый SKIP_AUTH visual-проект до реализации VIS-1.

## Неблокирующая находка

`main.yml` целиком нормализован по EOL, поэтому малое изменение даёт diff почти на тысячу строк. Скрытых содержательных правок не найдено (`--ignore-space-at-eol` это подтверждает), но при исправлении желательно вернуть минимальный diff, если это можно сделать без переписывания истории.

## Что нужно для следующего раунда

1. Восстановить отдельные терминальные строки CI-1 и CI-2 на доске.
2. Удалить дублирующую табличную строку из prose-раздела «Обмен репликами», вернув связный текст.
3. Обновить `docs/ARGOS.md` и `docs/TESTING.md` на `ARGOS_REFERENCE_BRANCH=main` и зафиксировать временное сохранение одного SKIP_AUTH-проекта до VIS-1.
4. Повторить `docker compose ... config --quiet`, `playwright --project=visual --list` и поиск устаревшего `ai-agents`.
