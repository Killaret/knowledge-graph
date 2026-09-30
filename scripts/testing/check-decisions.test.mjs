// Mutation tests for check-decisions.mjs rule 4 (BOARD-3).
//
// Terminal statuses are "принято" and "отменено" only; such a row on the
// board is an error the moment it appears - it belongs in
// docs/archive/board/YYYY-MM.md. "отклонено" is rework, not closure: a board
// holding it must stay green.
//
// The fixtures replace only the board (--board=); rules 1-3 still run against
// the real repository, so the test suite must run from the repo root.

import { spawnSync } from "node:child_process";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const guard = join(here, "check-decisions.mjs");
const repoRoot = join(here, "..", "..");
const fixtures = join(here, "fixtures", "decisions");

const cases = [
    {
        name: "accepted row still on the board",
        fixture: "terminal-on-board.md",
        expectExit: 1,
        stderrIncludes: ["terminal board row", "docs/archive/board/", "CC-9"],
    },
    {
        name: "rejected row stays green (rework, not terminal)",
        fixture: "rejected-on-board.md",
        expectExit: 0,
        stdoutIncludes: "Decisions OK",
    },
    {
        // CHECK-DECISIONS-2: a marker for a new decision quoting an existing
        // id (MODEL-2) but carrying a date no index row has must fail —
        // before the fix it slipped through on a shared/used row.
        name: "new-decision marker with borrowed id has no row",
        fixture: "rejected-on-board.md",
        tasks: "tasks-probe",
        expectExit: 1,
        stderrIncludes: ["CHECK-DECISIONS-2-probe.md", "2026-10-01", "no matching row"],
    },
    {
        // CHECK-DECISIONS-2 rework: the real-path hole — a file linked by
        // TWO index rows. The same-day markers claim the first row, the
        // second stays free, and the new-date marker rode its file link
        // straight through. With the date check on the strong match it must
        // be red.
        name: "new-decision marker riding a second file-linked row",
        fixture: "rejected-on-board.md",
        tasks: "tasks-probe-tworrow",
        decisions: "DECISIONS-tworrow.md",
        expectExit: 1,
        stderrIncludes: ["TWOROW-1-probe.md", "2026-10-01", "no matching row"],
    },
    {
        // Control: drop the new-date marker — the two citations alone must
        // stay green under the same fixture index.
        name: "same-day citations on a two-row file stay green",
        fixture: "rejected-on-board.md",
        tasks: "tasks-probe-tworrow-green",
        decisions: "DECISIONS-tworrow.md",
        expectExit: 0,
        stdoutIncludes: "Decisions OK",
    },
];

let failed = false;
for (const c of cases) {
    const board = join(fixtures, c.fixture);
    const spawnArgs = [guard, repoRoot, `--board=${board}`];
    if (c.tasks) spawnArgs.push(`--tasks=${join(fixtures, c.tasks)}`);
    if (c.decisions) spawnArgs.push(`--decisions=${join(fixtures, c.decisions)}`);
    const res = spawnSync("node", spawnArgs, {
        encoding: "utf8",
    });
    const code = res.status ?? 1;
    const output = `${res.stdout ?? ""}${res.stderr ?? ""}`;

    const problems = [];
    if (code !== c.expectExit) {
        problems.push(`exit ${code}, expected ${c.expectExit}`);
    }
    for (const needle of c.stdoutIncludes ? [c.stdoutIncludes] : []) {
        if (!output.includes(needle)) problems.push(`stdout missing "${needle}"`);
    }
    for (const needle of c.stderrIncludes ?? []) {
        if (!output.includes(needle)) problems.push(`output missing "${needle}"`);
    }

    if (problems.length) {
        failed = true;
        console.error(`FAIL ${c.name}: ${problems.join("; ")}`);
        console.error(`  output: ${output.trim().slice(0, 400)}`);
    } else {
        console.log(`PASS ${c.name}`);
    }
}

process.exit(failed ? 1 : 0);
