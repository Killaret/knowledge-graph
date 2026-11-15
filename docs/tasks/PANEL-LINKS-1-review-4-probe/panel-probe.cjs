// PANEL-LINKS-1 live probe, review no. 4 (Claude Code, 2026-11-15). Test stack with SKIP_AUTH=false.
// Run from a checkout with frontend/node_modules: NODE_PATH=frontend/node_modules node <path>/panel-probe.cjs
// Note panel: the link list renders (not "Links (undefined)"); a manual link is deleted without a warning;
// an auto link (gamma) shows the canvas warning, Cancel keeps it, Confirm deletes it.
const { chromium, request } = require("@playwright/test");

const FE = "http://127.0.0.1:3002";
const BE = "http://127.0.0.1:18083";
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

(async () => {
    const api = await request.newContext();
    const login = `panel${Date.now()}`;
    const password = "ProbePassword123!";
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
    const linksOf = async (id) => {
        const d = (
            await (
                await api.get(`${BE}/api/v1/notes/${id}/links`, { headers: H })
            ).json()
        ).data;
        return Array.isArray(d)
            ? d
            : [...(d.outgoing ?? []), ...(d.incoming ?? [])];
    };

    const body =
        "Kubernetes pods are scheduled onto nodes by the kube-scheduler, which scores nodes by resource requests, affinity rules and taints.";
    const K1 = await mk("Kubernetes scheduling", body);
    const K2 = await mk(
        "How kube-scheduler places pods",
        body + " Scheduling considers affinity and taints.",
    );
    const M = await mk(
        "Manual partner",
        "an unrelated note about gardening tomatoes",
    );
    await api.post(`${BE}/api/v1/links`, {
        headers: H,
        data: {
            source_note_id: K1,
            target_note_id: M,
            link_type: "related",
            weight: 0.7,
        },
    });
    // wait for the worker's auto link K1–K2
    let gamma;
    for (let i = 0; i < 40 && !gamma; i++) {
        await sleep(500);
        gamma = (await linksOf(K1)).find((l) => l.source_type === "gamma");
    }
    const before = await linksOf(K1);
    console.log(
        `API links of K1: ${before.map((l) => `${l.source_type}:${l.link_type}${l.metadata?.gamma ? "+gammaMeta" : ""}`).join(", ")}`,
    );
    if (!gamma) console.log("no gamma link appeared in 20 s");

    const browser = await chromium.launch();
    const page = await browser.newPage({
        viewport: { width: 1500, height: 950 },
    });
    page.on("pageerror", (e) =>
        console.log("pageerror:", e.message.slice(0, 200)),
    );
    await page.addInitScript((t) => {
        window.__ACCESS_TOKEN__ = t;
    }, token);
    await page.goto(`${FE}/`, { waitUntil: "networkidle", timeout: 60000 });
    await page.waitForFunction(
        () => (window.__graphCanvas?.getSimulationNodes?.() ?? []).length >= 3,
        null,
        {
            timeout: 30000,
        },
    );
    await sleep(5000);
    const p = await page.evaluate(() => {
        const a = window.__graphCanvas;
        const n = a
            .getSimulationNodes()
            .find((x) => x.title === "Kubernetes scheduling");
        const r = document
            .querySelector('[data-testid="graph-canvas"]')
            .getBoundingClientRect();
        return {
            x: r.left + n.x * a.transform.k + a.transform.x,
            y: r.top + n.y * a.transform.k + a.transform.y,
        };
    });
    const hole = await page.evaluate(() => {
        const h = window.__graphCanvas.getBlackHole?.();
        const r = document
            .querySelector('[data-testid="graph-canvas"]')
            .getBoundingClientRect();
        return h ? { x: r.left + h.x, y: r.top + h.y, radius: h.radius } : null;
    });
    console.log(
        `click at node (${p.x.toFixed(0)}, ${p.y.toFixed(0)}); black hole at (${hole?.x.toFixed(0)}, ${hole?.y.toFixed(0)}) r=${hole?.radius.toFixed(0)}, distance=${hole ? Math.hypot(p.x - hole.x, p.y - hole.y).toFixed(0) : "?"}`,
    );
    await page.mouse.click(p.x, p.y);
    const header = page.locator(".links-section h3");
    await header.waitFor({ timeout: 10000 });
    await sleep(1500);
    const rows = page.locator(".links-section .link-item");
    console.log(
        `panel header: "${(await header.innerText()).trim()}", rows: ${await rows.count()}`,
    );
    for (let i = 0; i < (await rows.count()); i++)
        console.log(
            `  row ${i}: ${(await rows.nth(i).innerText()).replace(/\s+/g, " ").trim().slice(0, 90)}`,
        );

    if (process.env.PROBE_SHOT)
        await page.screenshot({ path: process.env.PROBE_SHOT });
    const openDialogs = await page.locator(".modal-overlay").count();
    console.log(
        `open modal overlays before any delete: ${openDialogs}; ${(await page.locator('[role="dialog"]').allInnerTexts()).map((t) => t.replace(/\s+/g, " ").slice(0, 120)).join(" || ")}`,
    );
    const modal = page.locator('[data-testid="confirm-modal-confirm"]');
    // Rows do not name the partner note; the auto link carries the "Recommended" badge.
    const manualRow = () =>
        rows.filter({ hasNotText: /Recommended|Рекоменд/ }).first();
    const gammaRow = () =>
        rows.filter({ hasText: /Recommended|Рекоменд/ }).first();
    const del = (row) =>
        row.getByRole("button", { name: /Delete link|Удалить связь/ });

    // 1) manual link: deleted at once, no warning
    await del(manualRow()).click();
    const warnedManual = await modal
        .isVisible({ timeout: 1500 })
        .catch(() => false);
    await sleep(1000);
    console.log(
        `manual: warning shown=${warnedManual}; API K1–M link left=${(await linksOf(K1)).some((l) => [l.source_note_id, l.target_note_id].includes(M))}`,
    );

    // 2) auto link: warning, cancel keeps it, confirm deletes it
    if (gamma) {
        const gammaRows = await rows
            .filter({ hasText: /Recommended|Рекоменд/ })
            .count();
        console.log(
            `gamma rows in panel: ${gammaRows}; API gamma links of K1: ${(await linksOf(K1)).filter((l) => l.source_type === "gamma").length}`,
        );
        await del(gammaRow()).click();
        const warned = await modal
            .isVisible({ timeout: 3000 })
            .catch(() => false);
        const text = warned
            ? (
                  await page.locator('[role="dialog"]').last().innerText()
              ).replace(/\s+/g, " ")
            : "";
        console.log(
            `gamma: warning shown=${warned}; text: ${text.slice(0, 160)}`,
        );
        await page.locator('[data-testid="confirm-modal-cancel"]').click();
        await sleep(800);
        const g1 = (await linksOf(K1)).filter(
            (l) => l.source_type === "gamma",
        ).length;
        console.log(`gamma after Cancel: gamma links via API=${g1}`);
        await del(gammaRow()).click();
        await modal.click();
        await sleep(1500);
        console.log(
            `gamma after Confirm: gamma links via API=${(await linksOf(K1)).filter((l) => l.source_type === "gamma").length}`,
        );
        console.log(`panel header now: "${(await header.innerText()).trim()}"`);
    }
    console.log(`ids: K1=${K1} K2=${K2}`);
    await browser.close();
    await api.dispose();
})().catch((e) => {
    console.error(e);
    process.exit(1);
});
