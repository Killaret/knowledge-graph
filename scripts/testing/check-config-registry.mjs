// CONFIG-AUDIT-1 registry guard — mirrors check-spec-audit-1-register.mjs.
//
// 1. Drift: docs/operations/CONFIG_REGISTRY.md must equal the output of
//    `node scripts/testing/generate-config-registry.mjs . --check`.
// 2. Completeness: every flattened leaf key of config/*.json must appear as a
//    registry row, and every registry row's key must exist in config/*.json.
// 3. Evidence: rows not marked «мёртвый» must name a reader; «мёртвый» is a
//    deliberate verdict — the owner decides keep-or-remove (spec item 4).
//
// Usage: node scripts/testing/check-config-registry.mjs [repoRoot]

import { readFileSync, readdirSync } from "node:fs";
import { join, resolve } from "node:path";
import { execSync } from "node:child_process";

const repoRoot = resolve(process.argv[2] ?? ".");
const registryPath = join(repoRoot, "docs", "operations", "CONFIG_REGISTRY.md");

function flat(obj, prefix = "", out = new Set()) {
  for (const [k, v] of Object.entries(obj)) {
    if (v && typeof v === "object" && !Array.isArray(v)) flat(v, prefix + k + ".", out);
    else out.add(prefix + k);
  }
  return out;
}

const configKeys = new Set();
for (const f of readdirSync(join(repoRoot, "config")).filter((n) => n.endsWith(".json"))) {
  flat(JSON.parse(readFileSync(join(repoRoot, "config", f), "utf8")), "", configKeys);
}

const registry = readFileSync(registryPath, "utf8");
const rowKeys = new Set();
let missingReader = [];
for (const line of registry.split("\n")) {
  const m = line.match(/^\| `([a-z0-9_.\-]+)` \|/);
  if (!m) continue;
  rowKeys.add(m[1]);
  if (!line.includes("мёртвый")) {
    const cells = line.split("|").map((c) => c.trim());
    if (!cells[3] || cells[3] === "—") missingReader.push(m[1]);
  }
}

const missingFromRegistry = [...configKeys].filter((k) => !rowKeys.has(k));
const phantom = [...rowKeys].filter((k) => !configKeys.has(k));

let failed = false;
if (missingFromRegistry.length) {
  failed = true;
  console.log(`CONFIG-REGISTRY: ${missingFromRegistry.length} key(s) missing from the registry:`);
  for (const k of missingFromRegistry) console.log(`  - ${k}`);
}
if (phantom.length) {
  failed = true;
  console.log(`CONFIG-REGISTRY: ${phantom.length} row(s) reference keys not in config/*.json:`);
  for (const k of phantom) console.log(`  - ${k}`);
}
if (missingReader.length) {
  failed = true;
  console.log(`CONFIG-REGISTRY: ${missingReader.length} row(s) lack a reader and are not marked мёртвый:`);
  for (const k of missingReader) console.log(`  - ${k}`);
}

let expected;
try {
  expected = execSync("node scripts/testing/generate-config-registry.mjs . --check", {
    cwd: repoRoot,
    encoding: "utf8",
    maxBuffer: 64 * 1024 * 1024,
  });
} catch (e) {
  console.log("CONFIG-REGISTRY: generator failed:", e.message);
  process.exit(1);
}
if (expected !== registry) {
  failed = true;
  const e = expected.split("\n"), a = registry.split("\n");
  for (let i = 0; i < Math.max(e.length, a.length); i++) {
    if (e[i] !== a[i]) {
      console.log(`CONFIG-REGISTRY: drift at line ${i + 1}:`);
      console.log(`  expected: ${e[i] ?? "<none>"}`);
      console.log(`  actual:   ${a[i] ?? "<none>"}`);
      break;
    }
  }
  console.log("  Run scripts/testing/generate-config-registry.mjs .");
}

if (failed) process.exit(1);
console.log(`CONFIG-REGISTRY OK: ${rowKeys.size} keys registered, no drift, every live key names a reader.`);
