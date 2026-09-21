// Git hooks activation check for AUTHOR-2.
//
// The pre-push authorship guard only runs when core.hooksPath points at
// husky's directory, and `npm run prepare` is the only thing that sets it.
// A clone where the hook stays silent must say so itself.
//
// On CI runners the check skips itself: GitHub Actions does not push from the
// checkout, so hooks are unnecessary there — authorship is caught directly by
// check-commit-authorship.mjs.

import { execFileSync } from "node:child_process";
import { existsSync } from "node:fs";
import { join, resolve } from "node:path";

const repoRoot = resolve(process.argv[2] ?? ".");

if (process.env.CI) {
    console.log(
        "Hooks activation check skipped: CI runners do not push; " +
            "authorship is checked by check-commit-authorship.mjs directly.",
    );
    process.exit(0);
}

let hooksPath = "";
try {
    hooksPath = execFileSync(
        "git",
        ["config", "--get", "core.hooksPath"],
        { cwd: repoRoot, encoding: "utf8" },
    ).trim();
} catch {
    hooksPath = ""; // unset -> git exits 1
}

const normalized = hooksPath.replace(/\\/g, "/");
const pointsAtHusky =
    normalized === ".husky/_" ||
    normalized === ".husky/_/" ||
    normalized.endsWith("/.husky/_") ||
    normalized.endsWith("/.husky/_/");

const problems = [];
if (!pointsAtHusky) {
    problems.push(
        `core.hooksPath is "${hooksPath || "(unset)"}", expected ".husky/_"`,
    );
}
// Husky creates the `_/pre-push` stub in every prepared clone; the real guard
// script is `.husky/pre-push`. Both must exist — a stub without the script
// means git runs a no-op wrapper and the authorship guard stays silent.
if (!existsSync(join(repoRoot, ".husky", "_", "pre-push"))) {
    problems.push(".husky/_/pre-push is missing");
}
if (!existsSync(join(repoRoot, ".husky", "pre-push"))) {
    problems.push(
        ".husky/pre-push is missing (the stub alone is a no-op; on main " +
            "the file arrives only when ai-agents merges - AUTHOR-2)",
    );
}

if (problems.length > 0) {
    console.error("Hooks activation check failed:");
    for (const problem of problems) {
        console.error(`  ${problem}`);
    }
    console.error(
        "Run `npm run prepare` in this clone (husky creates .husky/_ " +
            "per worktree). The pre-push authorship guard stays silent until then.",
    );
    process.exit(1);
}

console.log(`Hooks active: core.hooksPath=${hooksPath}, .husky/_/pre-push present.`);
