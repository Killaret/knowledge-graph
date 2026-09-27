# Probe fixture for the CHECK-DECISIONS-2 rework

Reproduces the real-path hole the review found: the file is linked by two
index rows (DECISIONS-tworrow.md rows 1 and 2). The two same-day markers
are legitimate citations — both claim the first 2026-09-26 row, leaving the
second free. The 2026-10-01 marker is a new decision with no index row; it
must not ride the free row's file link.

**Решение владельца (2026-09-26, TWOROW-1):** citation of the first decision.

**Решение владельца (2026-09-26, TWOROW-1):** citation of the second decision — same id, same day, free row.

**Решение владельца (2026-10-01, TWOROW-1):** probe — no 2026-10-01 row exists.
