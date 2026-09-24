import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { findForeignContainers, normalizeWorkingDir } from "./check-test-stack-ownership.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const guardSource = fs.readFileSync(
  path.join(root, "scripts/testing/check-test-stack-ownership.mjs"),
  "utf8"
);

assert.ok(guardSource.includes('"name=^/kg-test-"'));
assert.equal(normalizeWorkingDir("D:\\Repo\\Tree\\"), "d:/repo/tree");
assert.equal(normalizeWorkingDir("/work/repo/"), "/work/repo");

const sameOwner = [
  { id: "1", name: "kg-test-postgres", workingDir: root, startedAt: "now" },
];
assert.deepEqual(findForeignContainers(sameOwner, root), []);

const foreignOwner = [
  { id: "2", name: "kg-test-postgres", workingDir: path.join(root, "..", "other-tree"), startedAt: "then" },
];
assert.deepEqual(findForeignContainers(foreignOwner, root), foreignOwner);

const missingOwner = [
  { id: "3", name: "kg-test-postgres", workingDir: "", startedAt: "unknown" },
];
assert.deepEqual(findForeignContainers(missingOwner, root), missingOwner);

const ps1 = fs.readFileSync(path.join(root, "scripts/testing/start-test.ps1"), "utf8");
const psGuard = ps1.indexOf("& node @guardArgs");
assert.ok(ps1.includes("[switch]$Force"), "PowerShell start must expose explicit -Force takeover");
assert.ok(psGuard >= 0, "PowerShell start must execute the ownership guard");
assert.ok(psGuard < ps1.indexOf("down -v"), "PowerShell guard must run before compose down");
assert.ok(psGuard < ps1.indexOf("docker rm -f"), "PowerShell guard must run before orphan removal");
assert.ok(ps1.includes('name=^/kg-test-'), "PowerShell cleanup must use the guarded name prefix");

const shell = fs.readFileSync(path.join(root, "scripts/testing/start-test.sh"), "utf8");
const shellGuard = shell.indexOf('node scripts/testing/check-test-stack-ownership.mjs "$repo_dir"');
assert.ok(shellGuard >= 0, "shell start must execute the ownership guard");
assert.ok(shellGuard < shell.indexOf("down -v"), "shell guard must run before compose down");
assert.ok(!shell.includes("--force"), "shell start must not silently force takeover");

console.log("Test stack ownership guard tests passed.");
