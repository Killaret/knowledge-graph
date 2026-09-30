# Probe fixture for CHECK-DECISIONS-2

Reproduces the reviewer's probe from `DOC-AUDIT-2-review-findings.md`: a
marker for a NEW owner decision that cites an existing task id (MODEL-2)
must have its own row in `DECISIONS.md`. The citation marker below copies
the original decision's date and is legitimate; the 2026-10-01 marker has
no 2026-10-01 row and must fail.

**Решение владельца (2026-09-24, MODEL-2):** citation of the existing MODEL-2 decision.

**Решение владельца (2026-10-01, MODEL-2):** probe — no 2026-10-01 row exists.
