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
        return { "docs/AI_LOG.md": readFileSync(p, "utf8") + `| 2026-09-22 | ${agent} | row ${++seq} [t](old.md) |\n` };
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
            // DOC-REORG-1 moved files and rewrote link targets inside old
            // journal rows — the row itself is unchanged, so this is not a
            // new claim and must not require the row's agent as author.
            name: "link rewrite inside Claude row authored by Devin",
            author: DEVIN,
            files: () => ({
                "docs/AI_LOG.md": readFileSync(join(repo, "docs/AI_LOG.md"), "utf8")
                    .replaceAll("old.md", "moved/old.md"),
            }),
            expectExit: 0,
        },
        {
            name: "edited Claude row text authored by Devin",
            author: DEVIN,
            files: () => ({
                "docs/AI_LOG.md": readFileSync(join(repo, "docs/AI_LOG.md"), "utf8")
                    .replace(/\| Claude Code \| row \d+/, "| Claude Code | edited row"),
            }),
            expectExit: 1,
            stderrIncludes: "claims agent: Claude Code",
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
    const check = (name, res, expectExit, stderrIncludes) => {
        const problems = [];
        if (res.code !== expectExit) problems.push(`exit ${res.code}, expected ${expectExit}`);
        if (stderrIncludes && !res.output.includes(stderrIncludes)) {
            problems.push(`output missing "${stderrIncludes}"`);
        }
        if (problems.length) {
            failed = true;
            console.error(`FAIL ${name}: ${problems.join("; ")}`);
            console.error(`  output: ${res.output.trim().slice(0, 400)}`);
        } else {
            console.log(`PASS ${name}`);
        }
    };
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
        check(c.name, res, c.expectExit, c.stderrIncludes);
    }

    // Merge commits: conflict resolution legitimately re-adds the second
    // parent's lines — those must not count as new claims by the merger,
    // while a forged agent claim absent from BOTH parents must still fail.
    const commitAs = (author, msg) =>
        execFileSync("git", ["-c", "user.name=tester", "-c", "user.email=tester@test.local",
                             "add", "-A"], { cwd: repo }) &&
        execFileSync("git", ["-c", "user.name=tester", "-c", "user.email=tester@test.local",
                             "commit", "-m", msg, `--author=${author}`], { cwd: repo });

    // Claude's branch: his journal row and read marker.
    const mainTip = git(repo, ["rev-parse", "HEAD"]).trim();
    git(repo, ["checkout", "-b", "theirs"]);
    writeFileSync(join(repo, "docs/AI_LOG.md"),
        readFileSync(join(repo, "docs/AI_LOG.md"), "utf8") +
            "| 2026-09-25 | Claude Code | their row |\n");
    writeFileSync(join(repo, "docs/AI_HANDOFF.md"),
        readFileSync(join(repo, "docs/AI_HANDOFF.md"), "utf8") +
            "Прочитано: Claude Code — 2026-09-25 — cafe123\n");
    commitAs(CLAUDE, "their work");

    // Devin's main: his own marker.
    git(repo, ["checkout", "main"]);
    writeFileSync(join(repo, "docs/AI_HANDOFF.md"),
        readFileSync(join(repo, "docs/AI_HANDOFF.md"), "utf8") +
            "Прочитано: Devin — 2026-09-25 — beef456\n");
    commitAs(DEVIN, "my work");

    // Legit merge: resolution keeps both sides — every Claude line exists in
    // the second parent, so the merge authored by Devin must pass.
    spawnSync("git", ["merge", "--no-commit", "--no-ff", "theirs"], { cwd: repo });
    writeFileSync(join(repo, "docs/AI_HANDOFF.md"),
        "Прочитано: Claude Code — 2026-09-25 — cafe123\n" +
        "Прочитано: Devin — 2026-09-25 — beef456\n");
    commitAs(DEVIN, "merge theirs");
    check("merge carrying Claude lines authored by Devin", runGuard(repo), 0);

    // Forged merge: Devin merges a second Claude branch but slips in a Claude
    // marker that exists in NEITHER parent — must be flagged.
    const mainTip2 = git(repo, ["rev-parse", "HEAD"]).trim();
    git(repo, ["checkout", "-b", "theirs2", mainTip]);
    writeFileSync(join(repo, "docs/AI_LOG.md"),
        readFileSync(join(repo, "docs/AI_LOG.md"), "utf8") +
            "| 2026-09-26 | Claude Code | second row |\n");
    commitAs(CLAUDE, "their second work");
    git(repo, ["checkout", "main"]);
    spawnSync("git", ["merge", "--no-commit", "--no-ff", "theirs2"], { cwd: repo });
    writeFileSync(join(repo, "docs/AI_HANDOFF.md"),
        readFileSync(join(repo, "docs/AI_HANDOFF.md"), "utf8") +
            "Прочитано: Claude Code — 2026-09-26 — forged0\n");
    commitAs(DEVIN, "merge theirs2 with forged marker");
    check("forged Claude marker inside merge resolution", runGuard(repo), 1,
        "claims agent: Claude Code");

    process.exit(failed ? 1 : 0);
} finally {
    rmSync(repo, { recursive: true, force: true });
}
