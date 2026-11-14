import { test, expect } from "@playwright/test";
import { loginOrCreateTestUser } from "./helpers/auth";

// SMART-CREATE-1 (review blocker 1): a document submit must reach the backend
// through the SvelteKit server on baseURL (test stack :3002). The first
// implementation sent multipart/form-data, which SvelteKit's CSRF check
// rejected with 403 for every Origin because the frontend container sets no
// ORIGIN — the request never reached the API. The client now sends a raw
// application/octet-stream body (not a form content type), so the request
// passes the check on any origin. A regression back to a form content type
// turns this spec red with a 403.
test.describe("Smart create document upload @auth-real", () => {
  test("POST /api/v1/import/document passes the SvelteKit proxy to the backend", async ({
    request,
  }) => {
    const token = await loginOrCreateTestUser(request);

    // Relative URL resolves against use.baseURL (FRONTEND_URL, :3002 on the
    // test stack) — this is the SvelteKit proxy path a browser takes.
    const res = await request.post("/api/v1/import/document?type=TEXT", {
      data: "e2e smart-create body",
      headers: {
        authorization: `Bearer ${token}`,
        "content-type": "application/octet-stream",
      },
    });

    // 403 would mean the request died at SvelteKit CSRF before the backend.
    // The test stack runs no Java listener, so a reached backend answers 503
    // with the contract code.
    expect(res.status()).toBe(503);
    const body = await res.json();
    expect(body.code).toBe("source_text_handler_unavailable");
  });
});
