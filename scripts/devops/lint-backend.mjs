// Backend lint task for lint-staged (AUTHOR-2 fix).
//
// lint-staged spawns commands through tinyexec, which does not run a shell:
// `cd backend && golangci-lint ...` cannot be parsed and fails on Windows
// with "The filename, directory name, or volume label syntax is incorrect".
// This wrapper does the directory change in-process instead, keeping the
// command itself shell-free.

import { spawnSync } from "node:child_process";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const backendDir = join(
    dirname(fileURLToPath(import.meta.url)),
    "..",
    "..",
    "backend",
);

const args = ["run", "--new-from-rev=HEAD~1"];
let result = spawnSync("golangci-lint", args, {
    cwd: backendDir,
    stdio: "inherit",
    shell: false,
});

// Not on PATH? Try the Go workspace bin directory — Go installs tools there
// and many clones never add it to PATH.
if (result.error && result.error.code === "ENOENT") {
    const go = spawnSync("go", ["env", "GOPATH"], {
        encoding: "utf8",
        shell: false,
    });
    const goPath = (go.stdout ?? "").trim();
    if (go.status === 0 && goPath) {
        const bin = join(
            goPath,
            "bin",
            process.platform === "win32" ? "golangci-lint.exe" : "golangci-lint",
        );
        result = spawnSync(bin, args, {
            cwd: backendDir,
            stdio: "inherit",
            shell: false,
        });
    }
}

if (result.error) {
    console.error(
        `golangci-lint failed to start: ${result.error.message}. ` +
            "Install it (see COMMANDS.md) or put it on PATH.",
    );
    process.exit(1);
}
process.exit(result.status ?? 1);
