import { execFileSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

export function normalizeWorkingDir(value) {
  return value.trim().replaceAll("\\", "/").replace(/\/+$/, "").toLowerCase();
}

export function findForeignContainers(records, repoDir) {
  const owner = normalizeWorkingDir(path.resolve(repoDir));
  return records.filter((record) => normalizeWorkingDir(record.workingDir) !== owner);
}

function docker(args) {
  return execFileSync("docker", args, { encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }).trim();
}

export function inspectTestContainers() {
  const ids = docker(["ps", "-aq", "--filter", "name=^/kg-test-"])
    .split(/\r?\n/)
    .map((id) => id.trim())
    .filter(Boolean);

  return ids.map((id) => {
    const output = docker([
      "inspect",
      "--format",
      '{{ index .Config.Labels "com.docker.compose.project.working_dir" }}\t{{.State.StartedAt}}\t{{.Name}}',
      id,
    ]);
    const [workingDir = "", startedAt = "unknown", name = id] = output.split("\t");
    return { id, workingDir, startedAt, name: name.replace(/^\//, "") };
  });
}

export function checkOwnership(repoDir, force = false, records = inspectTestContainers()) {
  const foreign = findForeignContainers(records, repoDir);
  if (foreign.length === 0 || force) return foreign;

  console.error("ERROR: test stack belongs to another working tree; nothing was removed.");
  for (const item of foreign) {
    console.error(`  ${item.name}: owner=${item.workingDir || "<missing label>"}, started=${item.startedAt}`);
  }
  console.error("Stop it from its owning tree. On Windows, start-test.ps1 -Force takes ownership explicitly; the shell script has no force option.");
  process.exitCode = 1;
  return foreign;
}

const isMain = process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url);
if (isMain) {
  const args = process.argv.slice(2);
  const force = args.includes("--force");
  const repoDir = args.find((arg) => arg !== "--force");
  if (!repoDir) {
    console.error("Usage: node check-test-stack-ownership.mjs <repo-dir> [--force]");
    process.exit(2);
  }
  checkOwnership(repoDir, force);
}
