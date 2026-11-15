// Graph write-path event guard — SYNC-1 stage A2 (owner decision 71).
//
// Stage A required a manual Publish* call next to every repository write;
// stage A2 makes manual publishing impossible by design: the outbox
// decorators (internal/infrastructure/outbox) write a graph_outbox row inside
// the repository transaction, and only the relayer talks to Redis. This guard
// enforces the two invariants that keep the outbox airtight:
//
//   1. No direct PublishNote*/PublishLink* calls outside the relay
//      (infrastructure/outbox) and the publisher type (infrastructure/events).
//      A manual call bypasses the table — the event is gone if the process
//      dies between write and publish, and nothing re-delivers it.
//   2. Every postgres.NewNoteRepository / NewLinkRepository construction is
//      the direct argument of the matching outbox.New*Repository decorator.
//      An undecorated repository writes without an event — silently.
//
// A genuinely event-free construction (e.g. a fixture) must be allowlisted
// below with a reason, which forces a review of the choice.

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

const PUBLISH_RE = /\bPublish(?:Note|Link)\w*\s*\(/g;
const CONSTRUCT_RE = /postgres\.New(Note|Link)Repository\s*\(/g;
const FUNC_RE = /^func\s+(?:\([^)]*\)\s*)?(\w+)\s*\(/gm;

// enclosingFunc finds which parsed function contains the byte offset.
function enclosingFunc(text, offset) {
    let current = "";
    for (const m of text.matchAll(FUNC_RE)) {
        if (m.index > offset) break;
        current = m[1];
    }
    return current;
}

// Directories where Publish* calls are legal: the relay (the only legitimate
// publisher) and the publisher type itself.
const PUBLISH_ALLOWED = new Set([
    "internal/infrastructure/outbox",
    "internal/infrastructure/events",
]);

// file (relative to backendDir) -> reason. Entries here bypass rule 2 —
// construction without the decorator. Rule 1 has no exceptions.
const CONSTRUCT_ALLOWLIST = new Map([]);

function fail(errors, message) {
    errors.push(message);
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
let publishSites = 0;
let constructSites = 0;

for (const file of files) {
    const text = readFileSync(file, "utf8");
    const rel = relative(backendDir, file).split("\\").join("/");
    const dir = rel.split("/").slice(0, -1).join("/");
    const publishAllowed = [...PUBLISH_ALLOWED].some(
        (d) => dir === d || dir.startsWith(d + "/"),
    );

    if (!publishAllowed) {
        for (const m of text.matchAll(PUBLISH_RE)) {
            publishSites += 1;
            const line = text.slice(0, m.index).split("\n").length;
            const fn = enclosingFunc(text, m.index);
            fail(
                errors,
                `manual graph event publish: ${rel}:${line} ` +
                    `(${fn || "file scope"} calls ${m[0].trim()}) ` +
                    `— events must leave through graph_outbox and the relayer, ` +
                    `never through a direct Publish* call`,
            );
        }
    }

    if (!CONSTRUCT_ALLOWLIST.has(rel)) {
        for (const m of text.matchAll(CONSTRUCT_RE)) {
            constructSites += 1;
            const before = text.slice(0, m.index);
            const wrapped = new RegExp(
                `outbox\\.New${m[1]}Repository\\(\\s*(?:/\\*[^*]*\\*/\\s*)?$`,
            ).test(before);
            if (wrapped) continue;
            const line = before.split("\n").length;
            const fn = enclosingFunc(text, m.index);
            fail(
                errors,
                `undecorated repository: ${rel}:${line} ` +
                    `(${fn || "file scope"} constructs postgres.New${m[1]}Repository ` +
                    `without the outbox decorator — writes would commit without a graph event)`,
            );
        }
    }
}

if (errors.length) {
    console.error("Write-path guard: FAIL");
    for (const e of errors) console.error(`  - ${e}`);
    process.exit(1);
}
console.log(
    `Write-path guard OK: ${publishSites + constructSites} inspected sites pass ` +
        `(no manual publishes outside the relay, all repositories wrapped).`,
);
