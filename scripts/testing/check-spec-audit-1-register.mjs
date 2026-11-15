#!/usr/bin/env node
/**
 * SPEC-AUDIT-1 completeness guard.
 *
 * Every file in docs/tasks/ (except the generated README.md index and the
 * register itself) must appear in the first column of a stage table inside
 * docs/tasks/SPEC-AUDIT-1-register.md. Reviewer demand: the completeness
 * claim must be repeatable, not a one-off diff.
 *
 * Usage: node scripts/testing/check-spec-audit-1-register.mjs [repoRoot]
 */
import { readdirSync, readFileSync } from "node:fs";
import { join, resolve } from "node:path";

const root = resolve(process.argv[2] ?? ".");
const tasksDir = join(root, "docs", "tasks");
const registerPath = join(tasksDir, "SPEC-AUDIT-1-register.md");

const EXCLUDED = new Set(["README.md", "SPEC-AUDIT-1-register.md"]);

const files = readdirSync(tasksDir)
  .filter((f) => f.endsWith(".md") && !EXCLUDED.has(f))
  .sort();

const register = readFileSync(registerPath, "utf8");

// First-column entries look like `| FILE.md | ...` inside stage tables.
const listed = new Set();
for (const line of register.split("\n")) {
  const m = line.match(/^\|\s*([A-Za-z0-9_.-]+\.md)\s*\|/);
  if (m) listed.add(m[1]);
}

const missing = files.filter((f) => !listed.has(f));
const phantom = [...listed].filter((f) => !files.includes(f));

let failed = false;
if (missing.length) {
  failed = true;
  console.log(`SPEC-AUDIT-1 register: ${missing.length} file(s) in docs/tasks/ missing from the register:`);
  for (const f of missing) console.log(`  - ${f}`);
}
if (phantom.length) {
  failed = true;
  console.log(`SPEC-AUDIT-1 register: ${phantom.length} entr(ies) in the register do not exist in docs/tasks/:`);
  for (const f of phantom) console.log(`  - ${f}`);
}

if (failed) process.exit(1);
console.log(
  `SPEC-AUDIT-1 register completeness OK: ${files.length} task files, all listed in the register.`,
);
