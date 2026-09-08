// Documentation hygiene: local markdown links must resolve, and `npm run`
// targets quoted as instructions must exist.
//
// Why this exists: COMMANDS.md spent months pointing at three scripts that had
// been deleted and one npm target that never existed, and nobody noticed
// because nothing checked. A one-time cleanup rots; this does not.
//
// Historical documents are exempt. They quote the state of the world on the day
// they were written — a journal entry naming a since-renamed file is correct,
// not broken — so rewriting them would erase history. The exempt set is listed
// below and deliberately narrow: everything else is expected to be true today.

import { readFileSync, readdirSync, statSync, existsSync } from "node:fs";
import { join, dirname, resolve, relative, sep } from "node:path";

const repoRoot = resolve(process.argv[2] ?? ".");

const SKIP_DIRS = new Set(["node_modules", ".git", ".svelte-kit", "dist", "build", "coverage"]);

// Documents that record the past rather than describe the present.
const HISTORICAL = [
    "docs/archive",
    "docs/3d-archive",
    "docs/tasks",
    "docs/AI_LOG.md",
    "docs/AI_HANDOFF.md",
    "docs/EXTERNAL_AUDIT_2026-09.md",
    "docs/AI_PROCESS_AUDIT.md",
    "docs/MANUAL_TEST_FEEDBACK.md",
    "CHANGELOG.md",
];

const isHistorical = (rel) =>
    HISTORICAL.some((h) => rel === h || rel.startsWith(h + "/"));

function walk(dir, out = []) {
    for (const entry of readdirSync(dir)) {
        if (SKIP_DIRS.has(entry)) continue;
        const full = join(dir, entry);
        if (statSync(full).isDirectory()) walk(full, out);
        else if (entry.endsWith(".md")) out.push(full);
    }
    return out;
}

// Links and commands need opposite treatment, which is why there are two passes.
//
// Links: code spans hold examples, not links — a `[title](url)` describing a
// return value is text. Blanking them avoids false positives, and a check that
// cries wolf gets ignored.
//
// Commands: documented commands live *inside* fenced blocks, so scanning the
// stripped text would leave this half of the check dead on arrival. Found by
// mutation — an npm target inside a bash fence went unnoticed.
function stripCodeForLinks(text) {
    return text
        .replace(/```[\s\S]*?```/g, "")
        .replace(/`[^`\n]*`/g, "");
}

const npmScripts = new Set();
for (const pkg of ["package.json", "frontend/package.json", "tests/package.json"]) {
    const full = join(repoRoot, pkg);
    if (!existsSync(full)) continue;
    const parsed = JSON.parse(readFileSync(full, "utf8"));
    for (const name of Object.keys(parsed.scripts ?? {})) npmScripts.add(name);
}

const linkPattern = /\[[^\]]*\]\(([^)\s]+?)(?:#[^)]*)?\)/g;
const npmPattern = /npm run ([a-zA-Z0-9:_-]+)/g;

const brokenLinks = [];
const missingScripts = [];

for (const file of walk(repoRoot)) {
    const rel = relative(repoRoot, file).split(sep).join("/");
    const raw = readFileSync(file, "utf8");
    const body = stripCodeForLinks(raw);

    for (const match of body.matchAll(linkPattern)) {
        const target = match[1].trim();
        if (/^(https?:|mailto:|tel:|#)/.test(target)) continue;
        if (!existsSync(resolve(dirname(file), decodeURIComponent(target)))) {
            brokenLinks.push(`${rel} -> ${target}`);
        }
    }

    if (isHistorical(rel)) continue;

    for (const match of raw.matchAll(npmPattern)) {
        // `npm run test:*` names a family of scripts, not one of them.
        const next = raw[match.index + match[0].length];
        if (next === "*" || match[1].endsWith(":")) continue;
        if (!npmScripts.has(match[1])) {
            missingScripts.push(`${rel} -> npm run ${match[1]}`);
        }
    }
}

if (brokenLinks.length || missingScripts.length) {
    console.error("Documentation drift detected.");
    if (brokenLinks.length) {
        console.error(`Broken local links (${brokenLinks.length}):`);
        for (const item of brokenLinks) console.error(`  ${item}`);
    }
    if (missingScripts.length) {
        console.error(`npm targets that do not exist (${missingScripts.length}):`);
        for (const item of missingScripts) console.error(`  ${item}`);
    }
    process.exit(1);
}

console.log("Docs OK: every local link resolves and every documented npm target exists.");
