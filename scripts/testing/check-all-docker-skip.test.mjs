// CHECK-ALL-2 regression: when `docker info` fails on a machine where docker
// is "installed" but the daemon answer is transient or broken, the skip reason
// must name the command and show its output — not just "Docker daemon is
// unavailable", which hid live-stack phases behind a silent skip.
//
// A stub `docker` on PATH simulates the failure; a one-phase manifest keeps
// the run fast (the workflow sync check is skipped for custom manifests).

import { spawnSync } from "node:child_process";
import { chmodSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { join, dirname } from "node:path";
import { tmpdir } from "node:os";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const ps1 = join(here, "check-all.ps1");
const sh = join(here, "check-all.sh");

const work = mkdtempSync(join(tmpdir(), "check-all-docker-"));
const stubDir = join(work, "stub");
mkdirSync(stubDir, { recursive: true });
const manifest = join(work, "manifest.tsv");

const FAIL_MARKER = "simulated-npipe-failure-CHECK-ALL-2";

// One docker-requiring phase whose command trivially succeeds — so PASS vs
// SKIP depends purely on the docker probe.
writeFileSync(
    manifest,
    "id\tname\tjob\tworkflow_step\tcwd\tcommand\tsignature\tquick_skip\ttool\n" +
        "fixture-docker\tFixture docker phase\tfixture\tFixture step\t.\tnode --version\tnode --version\t0\tdocker+node\n",
    "utf8",
);

let failures = 0;
const check = (name, ok, detail) => {
    if (ok) {
        console.log(`PASS  ${name}`);
    } else {
        failures++;
        console.log(`FAIL  ${name}\n${detail}`);
    }
};

function writeDockerStub(fails) {
    const cmd = join(stubDir, "docker.cmd");
    const shell = join(stubDir, "docker");
    if (fails) {
        writeFileSync(cmd, `@echo off\r\necho ${FAIL_MARKER} 1>&2\r\nexit /b 1\r\n`);
        writeFileSync(shell, `#!/bin/sh\necho ${FAIL_MARKER} >&2\nexit 1\n`);
    } else {
        writeFileSync(cmd, "@echo off\r\necho ServerVersion: stub\r\nexit /b 0\r\n");
        writeFileSync(shell, "#!/bin/sh\necho ServerVersion: stub\nexit 0\n");
    }
    try { chmodSync(shell, 0o755); } catch { /* non-POSIX fs */ }
}

function runWithStub(fails, runner) {
    writeDockerStub(fails);
    const env = { ...process.env, PATH: `${stubDir}${process.platform === "win32" ? ";" : ":"}${process.env.PATH ?? ""}` };
    return runner(env);
}

try {
    // --- check-all.ps1 (Windows PowerShell) ---
    if (process.platform === "win32") {
        const runPs = (env) => spawnSync(
            "powershell",
            ["-NoProfile", "-ExecutionPolicy", "Bypass", "-File", ps1, "-Manifest", manifest],
            { encoding: "utf8", env, timeout: 120000 },
        );

        let res = runWithStub(true, runPs);
        let out = `${res.stdout ?? ""}${res.stderr ?? ""}`;
        check(
            "ps1: failing docker stub -> SKIP naming command and output",
            out.includes("[SKIP] Fixture docker phase") &&
                out.includes("docker info") &&
                out.includes(FAIL_MARKER),
            `exit=${res.status}\n${out}`,
        );

        res = runWithStub(false, runPs);
        out = `${res.stdout ?? ""}${res.stderr ?? ""}`;
        check(
            "ps1: healthy docker stub -> phase runs",
            out.includes("[PASS] Fixture docker phase") && !out.includes("[SKIP] Fixture docker phase"),
            `exit=${res.status}\n${out}`,
        );
    }

    // --- check-all.sh ---
    const bashCheck = spawnSync("bash", ["-lc", "command -v bash"], { encoding: "utf8" });
    if (bashCheck.status === 0) {
        const runSh = (env) => spawnSync("bash", [sh, "--manifest", manifest], {
            encoding: "utf8",
            env,
            timeout: 120000,
        });

        let res = runWithStub(true, runSh);
        let out = `${res.stdout ?? ""}${res.stderr ?? ""}`;
        check(
            "sh: failing docker stub -> SKIP naming command and output",
            out.includes("[SKIP] Fixture docker phase") &&
                out.includes("docker info") &&
                out.includes(FAIL_MARKER),
            `exit=${res.status}\n${out}`,
        );

        res = runWithStub(false, runSh);
        out = `${res.stdout ?? ""}${res.stderr ?? ""}`;
        check(
            "sh: healthy docker stub -> phase runs",
            out.includes("[PASS] Fixture docker phase") &&
                !out.includes("[SKIP] Fixture docker phase"),
            `exit=${res.status}\n${out}`,
        );
    } else {
        console.log("SKIP  bash not found — sh variant not exercised");
    }
} finally {
    rmSync(work, { recursive: true, force: true });
}

if (failures > 0) {
    process.exit(1);
}
console.log("All checks passed.");
