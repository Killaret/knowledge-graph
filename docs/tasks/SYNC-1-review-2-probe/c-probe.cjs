// SYNC-1 C live probe, review no. 2: worker autolink delivery and SSE reconnect after a graph-service restart.
// Test stack with SKIP_AUTH=false. Run from the repo root: NODE_PATH=frontend/node_modules node docs/tasks/SYNC-1-review-2-probe/c-probe.cjs
const { chromium, request } = require("@playwright/test");
const { execSync } = require("child_process");
const FE = "http://127.0.0.1:3002",
    BE = "http://127.0.0.1:18083";
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const ts = () => new Date().toISOString().slice(11, 23);
(async () => {
    const api = await request.newContext();
    const login = `sync1c${Date.now()}`,
        password = "ProbePassword123!";
    await api.post(`${BE}/api/v1/auth/register`, {
        data: { login, email: `${login}@example.com`, password },
    });
    const token = (
        await (
            await api.post(`${BE}/api/v1/auth/login`, {
                data: { login, password },
            })
        ).json()
    ).access_token;
    const H = { Authorization: `Bearer ${token}` };
    const mk = async (title, content) =>
        (
            await (
                await api.post(`${BE}/api/v1/notes`, {
                    headers: H,
                    data: { title, content, type: "star" },
                })
            ).json()
        ).data.id;
    await mk("Seed", "unrelated seed note about gardening tomatoes");
    const browser = await chromium.launch();
    const page = await browser.newPage();
    const log = [];
    page.on("request", (q) => {
        if (/graph\/(events|delta)/.test(q.url()))
            log.push(`${ts()} REQ ${q.url().replace(/\?.*/, "")}`);
    });
    page.on("requestfailed", (q) => {
        if (/graph\/events/.test(q.url()))
            log.push(`${ts()} FAILED events ${q.failure()?.errorText}`);
    });
    page.on("response", async (resp) => {
        if (!/graph\/delta/.test(resp.url())) return;
        try {
            const j = await resp.json();
            const g = (j.added_links ?? []).map(
                (l) => `${l.source_type ?? "?"}:${l.link_type}`,
            );
            if (g.length || (j.added_nodes ?? []).length)
                log.push(
                    `${ts()}   delta added_nodes=${(j.added_nodes ?? []).length} added_links=[${g}]`,
                );
        } catch {}
    });
    await page.addInitScript((t) => {
        window.__ACCESS_TOKEN__ = t;
    }, token);
    await page.goto(`${FE}/`, { waitUntil: "networkidle" });
    await page.waitForFunction(
        () => (window.__graphCanvas?.getSimulationNodes?.() ?? []).length >= 1,
    );
    await sleep(3000);
    log.push(`${ts()} --- create two near-duplicate notes via API`);
    const body =
        "Kubernetes pods are scheduled onto nodes by the kube-scheduler, which scores nodes by resource requests, affinity rules and taints.";
    await mk("Kubernetes scheduling", body);
    await mk(
        "How kube-scheduler places pods",
        body + " Scheduling decisions consider affinity and taints.",
    );
    await sleep(25000);
    log.push(`${ts()} --- restart graph-service container`);
    execSync("docker restart kg-test-graph-service", { stdio: "ignore" });
    log.push(`${ts()} --- graph-service restarted`);
    await sleep(12000);
    log.push(`${ts()} --- create a note after restart`);
    await mk("After restart", "a note created after the SSE reconnect");
    await sleep(6000);
    console.log(log.join("\n"));
    await browser.close();
    await api.dispose();
})().catch((e) => {
    console.error(e);
    process.exit(1);
});
