// Agent session start guard.
//
// Rule: an agent session must start from a clean working tree. If another
// agent's uncommitted work is present, the current agent must not begin, or it
// risks capturing that work under the wrong authorship.
//
// This is the second signal for AUTHOR-1: the commit-trailer check catches
// wrong attribution after the fact; this check removes the condition that lets
// it happen.

import { execFileSync } from "node:child_process";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = process.argv[2] ?? ".";
const hooksCheck = join(
  dirname(fileURLToPath(import.meta.url)),
  "check-hooks-active.mjs",
);

// AUTHOR-2: a hook that is not active must say so itself. Delegated to
// check-hooks-active.mjs so the check-all phase and this guard share one rule.
try {
  execFileSync("node", [hooksCheck, repoRoot], { stdio: "inherit" });
} catch {
  process.exit(1);
}

let status;
try {
  status = execFileSync(
    "git",
    ["status", "--porcelain", "--untracked-files=all"],
    { cwd: repoRoot, encoding: "utf8" },
  );
} catch (err) {
  console.error(`Failed to read git status: ${err.message}`);
  process.exit(2);
}

const lines = status.split(/\r?\n/).filter(Boolean);

if (lines.length > 0) {
  console.error(
    "Agent session cannot start: the working tree is not clean.",
  );
  console.error(
    "Another agent's uncommitted changes are present. " +
      "Commit, stash, or clean them before starting a session.",
  );
  console.error("Dirty files:");
  for (const line of lines) {
    console.error(`  ${line}`);
  }
  process.exit(1);
}

console.log("Working tree is clean; agent session may start.");
