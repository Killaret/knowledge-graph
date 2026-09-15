# COVERAGE-1. Согласовать backend coverage: 70% target vs 64.8% enforced

Вопрос владельцу: backend unit coverage сейчас **target 70%**, но **enforced min 64.8%**. Нужно ли поднять enforced min до 70% и привести все документы к одной цифре, или 64.8% остаётся рабочим минимумом, пока coverage не выросла?

## Что сейчас в репозитории

| Источник | Что говорит |
|---|---|
| `.windsurfrules` | Go backend: **Target 70% (enforced min 64.8%, measured 66.8% on 2026-09-07)** |
| `docs/PROJECT_REVIEW_AI_AGENTS.md` §8 | Go unit: **target 70%, min 60%** |
| `docs/TESTING.md` | Backend statements: **70% (min 60%)** (данные устарели — 2026-07-20) |
| `scripts/testing/core-checks.tsv` | `backend-coverage`: **Backend coverage >= 64.8%** |
| `.github/workflows/_core-checks.yml` | `Check backend coverage threshold (min 64.8%)` |
| `.devin/prompts/MASTER_PROMPT.md` и `MASTER_PROMPT_RU.md` | Go unit: **Target 70% coverage, min 60%** |
| `docs/DECISIONS.md` #4 | **AUD-7b: … порог 70 %**, но формулировка про `src/**` — это frontend, не backend |
| Последний локальный прогон | **66.4%** (порог 64.8% пройден, 70% не пройден) |

## Варианты решения

1. **Enforced min = 70% сейчас.** `check-all` и CI будут красными, пока coverage не поднимется. Мотивирует добирать тесты, но блокирует зелёный CI.
2. **Enforced min оставить 64.8%, target — 70%.** Дрейф сохраняется; зелёный CI, но нет гарантии, что 70% достигнут.
3. **Принять поэтапный план:** min 70% через N дней/итераций, с планом доработки coverage (какие пакеты добирать).

## Что нужно от владельца

- Какой вариант выбрать?
- Если 70% — согласен ли на красный CI до тех пор, пока coverage не выросла, или нужен план добора тестов?
- Если 64.8% — зафиксировать это как временное решение в `DECISIONS.md` и привести `.devin/prompts/`, `TESTING.md`, `PROJECT_REVIEW_AI_AGENTS.md` к одной формулировке.

## Статус

Ждёт решения владельца / обсуждение с Devin.
