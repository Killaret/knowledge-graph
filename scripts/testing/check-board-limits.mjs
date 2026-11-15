// Board limits guard for BOARD-2.
//
// The board answers one question — "what is being done right now and whose
// turn it is". It cannot answer it when an agent's section holds dozens of
// rows, so active work is limited and everything else lives in the backlog.
//
// THE LIMITS ARE OWNER DECISIONS, NOT TUNABLES. BOARD-2 forbids changing them
// to make a run pass; only the owner may change them, and only on the record.
// 2026-09-21: "пять задач в работе, остальные должны быть в бэклоге",
// "на Devin — три". See decision 48 in docs/DECISIONS.md and the spec at
// docs/tasks/BOARD-2-board-size-and-backlog-policy.md. An agent moving these
// numbers on its own authority is loosening a guard on itself - do not.
//
// Rules (spec table):
//   1. The first bold word of a row's status must come from the status
//      dictionary, in agent sections, "На человеке", and the backlog.
//   2. `в работе` + `отклонено` per agent section <= 3.
//   3. `на ревью` across the whole board <= 5.
//   4. `бэклог` and `решает владелец` statuses are forbidden inside agent
//      sections - they belong in "Бэклог" / "На человеке".
//   5. A backlog row is at most 500 bytes.
//   6. "На человеке" holding more than 7 rows is a warning, not an error.
//   7. Reply retention (BOARD-3, owner decision 68): a reply in
//      "Обмен репликами" is at most 600 characters and at most three days
//      old by the date in its heading. Tests pin "today" via --today=YYYY-MM-DD.
//
// Terminal-row retention is not duplicated here; check-decisions.mjs owns it.
// The archive itself is a directory - docs/archive/board/ - not a section.

import { readFileSync } from "node:fs";
import { resolve, join } from "node:path";

const args = process.argv.slice(2);
let repoRoot = ".";
let boardPath = null;
let today = null;
for (const arg of args) {
    if (arg.startsWith("--board=")) {
        boardPath = resolve(arg.slice("--board=".length));
    } else if (arg.startsWith("--today=")) {
        today = arg.slice("--today=".length);
    } else {
        repoRoot = arg;
    }
}
repoRoot = resolve(repoRoot);
if (!boardPath) boardPath = join(repoRoot, "docs", "AI_HANDOFF.md");

const AGENT_SECTIONS = ["На Devin", "На Claude Code"];
const OWNER_SECTION = "На человеке";
const BACKLOG_SECTION = "Бэклог";
const REPLIES_SECTION = "Обмен репликами";
const UNTRACKED_SECTIONS = ["Обмен репликами", "Решения владельца"];

const STATUS_WORDS = [
    "в работе",
    "на ревью",
    "отклонено",
    "принято",
    "отменено",
    "бэклог",
    "решает владелец",
];
const BACKLOG_ONLY_STATUSES = ["бэклог", "решает владелец"];

const MAX_ACTIVE_PER_AGENT = 3;
const MAX_ON_REVIEW_TOTAL = 5;
const MAX_BACKLOG_ROW_BYTES = 500;
const OWNER_SECTION_SOFT_LIMIT = 7;
const MAX_REPLY_CHARS = 600;
const MAX_REPLY_AGE_DAYS = 3;
const REPLY_DATE_RE = /\d{4}-\d{2}-\d{2}/;

const errors = [];
const warnings = [];

function firstStatusWord(statusCell) {
    const plain = statusCell
        .replace(/\*\*/g, "")
        .replace(/[`*_]/g, "")
        .trim()
        .toLowerCase();
    for (const word of STATUS_WORDS) {
        if (plain === word || plain.startsWith(word + " ") || plain.startsWith(word + " ")
            || plain.startsWith(word + "(")) {
            return word;
        }
    }
    return plain.split(/\s|—|-/)[0] ?? "";
}

// ---------------------------------------------------------------------------
// Parse the board into sections
// ---------------------------------------------------------------------------

const text = readFileSync(boardPath, "utf8");
const lines = text.split(/\r?\n/);

const sections = {}; // name -> [{lineNo, line, parts}]
let currentSection = null;
for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const heading = line.match(/^##\s+(.+?)\s*$/);
    if (heading) {
        currentSection = heading[1];
        continue;
    }
    if (currentSection === null) continue;
    if (UNTRACKED_SECTIONS.includes(currentSection)) continue;
    if (!line.startsWith("|")) continue;
    if (/^\|[-\s|]+\|/.test(line)) continue; // separator
    const parts = line
        .replace(/^\|+/, "")
        .replace(/\|+$/, "")
        .split("|")
        .map((s) => s.trim());
    if (parts.length < 4) continue;
    if (parts.some((p) => /^статус$/i.test(p))) continue; // header row
    (sections[currentSection] ??= []).push({ lineNo: i + 1, line, parts });
}

function statusWord(row) {
    return firstStatusWord(row.parts[row.parts.length - 2] ?? "");
}

function shortRow(row) {
    const firstCell = row.parts[0].replace(/\*\*/g, "");
    return `line ${row.lineNo} (${firstCell.slice(0, 60)})`;
}

// ---------------------------------------------------------------------------
// Rule 1: status dictionary
// ---------------------------------------------------------------------------

for (const [section, rows] of Object.entries(sections)) {
    for (const row of rows) {
        const word = statusWord(row);
        if (!STATUS_WORDS.includes(word)) {
            errors.push(
                `${section}: ${shortRow(row)}: status outside the dictionary ` +
                    `(got "${word}", allowed: ${STATUS_WORDS.join(", ")})`,
            );
        }
    }
}

// ---------------------------------------------------------------------------
// Rule 2: active work per agent <= MAX_ACTIVE_PER_AGENT
// ---------------------------------------------------------------------------

for (const section of AGENT_SECTIONS) {
    const rows = sections[section] ?? [];
    const active = rows.filter((r) => ["в работе", "отклонено"].includes(statusWord(r)));
    if (active.length > MAX_ACTIVE_PER_AGENT) {
        errors.push(
            `${section}: ${active.length} rows in "в работе"/"отклонено", ` +
                `limit is ${MAX_ACTIVE_PER_AGENT}: ` +
                active.map(shortRow).join("; "),
        );
    }
}

// ---------------------------------------------------------------------------
// Rule 3: rows under review across the board <= MAX_ON_REVIEW_TOTAL
// ---------------------------------------------------------------------------

const reviewRows = [];
for (const [section, rows] of Object.entries(sections)) {
    for (const row of rows) {
        if (statusWord(row) === "на ревью") reviewRows.push({ section, row });
    }
}
if (reviewRows.length > MAX_ON_REVIEW_TOTAL) {
    errors.push(
        `${reviewRows.length} rows in "на ревью" across the board, ` +
            `limit is ${MAX_ON_REVIEW_TOTAL}: ` +
            reviewRows.map((r) => `${r.section} ${shortRow(r.row)}`).join("; "),
    );
}

// ---------------------------------------------------------------------------
// Rule 4: backlog-only statuses forbidden inside agent sections
// ---------------------------------------------------------------------------

for (const section of AGENT_SECTIONS) {
    for (const row of sections[section] ?? []) {
        const word = statusWord(row);
        if (BACKLOG_ONLY_STATUSES.includes(word)) {
            errors.push(
                `${section}: ${shortRow(row)}: status "${word}" does not belong ` +
                    `in an agent section - move the row to ` +
                    `"${word === "бэклог" ? "Бэклог" : "На человеке"}"`,
            );
        }
    }
}

// ---------------------------------------------------------------------------
// Rule 5: backlog row size <= MAX_BACKLOG_ROW_BYTES
// ---------------------------------------------------------------------------

for (const row of sections[BACKLOG_SECTION] ?? []) {
    const bytes = Buffer.byteLength(row.line, "utf8");
    if (bytes > MAX_BACKLOG_ROW_BYTES) {
        errors.push(
            `${BACKLOG_SECTION}: ${shortRow(row)}: row is ${bytes} bytes, ` +
                `limit is ${MAX_BACKLOG_ROW_BYTES} - details belong in the spec file`,
        );
    }
}

// ---------------------------------------------------------------------------
// Rule 6: owner section soft limit (warning only)
// ---------------------------------------------------------------------------

const ownerRows = sections[OWNER_SECTION] ?? [];
if (ownerRows.length > OWNER_SECTION_SOFT_LIMIT) {
    warnings.push(
        `${OWNER_SECTION}: ${ownerRows.length} rows, soft limit is ` +
            `${OWNER_SECTION_SOFT_LIMIT} - the owner should see the overload`,
    );
}

// ---------------------------------------------------------------------------
// Rule 7: reply watcher - replies are short pointers, at most three days old
// ---------------------------------------------------------------------------

const repliesThreshold = today ? new Date(today) : new Date();
repliesThreshold.setHours(0, 0, 0, 0);
repliesThreshold.setDate(repliesThreshold.getDate() - MAX_REPLY_AGE_DAYS);

function extractReplies() {
    const start = lines.findIndex((l) => /^##\s+Обмен репликами/.test(l));
    if (start === -1) return [];
    const replies = [];
    let paragraph = [];
    let startLine = 0;
    const flush = () => {
        const text = paragraph.join("\n").trim();
        paragraph = [];
        if (!text.startsWith("**")) return;
        const heading = text.split("**", 2)[1] ?? "";
        const m = heading.match(REPLY_DATE_RE) ?? text.match(REPLY_DATE_RE);
        if (!m) return; // section prose, not a reply
        replies.push({ lineNo: startLine, text, date: m[0] });
    };
    for (let i = start + 1; i < lines.length; i++) {
        const line = lines[i];
        if (/^##\s+/.test(line)) break;
        if (line.trim() === "") {
            flush();
        } else {
            if (paragraph.length === 0) startLine = i + 1;
            paragraph.push(line);
        }
    }
    flush();
    return replies;
}

for (const reply of extractReplies()) {
    const label = `line ${reply.lineNo} (${reply.date})`;
    if ([...reply.text].length > MAX_REPLY_CHARS) {
        errors.push(
            `${REPLIES_SECTION}: ${label}: reply is ${[...reply.text].length} ` +
                `characters, limit is ${MAX_REPLY_CHARS} - a reply is a pointer, ` +
                `not a retelling`,
        );
    }
    if (new Date(reply.date) < repliesThreshold) {
        errors.push(
            `${REPLIES_SECTION}: ${label}: reply is older than ` +
                `${MAX_REPLY_AGE_DAYS} days - its trace stays in AI_LOG.md and git`,
        );
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

if (errors.length) {
    console.error("Board limits guard failed:");
    for (const err of errors) console.error(`  ${err}`);
    process.exit(1);
}

for (const warn of warnings) console.warn(`Board limits warning: ${warn}`);

const devinActive = (sections["На Devin"] ?? []).filter((r) =>
    ["в работе", "отклонено"].includes(statusWord(r)),
).length;
const backlogCount = (sections[BACKLOG_SECTION] ?? []).length;
console.log(
    `Board limits OK: Devin ${devinActive}/${MAX_ACTIVE_PER_AGENT} в работе, ` +
        `на ревью ${reviewRows.length}/${MAX_ON_REVIEW_TOTAL}, ` +
        `бэклог ${backlogCount} строк.`,
);
