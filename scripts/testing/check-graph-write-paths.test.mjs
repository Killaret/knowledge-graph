// Tests for check-graph-write-paths.mjs (SYNC-1).
//
// The real backend is green; a fixture backend containing a write path
// without a Publish call must turn the guard red, and a fully covered or
// allowlisted backend stays green.

import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const here = dirname(fileURLToPath(import.meta.url));
const repoRoot = join(here, "..", "..");
const guard = join(here, "check-graph-write-paths.mjs");
const fixtureBackend = join(here, "fixtures", "graph-write-paths");

const cases = [
    {
        name: "uncovered write path is red",
        backend: fixtureBackend,
        expectExit: 1,
        stderrIncludes: ["UncoveredSave", "write path without a graph event"],
    },
    {
        name: "real backend stays green",
        backend: null,
        expectExit: 0,
        stdoutIncludes: "Write-path guard OK",
    },
];

let failed = false;
for (const c of cases) {
    const args = [guard, repoRoot];
    if (c.backend) args.push(`--backend=${c.backend}`);
    const res = spawnSync("node", args, { encoding: "utf8" });
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
