// Drift guard for docs/tasks/README.md.
//
// Detects:
// - task files without a known board/journal identifier;
// - a generated index that differs from the committed one;
// - index rows pointing to missing files;
// - board/journal links to task files that do not exist.

import { readFileSync, readdirSync, existsSync } from "node:fs";
import { join, resolve, relative } from "node:path";
import { generateTaskIndex, renderReadme } from "./generate-tasks-index.mjs";

const repoRoot = resolve(process.argv[2] ?? ".");
const tasksDir = join(repoRoot, "docs", "tasks");
const readmePath = join(tasksDir, "README.md");
const handoffPath = join(repoRoot, "docs", "AI_HANDOFF.md");
const logPath = join(repoRoot, "docs", "AI_LOG.md");

const fileLinkRe = /\[([^\]]*)\]\(tasks\/([^)\s]+?)(?:#[^)]*)?\)/g;

function checkIndexDrift(expected) {
    if (!existsSync(readmePath)) {
        console.error("docs/tasks/README.md does not exist.");
        return ["docs/tasks/README.md missing"];
    }
    const actual = readFileSync(readmePath, "utf8");
    if (actual === expected) return [];

    const expLines = expected.split(/\r?\n/);
    const actLines = actual.split(/\r?\n/);
    for (let i = 0; i < Math.max(expLines.length, actLines.length); i++) {
        if (expLines[i] !== actLines[i]) {
            return [
                `docs/tasks/README.md drift at line ${i + 1}: expected "${expLines[i] ?? "(missing)"}" but found "${actLines[i] ?? "(missing)"}". Run scripts/testing/generate-tasks-index.mjs .`,
            ];
        }
    }
    return [];
}

function checkMissingFiles(entries) {
    const errors = [];
    const files = new Set(readdirSync(tasksDir).filter((n) => n.endsWith(".md")));
    for (const e of entries) {
        if (!files.has(e.file)) {
            errors.push(`Index row references missing file: ${e.file}`);
        }
    }
    for (const f of files) {
        if (f === "README.md") continue;
        if (!entries.some((e) => e.file === f)) {
            errors.push(`Task file missing from index: ${f}`);
        }
    }
    return errors;
}

function checkBoardLinks() {
    const errors = [];
    for (const source of [handoffPath, logPath]) {
        const text = readFileSync(source, "utf8");
        let m;
        const re = new RegExp(fileLinkRe.source, "g");
        while ((m = re.exec(text)) !== null) {
            const filename = m[2];
            if (!filename.endsWith(".md")) continue;
            const full = join(tasksDir, filename);
            if (!existsSync(full)) {
                errors.push(
                    `${relative(repoRoot, source)} links to missing task file: ${filename}`,
                );
            }
        }
    }
    return errors;
}

async function main() {
    const { entries, unknown } = await generateTaskIndex();
    const expected = renderReadme(entries);

    const errors = [];
    if (unknown.length) {
        for (const u of [...new Set(unknown)].sort()) {
            errors.push(`Unknown identifier: ${u}`);
        }
    }

    errors.push(...checkMissingFiles(entries));
    errors.push(...checkIndexDrift(expected));
    errors.push(...checkBoardLinks());

    if (errors.length) {
        console.error("Task index guard failed:");
        for (const e of errors) console.error(`  - ${e}`);
        process.exit(1);
    }

    console.log(`Task index OK: ${entries.length} entries, no drift, no broken board links.`);
}

main();
