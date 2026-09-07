import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const [manifestArg, workflowArg] = process.argv.slice(2);
if (!manifestArg || !workflowArg) {
    console.error(
        "Usage: node check-core-workflow-sync.mjs <manifest.tsv> <workflow.yml>",
    );
    process.exit(2);
}

const targetJobs = new Set([
    "backend-checks",
    "backend-integration-tests",
    "graph-service-checks",
    "frontend-checks",
    "nlp-checks",
]);
const ignoredSteps = new Set([
    "Download dependencies",
    "Install dependencies",
    "Clean ESLint cache",
    "Check requirements exist",
]);

const manifestLines = readFileSync(resolve(manifestArg), "utf8")
    .trim()
    .split(/\r?\n/);
const headers = manifestLines.shift().split("\t");
const manifest = manifestLines.map((line) => {
    const values = line.split("\t");
    return Object.fromEntries(
        headers.map((header, index) => [header, values[index] ?? ""]),
    );
});

const duplicateIds = manifest
    .map((entry) => entry.id)
    .filter((id, index, ids) => ids.indexOf(id) !== index);
if (duplicateIds.length > 0) {
    console.error(
        `Duplicate manifest ids: ${[...new Set(duplicateIds)].join(", ")}`,
    );
    process.exit(1);
}

const workflowLines = readFileSync(resolve(workflowArg), "utf8").split(/\r?\n/);
const workflowSteps = [];
let currentJob = "";
let currentStep = null;

function finishStep() {
    if (
        !currentStep ||
        !targetJobs.has(currentJob) ||
        ignoredSteps.has(currentStep.name)
    )
        return;
    const body = currentStep.lines.join("\n");
    if (/^\s+run:/m.test(body) || /golangci\/golangci-lint-action/.test(body)) {
        workflowSteps.push({ job: currentJob, name: currentStep.name, body });
    }
}

for (const line of workflowLines) {
    const jobMatch = line.match(/^  ([a-z0-9-]+):\s*$/);
    if (jobMatch) {
        finishStep();
        currentStep = null;
        currentJob = jobMatch[1];
        continue;
    }
    const stepMatch = line.match(/^      - name:\s*(.+?)\s*$/);
    if (stepMatch) {
        finishStep();
        currentStep = {
            name: stepMatch[1].replace(/^['"]|['"]$/g, ""),
            lines: [line],
        };
        continue;
    }
    if (currentStep) currentStep.lines.push(line);
}
finishStep();

const key = (entry) => `${entry.job}::${entry.workflow_step ?? entry.name}`;
const manifestKeys = new Set(manifest.map(key));
const workflowKeys = new Set(
    workflowSteps.map((step) => `${step.job}::${step.name}`),
);
const missingLocally = [...workflowKeys].filter(
    (entry) => !manifestKeys.has(entry),
);
const missingInWorkflow = [...manifestKeys].filter(
    (entry) => !workflowKeys.has(entry),
);
const commandDrift = [];
const localCommandDrift = [];

for (const entry of manifest) {
    const step = workflowSteps.find(
        (candidate) =>
            candidate.job === entry.job &&
            candidate.name === entry.workflow_step,
    );
    if (!step) continue;
    const normalizedBody = step.body.replace(/\s+/g, " ");
    const fragments = entry.signature.split(" && ");
    if (fragments.some((fragment) => !normalizedBody.includes(fragment))) {
        commandDrift.push(
            `${entry.job}/${entry.workflow_step}: expected ${entry.signature}`,
        );
    }

    if (!entry.command.startsWith("@") && entry.id !== "backend-lint") {
        const normalizedLocal = entry.command
            .replace(/['"]/g, "")
            .replace(/\s+/g, " ")
            .trim();
        if (normalizedLocal !== entry.signature) {
            localCommandDrift.push(
                `${entry.id}: runs ${entry.command}, signature is ${entry.signature}`,
            );
        }
    }
}

if (
    missingLocally.length ||
    missingInWorkflow.length ||
    commandDrift.length ||
    localCommandDrift.length
) {
    console.error("Core-check workflow drift detected.");
    if (missingLocally.length) {
        console.error(`Workflow-only steps: ${missingLocally.join(", ")}`);
    }
    if (missingInWorkflow.length) {
        console.error(`Local-only steps: ${missingInWorkflow.join(", ")}`);
    }
    if (commandDrift.length) {
        console.error(`Command drift: ${commandDrift.join("; ")}`);
    }
    if (localCommandDrift.length) {
        console.error(`Local command drift: ${localCommandDrift.join("; ")}`);
    }
    process.exit(1);
}

console.log(
    `Workflow sync OK: ${manifest.length} local phases match ${workflowSteps.length} CI steps.`,
);
