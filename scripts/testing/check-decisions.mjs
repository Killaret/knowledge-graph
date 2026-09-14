// Owner-decision index guard.
//
// Enforces four rules from DECISIONS-1:
// 1. Every link in docs/DECISIONS.md resolves.
// 2. Every owner-decision marker in docs/tasks/*.md and in the
//    "## Решения владельца" section of docs/AI_HANDOFF.md has a matching row in
//    docs/DECISIONS.md (matched by task identifier and/or date).
// 3. Every decision marked as requiring code has at least one commit that names
//    its task identifier in the subject or body.
// 4. Terminal board rows in docs/AI_HANDOFF.md older than three days are archived.

import { readFileSync, readdirSync, existsSync, statSync } from "node:fs";
import { join, resolve, dirname, relative } from "node:path";
import { execSync } from "node:child_process";

const repoRoot = resolve(process.argv[2] ?? ".");

const DECISIONS_PATH = join(repoRoot, "docs", "DECISIONS.md");
const HANDOFF_PATH = join(repoRoot, "docs", "AI_HANDOFF.md");
const TASKS_DIR = join(repoRoot, "docs", "tasks");

const DATE_RE = /\d{4}-\d{2}-\d{2}/;
const ID_RE = /[A-Z][A-Z0-9]*(?:-[A-Z0-9]+)+/g; // e.g. PUB-2, P11-1, DEPENDABOT-79, AUTHOR-1
const TERMINAL_STATUSES = new Set(["принято", "отклонено", "отменено"]);

const errors = [];

function fail(message) {
    errors.push(message);
}

function readText(path) {
    return readFileSync(path, "utf8");
}

function resolveTarget(sourceFile, target) {
    const base = dirname(sourceFile);
    const decoded = decodeURIComponent(target);
    return resolve(base, decoded);
}

// ---------------------------------------------------------------------------
// 1. Parse docs/DECISIONS.md
// ---------------------------------------------------------------------------

const decisionsText = readText(DECISIONS_PATH);
const decisionRows = [];
let decisionsTableStarted = false;
let codeColumnIndex = -1;

function parseDecisionsTable() {
    const lines = decisionsText.split(/\r?\n/);
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        if (!decisionsTableStarted) {
            if (/^\|\s*№\s*\|/.test(line)) {
                const headers = line
                    .split("|")
                    .map((h) => h.trim().toLowerCase());
                codeColumnIndex = headers.indexOf("код");
                decisionsTableStarted = true;
            }
            continue;
        }
        if (/^\s*##?\s/.test(line) || line.startsWith("---")) {
            // End of the decision table.
            break;
        }
        if (!/^\|\s*\d+\s*\|/.test(line)) continue;

        const parts = line
            .replace(/^\|+/, "")
            .replace(/\|+$/, "")
            .split("|")
            .map((s) => s.trim());

        const number = parts[0];
        const date = parts[1];
        const decisionText = parts[2] ?? "";
        const reasonText = parts[3] ?? "";
        const refsText = parts[4] ?? "";
        const codeFlag = codeColumnIndex >= 0 ? (parts[codeColumnIndex] ?? "").trim() : "";

        const refs = [];
        let m;
        const linkPattern = /\[[^\]]*\]\(([^)\s]+?)(?:#[^)]*)?\)/g;
        while ((m = linkPattern.exec(refsText)) !== null) {
            refs.push({
                raw: m[1].trim(),
                resolved: resolveTarget(DECISIONS_PATH, m[1].trim()),
                isLink: true,
            });
        }
        // Also collect plain references like "ADR-018" or "TD-TLS".
        const plainPattern = /\b(ADR-\d+|TD-[A-Z]+|BACKLOG|—)\b/g;
        while ((m = plainPattern.exec(refsText)) !== null) {
            refs.push({ raw: m[1], resolved: null, isLink: false });
        }

        // Derive candidate task identifiers from references, decision text, and reason.
        const ids = new Set();
        for (const ref of refs) {
            const base = ref.raw.split("/").pop()?.replace(/\.md$/, "") ?? "";
            const fromFile = extractLeadingId(base);
            if (fromFile) ids.add(fromFile);
        }
        for (const text of [decisionText, reasonText]) {
            const fromText = extractAllIds(text);
            for (const id of fromText) ids.add(id);
        }

        decisionRows.push({
            number,
            date,
            decisionText,
            reasonText,
            refs,
            codeFlag,
            raw: line,
            ids: [...ids],
            used: false,
        });
    }
}

function extractLeadingId(name) {
    const match = name.match(/^([A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*)/);
    return match ? match[1] : null;
}

function extractAllIds(text) {
    const ids = [];
    let m;
    while ((m = ID_RE.exec(text)) !== null) {
        ids.push(m[0]);
    }
    return ids;
}

parseDecisionsTable();

// ---------------------------------------------------------------------------
// Rule 1: links in DECISIONS.md resolve
// ---------------------------------------------------------------------------

for (const row of decisionRows) {
    for (const ref of row.refs) {
        if (ref.isLink && !existsSync(ref.resolved)) {
            fail(`DECISIONS.md row ${row.number}: broken link "${ref.raw}"`);
        }
    }
}

// ---------------------------------------------------------------------------
// Rule 2: every owner-decision marker has a matching index row
// ---------------------------------------------------------------------------

function hasCommonId(marker, row) {
    if (marker.ids.length === 0 || row.ids.length === 0) return false;
    return marker.ids.some((id) =>
        row.ids.some((rid) => rid.toLowerCase() === id.toLowerCase()),
    );
}

function dateCompatible(marker, row) {
    if (!marker.date) return true;
    if (!row.date) return true;
    return marker.date === row.date;
}

function findMatchingRow(markers, sourceType) {
    // Match by task / reference identifier first, then by date.
    // A task marker can match a row that links to the same file even when the
    // marker's prose does not repeat the task id.
    for (const marker of markers) {
        let matched = false;

        // 1. Identifier match (strong).
        for (const row of decisionRows) {
            if (row.used) continue;
            if (hasCommonId(marker, row)) {
                row.used = true;
                matched = true;
                break;
            }
        }

        // 2. For task files, match by the source file name against row ids or
        //    resolved row references.
        if (!matched && sourceType === "task" && marker.taskFileIdentifier) {
            for (const row of decisionRows) {
                if (row.used) continue;
                if (
                    row.ids.some(
                        (id) => id.toLowerCase() === marker.taskFileIdentifier.toLowerCase(),
                    )
                ) {
                    row.used = true;
                    matched = true;
                    break;
                }
                if (
                    marker.filePath &&
                    row.refs.some(
                        (ref) => ref.isLink && ref.resolved === marker.filePath,
                    )
                ) {
                    row.used = true;
                    matched = true;
                    break;
                }
            }
        }

        // 3. Date-only fallback, but only when no ids and no task file are
        //    available and the candidate row has no strong identifier either
        //    (or the marker text mentions one of the row refs).
        if (
            !matched &&
            marker.date &&
            marker.ids.length === 0 &&
            (sourceType !== "task" || !marker.taskFileIdentifier)
        ) {
            for (const row of decisionRows) {
                if (row.used) continue;
                if (row.date !== marker.date) continue;
                // Accept the date match if the marker body mentions one of the
                // row's identifiers or references.
                const markerText = (marker.line || "") + "\n" + (marker.body || "");
                const mentioned =
                    row.ids.some((id) =>
                        markerText.toLowerCase().includes(id.toLowerCase()),
                    ) ||
                    row.refs.some((ref) =>
                        markerText.toLowerCase().includes(ref.raw.toLowerCase()),
                    );
                if (mentioned || row.ids.length === 0) {
                    row.used = true;
                    matched = true;
                    break;
                }
            }
        }

        if (!matched) {
            fail(
                `${sourceType === "task" ? marker.file : "AI_HANDOFF.md"}: owner decision ` +
                    (marker.date ? `(${marker.date}) ` : "") +
                    `has no matching row in DECISIONS.md`,
            );
        }
    }
}

function extractDecisionMarkersFromFile(filePath, relPath) {
    const text = readText(filePath);
    const lines = text.split(/\r?\n/);
    const fileId = extractLeadingId(
        relPath.replace(/^docs\/tasks\//, "").replace(/\.md$/, ""),
    );
    const markers = [];

    for (const line of lines) {
        // Ignore the DECISIONS-1 spec file's examples and bullet examples.
        if (
            /вкраплениями в\s+\d+\s+файлов/.test(line) ||
            /Добавить в любую задачу/.test(line)
        ) {
            continue;
        }

        // Match markers that start the line (paragraphs, headings, bold lines, bullets).
        const markerMatch = line.match(
            /^(?:##\s+|\*\*|[-*]\s+\*\*)?\s*Решени[ея] владельца/i,
        );
        if (!markerMatch) continue;

        const dateMatch = line.match(DATE_RE);
        const date = dateMatch ? dateMatch[0] : null;

        const ids = extractAllIds(line);

        markers.push({
            file: relPath,
            filePath,
            taskFileIdentifier: fileId,
            line: line.trim(),
            date,
            ids,
            body: "",
        });
    }

    return markers;
}

function extractHandoffDecisionBlocks() {
    const text = readText(HANDOFF_PATH);
    const lines = text.split(/\r?\n/);
    const blocks = [];
    let inOwnerSection = false;
    let currentBlock = null;

    for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        if (/^##\s+Решения\s+владельца\b/i.test(line)) {
            inOwnerSection = true;
            continue;
        }
        if (/^##\s+/.test(line) && inOwnerSection) {
            // End of owner decisions section.
            if (currentBlock) blocks.push(currentBlock);
            break;
        }
        if (!inOwnerSection) continue;

        const headingMatch = line.match(
            /^(\*\*?)Решени[ея] владельца(.*)$/i,
        );
        if (headingMatch) {
            if (currentBlock) blocks.push(currentBlock);
            currentBlock = {
                heading: line,
                lines: [line],
            };
            continue;
        }
        if (currentBlock) {
            currentBlock.lines.push(line);
        }
    }
    if (currentBlock) blocks.push(currentBlock);

    const markers = [];
    for (const block of blocks) {
        const heading = block.heading;
        const dateMatch = heading.match(DATE_RE);
        const date = dateMatch ? dateMatch[0] : null;
        const body = block.lines.join("\n");
        const ids = extractAllIds(body);

        // Require a date or a known identifier; otherwise it's not an indexed decision.
        if (!date && ids.length === 0) continue;

        markers.push({
            file: "docs/AI_HANDOFF.md",
            line: heading.trim(),
            body,
            date,
            ids,
        });
    }

    return markers;
}

// Reset used flags before matching.
for (const row of decisionRows) row.used = false;

// Task files.
const taskFiles = readdirSync(TASKS_DIR)
    .filter((name) => name.endsWith(".md"))
    .map((name) => join(TASKS_DIR, name));

const taskMarkers = [];
for (const filePath of taskFiles) {
    const relPath = relative(repoRoot, filePath).replace(/\\/g, "/");
    if (relPath === "docs/tasks/DECISIONS-1-decision-index-and-guard.md") continue;
    const markers = extractDecisionMarkersFromFile(filePath, relPath);
    taskMarkers.push(...markers);
}
findMatchingRow(taskMarkers, "task");

// AI_HANDOFF.md "Решения владельца" blocks.
const handoffMarkers = extractHandoffDecisionBlocks();
for (const row of decisionRows) row.used = false;
findMatchingRow(handoffMarkers, "handoff");

// ---------------------------------------------------------------------------
// Rule 3: decisions marked as requiring code have a commit that names the id
// ---------------------------------------------------------------------------

let allCommitsText = "";
try {
    allCommitsText = execSync(
        'git log --all --format="%H%x00%s%n%b%x00"',
        { cwd: repoRoot, encoding: "utf8", maxBuffer: 50 * 1024 * 1024 },
    );
} catch (err) {
    fail(`Could not read git history: ${err.message}`);
}

const commits = allCommitsText
    .split("\0")
    .filter((s) => s.trim())
    .map((s) => s.replace(/^\n+/, ""));

// A no-code marker in the row text: (кода не требует), (no code required), etc.
const NOCODE_RE = /\(\s*(?:кода?\s+не\s+требует(?:ся)?|no\s+code(?:\s+required)?)\s*\)/i;

for (const row of decisionRows) {
    if (NOCODE_RE.test(row.raw)) {
        continue; // explicitly marked as no-code
    }
    if (row.ids.length === 0) {
        fail(
            `DECISIONS.md row ${row.number}: cannot determine task identifier to look for commits`,
        );
        continue;
    }
    let found = false;
    for (const id of row.ids) {
        if (commits.some((commit) => commit.toLowerCase().includes(id.toLowerCase()))) {
            found = true;
            break;
        }
    }
    if (!found) {
        fail(
            `DECISIONS.md row ${row.number} (${row.ids.join(", ")}): no commit names the task identifier`,
        );
    }
}

// ---------------------------------------------------------------------------
// Rule 4: terminal board rows older than three days
// ---------------------------------------------------------------------------

const today = new Date();
const threshold = new Date(today);
threshold.setDate(threshold.getDate() - 3);
threshold.setHours(0, 0, 0, 0);

function parseBoardRows() {
    const text = readText(HANDOFF_PATH);
    const rows = [];
    let inArchive = false;
    for (const line of text.split(/\r?\n/)) {
        if (/^\s*<!--\s*archive\s*-->/i.test(line) || /^##\s+Архив/i.test(line)) {
            inArchive = true;
            break;
        }
        if (inArchive) continue;
        if (!/^\|/.test(line)) continue;
        if (/^\|[-\s|]+\|/.test(line)) continue; // separator
        const parts = line
            .replace(/^\|+/, "")
            .replace(/\|+$/, "")
            .split("|")
            .map((s) => s.trim());
        if (parts.length < 4) continue;
        const date = parts[parts.length - 1];
        const status = parts[parts.length - 2];
        if (!DATE_RE.test(date)) continue;

        const normalizedStatus = status
            .toLowerCase()
            .replace(/\*\*/g, "")
            .replace(/[\s\u2014\-–].*/, "");

        if (TERMINAL_STATUSES.has(normalizedStatus)) {
            rows.push({ line: line.trim(), date });
        }
    }
    return rows;
}

const boardRows = parseBoardRows();
for (const { line, date } of boardRows) {
    const d = new Date(date);
    if (d < threshold) {
        fail(`AI_HANDOFF.md board row is stale (older than 3 days): ${line}`);
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

if (errors.length) {
    console.error("Decision index guard failed:");
    for (const err of errors) {
        console.error(`  ${err}`);
    }
    process.exit(1);
}

console.log("Decisions OK: index, links, markers, code evidence, and board retention are consistent.");
