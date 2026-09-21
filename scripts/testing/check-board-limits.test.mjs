// Mutation tests for check-board-limits.mjs (BOARD-2).
//
// Each fixture under fixtures/board-limits/ carries one violation; the guard
// must fail and name the offending row. The valid board must pass, and the
// owner-section overflow must pass with a warning (soft limit, exit 0).

import { spawnSync } from "node:child_process";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const guard = join(here, "check-board-limits.mjs");
const fixtures = join(here, "fixtures", "board-limits");

const cases = [
    {
        name: "valid board",
        fixture: "valid.md",
        expectExit: 0,
        stdoutIncludes: "Board limits OK",
    },
    {
        name: "fourth active row in an agent section",
        fixture: "m1-fourth-active.md",
        expectExit: 1,
        stderrIncludes: ['На Devin: 4 rows', "limit is 3", "DEV-4"],
    },
    {
        name: "sixth row under review board-wide",
        fixture: "m2-sixth-review.md",
        expectExit: 1,
        stderrIncludes: ["6 rows in \"на ревью\"", "limit is 5"],
    },
    {
        name: "status outside the dictionary",
        fixture: "m3-bad-status-word.md",
        expectExit: 1,
        stderrIncludes: ["status outside the dictionary", "обсуждается", "DEV-2"],
    },
    {
        name: "backlog status inside an agent section",
        fixture: "m4-backlog-status-in-agent.md",
        expectExit: 1,
        stderrIncludes: ["does not belong in an agent section", "DEV-2"],
    },
    {
        name: "backlog row over 500 bytes",
        fixture: "m5-backlog-row-501.md",
        expectExit: 1,
        stderrIncludes: ["limit is 500", "BL-1"],
    },
    {
        name: "eight rows for the owner (warning only)",
        fixture: "m6-owner-8-rows.md",
        expectExit: 0,
        stderrIncludes: ["На человеке: 8 rows", "soft limit is 7"],
    },
];

let failed = false;
for (const c of cases) {
    const board = join(fixtures, c.fixture);
    const res = spawnSync("node", [guard, ".", `--board=${board}`], {
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
