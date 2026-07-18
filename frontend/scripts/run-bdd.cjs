const { spawn } = require("child_process");
const path = require("path");

const root = path.resolve(__dirname, "..", "..");
const frontend = path.resolve(__dirname, "..");
let devServer;

const isDevServerReady = () => {
    return new Promise((resolve) => {
        const http = require("http");
        const req = http.get("http://localhost:5173/", (res) => {
            resolve(res.statusCode === 200);
        });
        req.on("error", () => resolve(false));
        req.setTimeout(2000, () => {
            req.destroy();
            resolve(false);
        });
    });
};

const startDevServer = () => {
    devServer = spawn("npm", ["run", "dev"], {
        cwd: frontend,
        stdio: "pipe",
        env: { ...process.env, SKIP_AUTH: "true" },
        shell: true,
    });
    devServer.stdout.on("data", (data) => process.stdout.write(data));
    devServer.stderr.on("data", (data) => process.stderr.write(data));
};

const waitForDevServer = async () => {
    if (await isDevServerReady()) return;
    startDevServer();
    const start = Date.now();
    while (Date.now() - start < 120000) {
        if (await isDevServerReady()) return;
        await new Promise((r) => setTimeout(r, 1000));
    }
    throw new Error("Dev server did not start in 120s");
};

const runCucumber = () => {
    return new Promise((resolve, reject) => {
        const args = [
            "--import",
            "./frontend/node_modules/tsx/dist/loader.mjs",
            "./frontend/node_modules/@cucumber/cucumber/bin/cucumber.js",
            "--config",
            "./cucumber.mjs",
            ...process.argv.slice(2),
        ];
        const cucumber = spawn("node", args, {
            cwd: root,
            stdio: "inherit",
            shell: true,
        });
        cucumber.on("exit", (code) => resolve(code));
        cucumber.on("error", reject);
    });
};

(async () => {
    try {
        console.log("[BDD] Ensuring dev server is ready...");
        await waitForDevServer();
        console.log("[BDD] Dev server ready");
        const exitCode = await runCucumber();
        process.exit(exitCode ?? 0);
    } catch (e) {
        console.error("[BDD] Error:", e);
        process.exit(1);
    } finally {
        if (devServer) devServer.kill();
    }
})();
