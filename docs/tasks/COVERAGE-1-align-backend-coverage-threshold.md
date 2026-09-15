# COVERAGE-1. Выровнять backend coverage — 70% как цель и enforced min

## Решение владельца

- **Все пороги unit coverage — 70%** (frontend и backend). Цель и enforced min совпадают.
- **Backend** измеряется по unit-testable пакетам с исключённым из знаменателя кодом, который не покрывается unit-тестами: CLI main-файлы (`cmd/cli`, `cmd/embed-recompute`, `cmd/rotate-api-keys`, `cmd/seed`, `cmd/checkmigrations`, `cmd/worker`), generated gRPC client (`internal/infrastructure/graph`), test helpers (`internal/testutil`, `internal/domain/cache/cachetest`) и `scripts`.
- Порог 64.8% отменён; все упоминания приведены к 70%.

## Выполнено

- В `.windsurfrules`, `docs/TESTING.md`, `docs/PROJECT_REVIEW_AI_AGENTS.md` и `.devin/prompts/MASTER_PROMPT*.md` пороги исправлены на 70%.
- `scripts/testing/core-checks.tsv` и `.github/workflows/_core-checks.yml` теперь требуют **backend coverage >= 70%**.
- Локальные `check-all.ps1` и `check-all.sh` используют тот же порог и фильтр.
- Создан `scripts/testing/backend-coverage-total.py` — скрипт, который считает покрытие по `cover.out`, исключая не-unit пакеты (список — `scripts/testing/backend-coverage-excludes.txt`).
- Решение зафиксировано в `docs/DECISIONS.md` (#38).

## Текущее измерение

- Backend unit coverage (filtered): **72.2%**.
- Frontend coverage: **statements 82.21%**, **branches 70.7%**, **functions 82.38%**, **lines 84.01%** — всё выше 70%.

## Исключения из backend-знаменателя

CLI main-файлы, сгенерированный gRPC client, test helpers и ad-hoc скрипты исключены, так как они либо покрываются интеграционными/E2E-тестами, либо являются частью инфраструктуры тестов. Если владелец решит включить их — нужен отдельный план добора тестов.

## Статус

Принято владельцем; реализация на ревью у Claude Code.
