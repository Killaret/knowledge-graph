// UX-1 live probe, review no. 3 (Claude Code, 2026-11-15). Test stack with SKIP_AUTH=false.
// Run from the repo root: NODE_PATH=frontend/node_modules node docs/tasks/UX-1-review-3-probe/ux1-probe.cjs
// UX1_WORKAROUND=1 bypasses blocker 1 (picker closed by its own opening click) to check the rest of the flow.
// 1) "Link to…" from the node context menu on / and /graph creates a link between existing notes;
// 2) deleting a note on / and /graph: does the canvas ever go blank (sampled every 30 ms)?
const { chromium, request } = require("@playwright/test");

const FE = "http://127.0.0.1:3002";
const BE = "http://127.0.0.1:18083";
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const OUT = process.env.PROBE_OUT || ".";

(async () => {
    const api = await request.newContext();
    const login = `ux1p${Date.now()}`;
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
    const mk = async (title) =>
        (
            await (
                await api.post(`${BE}/api/v1/notes`, {
                    headers: H,
                    data: {
                        title,
                        content: `${title} distinct text ${Math.random()}`,
                        type: "star",
                    },
                })
            ).json()
        ).data.id;
    const ids = {};
    for (const t of ["Alpha", "Bravo", "Charlie", "Delta", "Echo"])
        ids[t] = await mk(t);
    const linksOf = async (id) => {
        const j = await (
            await api.get(`${BE}/api/v1/notes/${id}/links`, { headers: H })
        ).json();
        const d = j.data ?? j;
        return [
            ...(d.outgoing ?? []),
            ...(d.incoming ?? []),
            ...(Array.isArray(d) ? d : []),
        ];
    };

    const browser = await chromium.launch();
    const page = await browser.newPage({
        viewport: { width: 1400, height: 900 },
    });
    page.on("pageerror", (e) =>
        console.log("pageerror:", e.message.slice(0, 200)),
    );
    await page.addInitScript((t) => {
        window.__ACCESS_TOKEN__ = t;
    }, token);

    const waitGraph = async (n) => {
        await page.waitForFunction(
            (n) => {
                const s = window.__graphCanvas?.getSimulationNodes?.() ?? [];
                return s.length >= n && s.every((x) => x.x != null);
            },
            n,
            { timeout: 30000 },
        );
        await sleep(6000); // let the layout settle
    };
    const nodeScreen = (title) =>
        page.evaluate((title) => {
            const api = window.__graphCanvas;
            const n = api.getSimulationNodes().find((x) => x.title === title);
            const r = document
                .querySelector('[data-testid="graph-canvas"]')
                .getBoundingClientRect();
            return {
                x: r.left + n.x * api.transform.k + api.transform.x,
                y: r.top + n.y * api.transform.k + api.transform.y,
            };
        }, title);

    async function linkVia(route, from, to) {
        await page.goto(`${FE}${route}`, {
            waitUntil: "networkidle",
            timeout: 60000,
        });
        await waitGraph(2);
        if (process.env.UX1_WORKAROUND) {
            // Review workaround for blocker 1: keep the opening click from reaching `window`,
            // where the picker's own click-outside handler would close it at once.
            await page.evaluate(() =>
                document.addEventListener("click", (e) => {
                    if (
                        e.target.closest?.(
                            '[data-testid="context-menu-link-to"]',
                        )
                    )
                        e.stopPropagation();
                }),
            );
        }
        const p = await nodeScreen(from);
        await page.mouse.click(p.x, p.y, { button: "right" });
        const item = page.locator('[data-testid="context-menu-link-to"]');
        const itemVisible = await item
            .isVisible({ timeout: 3000 })
            .catch(() => false);
        console.log(
            `[${route}] right-click ${from}: "Link to" item visible=${itemVisible}`,
        );
        if (!itemVisible) {
            await page.screenshot({
                path: `${OUT}/ux1-${route.replace(/\W/g, "") || "home"}-menu.png`,
            });
            return;
        }
        await item.click();
        const picker = page.locator('[data-testid="link-target-picker"]');
        await sleep(500);
        const pickerVisible = await picker.isVisible();
        console.log(
            `[${route}] picker visible 500 ms after "Link to"=${pickerVisible}`,
        );
        if (!pickerVisible) return;
        const opts = await page
            .locator('[data-testid="link-target-option"]')
            .allInnerTexts();
        console.log(
            `[${route}] options: ${opts.map((s) => s.replace(/\s+/g, " ").trim()).join(" | ")}`,
        );
        await page.fill('[data-testid="link-target-search"]', to.slice(0, 3));
        const filtered = await page
            .locator('[data-testid="link-target-option"]')
            .allInnerTexts();
        console.log(
            `[${route}] after search "${to.slice(0, 3)}": ${filtered.length} option(s)`,
        );
        await page
            .locator('[data-testid="link-target-option"]', { hasText: to })
            .first()
            .click();
        const form = page.locator('[data-testid="link-form"]');
        console.log(
            `[${route}] link form visible=${await form.isVisible({ timeout: 3000 }).catch(() => false)}`,
        );
        await page.locator('[data-testid="link-form-create"]').click();
        await sleep(2500);
        const links = await linksOf(ids[from]);
        console.log(`[${route}] API: links of ${from} now ${links.length}`);
        await page.screenshot({
            path: `${OUT}/ux1-${route.replace(/\W/g, "") || "home"}-after-link.png`,
        });
    }

    async function deleteVia(route, title) {
        await page.goto(`${FE}${route}`, {
            waitUntil: "networkidle",
            timeout: 60000,
        });
        await waitGraph(2);
        // Sample the canvas while the app processes a deletion made in the app itself.
        await page.evaluate(() => {
            window.__ux1Samples = [];
            window.__ux1Timer = setInterval(() => {
                const c = document.querySelector(
                    '[data-testid="graph-canvas"]',
                );
                const api = window.__graphCanvas;
                window.__ux1Samples.push({
                    t: performance.now(),
                    canvas: !!c,
                    nodes: api?.getSimulationNodes?.().length ?? -1,
                    overlay: !!document.querySelector(
                        '[data-testid="loading-overlay"]',
                    ),
                    chip: !!document.querySelector(
                        '[data-testid="refresh-chip"], .loading-chip',
                    ),
                });
            }, 30);
        });
        // Delete the way a user does on the canvas: select the node, press Delete, confirm.
        const p = await nodeScreen(title);
        await page.mouse.click(p.x, p.y);
        await sleep(300);
        await page.keyboard.press("Delete");
        const confirm = page.locator('[data-testid="confirm-modal-confirm"]');
        const asked = await confirm
            .isVisible({ timeout: 3000 })
            .catch(() => false);
        console.log(`[${route}] confirm dialog shown=${asked}`);
        if (asked) await confirm.click();
        await sleep(5000);
        const s = await page.evaluate(() => {
            clearInterval(window.__ux1Timer);
            return window.__ux1Samples;
        });
        const blank = s.filter((x) => !x.canvas || x.nodes === 0 || x.overlay);
        const nodesSeq = [
            ...new Set(
                s.map(
                    (x) =>
                        `${x.canvas ? "c" : "-"}${x.nodes}${x.overlay ? "O" : ""}${x.chip ? "+chip" : ""}`,
                ),
            ),
        ];
        console.log(
            `[${route}] delete ${title}: samples=${s.length}, blank-or-covered=${blank.length} (~${blank.length * 30} ms), states=${nodesSeq.join(" → ")}`,
        );
        const r = await api.get(`${BE}/api/v1/notes/${ids[title]}`, {
            headers: H,
        });
        console.log(`[${route}] API GET deleted note → ${r.status()}`);
    }

    await linkVia("/", "Alpha", "Bravo");
    await linkVia("/graph", "Charlie", "Delta");
    await deleteVia("/", "Echo");
    await deleteVia("/graph", "Delta");

    await browser.close();
    await api.dispose();
})().catch((e) => {
    console.error(e);
    process.exit(1);
});
