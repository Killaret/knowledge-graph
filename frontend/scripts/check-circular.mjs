#!/usr/bin/env node
// Circular dependency check using madge.
// SvelteKit aliases are resolved through tsconfig.madge.json.
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import madge from "madge";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");

const EXCLUDE_RE = /(?:\.(?:test|spec)\.[tj]s|__mocks__|\.d\.ts|node_modules)$/;
const INCLUDE_RE = /\.(?:[tj]s|svelte)$/;

function collectFiles(dir, files = []) {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
        const full = path.join(dir, entry.name);
        if (entry.isDirectory()) {
            if (
                entry.name === "node_modules" ||
                entry.name === ".svelte-kit" ||
                entry.name === "__mocks__"
            ) {
                continue;
            }
            collectFiles(full, files);
        } else if (INCLUDE_RE.test(entry.name) && !EXCLUDE_RE.test(full)) {
            files.push(full);
        }
    }
    return files;
}

const srcDir = path.join(root, "src");
const files = collectFiles(srcDir);

if (files.length === 0) {
    console.error("No source files found for circular dependency check.");
    process.exit(0);
}

const result = await madge(files, {
    fileExtensions: ["js", "ts", "svelte"],
    tsConfig: path.join(root, "tsconfig.madge.json"),
    excludeRegExp: [EXCLUDE_RE],
});

const circular = result.circular();

if (circular.length) {
    console.error("Found circular dependencies:");
    for (const cycle of circular) {
        console.error("  " + cycle.join(" > "));
    }
    process.exit(1);
}

console.log(`No circular dependencies found (${files.length} files checked).`);
