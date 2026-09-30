#!/usr/bin/env node
/**
 * DOC-AUDIT-2 register integrity guard.
 *
 * Reviewer demand (2026-09-27): every file cited as evidence in
 * docs/tasks/DOC-AUDIT-2-register.md must exist in the repository —
 * the previous pass cited files that were never there
 * (`queue/asynq.go`, `application/recommendation/service.go`, …).
 *
 * Rule: every backticked token that looks like a file path
 * (`dir/name.ext`, `name.ext`, `name.ext:LINE(-LINE)`, a trailing-slash
 * directory, or a glob with `*`/`{…}`) must resolve. Resolution order:
 * exact path from repo root → suffix match (`path` is the tail of a real
 * file) → basename match. Globs are turned into a regex over the file
 * index. Tokens without a file extension, commands, URLs, env vars and
 * identifiers (`graph:events`) are ignored by the extension whitelist.
 *
 * Usage: node scripts/testing/check-doc-audit-2-register.mjs [repoRoot]
 */
import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { join, resolve } from "node:path";

const root = resolve(process.argv[2] ?? ".");
const registerPath = join(root, "docs", "tasks", "DOC-AUDIT-2-register.md");

const SKIP_DIRS = new Set([
  ".git",
  "node_modules",
  "dist",
  "build",
  "coverage",
  ".next",
  ".turbo",
  ".venv",
  "venv",
  "__pycache__",
  "work-model2",
]);

const FILE_EXT = new Set([
  "go", "ts", "tsx", "svelte", "js", "mjs", "cjs", "py", "md", "sql",
  "yaml", "yml", "json", "ps1", "sh", "puml", "txt", "css", "html",
  "env", "toml", "proto",
]);

// ---------------------------------------------------------------- index ---

const files = [];
const dirs = new Set();
function walk(dir) {
  for (const name of readdirSync(dir)) {
    if (SKIP_DIRS.has(name)) continue;
    const full = join(dir, name);
    let st;
    try {
      st = statSync(full);
    } catch {
      continue;
    }
    if (st.isDirectory()) {
      const rel = full.slice(root.length + 1).replace(/\\/g, "/") + "/";
      dirs.add(rel);
      walk(full);
    } else {
      files.push(full.slice(root.length + 1).replace(/\\/g, "/"));
    }
  }
}
walk(root);
const basenames = new Map();
for (const f of files) {
  const base = f.slice(f.lastIndexOf("/") + 1);
  if (!basenames.has(base)) basenames.set(base, []);
  basenames.get(base).push(f);
}

// -------------------------------------------------------------- matching ---

function globToRegex(pattern) {
  // Expand at most one {a,b,c} group, then turn * into [^/]*.
  const expanded = [pattern];
  const brace = pattern.match(/\{([^}]+)\}/);
  if (brace) {
    expanded.length = 0;
    for (const alt of brace[1].split(",")) {
      expanded.push(pattern.replace(brace[0], alt));
    }
  }
  return expanded.map(
    (p) =>
      new RegExp(
        "(^|/)" +
          p
            .split("*")
            .map((seg) => seg.replace(/[.+?^${}()|[\]\\]/g, "\\$&"))
            .join("[^/]*") +
          "$",
      ),
  );
}

// Globs may name directories (cmd/*-recompute) as well as files — test both.
const dirList = [...dirs].map((d) => d.replace(/\/$/, ""));

function resolveRef(ref) {
  if (existsSync(join(root, ref))) return true;
  if (ref.endsWith("/")) {
    return dirs.has(ref) || [...dirs].some((d) => d.endsWith(ref));
  }
  if (files.includes(ref)) return true;
  if (files.some((f) => f.endsWith("/" + ref))) return true;
  if (!ref.includes("/")) {
    return basenames.has(ref);
  }
  return false;
}

function checkToken(raw) {
  // strip line anchors :123, :123-140
  const ref = raw.replace(/:\d+(-\d+)?$/, "").replace(/[()\[\]]/g, "");
  const isDir = ref.endsWith("/");
  const hasGlob = /[*{]/.test(ref);
  const base = ref.replace(/\/+$/, "").split("/").pop();
  const ext = base.includes(".") ? base.split(".").pop().toLowerCase() : null;
  if (hasGlob) {
    // Symbol globs like `TestX_*` carry no path separator — not file refs.
    if (!ref.includes("/") && !ref.includes(".")) return null;
    const ok = globToRegex(ref).some(
      (re) => files.some((f) => re.test(f)) || dirList.some((d) => re.test(d)),
    );
    return ok ? null : raw;
  }
  if (!isDir && !FILE_EXT.has(ext)) return null; // not a file ref
  return resolveRef(ref) ? null : raw;
}

// --------------------------------------------------------------- parsing ---

// Only the «Доказательство» column is checked: it is the one that must point
// at real code. The claim and history columns legitimately cite paths that
// describe what was documented (and was wrong or absent) — checking those
// would flag the findings themselves.
const register = readFileSync(registerPath, "utf8");
const bad = new Map(); // token -> line number
const noLineRef = []; // rows whose verdict is «верно» but evidence lacks file:line
let evidenceCol = -1;
let verdictCol = -1;
register.split("\n").forEach((line, i) => {
  if (!line.trimStart().startsWith("|")) {
    evidenceCol = -1;
    verdictCol = -1;
    return;
  }
  const cells = line.split("|").slice(1, -1).map((c) => c.trim());
  const headerIdx = cells.findIndex((c) => c === "Доказательство");
  if (headerIdx >= 0) {
    evidenceCol = headerIdx;
    verdictCol = cells.findIndex((c) => c === "Вердикт");
    return;
  }
  if (evidenceCol < 0 || cells.every((c) => /^-+$/.test(c))) return;
  const cell = cells[evidenceCol] ?? "";
  for (const m of cell.matchAll(/`([^`\n]+)`/g)) {
    // Split multi-token spans on whitespace only — commas inside {a,b}
    // brace groups must not split the glob.
    for (const tok of m[1].split(/\s+/).filter(Boolean)) {
      const miss = checkToken(tok.replace(/^,|,$/g, ""));
      if (miss && !bad.has(miss)) bad.set(miss, i + 1);
    }
  }
  // Reviewer demand (2026-09-29): a «верно» verdict must cite `file:line`,
  // not a package name or «соответствует».
  const verdict = verdictCol >= 0 ? (cells[verdictCol] ?? "") : "";
  // Plans, historical notes and contract descriptions have nothing in code
  // to point at — forcing file:line there would fabricate evidence.
  const noCodeClaim = /план|историч|контракт|иде|спека|проверяется конфигурацией/.test(verdict);
  if (verdict.includes("верно") && !noCodeClaim && !/`[^\s`]+:\d+/.test(cell)) {
    noLineRef.push(i + 1);
  }
});

let failed = false;
if (bad.size) {
  console.log(
    `DOC-AUDIT-2 register: ${bad.size} referenced path(s) do not exist:`,
  );
  for (const [tok, line] of bad) {
    console.log(`  - line ${line}: ${tok}`);
  }
  failed = true;
}
if (noLineRef.length) {
  console.log(
    `DOC-AUDIT-2 register: ${noLineRef.length} «верно» row(s) lack file:line evidence:`,
  );
  for (const line of noLineRef) console.log(`  - line ${line}`);
  failed = true;
}
if (failed) process.exit(1);
console.log(
  `DOC-AUDIT-2 register integrity OK: ${files.length} files indexed, all references resolve, all «верно» rows carry file:line evidence.`,
);
