import { test, expect, type APIRequestContext } from "@playwright/test";

const FRONTEND_URL = process.env.FRONTEND_URL || "http://127.0.0.1:3002";
const BACKEND_URL = process.env.BACKEND_URL || "http://127.0.0.1:18083";
const GRAPH_URL = process.env.GRAPH_SERVICE_URL || "http://127.0.0.1:29091";

/**
 * PROMISES-1 — сквозные тесты обещаний пользователю.
 * Каждый тест несёт @promise и проверяет обещание из docs/product/USER_PROMISES.md
 * целиком на живом тест-стеке (self-seeding через API). Сторож
 * scripts/testing/check-promises.mjs сверяет каталог с этими тестами.
 */

interface GraphPayload {
  nodes: Array<{ id: string; title: string; type: string }>;
  links: Array<{
    source: string;
    target: string;
    link_type?: string;
    gamma_origin?: boolean;
  }>;
}

async function createNote(
  request: APIRequestContext,
  title: string,
  content = ""
): Promise<string> {
  const res = await request.post(`${BACKEND_URL}/api/v1/notes`, {
    data: { title, content },
  });
  expect(res.ok()).toBeTruthy();
  return (await res.json()).data.id as string;
}

async function graphFull(request: APIRequestContext): Promise<GraphPayload> {
  const res = await request.get(`${GRAPH_URL}/api/v1/graph/full?limit=0&nocache=true`);
  expect(res.ok()).toBeTruthy();
  const body = await res.json();
  return (body.data ?? body) as GraphPayload;
}

async function pollUntil(
  fn: () => Promise<boolean>,
  timeoutMs: number,
  stepMs = 2_000
): Promise<boolean> {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    try {
      if (await fn()) return true;
    } catch {
      /* transient — retry */
    }
    await new Promise((r) => setTimeout(r, stepMs));
  }
  return false;
}

async function noteTitles(request: APIRequestContext): Promise<string[]> {
  const res = await request.get(`${BACKEND_URL}/api/v1/notes?limit=200`);
  expect(res.ok()).toBeTruthy();
  const body = await res.json();
  const items = body.notes ?? body.data?.items ?? body.data ?? body.items ?? [];
  return items.map((n: { title: string }) => n.title);
}

test.describe("USER_PROMISES @promise", () => {
  test.setTimeout(240_000);

  test("P-01: a manually created note appears in the list and on the graph", async ({
    request,
  }) => {
    const title = `P-01 promise ${Date.now().toString(36)}`;
    const id = await createNote(request, title);

    expect(await noteTitles(request)).toContain(title);
    const onGraph = await pollUntil(
      async () => (await graphFull(request)).nodes.some((n) => n.id === id),
      20_000
    );
    expect(onGraph).toBeTruthy();
  });

  test("P-03/P-08: auto-link appears without reload; manual link keeps gamma provenance", async ({
    request,
  }) => {
    const run = Date.now().toString(36);
    const content = `Promise chain content ${run} — two near-identical notes must attract an auto link`;
    const a = await createNote(request, `P-08 A ${run}`, content);
    const b = await createNote(request, `P-08 B ${run}`, content + " (twin)");

    // P-03: the NLP pipeline adds a gamma link between the pair.
    const autoAppeared = await pollUntil(
      async () => {
        const g = await graphFull(request);
        return g.links.some(
          (l) =>
            ((l.source === a && l.target === b) || (l.source === b && l.target === a)) &&
            l.gamma_origin === true
        );
      },
      120_000,
      5_000
    );
    expect(autoAppeared).toBeTruthy();

    // P-08: a manual link over the auto pair keeps its origin recorded.
    const res = await request.post(`${BACKEND_URL}/api/v1/links`, {
      data: { source_note_id: a, target_note_id: b, link_type: "related" },
    });
    expect(res.ok()).toBeTruthy();

    const stillRecorded = await pollUntil(async () => {
      const g = await graphFull(request);
      return g.links.some(
        (l) =>
          ((l.source === a && l.target === b) || (l.source === b && l.target === a)) &&
          l.gamma_origin === true
      );
    }, 15_000);
    expect(stillRecorded).toBeTruthy();
  });

  test("P-04: a rename is visible on the graph without reload", async ({ request }) => {
    const run = Date.now().toString(36);
    const id = await createNote(request, `P-04 old ${run}`);
    const newTitle = `P-04 renamed ${run}`;
    const res = await request.put(`${BACKEND_URL}/api/v1/notes/${id}`, {
      data: { title: newTitle },
    });
    expect(res.ok()).toBeTruthy();

    const renamed = await pollUntil(
      async () => (await graphFull(request)).nodes.some((n) => n.id === id && n.title === newTitle),
      20_000
    );
    expect(renamed).toBeTruthy();
  });

  test("P-05/P-11: deleted note leaves the graph; restore brings it back", async ({ request }) => {
    const title = `P-05 promise ${Date.now().toString(36)}`;
    const id = await createNote(request, title);
    await pollUntil(async () => (await graphFull(request)).nodes.some((n) => n.id === id), 20_000);

    expect((await request.delete(`${BACKEND_URL}/api/v1/notes/${id}`)).ok()).toBeTruthy();
    const gone = await pollUntil(
      async () => !(await graphFull(request)).nodes.some((n) => n.id === id),
      20_000
    );
    expect(gone).toBeTruthy();
    expect(await noteTitles(request)).not.toContain(title);

    // P-11: restore within the retention window brings it back.
    expect((await request.post(`${BACKEND_URL}/api/v1/notes/${id}/restore`)).ok()).toBeTruthy();
    const back = await pollUntil(
      async () => (await graphFull(request)).nodes.some((n) => n.id === id),
      20_000
    );
    expect(back).toBeTruthy();
    expect(await noteTitles(request)).toContain(title);
  });

  test("P-06: the note list works while the graph service is unreachable", async ({
    page,
    request,
  }) => {
    const title = `P-06 promise ${Date.now().toString(36)}`;
    await createNote(request, title);

    // Block every graph-service call — the UI must still serve the notes.
    await page.route("**/graph-service/**", (route) => route.abort());
    await page.goto(`${FRONTEND_URL}/`, { timeout: 60_000, waitUntil: "domcontentloaded" });
    await expect(page.locator('[data-testid="view-toggle-list"]')).toBeVisible({
      timeout: 30_000,
    });
    await page.locator('[data-testid="view-toggle-list"]').click();
    await expect(page.locator("body")).toContainText(title, { timeout: 30_000 });
  });

  test("P-07: search finds a note by a title word and by a content word", async ({ request }) => {
    const run = Date.now().toString(36);
    const titleWord = `promisetitle${run}`;
    const contentWord = `promisecontent${run}`;
    await createNote(request, `${titleWord} heading`, `body mentions ${contentWord} here`);

    for (const q of [titleWord, contentWord]) {
      const res = await request.get(
        `${BACKEND_URL}/api/v1/notes/search?q=${encodeURIComponent(q)}`
      );
      expect(res.ok()).toBeTruthy();
      const body = await res.json();
      const items = body.data?.items ?? body.data ?? [];
      expect(
        items.some((n: { title: string }) => n.title.includes(titleWord)),
        `search for ${q}`
      ).toBeTruthy();
    }
  });

  test("P-10: with nlp.quality disabled the feature reports itself off", async ({ request }) => {
    const id = await createNote(request, `P-10 promise ${Date.now().toString(36)}`);
    const res = await request.get(`${BACKEND_URL}/api/v1/notes/${id}/quality`);
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.enabled ?? body.data?.enabled).toBe(false);
  });
});
