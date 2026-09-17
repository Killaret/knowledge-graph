// Board size guard for BOARD-1.
// The handoff board is not a log: if it grows past the threshold, old terminal
// rows and replies in "Обмен репликами" have not been cleaned up.
//
// THRESHOLD IS AN OWNER DECISION, NOT A TUNABLE. BOARD-1 forbids raising it to
// make a run pass; only the owner may change it, and only on the record.
// 2026-09-14: raised 40 KB -> 120 KB by the owner, who looked at the cleaned
// board and judged three times its current size to be the right room. See
// decision 36 in docs/DECISIONS.md. An agent moving this number on its own
// authority is loosening a guard on itself - do not.

import { statSync } from "node:fs";
import { resolve, join } from "node:path";

const repoRoot = resolve(process.argv[2] ?? ".");
const handoffPath = join(repoRoot, "docs", "AI_HANDOFF.md");
const SIZE_THRESHOLD_BYTES = 120 * 1024;

const { size } = statSync(handoffPath);
const sizeKb = (size / 1024).toFixed(1);
const thresholdKb = (SIZE_THRESHOLD_BYTES / 1024).toFixed(0);

if (size > SIZE_THRESHOLD_BYTES) {
    console.error(
        `Board size guard failed: docs/AI_HANDOFF.md is ${sizeKb} KB, ` +
            `threshold is ${thresholdKb} KB.`,
    );
    console.error(
        "Remove terminal board rows and replies in 'Обмен репликами' " +
            "older than three days. The board is not a log; keep it under " +
            `the ${thresholdKb} KB threshold.`,
    );
    process.exit(1);
}

console.log(
    `Board size OK: ${sizeKb} KB <= ${thresholdKb} KB.`,
);
