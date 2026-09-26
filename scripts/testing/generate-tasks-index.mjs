// Generated task index for docs/tasks/README.md.
//
// Scans docs/tasks/*.md, docs/AI_HANDOFF.md (board), docs/archive/board/*.md
// (closed rows, BOARD-3) and docs/AI_LOG.md (journal) to build a
// machine-readable, human-usable index. Run after adding, removing or
// renaming task files.

import { readFileSync, readdirSync, writeFileSync, existsSync } from "node:fs";
import { join, resolve, relative, basename } from "node:path";
import { exec, execSync } from "node:child_process";
import { promisify } from "node:util";

const execAsync = promisify(exec);

const repoRoot = resolve(process.argv[2] ?? ".");

const tasksDir = join(repoRoot, "docs", "tasks");
const handoffPath = join(repoRoot, "docs", "AI_HANDOFF.md");
const boardArchiveDir = join(repoRoot, "docs", "archive", "board");
const logPath = join(repoRoot, "docs", "AI_LOG.md");
const readmePath = join(repoRoot, "docs", "tasks", "README.md");

// Project task identifier: starts with a letter, may contain numbers,
// hyphen-separated segments. A trailing lowercase letter (AUD-7a) or a
// number-only final segment (BATCH-DDD-1) are both valid.
const ID_RE = /[A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*(?:[a-z])?(?![A-Za-z0-9])/g;

// A more lenient pattern that also captures lowercase suffixes glued to a
// number, used when scanning known text rather than filenames.
const ID_RE_WITH_LC = /[A-Z][A-Z0-9]*(?:-[A-Z0-9]+[a-z]?)*(?![A-Za-z0-9])/g;

// Board rows link tasks as `tasks/x.md` (relative to docs/); archived rows in
// docs/archive/board/ link them as `../../tasks/x.md`. Both resolve here.
const fileLinkRe = /\[([^\]]*)\]\((?:\.\.\/)*tasks\/([^)\s]+?)(?:#[^)]*)?\)/g;

const isTaskFile = (name) => name.endsWith(".md");

function readText(path) {
    return readFileSync(path, "utf8");
}

async function gitLastCommitDateAsync(file) {
    const run = async (args) => {
        try {
            const { stdout } = await execAsync('git log -1 ' + args + ' --format=%cs -- "' + file + '"', {
                cwd: repoRoot,
                encoding: "utf8",
            });
            return stdout.trim();
        } catch {
            return "";
        }
    };
    let out = await run("");
    if (!out) {
        // A freshly renamed file may not have its own path in git history yet,
        // so follow renames to keep the date until the change is committed.
        out = await run("--follow");
    }
    return out || "—";
}

function extractHeading(path) {
    const text = readText(path);
    const m = text.match(/^#\s*(.+)$/m);
    return m ? m[1].trim() : "—";
}

function escapeRegExp(s) {
    return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

// Extract the longest known task identifier that is a prefix of the name.
function chooseIdForFilename(name, knownIds) {
    // Strip the .md extension for matching.
    const base = name.replace(/\.md$/, "");
    let best = null;
    for (const id of knownIds) {
        // The id must be a prefix, followed by '-' or end-of-string.
        const re = new RegExp("^" + escapeRegExp(id) + "(?:-(?=[A-Za-z0-9])|$)");
        if (re.test(base)) {
            if (!best || id.length > best.length) best = id;
        }
    }
    return best;
}

function parseTableLines(text) {
    return text
        .split(/\r?\n/)
        .filter((line) => /^\|/.test(line) && !/^\|[-\s|]+\|/.test(line))
        .map((line) =>
            line
                .replace(/^\|+/, "")
                .replace(/\|+$/, "")
                .split("|")
                .map((s) => s.trim()),
        );
}

function extractIdsFromCell(cell) {
    const out = new Set();
    if (!cell) return out;
    // Remove markdown bold/italic markers and code spans.
    const cleaned = cell.replace(/\*\*/g, "").replace(/\*/g, "").replace(/`/g, "");
    const idPattern = /^[A-Z][A-Z0-9]*(?:-[A-Z0-9]+[a-z]?)*(?:-[A-Z0-9]+)*(?:[a-z])?$/;
    // Split by common separators used to list multiple IDs and also by whitespace.
    for (const token of cleaned.split(/[\/\,\;\&\s]/)) {
        const trimmed = token.replace(/[.:;!]+$/, "").trim();
        if (!trimmed) continue;
        // A task id always carries a hyphen (PUB-2, BOARD-1); a bare all-caps
        // word like "API" or "NLP" is prose, not an identifier - letting it
        // through once made BACKUP-1 inherit NOTE-QUALITY-1's board status.
        if (!trimmed.includes("-")) continue;
        const m = trimmed.match(idPattern);
        if (m) out.add(m[0]);
    }
    return out;
}

function collectBoardRows(text, known, boardStatus, ids) {
    for (const parts of parseTableLines(text)) {
        if (parts.length < 4) continue;
        // First column carries the task identifier(s).
        const rowIds = extractIdsFromCell(parts[0]);
        for (const id of rowIds) ids.add(id);

        const status = parts[parts.length - 2] || "";
        const rowText = parts.join(" ");

        let m;
        const re = new RegExp(fileLinkRe.source, "g");
        while ((m = re.exec(rowText)) !== null) {
            const filename = m[2];
            for (const id of rowIds) {
                if (!known.has(filename)) known.set(filename, new Set());
                known.get(filename).add(id);
                if (!boardStatus.has(id)) boardStatus.set(id, status);
            }
        }
    }
}

function parseHandoff() {
    const known = new Map(); // filename -> Set(ids)
    const boardStatus = new Map(); // id -> status text
    const ids = new Set();

    collectBoardRows(readText(handoffPath), known, boardStatus, ids);

    // Closed rows live in the board archive (BOARD-3); their statuses count
    // too, but live board rows win for ids present in both places.
    if (existsSync(boardArchiveDir)) {
        const months = readdirSync(boardArchiveDir)
            .filter((name) => /^\d{4}-\d{2}\.md$/.test(name))
            .sort();
        for (const name of months) {
            collectBoardRows(
                readText(join(boardArchiveDir, name)),
                known,
                boardStatus,
                ids,
            );
        }
    }

    return { known, boardStatus, ids };
}

function parseLog() {
    const known = new Map();
    const ids = new Set();

    for (const parts of parseTableLines(readText(logPath))) {
        if (parts.length < 5) continue;
        // Task and status columns may carry identifiers and links.
        const rowIds = new Set([
            ...extractIdsFromCell(parts[2]),
            ...extractIdsFromCell(parts[3]),
        ]);
        for (const id of rowIds) ids.add(id);

        const rowText = parts.join(" ");
        let m;
        const re = new RegExp(fileLinkRe.source, "g");
        while ((m = re.exec(rowText)) !== null) {
            const filename = m[2];
            for (const id of rowIds) {
                if (!known.has(filename)) known.set(filename, new Set());
                known.get(filename).add(id);
            }
        }
    }
    return { known, ids };
}

function buildKnownIdSet() {
    const handoff = parseHandoff();
    const log = parseLog();
    const all = new Set([...handoff.ids, ...log.ids]);

    // Merge known maps: a file may be referenced from both board and journal.
    const known = new Map();
    for (const [filename, set] of handoff.known) {
        if (!known.has(filename)) known.set(filename, new Set());
        for (const id of set) known.get(filename).add(id);
    }
    for (const [filename, set] of log.known) {
        if (!known.has(filename)) known.set(filename, new Set());
        for (const id of set) known.get(filename).add(id);
    }

    return { all, known, boardStatus: handoff.boardStatus };
}

export async function generateTaskIndex() {
    const { all: knownIds, known, boardStatus } = buildKnownIdSet();

    const files = readdirSync(tasksDir)
        .filter(isTaskFile)
        .sort((a, b) => a.localeCompare(b))
        .filter((f) => f !== "README.md");

    const headingMap = new Map();
    const relPaths = new Map();
    for (const file of files) {
        const fullPath = join(tasksDir, file);
        const relPath = relative(repoRoot, fullPath).replace(/\\/g, "/");
        relPaths.set(file, relPath);
        headingMap.set(file, extractHeading(fullPath));
    }

    // Resolve all last-commit dates in parallel.
    const datePromises = files.map((file) => gitLastCommitDateAsync(relPaths.get(file)));
    const dates = await Promise.all(datePromises);

    const entries = [];
    const unknown = [];

    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        const heading = headingMap.get(file);
        const date = dates[i];

        const ids = known.get(file) || new Set();
        let id = chooseIdForFilename(file, new Set([...ids, ...knownIds]));

        if (!id) {
            // Fallback: the leading token of the filename itself.
            const m = file.match(/^([A-Z][A-Z0-9]*(?:-[A-Z0-9]+[a-z]?)*(?:-[A-Z0-9]+)*(?:[a-z])?)\b/);
            id = m ? m[1] : null;
        }

        let status = "—";
        if (id) {
            // Prefer the status of the id that matches the file prefix; otherwise
            // any known status.
            let statusId = null;
            for (const candidate of ids) {
                if (boardStatus.has(candidate)) {
                    statusId = candidate;
                    break;
                }
            }
            if (!statusId && ids.size) statusId = [...ids][0];
            if (statusId && boardStatus.has(statusId)) {
                status = boardStatus.get(statusId);
            }
        }

        entries.push({ file, heading, id: id ?? "—", status, date });

        if (!id || (!knownIds.has(id) && !ids.has(id))) {
            if (!/^TASKS-INDEX-1/.test(file)) unknown.push(file);
        }
    }

    return { entries, unknown };
}

function normalizeStatusLinks(status) {
    // README.md lives in docs/tasks/, so links to `tasks/X.md` (board) or
    // `../../tasks/X.md` (archive files) should both be `X.md`.
    return status.replace(
        /\[([^\]]*)\]\((?:\.\.\/)*tasks\/([^)\s]+?)(?:#[^)]*)?\)/g,
        "[$1]($2)",
    );
}

export function renderReadme(entries) {
    let out = "# Generated task index\n\n";
    out += "This file is generated by `scripts/testing/generate-tasks-index.mjs`.\n";
    out += "Do not edit it manually; run the generator after changing the task directory.\n\n";
    out += "| Identifier | Title | File | Board status | Last commit |\n";
    out += "|---|---|---|---|---|\n";

    for (const e of entries) {
        const fileLink = `[${e.file}](${e.file})`;
        const status = normalizeStatusLinks(e.status);
        out += `| ${e.id} | ${e.heading} | ${fileLink} | ${status} | ${e.date} |\n`;
    }
    return out;
}

async function main() {
    const { entries, unknown } = await generateTaskIndex();
    const out = renderReadme(entries);

    writeFileSync(readmePath, out, "utf8");
    console.log(`Generated ${entries.length} task index entries at ${relative(repoRoot, readmePath)}.`);

    if (unknown.length) {
        console.warn(
            "The following files have no identifier known to the board or journal:",
        );
        for (const u of [...new Set(unknown)].sort()) console.warn(`  ${u}`);
        console.warn("Assign an identifier and update docs/AI_HANDOFF.md or docs/AI_LOG.md.");
        process.exit(1);
    }
}

if (process.argv[1] && basename(process.argv[1]) === basename(import.meta.url)) {
    main();
}
