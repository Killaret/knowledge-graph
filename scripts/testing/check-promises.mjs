// Promises guard for PROMISES-1.
//
// docs/product/USER_PROMISES.md is the contract with the user: every promise
// must be held by a real, named test. A promise without a test is a promise
// nobody is keeping — the guard fails when a row has no test reference or
// when the referenced test is gone from the spec file.
//
// Catalog row format (markdown table):
//   | P-01 | <promise> | <condition> | <deadline> | `file.spec.ts` · `test name fragment` |
//
// Test cell rules:
//   - `promises.spec.ts` rows: the spec must contain a test whose name
//     includes the promise id (e.g. `P-01`).
//   - other `*.spec.ts` rows: the file must exist and contain the quoted
//     test name fragment.
//   - a promise deliberately outside PR-CI must say so: the cell contains
//     "вне PR-CI" or "регрессионный цикл" — silence is not an exemption.
//
// Usage: node scripts/testing/check-promises.mjs [--catalog=<path>] [--root=<dir>]

import { readFileSync, existsSync } from "node:fs";
import { resolve, join, dirname } from "node:path";

const args = process.argv.slice(2);
let repoRoot = ".";
let catalogPath = null;
for (const arg of args) {
    if (arg.startsWith("--catalog=")) {
        catalogPath = resolve(arg.slice("--catalog=".length));
    } else if (arg.startsWith("--root=")) {
        repoRoot = arg.slice("--root=".length);
    } else {
        repoRoot = arg;
    }
}
repoRoot = resolve(repoRoot);
if (!catalogPath) catalogPath = join(repoRoot, "docs", "product", "USER_PROMISES.md");

const PROMISE_ROW = /^\|\s*(P-\d{2})\s*\|/;
const EXEMPTION = /вне PR-CI|регрессионный цикл|ручной прогон/;
const BACKTICK = /`([^`]+)`/g;

const errors = [];

if (!existsSync(catalogPath)) {
    console.error(`Promises guard failed: catalog not found at ${catalogPath}`);
    process.exit(1);
}

const rows = readFileSync(catalogPath, "utf8").split("\n");
const seen = new Set();

for (const [i, line] of rows.entries()) {
    const m = line.match(PROMISE_ROW);
    if (!m) continue;
    const id = m[1];
    if (seen.has(id)) {
        errors.push(`USER_PROMISES.md:${i + 1}: duplicate promise id ${id}`);
        continue;
    }
    seen.add(id);

    const cells = line.split("|").map((c) => c.trim());
    const testCell = cells[5] ?? "";
    const label = `${id} (USER_PROMISES.md:${i + 1})`;

    if (EXEMPTION.test(testCell) && !testCell.includes(".spec.ts")) continue;

    const refs = [...testCell.matchAll(BACKTICK)].map((g) => g[1]);
    if (refs.length === 0) {
        errors.push(`${label}: no test reference — name a spec file and test, or mark the cell "вне PR-CI"`);
        continue;
    }

    const specFile = refs.find((r) => r.endsWith(".spec.ts"));
    if (!specFile) {
        errors.push(`${label}: test cell has no *.spec.ts reference`);
        continue;
    }
    const specPath = join(repoRoot, "frontend", "tests", specFile);
    if (!existsSync(specPath)) {
        errors.push(`${label}: spec file not found: frontend/tests/${specFile}`);
        continue;
    }
    const spec = readFileSync(specPath, "utf8");

    if (dirname(specPath).endsWith("tests") && specFile === "promises.spec.ts") {
        // The promise spec keys tests by id — the test name must carry it.
        const testName = new RegExp(`test\\(["'\`][^"'\`]*${id}[^"'\`]*["'\`]`);
        if (!testName.test(spec)) {
            errors.push(`${label}: no test in ${specFile} names ${id}`);
        }
        continue;
    }

    const fragment = refs.find((r) => r !== specFile);
    if (!fragment) {
        errors.push(`${label}: name the test inside ${specFile} (second backticked fragment)`);
        continue;
    }
    if (!spec.includes(fragment)) {
        errors.push(`${label}: ${specFile} no longer contains a test named "${fragment}"`);
    }
}

if (seen.size === 0) {
    errors.push("USER_PROMISES.md: no P-xx rows found — the catalog is empty or misformatted");
}

if (errors.length) {
    console.error("Promises guard failed:");
    for (const err of errors) console.error(`  ${err}`);
    process.exit(1);
}

console.log(`Promises OK: ${seen.size} promises catalogued, every testable one names a live test.`);
