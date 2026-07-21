import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { api } from "./client";
import {
  getApiKey,
  saveTokens,
  clearAuthState,
} from "$shared/stores/auth-session.svelte";
import { goto } from "$app/navigation";

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

vi.mock("$shared/stores/auth-session.svelte", () => ({
  getApiKey: vi.fn(() => null),
  saveTokens: vi.fn(),
  clearAuthState: vi.fn(),
}));

function createJsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function getRequestUrl(input: RequestInfo | URL): string {
  if (input instanceof Request) return input.url;
  if (input instanceof URL) return input.toString();
  return input;
}

describe("API Client", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.clearAllMocks();
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("adds X-API-Key header when api key is set", async () => {
    vi.mocked(getApiKey).mockReturnValue("test-api-key");
    fetchMock.mockResolvedValue(createJsonResponse({ ok: true }));

    await api.get("test").text();

    const request = fetchMock.mock.calls[0][0] as Request;
    expect(request.headers.get("X-API-Key")).toBe("test-api-key");
  });

  it("refreshes token on 401 and retries the original request", async () => {
    let callCount = 0;
    fetchMock.mockImplementation(async (input: RequestInfo | URL) => {
      const url = getRequestUrl(input);
      const path = new URL(url).pathname;
      callCount++;

      if (path === "/api/test") {
        if (callCount === 1) {
          return createJsonResponse({ error: "Unauthorized" }, 401);
        }
        return createJsonResponse({ ok: true });
      }

      if (path === "/api/v1/auth/refresh") {
        return createJsonResponse({
          access_token: "new-access",
          refresh_token: "new-refresh",
          token_type: "Bearer",
          expires_at: "2025-01-01T00:00:00Z",
        });
      }

      return createJsonResponse({ error: "Not found" }, 404);
    });

    const response = await api.get("test");

    expect(response.ok).toBe(true);
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(saveTokens).toHaveBeenCalledWith(
      expect.objectContaining({ access_token: "new-access" }),
    );
    expect(clearAuthState).not.toHaveBeenCalled();
    expect(goto).not.toHaveBeenCalled();
  });

  it("redirects to login when token refresh fails", async () => {
    fetchMock.mockImplementation(async (input: RequestInfo | URL) => {
      const url = getRequestUrl(input);
      const path = new URL(url).pathname;

      if (path === "/api/test") {
        return createJsonResponse({ error: "Unauthorized" }, 401);
      }

      if (path === "/api/v1/auth/refresh") {
        return createJsonResponse({ error: "Invalid refresh token" }, 401);
      }

      return createJsonResponse({ error: "Not found" }, 404);
    });

    await expect(api.get("test")).rejects.toThrow();
    expect(clearAuthState).toHaveBeenCalled();
    expect(goto).toHaveBeenCalledWith("/auth/login");
  });

  it("does not refresh on a refresh request itself", async () => {
    fetchMock.mockResolvedValue(
      createJsonResponse({
        access_token: "token",
        refresh_token: "refresh",
        token_type: "Bearer",
        expires_at: "2025-01-01T00:00:00Z",
      }),
    );

    const response = await api
      .post("v1/auth/refresh")
      .json<{ access_token: string }>();

    expect(response.access_token).toBe("token");
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const request = fetchMock.mock.calls[0][0] as Request;
    expect(request.url).toMatch(/\/api\/v1\/auth\/refresh$/);
  });
});
