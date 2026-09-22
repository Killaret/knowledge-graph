// Commit authorship guard.
//
// Rule: if a commit has a Co-Authored-By trailer naming an agent, the commit
// author must be that same agent. A trailer says "participated", not "wrote",
// so the author field must match the agent who wrote the work. We do not
// rewrite history; we only check the commits that this branch adds.
//
// The agent identities are taken from docs/AI_AGENT_PROTOCOL.md. Email is the
// stable identifier; display names may include model versions.

import { execFileSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";

const repoRoot = process.argv[2] ?? ".";
const baseArg = process.argv[3];

function git(args, options = {}) {
  return execFileSync("git", args, {
    cwd: repoRoot,
    encoding: "utf8",
    maxBuffer: 10 * 1024 * 1024,
    ...options,
  });
}

function parseRangeArg() {
  if (baseArg) {
    return baseArg.includes("..") ? baseArg : `${baseArg}..HEAD`;
  }

  if (process.env.CI) {
    if (process.env.GITHUB_EVENT_NAME === "pull_request" && process.env.GITHUB_BASE_REF) {
      return `origin/${process.env.GITHUB_BASE_REF}..HEAD`;
    }
    if (process.env.GITHUB_EVENT_BEFORE && process.env.GITHUB_SHA) {
      return `${process.env.GITHUB_EVENT_BEFORE}..${process.env.GITHUB_SHA}`;
    }
    return "origin/main..HEAD";
  }

  try {
    git(["rev-parse", "--verify", "main"], { stdio: ["ignore", "pipe", "pipe"] });
    return "main..HEAD";
  } catch {
    return "origin/main..HEAD";
  }
}

function extractEmail(identity) {
  const match = identity.match(/<([^>]+)>/);
  return match ? match[1] : null;
}

// Load agent signatures from the protocol table. The table has rows like:
// || Claude Code | `Claude Opus 5 <noreply@anthropic.com>` |
const protocolPath = resolve(repoRoot, "docs/AI_AGENT_PROTOCOL.md");
const protocol = readFileSync(protocolPath, "utf8").replace(/\r\n/g, "\n");
const agentRows = [...protocol.matchAll(/^\|\s*([^|\n]+?)\s*\|\s*`([^`]+<[^`]+>)`\s*\|/gm)];
const agentSignatures = agentRows.map((m) => m[2].trim());
const agentEmails = new Set(agentSignatures.map(extractEmail).filter(Boolean));
// Agent display name -> signature email ("Devin" -> bot address, etc.).
const agentNameToEmail = new Map(
  agentRows
    .map((m) => [m[1].trim(), extractEmail(m[2].trim())])
    .filter(([, email]) => email),
);

if (agentEmails.size === 0) {
  console.error("Could not find agent signatures in docs/AI_AGENT_PROTOCOL.md");
  process.exit(2);
}

const range = parseRangeArg();

let logOutput;
try {
  logOutput = git(["log", `--format=%H%x1f%an <%ae>%x1f%(trailers:only,unfold)%x1e`, range]);
} catch (err) {
  console.error(`Failed to read git range ${range}: ${err.message}`);
  process.exit(2);
}

const violations = [];
const records = logOutput
  .split("\x1e")
  .map((r) => r.trim())
  .filter(Boolean);

for (const record of records) {
  const parts = record.split("\x1f");
  if (parts.length < 2) continue;
  const [hash, author, trailers] = parts;
  const authorEmail = extractEmail(author);

  const coauthors = trailers
    ? trailers
        .split(/\r?\n/)
        .map((line) => line.trim())
        .filter((line) => line.startsWith("Co-Authored-By:"))
        .map((line) => line.slice("Co-Authored-By:".length).trim())
    : [];

  const agentCoauthors = coauthors
    .map((c) => ({ raw: c, email: extractEmail(c) }))
    .filter((c) => c.email && agentEmails.has(c.email));

  const allMatchAuthor = agentCoauthors.every((c) => c.email === authorEmail);
  if (agentCoauthors.length > 0 && !allMatchAuthor) {
    violations.push({
      hash,
      author,
      trailers: agentCoauthors.map((c) => c.raw),
    });
  }

  // Second rule (AUTHOR-2 review, 2026-09-22): a commit that adds a journal
  // row for an agent (`| date | Agent | ... |` in docs/AI_LOG.md) or updates
  // that agent's "Прочитано:" marker in docs/AI_HANDOFF.md must be authored
  // by that agent. Sessions start and end with exactly such commits, so a
  // swapped identity fails on the very first one — even with no trailers.
  let diff;
  try {
    diff = git(
      ["show", "--format=", "--unified=0", hash, "--", "docs/AI_LOG.md", "docs/AI_HANDOFF.md"],
    );
  } catch {
    continue;
  }
  // Whole-file rewrites (CRLF normalization, table rebuilds) re-add every
  // existing row — those are not new claims. Only lines absent from the
  // parent version count.
  const parentLines = new Set();
  for (const path of ["docs/AI_LOG.md", "docs/AI_HANDOFF.md"]) {
    try {
      for (const l of git(["show", `${hash}^:${path}`]).split(/\r?\n/)) {
        parentLines.add(l);
      }
    } catch {
      // No parent or file absent there — every added line counts.
    }
  }
  const claimedAgents = new Set();
  for (const line of diff.split(/\r?\n/)) {
    if (!line.startsWith("+") || line.startsWith("+++")) continue;
    const text = line.slice(1);
    if (parentLines.has(text)) continue;
    const rowMatch = text.match(/^\|[^|\n]*\|\s*([^|\n]+?)\s*\|/);
    if (rowMatch && agentNameToEmail.has(rowMatch[1])) {
      claimedAgents.add(rowMatch[1]);
      continue;
    }
    const readMatch = text.match(/Прочитано:\s*([^—\n]+?)\s*—/);
    if (readMatch && agentNameToEmail.has(readMatch[1].trim())) {
      claimedAgents.add(readMatch[1].trim());
    }
  }
  const wrongClaims = [...claimedAgents].filter(
    (name) => agentNameToEmail.get(name) !== authorEmail,
  );
  if (wrongClaims.length > 0) {
    violations.push({ hash, author, journalRows: wrongClaims });
  }
}

// Acknowledged history: hashes listed in authorship-corrections.txt are
// journaled violations — history is not rewritten, so each must be recorded
// in docs/AI_LOG.md. A listed hash without a journal entry still fails.
const correctionsPath = resolve(repoRoot, "scripts/testing/authorship-corrections.txt");
let corrections = [];
if (existsSync(correctionsPath)) {
  corrections = readFileSync(correctionsPath, "utf8")
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter((l) => l && !l.startsWith("#"));
}
const acknowledged = new Set(corrections);

const logPath = resolve(repoRoot, "docs/AI_LOG.md");
const logText = existsSync(logPath) ? readFileSync(logPath, "utf8") : "";

const freshViolations = violations.filter((v) => !acknowledged.has(v.hash));
for (const hash of corrections) {
  if (!logText.includes(hash) && !logText.includes(hash.slice(0, 7))) {
    console.error(
      `Correction ${hash} is listed in authorship-corrections.txt but has no entry in docs/AI_LOG.md — journaled acknowledgement is required.`,
    );
    process.exit(1);
  }
}
violations.length = 0;
violations.push(...freshViolations);

if (violations.length > 0) {
  console.error("Commit authorship violations detected:");
  for (const v of violations) {
    console.error(`  ${v.hash}`);
    console.error(`    author: ${v.author}`);
    for (const trailer of v.trailers ?? []) {
      console.error(`    Co-Authored-By: ${trailer}`);
    }
    for (const name of v.journalRows ?? []) {
      console.error(`    journal/board marker claims agent: ${name}`);
    }
  }
  process.exit(1);
}

console.log(`Commit authorship OK for ${records.length} commit(s) in ${range}.`);
