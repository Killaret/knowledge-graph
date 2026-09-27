# Fixture decision index for CHECK-DECISIONS-2 rework

Two rows link the same probe file — the SYNC-1 shape from the review
findings (decisions 69 and 71 both point at the same task file). A marker
carrying a new date on that file must fail: a citation copies the original
date, so the new date can only be a new decision without its own row.

| № | Дата | Решение | Обоснование | Ссылки | Код |
|---|------|---------|-------------|--------|-----|
| 1 | 2026-09-26 | TWOROW-1 first decision | fixture | [probe](tasks-probe-tworrow/TWOROW-1-probe.md) | (кода не требует) |
| 2 | 2026-09-26 | TWOROW-1 second decision, same day | fixture | [probe](tasks-probe-tworrow/TWOROW-1-probe.md) | (кода не требует) |
