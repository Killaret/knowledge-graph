// Mutation tests for generate-config-registry.mjs (CONFIG-AUDIT-1).
//
// The fixture repo under fixtures/config-registry/ carries frontend keys with
// regex metacharacters (`+`, `*`). Every way a key reaches a search pattern —
// direct text, alias suffix, `export const` — must escape it fully; reverting
// to `key.replace(/\./g, "\\.")` (CodeQL js/incomplete-sanitization) turns the
// corresponding row «мёртвый» and this test goes red.

import { spawnSync } from "node:child_process";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const generator = join(here, "generate-config-registry.mjs");
const fixture = join(here, "fixtures", "config-registry");

const res = spawnSync("node", [generator, fixture, "--check"], {
    encoding: "utf8",
    maxBuffer: 16 * 1024 * 1024,
});
const out = `${res.stdout ?? ""}${res.stderr ?? ""}`;

const rowOf = (key) =>
    out.split("\n").find((l) => l.startsWith(`| \`${key}\` |`));

const cases = [
    {
        name: "direct key pattern escapes metachars",
        key: "frontend.a+b",
        reader: "reader-direct.ts",
    },
    {
        name: "alias suffix pattern escapes metachars",
        key: "frontend.c+d",
        reader: "reader-alias.ts",
    },
    {
        name: "export const pattern escapes metachars",
        key: "frontend.e*f",
        reader: "reader-export.ts",
    },
    {
        name: "control key without metachars is found",
        key: "frontend.normal.x",
        reader: "reader-direct.ts",
    },
];

if (res.status !== 0) {
    console.error(`FAIL generator exited ${res.status}: ${out.trim().slice(0, 400)}`);
    process.exit(1);
}

let failed = false;
for (const c of cases) {
    const row = rowOf(c.key);
    const problems = [];
    if (!row) {
        problems.push("row missing");
    } else {
        if (row.includes("мёртвый")) problems.push("marked мёртвый");
        if (!row.includes(c.reader)) problems.push(`reader ${c.reader} missing`);
    }
    if (problems.length) {
        failed = true;
        console.error(`FAIL ${c.name}: ${problems.join("; ")}`);
        console.error(`  row: ${(row ?? "<none>").slice(0, 200)}`);
    } else {
        console.log(`PASS ${c.name}`);
    }
}

process.exit(failed ? 1 : 0);
