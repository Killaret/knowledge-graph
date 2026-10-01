import { test, expect, type APIRequestContext } from "@playwright/test";

const FRONTEND_URL = process.env.FRONTEND_URL || "http://127.0.0.1:3002";
const BACKEND_URL = process.env.BACKEND_URL || "http://127.0.0.1:18083";

/**
 * COMET-1 stage F — live test-stack verification.
 * Self-seeding: creates its own comets through the API, then verifies
 * reminder delivery, .ics export, the upcoming list, archive suggestion
 * and captures graph screenshots for MANUAL_TEST_FEEDBACK.md.
 */

async function createComet(
  request: APIRequestContext,
  title: string,
  extra: Record<string, unknown> = {}
): Promise<string> {
  const res = await request.post(`${BACKEND_URL}/api/v1/notes`, {
    data: { title, type: "comet", ...extra },
  });
  expect(res.ok()).toBeTruthy();
  const body = await res.json();
  return body.data.id as string;
}

test.describe("COMET-1 stage F (live stack)", () => {
  test.setTimeout(180_000);

  test("reminder fires, .ics exports, archive suggestion converts to debris", async ({
    page,
    request,
  }) => {
    const now = Date.now();
    const run = now.toString(36); // unique titles — repeat runs leave comets behind
    const iso = (offsetMs: number) => new Date(now + offsetMs).toISOString();

    // Reminder comet: remind_at ≈ now + 10s (due −80s + 70s remind_before).
    const remindId = await createComet(request, `COMET-F live reminder ${run}`, {
      due_at: iso(80_000),
      remind_before_seconds: 70,
    });
    const pastTitle = `COMET-F live past ${run}`;
    const pastId = await createComet(request, pastTitle, { due_at: iso(-86_400_000) });
    await createComet(request, `COMET-F live soon ${run}`, { due_at: iso(6 * 3600_000) });
    await createComet(request, `COMET-F live far ${run}`, { due_at: iso(30 * 86_400_000) });
    await createComet(request, `COMET-F live undated ${run}`);

    // .ics export: VEVENT in UTC + VALARM.
    const ics = await request.get(`${BACKEND_URL}/api/v1/notes/${remindId}/calendar.ics`);
    expect(ics.ok()).toBeTruthy();
    expect(ics.headers()["content-type"]).toContain("text/calendar");
    const icsText = await ics.text();
    expect(icsText).toContain("BEGIN:VEVENT");
    expect(icsText).toContain("TRIGGER:-PT70S");

    // Notification appears once remind_at passes (worker polls; allow 90s).
    let delivered = false;
    for (let i = 0; i < 30 && !delivered; i++) {
      const res = await request.get(`${BACKEND_URL}/api/v1/notifications`);
      const body = await res.json();
      const items = (body.items ?? body.data?.items ?? []) as Array<{
        note_id: string;
        type: string;
      }>;
      delivered = items.some((n) => n.note_id === remindId && n.type === "comet_reminder");
      if (!delivered) await page.waitForTimeout(3_000);
    }
    expect(delivered).toBeTruthy();

    // UI: upcoming list groups, .ics button, archive suggestion.
    await page.goto(`${FRONTEND_URL}/`, { timeout: 60_000, waitUntil: "networkidle" });
    await expect(page.locator(".splash-screen")).toBeHidden({ timeout: 30_000 });

    const panel = page.locator('[data-testid="cockpit-left-panel"]');
    await expect(page.locator('[data-testid="comet-item-overdue"]').first()).toBeVisible({
      timeout: 30_000,
    });
    await expect(page.locator('[data-testid="comet-item-undated"]').first()).toBeVisible();
    await panel.screenshot({ path: "../docs/agents/screenshots/comet-1/upcoming-panel.png" });

    // Graph metaphor: canvas with urgent/far/past/undated comets.
    const canvas = page.locator('[data-testid="graph-canvas"]');
    await expect(canvas).toBeVisible({ timeout: 30_000 });
    await page.waitForTimeout(6_000); // let the force layout settle
    await canvas.screenshot({ path: "../docs/agents/screenshots/comet-1/graph-comets.png" });

    // Past-due comet offers archive; click converts to debris and closes the offer.
    // The overdue item may sit below the panel's internal scroll — dispatch the
    // click directly instead of relying on Playwright's viewport scroll.
    await page
      .locator('[data-testid="comet-item-overdue"]', { hasText: pastTitle })
      .first()
      .dispatchEvent("click");
    const details = page.locator('[data-testid="cockpit-note-details"]');
    const suggestion = page.locator('[data-testid="comet-archive-suggestion"]');
    await expect(suggestion).toBeVisible({ timeout: 15_000 });
    await details.screenshot({ path: "../docs/agents/screenshots/comet-1/archive-suggestion.png" });
    await page.locator('[data-testid="comet-archive-btn"]').click();
    await expect(suggestion).toBeHidden({ timeout: 15_000 });

    // The converted note left the comet groups.
    const comets = await request.get(`${BACKEND_URL}/api/v1/notes/comets`);
    const cometsBody = await comets.json();
    const groups = cometsBody.data ?? cometsBody;
    const allIds = [
      ...(groups.overdue ?? []),
      ...(groups.upcoming ?? []),
      ...(groups.undated ?? []),
    ].map((n: { id: string }) => n.id);
    expect(allIds).not.toContain(pastId);
  });
});
