// Mutation tests for check-promises.mjs (PROMISES-1).
//
// Each fixture root under fixtures/promises/ carries one catalog state; the
// guard must pass the valid one and fail — naming the promise — on the rest.

import { spawnSync } from "node:child_process";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const guard = join(here, "check-promises.mjs");
const fixtures = join(here, "fixtures", "promises");

const cases = [
    {
        name: "valid catalog",
        fixture: "valid-root",
        expectExit: 0,
        stdoutIncludes: "Promises OK: 3 promises",
    },
    {
        name: "promise whose test vanished",
        fixture: "missing-test-root",
        expectExit: 1,
        stderrIncludes: ["P-01", "no test in promises.spec.ts"],
    },
    {
        name: "promise without a test and without an exemption",
        fixture: "silent-exempt-root",
        expectExit: 1,
        stderrIncludes: ["P-09", "no test reference"],
    },
];

let failed = 0;
for (const c of cases) {
    const res = spawnSync(
        process.execPath,
        [guard, "--root", join(fixtures, c.fixture)],
        { encoding: "utf8" },
    );
    const problems = [];
    if (res.status !== c.expectExit) {
        problems.push(`exit ${res.status} (want ${c.expectExit})`);
    }
    for (const needle of c.stdoutIncludes ?? []) {
        if (!res.stdout.includes(needle)) problems.push(`stdout missing "${needle}"`);
    }
    for (const needle of c.stderrIncludes ?? []) {
        if (!res.stderr.includes(needle)) problems.push(`stderr missing "${needle}"`);
    }
    if (problems.length) {
        failed++;
        console.error(`FAIL ${c.name}: ${problems.join("; ")}`);
        if (res.stderr) console.error(res.stderr.trim());
    } else {
        console.log(`ok ${c.name}`);
    }
}

if (failed) {
    console.error(`${failed}/${cases.length} fixture cases failed`);
    process.exit(1);
}
console.log(`check-promises fixtures: ${cases.length}/${cases.length} pass`);
