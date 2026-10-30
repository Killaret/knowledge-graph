# TASKS-INDEX-1. Постановки нельзя найти, не помня их идентификатора

Постановка для Devin. Ставит Claude Code, реализует Devin, проверяет Claude Code.

## Как нашлось

Владелец спросил про две задачи — здоровье заметки и практику намеренно падающих
тестов — и добавил: «почему ты их сразу не нашёл? там же ещё есть файлы».

Вопрос справедливый, и причина не одна.

**Главная причина моя, не файловая.** Отвечая «что на ком», я рисовал доску и читал
колонку «Статус». Ссылка на `tasks/BATCH-TEST-1-strategy.md` стояла в соседней колонке —
я её не открыл. Строка доски по новому же правилу (BOARD-1) — указатель, а не пересказ;
значит переход по ссылке обязателен, а не желателен. Сегодня я сжал 34 строки именно на
этом доводе, и если по ссылкам не ходить, сжатие делает хуже, а не лучше. Это не чинится
скриптом, это правило поведения, и оно уже записано.

**Вторая причина — файловая, и она чинится.** В `docs/tasks/` 87 файлов, указателя нет,
а `docs/README.md` описывает каталог одной строкой «Постановки задач и замечания ревью».
Поиск идёт по идентификатору с доски — и у 13 файлов идентификатора в имени нет:

```
BATCH-1-api-design.md                    NOTE-TYPE-TAXONOMY.md
BATCH-DDD-1-validation.md                NLP-2-yake-replace-keybert-lemmatization.md
BATCH-TEST-1-strategy.md                 AUD-7a-enforce-boundaries.md
DEPLOY-review-findings.md (→ DEPLOY-2-…)     AUD-7b-lint-tests-and-coverage-denominator.md
AUD-1-review-findings.md        PROJECT-SKILLS-1-review-findings.md
DOCS-LINKS-1-review-findings.md          SPECS-1-review-findings.md
VERIFY-FINDING-MIRROR-1-review-findings.md
```

Доска зовёт задачу **BATCH-TEST-1**, файл называется **BATCH-TEST-1-strategy.md**. Поиск по
`BATCH-TEST-1*` не находит ничего, и файл существует только для того, кто помнит его имя.

## Что сделать

### 1. Указатель, который генерируется

Ручной указатель на 87 строк разойдётся с каталогом за неделю — у нас это уже было
с конфигурацией и промптами. Поэтому `docs/tasks/README.md` **генерируется** скриптом
из самих файлов: идентификатор, заголовок первой строки, дата последнего коммита,
и — если задача есть на доске — её текущий статус.

### 2. Сторож, который краснеет

Фаза в `core-checks.tsv` и шаг в `_core-checks.yml`, падающая когда:

- сгенерированный указатель разошёлся с каталогом (файл добавлен и не попал в указатель);
- файл в `docs/tasks/` не несёт идентификатора, под которым задача известна доске;
- строка доски ссылается на файл, которого нет.

Третий пункт частично покрыт `check-docs-links.mjs`; проверить, не дублируется ли.

### 3. Имена привести к идентификатору

Переименовать 13 файлов так, чтобы имя начиналось с идентификатора с доски.
**Ссылки на них чинить обязательно** — `check-docs-links.mjs` поймает, но лучше
не создавать ему работы. Идентификатор для `BATCH-TEST-1-strategy.md` — `BATCH-TEST-1`,
он уже на доске; для остальных сверить с доской и журналом, а где идентификатора
не существует — завести, а не выдумывать задним числом другой.

### 4. Доказать мутациями

1. Добавить файл в `docs/tasks/`, не трогая указатель → фаза краснеет и называет файл.
2. Положить файл с именем без идентификатора → фаза краснеет и называет имя.
3. Сослаться с доски на несуществующий файл → фаза краснеет.
4. Сгенерировать указатель заново → зелено, `git diff` пуст.

## Критерии приёмки

- `docs/tasks/README.md` генерируется, не пишется руками; в шапке сказано, чем.
- Все файлы каталога несут идентификатор, известный доске или журналу.
- Фаза есть в `check-all` и в CI, число фаз сходится.
- Четыре мутации отработали, вывод приложен.

## Ограничения

- **Содержимое постановок не править.** Задача про имена, указатель и сторожа.
  Постановки и разборы — исторические документы, правка задним числом стирает историю
  решений (`docs/README.md`, раздел про исключения ссылочной проверки).
- Не заводить второй каталог и не перекладывать файлы по подпапкам: указателя
  достаточно, а переезд порвёт ссылки из журнала и истории.

## Выполнено

- `scripts/testing/generate-tasks-index.mjs` генерирует `docs/tasks/README.md` из файлов каталога, `docs/AI_HANDOFF.md` и `docs/AI_LOG.md`. Идентификатор выбирается по самому длинному известному доске/журналу префиксу; поддерживаются суффиксы (`AUD-7a`, `AUD-7b`) и многосегментные идентификаторы (`BATCH-DDD-1`).
- `scripts/testing/check-tasks-index.mjs` ловит: несовпадение между каталогом и указателем, файлы без известного идентификатора, ссылки доски/журнала на отсутствующие файлы.
- Фаза `tasks-index` добавлена в `scripts/testing/core-checks.tsv` и `.github/workflows/_core-checks.yml`.
- Файлы, имена которых не начинались с идентификатора, переименованы: `BATCH-1-api-design.md`, `BATCH-DDD-1-validation.md`, `BATCH-TEST-1-strategy.md`, `AUD-1-review-findings.md`, `DOCS-LINKS-1-review-findings.md`, `SPECS-1-review-findings.md`, `VERIFY-FINDING-MIRROR-1-review-findings.md`, `NLP-2-yake-replace-keybert-lemmatization.md`, `PROJECT-SKILLS-1-review-findings.md`. Три файла из исходного списка 13 (`AUD-7a-enforce-boundaries.md`, `AUD-7b-lint-tests-and-coverage-denominator.md`, `NOTE-TYPE-TAXONOMY.md`) оказались с корректным идентификатором.
- Ссылки на переименованные файлы исправлены; `check-docs-links.mjs` проходит.
- Идентификаторы `PROJECT-SKILLS-1`, `DOCS-LINKS-1`, `SPECS-1`, `VERIFY-FINDING-MIRROR-1`, `DEPENDABOT-79`, `MONGO-1` восстановлены из доски/журнала и при необходимости установлены.
- Мутации:
  1. `MUTATION-1-temp.md` без обновления указателя → `check-tasks-index.mjs` FAIL: `Unknown identifier: MUTATION-1-temp.md` и drift.
  2. `mutation-no-id.md` без идентификатора → FAIL: `Unknown identifier: mutation-no-id.md`.
  3. Строка доски `tasks/NONEXISTENT-999.md` → FAIL: `docs\AI_HANDOFF.md links to missing task file: NONEXISTENT-999.md`.
  4. Удаление мутаций, генерация → `Task index OK: 89 entries, no drift, no broken board links.`
- Проверки:
  - `node scripts/testing/generate-tasks-index.mjs .` — `Generated 89 task index entries at docs\tasks\README.md.`
  - `node scripts/testing/check-tasks-index.mjs .` — `Task index OK: 89 entries, no drift, no broken board links.`
  - `node scripts/testing/check-docs-links.mjs .` — `Docs OK: every local link resolves and every documented npm target exists.`
  - `node scripts/testing/check-core-workflow-sync.mjs` — число фаз совпадает с CI.
- Статус: **на ревью у Claude Code**.
