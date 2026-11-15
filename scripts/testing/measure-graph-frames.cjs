// GRAPH-LIGHT-1: how often and how cheaply the 2D graph draws.
//
// Runs against the isolated test stack (http://127.0.0.1:3002, SKIP_AUTH=false)
// seeded with seed-test-data.ps1, logs in as its test user and measures four
// scenarios of four seconds each: idle, hover over notes near the centre,
// zoom with the wheel, dragging a note in a circle.
//
// The browser runs without a frame-rate limit, so a heavy frame shows up as
// fewer draws per second instead of hiding behind vsync. Per scenario:
//   drawsPerSec   — full redraws of the graph canvas (clearRect of the whole canvas)
//   framesPerSec  — animation frames the page produced
//   drawCostP50/95 — JS time of the frames in which the canvas was drawn, ms
//
// Usage (from the repository root):
//   node scripts/testing/measure-graph-frames.cjs [--all-notes] [classic] [light]
//
// --all-notes: the notes list is capped at 300 by the backend (NOTES-LIMIT-1),
// which also cuts the graph to 300 notes; with this flag the script fetches the
// list page by page and hands the page all of it.
"use strict";

const fs = require("fs");
const path = require("path");
const { createRequire } = require("module");

const root = path.resolve(__dirname, "..", "..");
const { chromium } = createRequire(path.join(root, "frontend", "package.json"))("@playwright/test");

const BASE = "http://127.0.0.1:3002";
const SECONDS = 4;

const argv = process.argv.slice(2);
const allNotes = argv.includes("--all-notes");
const styles = argv.filter((a) => !a.startsWith("--"));
if (!styles.length) styles.push("classic", "light");

const seedScript = fs.readFileSync(path.join(root, "scripts", "testing", "seed-test-data.ps1"), "utf8");
const userBlock = seedScript.slice(seedScript.indexOf("$testUser = @{"));
const login = /login\s*=\s*"([^"]+)"/.exec(userBlock)[1];
const password = /password\s*=\s*"([^"]+)"/.exec(userBlock)[1];

function instrument() {
  const perf = { on: false, frames: new Map(), drawn: new Set(), draws: 0 };
  window.__perf = perf;
  const raf = window.requestAnimationFrame.bind(window);
  let current = null;
  window.requestAnimationFrame = (cb) =>
    raf((ts) => {
      current = ts;
      const t0 = performance.now();
      try {
        cb(ts);
      } finally {
        if (perf.on) perf.frames.set(ts, (perf.frames.get(ts) ?? 0) + performance.now() - t0);
        current = null;
      }
    });
  const clear = CanvasRenderingContext2D.prototype.clearRect;
  CanvasRenderingContext2D.prototype.clearRect = function (x, y, w, h) {
    if (perf.on && this.canvas && this.canvas.dataset.testid === "graph-canvas" && x === 0 && y === 0) {
      perf.draws++;
      if (current !== null) perf.drawn.add(current);
    }
    return clear.call(this, x, y, w, h);
  };
}

function summary() {
  const p = window.__perf;
  const costs = [...p.drawn].map((ts) => p.frames.get(ts) ?? 0).sort((a, b) => a - b);
  const q = (x) => +(costs[Math.min(costs.length - 1, Math.floor(x * costs.length))] ?? 0).toFixed(1);
  return { draws: p.draws, frames: p.frames.size, p50: q(0.5), p95: q(0.95) };
}

async function measure(page, name, act) {
  await page.evaluate(() => {
    const p = window.__perf;
    p.frames.clear();
    p.drawn.clear();
    p.draws = 0;
    p.on = true;
  });
  const t0 = Date.now();
  await act();
  const elapsed = (Date.now() - t0) / 1000;
  await page.evaluate(() => (window.__perf.on = false));
  const s = await page.evaluate(summary);
  return {
    name,
    drawsPerSec: +(s.draws / elapsed).toFixed(1),
    framesPerSec: +(s.frames / elapsed).toFixed(1),
    drawCostP50: s.p50,
    drawCostP95: s.p95,
  };
}

async function notePoints(page) {
  return page.evaluate(() => {
    const g = window.__graphCanvas;
    const r = document.querySelector('[data-testid="graph-canvas"]').getBoundingClientRect();
    return g
      .getSimulationNodes()
      .map((n) => ({ x: r.left + n.x * g.transform.k + g.transform.x, y: r.top + n.y * g.transform.k + g.transform.y }))
      .filter((p) => p.x > r.left + 150 && p.x < r.right - 300 && p.y > r.top + 120 && p.y < r.bottom - 120);
  });
}

async function serveAllNotes(page) {
  await page.route(/\/api\/v1\/notes\?(.*&)?limit=10000/, async (route) => {
    const url = new URL(route.request().url());
    const notes = [];
    for (let offset = 0; offset < 100000; offset += 300) {
      url.searchParams.set("limit", "300");
      url.searchParams.set("offset", String(offset));
      const body = await (await route.fetch({ url: url.toString() })).json();
      notes.push(...body.notes);
      if (body.notes.length < 300) break;
    }
    await route.fulfill({ json: { notes, total: notes.length, limit: notes.length, offset: 0 } });
  });
}

(async () => {
  const browser = await chromium.launch({ args: ["--disable-gpu-vsync", "--disable-frame-rate-limit"] });
  for (const style of styles) {
    const page = await browser.newPage({ viewport: { width: 1600, height: 900 } });
    await page.addInitScript(instrument);
    if (allNotes) await serveAllNotes(page);
    await page.goto(`${BASE}/auth/login`);
    await page.fill('input[placeholder="Enter login"]', login);
    await page.fill('input[type="password"]', password);
    await page.keyboard.press("Enter");
    await page.waitForTimeout(3000);
    await page.goto(`${BASE}/?graphStyle=${style}`);
    await page.locator('[data-testid="graph-canvas"][data-test-stable="true"]').waitFor({ timeout: 180000 });
    await page.waitForTimeout(3000);
    const stats = (await page.locator('[data-testid="graph-stats"]').first().innerText()).replace(/\s+/g, " ");
    const report = (r) => console.log(JSON.stringify({ style, stats, ...r }));

    const points = await notePoints(page);
    const mid = points.reduce((a, p) => ({ x: a.x + p.x / points.length, y: a.y + p.y / points.length }), { x: 0, y: 0 });
    points.sort((a, b) => Math.hypot(a.x - mid.x, a.y - mid.y) - Math.hypot(b.x - mid.x, b.y - mid.y));
    points.splice(40);

    report(await measure(page, "idle", () => page.waitForTimeout(SECONDS * 1000)));
    report(
      await measure(page, "hover", async () => {
        const end = Date.now() + SECONDS * 1000;
        for (let i = 0; Date.now() < end; i++) {
          const p = points[i % points.length];
          await page.mouse.move(p.x, p.y, { steps: 2 });
          await page.waitForTimeout(300);
        }
      })
    );
    await page.mouse.move(5, 5);
    await page.waitForTimeout(800);

    const centre = await page.evaluate(() => {
      const r = document.querySelector('[data-testid="graph-canvas"]').getBoundingClientRect();
      return { x: r.left + r.width * 0.45, y: r.top + r.height * 0.5 };
    });
    report(
      await measure(page, "zoom", async () => {
        await page.mouse.move(centre.x, centre.y);
        const end = Date.now() + SECONDS * 1000;
        for (let i = 0; Date.now() < end; i++) {
          await page.mouse.wheel(0, i % 16 < 8 ? -60 : 60);
          await page.waitForTimeout(16);
        }
      })
    );
    await page.waitForTimeout(1500);

    const fresh = await notePoints(page);
    const grab = fresh[Math.floor(fresh.length / 2)];
    report(
      await measure(page, "drag", async () => {
        await page.mouse.move(grab.x, grab.y);
        await page.waitForTimeout(300);
        await page.mouse.down();
        const end = Date.now() + SECONDS * 1000;
        for (let t = 0; Date.now() < end; t += 0.15) {
          await page.mouse.move(grab.x + Math.sin(t) * 90, grab.y + Math.cos(t) * 90, { steps: 1 });
          await page.waitForTimeout(16);
        }
        await page.mouse.up();
      })
    );
    await page.close();
  }
  await browser.close();
})().catch((e) => {
  console.error("ERR", e.message);
  process.exit(1);
});
