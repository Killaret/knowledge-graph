import { test, expect, type Page, type Locator } from "@playwright/test";
import { loginAsTestUser } from "./helpers/auth";

test.describe.configure({ mode: "serial" });

interface SimNode {
  id: string;
  title: string;
  type: string;
  x?: number;
  y?: number;
}

interface GraphCanvasExposed {
  getSimulationNodes: () => SimNode[];
  transform: { x: number; y: number; k: number };
}

async function gotoGraph(page: Page) {
  // full=1 enables the full graph, nocache=1 bypasses the graph-service cache
  // so stats reflect the current DB state across test runs
  await page.goto("/graph?nocache=1&full=1", {
    timeout: 60000,
    waitUntil: "domcontentloaded",
  });
  await page.waitForLoadState("networkidle");
}

async function waitForGraphCanvas(page: Page): Promise<Locator> {
  const canvas = page.locator('[data-testid="graph-canvas"]');
  await expect(canvas).toBeVisible({ timeout: 20000 });
  const stats = page.locator('[data-testid="graph-stats"]');
  await expect(stats).toBeVisible({ timeout: 20000 });
  // Wait for graph data to actually load (non-zero node count in English or Russian)
  await expect(stats).toContainText(/[1-9]\d*\s*(?:nodes?|уз(?:лов|ел|ла|ьев)?)/i, {
    timeout: 20000,
  });
  // Wait for simulation to assign coordinates
  await page.waitForFunction(
    () => {
      const api = (window as any).__graphCanvas as GraphCanvasExposed | undefined;
      if (!api) return false;
      const nodes = api.getSimulationNodes();
      return nodes.length > 0 && nodes.some((n) => n.x != null && n.y != null);
    },
    { timeout: 20000 }
  );
  return canvas;
}

async function getNodePositions(
  page: Page,
  filterNonTechnical = true
): Promise<{
  nodes: SimNode[];
  transform: { x: number; y: number; k: number };
  box: { x: number; y: number };
}> {
  const box = await page.locator('[data-testid="graph-canvas"]').boundingBox();
  if (!box) throw new Error("Canvas bounding box not found");
  const { nodes, transform } = await page.evaluate(() => {
    const api = (window as any).__graphCanvas as GraphCanvasExposed;
    return { nodes: api.getSimulationNodes(), transform: api.transform };
  });
  return {
    nodes: filterNonTechnical ? nodes.filter((n) => n.type !== "technical") : nodes,
    transform,
    box,
  };
}

function toScreen(
  node: SimNode,
  transform: { x: number; y: number; k: number },
  box: { x: number; y: number }
) {
  return {
    x: box.x + (node.x ?? 0) * transform.k + transform.x,
    y: box.y + (node.y ?? 0) * transform.k + transform.y,
  };
}

async function clickNodeByIndex(page: Page, index: number) {
  await page.evaluate((idx: number) => {
    const api = (window as any).__graphCanvas as GraphCanvasExposed;
    const nodes = api.getSimulationNodes().filter((n: SimNode) => n.type !== "technical");
    const node = nodes[idx];
    if (!node) throw new Error(`Node at index ${idx} not found`);
    const canvas = document.querySelector('[data-testid="graph-canvas"]') as HTMLCanvasElement;
    const rect = canvas.getBoundingClientRect();
    const x = rect.left + (node.x ?? 0) * api.transform.k + api.transform.x;
    const y = rect.top + (node.y ?? 0) * api.transform.k + api.transform.y;
    canvas.dispatchEvent(
      new MouseEvent("click", {
        bubbles: true,
        clientX: x,
        clientY: y,
        button: 0,
      })
    );
  }, index);
}

async function getStatsCount(page: Page): Promise<{ nodes: number; links: number }> {
  const stats = page.locator('[data-testid="graph-stats"]');
  const text = (await stats.textContent()) || "";
  const nodesMatch = text.match(/(\d+)\s*(?:nodes?|уз(?:лов|ел|ла|ьев)?)/i);
  const linksMatch = text.match(/(\d+)\s*(?:links?|связ(?:ей|и|ь)?)/i);
  return {
    nodes: nodesMatch ? parseInt(nodesMatch[1], 10) : 0,
    links: linksMatch ? parseInt(linksMatch[1], 10) : 0,
  };
}

test.describe("Section 3 - Canvas Features", { tag: ["@manual", "@canvas", "@auth-real"] }, () => {
  test.beforeEach(async ({ page, request }) => {
    await loginAsTestUser(page, request);
    await gotoGraph(page);
    await waitForGraphCanvas(page);
  });

  test("canvas loads with nodes and links", async ({ page }) => {
    const stats = await getStatsCount(page);
    expect(stats.nodes).toBeGreaterThan(0);
    expect(stats.links).toBeGreaterThanOrEqual(0);
  });

  test("hotkeys open and close help, ghost form, and search", async ({ page }) => {
    // Help modal: ? then Esc
    await page.keyboard.press("Shift+Slash");
    await expect(page.locator('[data-testid="help-modal"]')).toBeVisible({
      timeout: 5000,
    });
    await page.keyboard.press("Escape");
    await expect(page.locator('[data-testid="help-modal"]')).toBeHidden({
      timeout: 5000,
    });

    // Ghost note form: N then Esc
    await page.keyboard.press("n");
    await expect(page.locator('[data-testid="ghost-note-form"]')).toBeVisible({ timeout: 5000 });
    await page.keyboard.press("Escape");
    await expect(page.locator('[data-testid="ghost-note-form"]')).toBeHidden({
      timeout: 5000,
    });

    // Search box: F then Esc
    await page.keyboard.press("f");
    await expect(page.locator('[data-testid="search-box"]')).toBeVisible({
      timeout: 5000,
    });
    await page.keyboard.press("Escape");
    await expect(page.locator('[data-testid="search-box"]')).toBeHidden({
      timeout: 5000,
    });
  });

  test("ghost node creation adds a node to the graph", async ({ page }) => {
    const beforeStats = await getStatsCount(page);
    const timestamp = Date.now();

    await page.keyboard.press("n");
    await expect(page.locator('[data-testid="ghost-note-form"]')).toBeVisible({ timeout: 5000 });

    await page.locator('[data-testid="ghost-note-title"]').fill(`Ghost Test ${timestamp}`);
    await page.locator('[data-testid="ghost-note-content"]').fill("Created via canvas hotkey");
    await page.locator('[data-testid="ghost-note-form"] [data-type="comet"]').click();

    const [response] = await Promise.all([
      page.waitForResponse(
        (res) => res.url().includes("/api/v1/notes") && res.request().method() === "POST"
      ),
      page.locator('[data-testid="ghost-note-create"]').click(),
    ]);
    expect(response.ok()).toBe(true);

    // After creation the graph should reload; wait for stats to reflect the new node
    await expect(async () => {
      const after = await getStatsCount(page);
      expect(after.nodes).toBe(beforeStats.nodes + 1);
    }).toPass({ timeout: 15000, intervals: [500] });
  });

  test("clicking a node opens the side panel", async ({ page }) => {
    await clickNodeByIndex(page, 0);
    await expect(page.locator('[data-testid="cockpit-right-panel"]')).toBeVisible({
      timeout: 10000,
    });
  });

  test("Delete key removes a selected node", async ({ page }) => {
    const beforeStats = await getStatsCount(page);
    await clickNodeByIndex(page, 0);
    await expect(page.locator('[data-testid="cockpit-right-panel"]')).toBeVisible({
      timeout: 10000,
    });

    const [response] = await Promise.all([
      page.waitForResponse(
        (res) => res.url().includes("/api/v1/notes/") && res.request().method() === "DELETE"
      ),
      (async () => {
        await page.keyboard.press("Delete");
        await page.locator('[data-testid="confirm-modal-confirm"]').click();
      })(),
    ]);
    expect(response.ok()).toBe(true);

    await expect(async () => {
      const after = await getStatsCount(page);
      expect(after.nodes).toBe(beforeStats.nodes - 1);
    }).toPass({ timeout: 15000, intervals: [500] });
  });

  test("drag-and-drop between nodes opens link form and creates a link", async ({ page }) => {
    const beforeStats = await getStatsCount(page);
    const { nodes, transform, box } = await getNodePositions(page);
    expect(nodes.length).toBeGreaterThan(1);

    const source = nodes[1];
    const target = nodes[0];
    const sourcePos = toScreen(source, transform, box);
    const targetPos = toScreen(target, transform, box);

    await page.mouse.move(sourcePos.x, sourcePos.y);
    await page.mouse.down();
    await page.mouse.move(targetPos.x, targetPos.y, { steps: 10 });
    await page.mouse.up();

    await expect(page.locator('[data-testid="link-form"]')).toBeVisible({
      timeout: 10000,
    });

    const [response] = await Promise.all([
      page.waitForResponse(
        (res) => res.url().includes("/api/v1/links") && res.request().method() === "POST"
      ),
      page.locator('[data-testid="link-form-create"]').click(),
    ]);
    expect(response.ok()).toBe(true);

    await expect(async () => {
      const after = await getStatsCount(page);
      expect(after.links).toBe(beforeStats.links + 1);
    }).toPass({ timeout: 15000, intervals: [500] });
  });
});
