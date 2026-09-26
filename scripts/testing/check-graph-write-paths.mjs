// Graph write-path event guard (SYNC-1, owner decision 69).
//
// Every code path that creates, updates, deletes or restores a note or a link
// must publish a graph event — otherwise the graph-service cache is not
// invalidated and the change stays invisible to open graphs until TTL
// (hole #4 in tasks/SYNC-1-graph-loading-and-sync-review.md: import created
// notes with zero events).
//
// The guard scans backend Go sources for repository write calls on the
// note/link repositories and fails when a call site has no Publish* event in
// the same function (or in a same-file helper it calls, e.g.
// postprocessCreatedNote) and is not explicitly allowlisted below. A new write
// path without an event is red immediately; a legitimately non-graph write
// must be allowlisted here with a reason, which forces a review of the choice.

import { readFileSync, readdirSync, statSync, existsSync } from "node:fs";
import { join, resolve, relative } from "node:path";

const args = process.argv.slice(2);
let repoRoot = ".";
let backendDir = null;
for (const arg of args) {
    if (arg.startsWith("--backend=")) {
        backendDir = resolve(arg.slice("--backend=".length));
    } else {
        repoRoot = arg;
    }
}
repoRoot = resolve(repoRoot);
backendDir = backendDir ?? join(repoRoot, "backend");

// Receivers carrying note/link repositories: h.repo/s.repo (handlers and
// services whose main repo is notes), linkRepo/noteRepo (explicit names).
const WRITE_RE =
    /\b(?:[a-zA-Z_]\w*\.)?(?:noteRepo|linkRepo|repo)\.(Save|SaveUserLink|Update|Delete|DeleteBySource|DeleteAndSuppress|DeleteBatch|Restore)\s*\(/g;
const PUBLISH_RE = /\bPublish(?:Note|Link)\w*\s*\(/;
const FUNC_RE = /^func\s+(?:\([^)]*\)\s*)?(\w+)\s*\(/gm;

// file (relative to backendDir) -> "*" for the whole file, or a list of
// function names. Every entry must state why the write needs no event.
const ALLOWLIST = new Map([
    // Drafts are a pre-publish workspace, not graph notes; the graph event
    // fires when the draft is published through the note handler.
    ["internal/application/draft/service.go", "*"],
    // User settings — not graph data.
    ["internal/application/user/settings_service.go", "*"],
    // User profile repository — not graph data.
    ["internal/interfaces/api/handlers/user/handler.go", "*"],
    [
        // The generator only writes; LinkCreated is published by the caller,
        // queue.Worker.generateGammaLinks (worker.go), which owns the userID.
        "internal/application/recommendation/gamma_link_generator.go",
        new Set(["saveMissingGammaLinks"]),
    ],
]);

function fail(errors, message) {
    errors.push(message);
}

// splitFunctions maps each top-level func (methods included) to its body —
// the text between this func's header and the next func header. Nested
// closures stay inside the enclosing chunk, which is what we want.
function splitFunctions(text) {
    const funcs = new Map();
    const matches = [...text.matchAll(FUNC_RE)];
    for (let i = 0; i < matches.length; i++) {
        const start = matches[i].index;
        const end = i + 1 < matches.length ? matches[i + 1].index : text.length;
        funcs.set(matches[i][1], text.slice(start, end));
    }
    return funcs;
}

// emitsPublish reports whether funcName's body publishes an event itself or
// calls a same-file function that does (one hop is enough: helpers like
// postprocessCreatedNote are the pattern this project uses).
function emitsPublish(funcs, funcName, depth = 0) {
    const body = funcs.get(funcName);
    if (!body) return false;
    if (PUBLISH_RE.test(body)) return true;
    if (depth >= 2) return false;
    for (const m of body.matchAll(/\b([a-zA-Z_]\w*)\s*\(/g)) {
        if (funcs.has(m[1]) && emitsPublish(funcs, m[1], depth + 1)) return true;
    }
    return false;
}

// enclosingFunc finds which parsed function contains the byte offset.
function enclosingFunc(text, offset) {
    let current = "";
    for (const m of text.matchAll(FUNC_RE)) {
        if (m.index > offset) break;
        current = m[1];
    }
    return current;
}

function walk(dir, out) {
    for (const entry of readdirSync(dir)) {
        const p = join(dir, entry);
        if (statSync(p).isDirectory()) {
            walk(p, out);
        } else if (p.endsWith(".go") && !p.endsWith("_test.go") && !/mock/i.test(entry)) {
            out.push(p);
        }
    }
}

const errors = [];

if (!existsSync(backendDir)) {
    console.error(`Write-path guard: backend dir not found: ${backendDir}`);
    process.exit(1);
}

const files = [];
walk(backendDir, files);
let sites = 0;

for (const file of files) {
    const text = readFileSync(file, "utf8");
    const rel = relative(backendDir, file).split("\\").join("/");
    const funcs = splitFunctions(text);
    const allow = ALLOWLIST.get(rel);

    for (const m of text.matchAll(WRITE_RE)) {
        sites += 1;
        const fn = enclosingFunc(text, m.index);
        if (allow === "*" || (allow instanceof Set && allow.has(fn))) continue;
        if (fn && emitsPublish(funcs, fn)) continue;
        const line = text.slice(0, m.index).split("\n").length;
        fail(
            errors,
            `write path without a graph event: ${rel}:${line} ` +
                `(${fn || "file scope"} calls .${m[1]} but never reaches a Publish* call)`,
        );
    }
}

if (errors.length) {
    console.error("Write-path guard: FAIL");
    for (const e of errors) console.error(`  - ${e}`);
    process.exit(1);
}
console.log(`Write-path guard OK: ${sites} repository write sites publish graph events or are allowlisted.`);
