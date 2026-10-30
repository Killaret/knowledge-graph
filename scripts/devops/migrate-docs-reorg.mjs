// DOC-REORG-1 migration tool — moves docs/ files into the new hierarchy and
// rewrites every link that points at a moved path.
//
// Two passes over every tracked text file:
//   1. Markdown links `[t](href)` and reference definitions `[t]: href`
//      (only in .md files and .windsurfrules): resolve href against the file's
//      OLD location, map the target through MOVE_MAP, emit the new relative
//      path computed from the file's NEW location. Depth changes and moved
//      targets are handled by the same mechanism.
//   2. Literal `docs/<path>` strings (all text files): catches references in
//      scripts, workflows, prompts and inline code that the link pass cannot
//      see. Runs after the link pass; already-rewritten links no longer match.
//
// Links that resolved before the move keep resolving after it; links that were
// already broken are reported but untouched. Run from the repo root:
//   node scripts/devops/migrate-docs-reorg.mjs --dry-run   # report only
//   node scripts/devops/migrate-docs-reorg.mjs             # rewrite + git mv

import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join, relative, resolve, sep } from "node:path";

const repoRoot = resolve(".");
const dryRun = process.argv.includes("--dry-run");

// ---------------------------------------------------------------- move map

const FILE_MOVES = {
    // agents/ — agent workflow documents that are not read every session
    "docs/AI_AGENT_SETUP.md": "docs/agents/AI_AGENT_SETUP.md",
    "docs/AI_PROCESS_AUDIT.md": "docs/agents/AI_PROCESS_AUDIT.md",
    "docs/MANUAL_TEST_FEEDBACK.md": "docs/agents/MANUAL_TEST_FEEDBACK.md",

    // architecture/
    "docs/ARCHITECTURE_SUMMARY.md": "docs/architecture/ARCHITECTURE_SUMMARY.md",
    "docs/ARCHITECTURE_EN.md": "docs/architecture/ARCHITECTURE_EN.md",
    "docs/ARCHITECTURE_PATTERNS.md": "docs/architecture/ARCHITECTURE_PATTERNS.md",
    "docs/FRONTEND_ARCHITECTURE_EN.md": "docs/architecture/FRONTEND_ARCHITECTURE_EN.md",
    "docs/SaaS_DATABASE_SCHEMA.md": "docs/architecture/SaaS_DATABASE_SCHEMA.md",
    "docs/GRAPH_SERVICE_AUTH.md": "docs/architecture/GRAPH_SERVICE_AUTH.md",
    "docs/GRAPH3D.md": "docs/architecture/GRAPH3D.md",
    "docs/RECOMMENDATION_ARCHITECTURE.md": "docs/architecture/RECOMMENDATION_ARCHITECTURE.md",
    "docs/RECOMMENDATION_TROUBLESHOOTING.md": "docs/architecture/RECOMMENDATION_TROUBLESHOOTING.md",

    // api/
    "docs/API_EN.md": "docs/api/API_EN.md",
    "docs/API_ERRORS_EN.md": "docs/api/API_ERRORS_EN.md",
    "docs/RECOMMENDATION_API.md": "docs/api/RECOMMENDATION_API.md",

    // operations/
    "docs/DOCKER.md": "docs/operations/DOCKER.md",
    "docs/DEPLOYMENT_EN.md": "docs/operations/DEPLOYMENT_EN.md",
    "docs/STACK_CONFIGURATION_COMPARISON.md": "docs/operations/STACK_CONFIGURATION_COMPARISON.md",
    "docs/CONFIGURATION_EN.md": "docs/operations/CONFIGURATION_EN.md",
    "docs/CONFIGURATION_RU.md": "docs/operations/CONFIGURATION_RU.md",
    "docs/BACKUP.md": "docs/operations/BACKUP.md",
    "docs/TESTING.md": "docs/operations/TESTING.md",
    "docs/TESTING_COMMANDS.md": "docs/operations/TESTING_COMMANDS.md",
    "docs/REGRESSION_TEST_PLAN.md": "docs/operations/REGRESSION_TEST_PLAN.md",
    "docs/ARGOS.md": "docs/operations/ARGOS.md",
    "docs/MANUAL_TEST_CHECKLIST_COCKPIT.md": "docs/operations/MANUAL_TEST_CHECKLIST_COCKPIT.md",
    "docs/MANUAL_TEST_CHECKLIST_MINIMAL.md": "docs/operations/MANUAL_TEST_CHECKLIST_MINIMAL.md",

    // product/
    "docs/LINK_TYPES.md": "docs/product/LINK_TYPES.md",
    "docs/LINK_TYPES_RU.md": "docs/product/LINK_TYPES_RU.md",
    "docs/LINKS_CHEATSHEET.md": "docs/product/LINKS_CHEATSHEET.md",
    "docs/GRAPH_LINKS_VISUALIZATION.md": "docs/product/GRAPH_LINKS_VISUALIZATION.md",
    "docs/CELESTIAL_BODY_SEMANTICS.md": "docs/product/CELESTIAL_BODY_SEMANTICS.md",
    "docs/ANOMALY_TYPES.md": "docs/product/ANOMALY_TYPES.md",
    "docs/BOOKMARKLET.md": "docs/product/BOOKMARKLET.md",
    "docs/bookmarklet.js": "docs/product/bookmarklet.js",
    "docs/OBSIDIAN_IMPORT_SPEC.md": "docs/product/OBSIDIAN_IMPORT_SPEC.md",
    "docs/UX_GUIDELINES_EN.md": "docs/product/UX_GUIDELINES_EN.md",
    "docs/IDEAS.md": "docs/product/IDEAS.md",
    "docs/BACKLOG.md": "docs/product/BACKLOG.md",
    "docs/FRONTEND_FEATURES.md": "docs/product/FRONTEND_FEATURES.md",
    "docs/UI_DUPLICATION_AND_NOTE_CREATION_ANALYSIS.md": "docs/product/UI_DUPLICATION_AND_NOTE_CREATION_ANALYSIS.md",
    "docs/NOTE_ERROR_CORRECTION_PLAN.md": "docs/product/NOTE_ERROR_CORRECTION_PLAN.md",

    // archive/ — stale plans, dated audits, superseded docs (content kept)
    "docs/ARCHITECTURE_ROADMAP.md": "docs/archive/ARCHITECTURE_ROADMAP.md",
    "docs/EXTERNAL_AUDIT_2026-09.md": "docs/archive/EXTERNAL_AUDIT_2026-09.md",
    "docs/AUTO_LINK_CREATION_PLAN.md": "docs/archive/AUTO_LINK_CREATION_PLAN.md",
    "docs/API_TEST_COVERAGE_PLAN.md": "docs/archive/API_TEST_COVERAGE_PLAN.md",
    "docs/UI_MODERNIZATION_ROADMAP.md": "docs/archive/UI_MODERNIZATION_ROADMAP.md",
    "docs/MASS_IMPORT_TEST_PLAN.md": "docs/archive/MASS_IMPORT_TEST_PLAN.md",
    "docs/TEST_PLAN_VALIDATION_AUTOMATION.md": "docs/archive/TEST_PLAN_VALIDATION_AUTOMATION.md",
    "docs/CRITICAL_FIXES.md": "docs/archive/CRITICAL_FIXES.md",
    "docs/MANUAL_TEST_CHECKLISTS_RU.md": "docs/archive/MANUAL_TEST_CHECKLISTS_RU.md",
    "docs/MANUAL_TEST_CHECKLIST_AI_AGENTS_3D_REFACTOR.md": "docs/archive/MANUAL_TEST_CHECKLIST_AI_AGENTS_3D_REFACTOR.md",
    "docs/YANDEX_DISK_BACKUP.md": "docs/archive/YANDEX_DISK_BACKUP.md",
    "docs/YANDEX_DISK_BACKUP_EN.md": "docs/archive/YANDEX_DISK_BACKUP_EN.md",
    "docs/CLOUD_BACKUP_SETUP.md": "docs/archive/CLOUD_BACKUP_SETUP.md",
    "docs/REGRESSION_TEST_PLAN_SUMMARY.md": "docs/archive/REGRESSION_TEST_PLAN_SUMMARY.md",
};

// Directory moves — matched as path prefixes.
const DIR_MOVES = {
    "docs/gordon": "docs/archive/gordon",
    "docs/3d-archive": "docs/archive/3d",
};

const toPosix = (p) => p.split(sep).join("/");

// Maps a repo-relative posix path to its new location, or null if unmoved.
function mapPath(rel) {
    if (FILE_MOVES[rel]) return FILE_MOVES[rel];
    for (const [from, to] of Object.entries(DIR_MOVES)) {
        if (rel === from || rel.startsWith(from + "/")) {
            return to + rel.slice(from.length);
        }
    }
    return null;
}

// ------------------------------------------------------------------ files

const tracked = execFileSync("git", ["ls-files"], { cwd: repoRoot, encoding: "utf8" })
    .split("\n")
    .map((s) => s.trim())
    .filter(Boolean);

const BINARY_EXT = new Set([
    ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".webp", ".gz", ".zip",
    ".pdf", ".woff", ".woff2", ".ttf", ".eot", ".jar", ".exe", ".dll",
]);
const isBinary = (p) => BINARY_EXT.has(p.slice(p.lastIndexOf(".")).toLowerCase());

const files = tracked.filter((f) => !isBinary(f));
const movedSet = new Set(Object.keys(FILE_MOVES));

// --------------------------------------------------------------- rewriting

const LINK_RE = /(!?)\[([^\]]*)\]\(([^)\s]+)((?:\s+"[^"]*")?)\)/g;
const REFDEF_RE = /^(\s*\[[^\]]+\]:\s*)(\S+)(.*)$/gm;

const SKIP_TARGET = /^(?:https?:|mailto:|ftp:|javascript:|#|data:)/i;

function newLocationOf(fileRel) {
    return mapPath(fileRel) ?? fileRel;
}

// Resolve a link target from the file's old location to a repo-relative path.
// Returns the rewritten target string, or null when the link must stay as is.
function rewriteTarget(fileRel, rawTarget) {
    let target = rawTarget.trim();
    if (target.startsWith("<") && target.endsWith(">")) target = target.slice(1, -1);
    if (SKIP_TARGET.test(target)) return null;

    const hashIdx = target.indexOf("#");
    const pathPart = hashIdx === -1 ? target : target.slice(0, hashIdx);
    const anchor = hashIdx === -1 ? "" : target.slice(hashIdx);
    if (!pathPart) return null; // pure anchor

    const oldDir = dirname(fileRel);
    let decoded = pathPart;
    try {
        decoded = decodeURIComponent(pathPart);
    } catch {
        /* malformed escape — resolve the raw path */
    }
    const absTarget = toPosix(resolve(repoRoot, oldDir, decoded));
    const relTarget = toPosix(absTarget).startsWith(toPosix(repoRoot))
        ? toPosix(absTarget).slice(toPosix(repoRoot).length + 1)
        : null;
    if (relTarget === null) return null; // escapes the repo — leave alone

    const mapped = mapPath(relTarget);
    const existsNow = existsSync(join(repoRoot, relTarget));
    if (!mapped && !existsNow) {
        broken.push({ file: fileRel, target: rawTarget });
        return null;
    }
    const newAbs = mapped ?? relTarget;
    const newDir = dirname(newLocationOf(fileRel));
    let relPath = toPosix(relative(join(repoRoot, newDir), join(repoRoot, newAbs)));
    if (!relPath) relPath = ".";
    // Keep the original style: bare target stays bare, ./-prefixed stays prefixed.
    const hadDotPrefix = pathPart.startsWith("./") || pathPart.startsWith("../");
    if (!relPath.startsWith(".") && hadDotPrefix) relPath = "./" + relPath;
    return relPath + anchor;
}

function rewriteMarkdown(content, fileRel) {
    let changed = 0;
    let out = content.replace(LINK_RE, (m, bang, text, target, title) => {
        const nt = rewriteTarget(fileRel, target);
        if (nt === null || nt === target) return m;
        changed++;
        return `${bang}[${text}](${nt}${title})`;
    });
    out = out.replace(REFDEF_RE, (m, prefix, target, rest) => {
        const nt = rewriteTarget(fileRel, target);
        if (nt === null || nt === target) return m;
        changed++;
        return `${prefix}${nt}${rest}`;
    });
    return { out, changed };
}

// Literal `docs/<old>` → `docs/<new>` for anything the link pass could not see
// (script strings, inline code, plain prose).
function rewriteLiterals(content) {
    let changed = 0;
    const keys = [...Object.keys(FILE_MOVES), ...Object.keys(DIR_MOVES)]
        .filter((k) => k.startsWith("docs/"))
        .sort((a, b) => b.length - a.length);
    for (const oldPath of keys) {
        const newPath = mapPath(oldPath);
        const isDir = oldPath in DIR_MOVES;
        const needle = isDir ? oldPath + "/" : oldPath;
        const replacement = isDir ? newPath + "/" : newPath;
        if (!content.includes(needle)) continue;
        // For files, require a word-ish boundary after the match so that
        // `docs/X.md` does not match inside `docs/X.md.bak`-style strings.
        const re = isDir
            ? new RegExp(escapeRe(needle), "g")
            : new RegExp(escapeRe(needle) + "(?![\\w.-])", "g");
        content = content.replace(re, () => {
            changed++;
            return replacement;
        });
    }
    return { out: content, changed };
}

function escapeRe(s) {
    return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

// -------------------------------------------------------------------- main

const broken = [];
const report = [];
let totalLinks = 0;
let totalLiterals = 0;

for (const rel of files) {
    const abs = join(repoRoot, rel);
    if (!existsSync(abs)) continue;
    let content;
    try {
        content = readFileSync(abs, "utf8");
    } catch {
        continue; // not utf8 — skip
    }
    if (content.includes(String.fromCharCode(0))) continue;

    let changedLinks = 0;
    let out = content;
    const isMd = rel.endsWith(".md") || rel === ".windsurfrules";
    if (isMd) {
        const r = rewriteMarkdown(out, rel);
        out = r.out;
        changedLinks = r.changed;
    }
    const lit = rewriteLiterals(out);
    out = lit.out;

    if (changedLinks || lit.changed) {
        report.push({ file: rel, links: changedLinks, literals: lit.changed });
        totalLinks += changedLinks;
        totalLiterals += lit.changed;
        if (!dryRun) writeFileSync(abs, out);
    }
}

// ------------------------------------------------------------------- moves

const moves = [];
for (const [from, to] of Object.entries(FILE_MOVES)) moves.push([from, to]);
for (const [from, to] of Object.entries(DIR_MOVES)) moves.push([from, to]);

console.log(`\n=== ${dryRun ? "DRY RUN" : "RUN"} ===`);
console.log(`rewritten link targets: ${totalLinks}`);
console.log(`rewritten literal paths: ${totalLiterals}`);
console.log(`files touched: ${report.length}`);
for (const r of report) console.log(`  ${r.file}: ${r.links} links, ${r.literals} literals`);

if (broken.length) {
    console.log(`\npre-existing broken link targets (${broken.length}):`);
    for (const b of broken.slice(0, 50)) console.log(`  ${b.file} -> ${b.target}`);
}

if (!dryRun) {
    for (const [from, to] of moves) {
        const absFrom = join(repoRoot, from);
        if (!existsSync(absFrom)) {
            console.log(`  skip (missing): ${from}`);
            continue;
        }
        mkdirSync(dirname(join(repoRoot, to)), { recursive: true });
        execFileSync("git", ["mv", from, to], { cwd: repoRoot });
        console.log(`  git mv ${from} -> ${to}`);
    }
    console.log("\ndone. Run check-docs-links.mjs and inspect `git status`.");
}
