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
import { existsSync } from "node:fs";
import { join, resolve, sep } from "node:path";

const repoRoot = process.argv[2] ?? ".";

// AUTHOR-2: a hook that is not active must say so itself. The pre-push
// authorship guard only runs when core.hooksPath points at husky's directory;
// `npm run prepare` sets that, and nothing else does.
function checkHooksActive() {
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

  const huskyDir = join(repoRoot, ".husky", "_");
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
  if (!existsSync(join(huskyDir, "pre-push"))) {
    problems.push(".husky/_/pre-push is missing");
  }

  if (problems.length > 0) {
    console.error(
      "Agent session cannot start: git hooks are not activated.",
    );
    for (const problem of problems) {
      console.error(`  ${problem}`);
    }
    console.error(
      "Run `npm run prepare` in this clone (husky creates .husky/_ " +
        "per worktree). The pre-push authorship guard stays silent until then.",
    );
    process.exit(1);
  }
}

checkHooksActive();

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
