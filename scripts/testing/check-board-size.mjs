// Board size guard for BOARD-1.
// The handoff board is not a log: if it grows past 40 KB, old terminal rows
// and replies in "Обмен репликами" have not been cleaned up.

import { statSync } from "node:fs";
import { resolve, join } from "node:path";

const repoRoot = resolve(process.argv[2] ?? ".");
const handoffPath = join(repoRoot, "docs", "AI_HANDOFF.md");
const SIZE_THRESHOLD_BYTES = 40 * 1024;

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
