// License inventory check (NLP-2).
// Verifies that every DIRECT dependency of the three services is listed in
// docs/LICENSES.md and that no copyleft license (GPL/LGPL/AGPL/SSPL/EUPL)
// is recorded. This is a process gate, not a transitive audit — see the
// header of docs/LICENSES.md.
//
// Usage: node scripts/testing/check-licenses.mjs

import { readFileSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..", "..");
const licensesPath = resolve(repoRoot, "docs", "LICENSES.md");

const COPYLEFT = /\b(GPL|LGPL|AGPL|SSPL|EUPL)\b/i;

// ---------- parse the inventory ----------

function parseInventory(path) {
    const map = new Map(); // name -> license
    const text = readFileSync(path, "utf8");
    for (const line of text.split(/\r?\n/)) {
        const m = line.match(/^\|(.+)\|$/);
        if (!m) continue;
        const cells = m[1].split("|").map((c) => c.trim());
        // Table layouts: | pkg | eco | version | license | source |
        //            and | pkg | eco | license | source |
        if (cells.length < 4) continue;
        const name = cells[0].replace(/`/g, "");
        if (!name || /^[-: ]+$/.test(name) || name === "Пакет") continue;
        const license = cells.length >= 5 ? cells[3] : cells[2];
        if (!license || /^[-: ]+$/.test(license) || license === "Лицензия") continue;
        map.set(name.toLowerCase(), license);
    }
    return map;
}

// ---------- collect direct deps ----------

function pythonDeps(path) {
    return readFileSync(path, "utf8")
        .split(/\r?\n/)
        .map((l) => l.trim())
        .filter((l) => l && !l.startsWith("#"))
        .map((l) => l.split(/[=<>!~\[]/)[0].trim().toLowerCase());
}

function goDeps(path) {
    const deps = [];
    let inRequire = false;
    for (const line of readFileSync(path, "utf8").split(/\r?\n/)) {
        const t = line.trim();
        if (/^require\s*\($/.test(t)) { inRequire = true; continue; }
        if (t === ")") { inRequire = false; continue; }
        const single = t.match(/^require\s+(\S+)\s+v[\d.]/);
        if (single) { deps.push(single[1].toLowerCase()); continue; }
        if (!inRequire || !t || t.startsWith("//") || t.includes("// indirect")) continue;
        const m = t.match(/^(\S+)\s+v[\d.]/);
        if (m && m[1].includes(".")) deps.push(m[1].toLowerCase());
    }
    return deps;
}

function npmDeps(path) {
    const pkg = JSON.parse(readFileSync(path, "utf8"));
    return [
        ...Object.keys(pkg.dependencies ?? {}),
        ...Object.keys(pkg.devDependencies ?? {}),
    ].map((n) => n.toLowerCase());
}

const depSources = [
    ["nlp-service requirements.txt", pythonDeps(resolve(repoRoot, "nlp-service", "requirements.txt"))],
    ["backend go.mod", goDeps(resolve(repoRoot, "backend", "go.mod"))],
    ["graph-service go.mod", goDeps(resolve(repoRoot, "services", "graph-service", "go.mod"))],
    ["frontend package.json", npmDeps(resolve(repoRoot, "frontend", "package.json"))],
];

const inventory = parseInventory(licensesPath);
if (inventory.size === 0) {
    console.error(`FAIL: no license entries parsed from ${licensesPath}`);
    process.exit(1);
}

const missing = [];
const copyleft = [];
for (const [label, deps] of depSources) {
    for (const dep of deps) {
        const license = inventory.get(dep);
        if (license === undefined) {
            missing.push(`${dep} (${label})`);
        } else if (COPYLEFT.test(license)) {
            copyleft.push(`${dep} (${label}) -> ${license}`);
        }
    }
}

for (const d of missing) console.error(`MISSING in LICENSES.md: ${d}`);
for (const d of copyleft) console.error(`COPYLEFT license: ${d}`);

if (missing.length || copyleft.length) {
    console.error(
        `\nFAIL: ${missing.length} missing, ${copyleft.length} copyleft — update docs/LICENSES.md`,
    );
    process.exit(1);
}

let total = 0;
for (const [, deps] of depSources) total += deps.length;
console.log(`License check OK: ${total} direct dependencies, ${inventory.size} inventory entries, no copyleft.`);
