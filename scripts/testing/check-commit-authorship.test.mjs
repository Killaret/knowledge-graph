// Mutation tests for check-commit-authorship.mjs (AUTHOR-2 rework).
//
// Builds a throwaway git repo with a minimal protocol table, then commits
// controlled changes under controlled authors and runs the guard on the
// single-commit range. Each scenario must produce the expected verdict.

import { execFileSync, spawnSync } from "node:child_process";
import { mkdtempSync, mkdirSync, writeFileSync, readFileSync, rmSync } from "node:fs";
import { join, dirname } from "node:path";
import { tmpdir } from "node:os";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const guard = join(here, "check-commit-authorship.mjs");

const DEVIN = "Devin <158243242+devin-ai-integration[bot]@users.noreply.github.com>";
const CLAUDE = "Claude Opus 5 <noreply@anthropic.com>";
const OWNER = "Yaroslav <owner@test.local>";

function git(repo, args) {
    return execFileSync("git", args, { cwd: repo, encoding: "utf8" });
}

function runGuard(repo) {
    const res = spawnSync("node", [guard, repo, "HEAD~1..HEAD"], { encoding: "utf8" });
    return { code: res.status ?? 1, output: `${res.stdout ?? ""}${res.stderr ?? ""}` };
}

const repo = mkdtempSync(join(tmpdir(), "authorship-guard-"));
try {
    git(repo, ["init", "-b", "main"]);
    mkdirSync(join(repo, "docs"), { recursive: true });
    writeFileSync(
        join(repo, "docs/AI_AGENT_PROTOCOL.md"),
        `| Agent | Signature |\n|---|---|\n| Devin | \`${DEVIN}\` |\n| Claude Code | \`${CLAUDE}\` |\n`,
        "utf8",
    );
    writeFileSync(join(repo, "docs/AI_LOG.md"), "# Log\n\n| Дата | Агент | Что | Статус | Где |\n|---|---|---|---|---|\n", "utf8");
    writeFileSync(join(repo, "docs/AI_HANDOFF.md"), "Прочитано: Devin — 2026-01-01 — 0000000\n", "utf8");
    execFileSync("git", ["-c", "user.name=tester", "-c", "user.email=tester@test.local", "add", "-A"], { cwd: repo });
    execFileSync("git", ["-c", "user.name=tester", "-c", "user.email=tester@test.local", "commit", "-m", "init", `--author=${OWNER}`], { cwd: repo });

    let seq = 0;
    const appendLogRow = (agent) => {
        const p = join(repo, "docs/AI_LOG.md");
        return { "docs/AI_LOG.md": readFileSync(p, "utf8") + `| 2026-09-22 | ${agent} | row ${++seq} |\n` };
    };

    const cases = [
        {
            name: "Devin journal row authored by Devin",
            author: DEVIN,
            files: () => appendLogRow("Devin"),
            expectExit: 0,
        },
        {
            name: "Devin journal row authored by Claude",
            author: CLAUDE,
            files: () => appendLogRow("Devin"),
            expectExit: 1,
            stderrIncludes: "claims agent: Devin",
        },
        {
            name: "Claude journal row authored by Claude",
            author: CLAUDE,
            files: () => appendLogRow("Claude Code"),
            expectExit: 0,
        },
        {
            name: "Devin read-marker update authored by Claude",
            author: CLAUDE,
            files: () => ({ "docs/AI_HANDOFF.md": "Прочитано: Devin — 2026-09-22 — abcdef1\n" }),
            expectExit: 1,
            stderrIncludes: "claims agent: Devin",
        },
        {
            name: "Devin read-marker update authored by Devin",
            author: DEVIN,
            files: () => ({ "docs/AI_HANDOFF.md": "Прочитано: Devin — 2026-09-23 — abcdef2\n" }),
            expectExit: 0,
        },
        {
            name: "owner commit with agent trailer (rule 1)",
            author: OWNER,
            files: () => ({ "code.txt": "x\n" }),
            trailer: `Co-Authored-By: ${DEVIN}`,
            expectExit: 1,
            stderrIncludes: "Co-Authored-By",
        },
        {
            name: "plain code commit by owner",
            author: OWNER,
            files: () => ({ "code.txt": "y\n" }),
            expectExit: 0,
        },
    ];

    let failed = false;
    for (const c of cases) {
        const files = c.files();
        for (const [path, content] of Object.entries(files)) {
            writeFileSync(join(repo, path), content, "utf8");
        }
        execFileSync("git", ["add", "-A"], { cwd: repo });
        const args = ["-c", "user.name=tester", "-c", "user.email=tester@test.local",
                      "commit", "--allow-empty", "-m", c.name, `--author=${c.author}`];
        execFileSync("git", args, { cwd: repo });
        if (c.trailer) {
            execFileSync("git", ["-c", "user.name=tester", "-c", "user.email=tester@test.local",
                                 "commit", "--amend", "-m", `${c.name}\n\n${c.trailer}`,
                                 `--author=${c.author}`], { cwd: repo });
        }

        const res = runGuard(repo);
        const problems = [];
        if (res.code !== c.expectExit) problems.push(`exit ${res.code}, expected ${c.expectExit}`);
        if (c.stderrIncludes && !res.output.includes(c.stderrIncludes)) {
            problems.push(`output missing "${c.stderrIncludes}"`);
        }
        if (problems.length) {
            failed = true;
            console.error(`FAIL ${c.name}: ${problems.join("; ")}`);
            console.error(`  output: ${res.output.trim().slice(0, 400)}`);
        } else {
            console.log(`PASS ${c.name}`);
        }
    }
    process.exit(failed ? 1 : 0);
} finally {
    rmSync(repo, { recursive: true, force: true });
}
