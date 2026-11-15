// SYNC-1 B/C live probe, review no. 2 (Claude Code, 2026-11-15). Test stack with SKIP_AUTH=false.
// Run from the repo root: NODE_PATH=frontend/node_modules node docs/tasks/SYNC-1-review-2-probe/sync1-probe.cjs [waitStableMs]
const { chromium, request } = require("@playwright/test");

const FE = "http://127.0.0.1:3002";
const BE = "http://127.0.0.1:18083";
const WAIT_STABLE = Number(process.argv[2] ?? 12000);
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

(async () => {
    const api = await request.newContext();
    const login = `sync1p${Date.now()}`;
    const password = "ProbePassword123!";
    let r = await api.post(`${BE}/api/v1/auth/register`, {
        data: { login, email: `${login}@example.com`, password },
    });
    console.log("register", r.status());
    r = await api.post(`${BE}/api/v1/auth/login`, {
        data: { login, password },
    });
    const token = (await r.json()).access_token;
    const H = { Authorization: `Bearer ${token}` };

    const mk = async (title) => {
        const res = await api.post(`${BE}/api/v1/notes`, {
            headers: H,
            data: { title, content: `${title} body text`, type: "star" },
        });
        const j = await res.json();
        return j.id ?? j.data?.id;
    };
    const A = await mk("Alpha");
    const B = await mk("Bravo");
    const C = await mk("Charlie");
    r = await api.post(`${BE}/api/v1/links`, {
        headers: H,
        data: {
            source_note_id: A,
            target_note_id: B,
            link_type: "related",
            weight: 0.8,
        },
    });
    console.log("seed link", r.status(), "notes", A, B, C);

    const browser = await chromium.launch();
    const page = await browser.newPage({
        viewport: { width: 1400, height: 900 },
    });
    const reqLog = [];
    page.on("request", (q) => {
        const u = q.url();
        if (/graph\/(events|delta|full|cached|fresh)|me\/graph/.test(u))
            reqLog.push(
                `${new Date().toISOString().slice(11, 23)} ${q.method()} ${u.replace(/access_token=[^&]+/, "access_token=…")}`,
            );
    });
    page.on("response", async (resp) => {
        if (!/graph\/delta/.test(resp.url())) return;
        try {
            const j = await resp.json();
            const c = (k) => (j[k] ?? []).length;
            reqLog.push(
                `    delta → added_nodes ${c("added_nodes")}, updated_nodes ${c("updated_nodes")}, removed_nodes ${c("removed_nodes")}, added_links ${c("added_links")}, removed_links ${c("removed_links")}${j.resync ? ", resync" : ""}${(j.updated_nodes ?? []).map((n) => " upd:" + n.title).join("")}`,
            );
        } catch {}
    });
    page.on("console", (m) => {
        if (m.type() === "error")
            console.log("console.error:", m.text().slice(0, 200));
    });
    await page.addInitScript((t) => {
        window.__ACCESS_TOKEN__ = t;
    }, token);
    await page.goto(`${FE}/`, { waitUntil: "networkidle", timeout: 60000 });
    await page.waitForFunction(
        () => {
            const n = window.__graphCanvas?.getSimulationNodes?.() ?? [];
            return n.length >= 3 && n.every((x) => x.x != null);
        },
        null,
        { timeout: 30000 },
    );
    console.log(`graph up; waiting ${WAIT_STABLE} ms for the layout to settle`);
    await sleep(WAIT_STABLE);

    const snap = () =>
        page.evaluate(() =>
            (window.__graphCanvas.getSimulationNodes() || []).map((n) => ({
                id: n.id,
                title: n.title,
                x: n.x,
                y: n.y,
                tagged: !!n.__probe,
            })),
        );
    const tag = () =>
        page.evaluate(() => {
            for (const n of window.__graphCanvas.getSimulationNodes())
                n.__probe = true;
        });
    const report = (label, before, after) => {
        const b = new Map(before.map((n) => [n.id, n]));
        const rows = after
            .filter((n) => b.has(n.id))
            .map((n) => {
                const o = b.get(n.id);
                return `${n.title}: moved ${Math.hypot(n.x - o.x, n.y - o.y).toFixed(1)} px, same object=${n.tagged}`;
            });
        console.log(
            `[${label}] nodes ${before.length}→${after.length}; ${rows.join("; ")}`,
        );
    };
    const waitFor = async (pred, ms = 15000) => {
        const t0 = Date.now();
        while (Date.now() - t0 < ms) {
            const s = await snap();
            if (pred(s)) return { s, dt: Date.now() - t0 };
            await sleep(200);
        }
        return { s: await snap(), dt: -1 };
    };

    // 1. add a note from "another tab" (API)
    await tag();
    let before = await snap();
    const D = await mk("Delta");
    let res = await waitFor((s) => s.some((n) => n.id === D));
    console.log(`add: appeared after ${res.dt} ms`);
    await sleep(1500);
    report("add", before, await snap());

    // 2. rename (count unchanged)
    await sleep(WAIT_STABLE);
    await tag();
    before = await snap();
    r = await api.put(`${BE}/api/v1/notes/${A}`, {
        headers: H,
        data: { title: "Alpha renamed" },
    });
    console.log("rename", r.status());
    res = await waitFor((s) =>
        s.some((n) => n.id === A && n.title === "Alpha renamed"),
    );
    console.log(`rename: visible in sim after ${res.dt} ms`);
    report("rename", before, await snap());

    // 3. link between existing nodes (count of nodes unchanged)
    await sleep(WAIT_STABLE);
    await tag();
    before = await snap();
    r = await api.post(`${BE}/api/v1/links`, {
        headers: H,
        data: {
            source_note_id: B,
            target_note_id: C,
            link_type: "related",
            weight: 0.8,
        },
    });
    const linkId = (await r.json()).id;
    console.log("link create", r.status());
    await sleep(4000);
    report("link add", before, await snap());

    // 4. delete note
    await sleep(WAIT_STABLE);
    await tag();
    before = await snap();
    r = await api.delete(`${BE}/api/v1/notes/${D}`, { headers: H });
    console.log("delete", r.status());
    res = await waitFor((s) => !s.some((n) => n.id === D));
    console.log(`delete: gone after ${res.dt} ms`);
    await sleep(1500);
    report("delete", before, await snap());

    console.log("--- graph requests ---\n" + reqLog.join("\n"));
    await browser.close();
    await api.dispose();
})().catch((e) => {
    console.error(e);
    process.exit(1);
});
