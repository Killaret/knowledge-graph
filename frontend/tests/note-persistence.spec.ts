import { test, expect } from "@playwright/test";
import { execSync } from "node:child_process";

const BACKEND_URL = process.env.BACKEND_URL || "http://127.0.0.1:18083";
const DB_CONTAINER = process.env.DB_CONTAINER || "kg-test-postgres";
const DB_NAME = process.env.DB_NAME || "knowledge_test";
const DB_USER = process.env.DB_USER || "kb_user";

interface LoginResponse {
  access_token: string;
  refresh_token: string;
}

interface CreateNoteResponse {
  data?: {
    id: string;
    title: string;
    content: string;
    type: string;
    created_at: string;
  };
  id?: string;
  title?: string;
}

test.describe("Note persistence", () => {
  test("note created via API is persisted in DB", async ({ request }) => {
    test.setTimeout(60000);
    const timestamp = Date.now();
    const title = `Persistence Test Note ${timestamp}`;

    // Use skip-auth for the test stack; no real login required.
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };

    if (process.env.SKIP_AUTH !== "true") {
      // 1. Login
      const loginResp = await request.post(`${BACKEND_URL}/api/v1/auth/login`, {
        data: { login: "testuser", password: "TestPassword123!" },
        headers: { "Content-Type": "application/json" },
      });
      expect(loginResp.ok()).toBe(true);
      const loginJson = (await loginResp.json()) as LoginResponse;
      const accessToken = loginJson.access_token;
      expect(accessToken).toBeTruthy();
      headers.Authorization = `Bearer ${accessToken}`;
    }

    // 2. Create note via API
    const createResp = await request.post(`${BACKEND_URL}/api/v1/notes`, {
      data: { title, content: "persistence check", type: "star", metadata: {} },
      headers,
    });
    expect(createResp.ok()).toBe(true);
    const createJson = (await createResp.json()) as CreateNoteResponse;
    const noteId = createJson.data?.id ?? createJson.id;
    const noteTitle = createJson.data?.title ?? createJson.title;
    expect(noteId).toBeTruthy();
    expect(noteTitle).toBe(title);

    // 3. Verify it can be fetched back
    const getHeaders: Record<string, string> = {};
    if (process.env.SKIP_AUTH !== "true" && headers.Authorization) {
      getHeaders.Authorization = headers.Authorization;
    }
    const getResp = await request.get(`${BACKEND_URL}/api/v1/notes/${noteId}`, {
      headers: getHeaders,
    });
    expect(getResp.ok()).toBe(true);
    const getJson = (await getResp.json()) as { data?: { title: string } };
    expect(getJson.data?.title ?? (getJson as any).title).toBe(title);

    // 4. Verify it exists in the database (works for test or personal stack via env vars)
    const dbQuery = `SELECT title FROM notes WHERE id = '${noteId}'`;
    const dbResult = execSync(
      `docker exec ${DB_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -t -A -c "${dbQuery}"`,
      { encoding: "utf-8", timeout: 15000 },
    ).trim();
    expect(dbResult).toBe(title);
  });
});
