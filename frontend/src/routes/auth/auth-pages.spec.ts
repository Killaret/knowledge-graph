import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/svelte";
import { writable } from "svelte/store";
import { goto } from "$app/navigation";
import { page } from "$app/stores";
import {
  isAuthenticated,
  skipAuthMode,
  initAuth,
  currentUser,
  isInitialized,
  handleYandexCallback,
  error,
} from "$shared/stores/auth.svelte";

vi.mock("$shared/services/PreloadService", async () => {
  const actual = await vi.importActual<typeof import("$shared/services/PreloadService")>(
    "$shared/services/PreloadService"
  );
  return {
    ...actual,
    startPreload: vi.fn().mockResolvedValue(undefined),
    isPreloadingData: vi.fn(() => false),
    getPreloadedGraph: vi.fn(() => null),
  };
});

vi.mock("$features/preload/hooks/usePreloadedData", () => ({
  getGraphWithPreload: vi.fn().mockResolvedValue({ nodes: [], links: [] }),
}));

vi.mock("$shared/stores/auth.svelte", async () => {
  const actual = await vi.importActual<typeof import("$shared/stores/auth.svelte")>(
    "$shared/stores/auth.svelte"
  );
  return {
    ...actual,
    isAuthenticated: vi.fn(() => false),
    skipAuthMode: vi.fn(() => false),
    currentUser: vi.fn(() => null),
    isInitialized: vi.fn(() => true),
    initAuth: vi.fn(),
    handleYandexCallback: vi.fn(),
    error: vi.fn(() => null),
  };
});

vi.mock("$app/stores", () => ({
  page: writable({
    url: new URL("http://localhost"),
    params: {},
    route: { id: null },
    status: 200,
    error: null,
    data: {},
    form: undefined,
  }),
  updated: { subscribe: vi.fn(() => () => {}), check: vi.fn() },
}));

function setPageUrl(urlString: string) {
  (page as any).set({
    url: new URL(urlString),
    params: {},
    route: { id: null },
    status: 200,
    error: null,
    data: {},
    form: undefined,
  });
}

describe("Auth route pages", () => {
  beforeEach(() => {
    vi.mocked(goto).mockClear();
    vi.mocked(isAuthenticated).mockReturnValue(false);
    vi.mocked(skipAuthMode).mockReturnValue(false);
    vi.mocked(currentUser).mockReturnValue(null);
    vi.mocked(isInitialized).mockReturnValue(true);
    vi.mocked(initAuth).mockClear();
    vi.mocked(handleYandexCallback).mockReset();
    vi.mocked(error).mockReturnValue(null);
    setPageUrl("http://localhost");
  });

  it("login page renders and redirects when authenticated", async () => {
    vi.mocked(isAuthenticated).mockReturnValue(true);
    const LoginPage = (await import("./login/+page.svelte")).default;
    render(LoginPage);
    expect(initAuth).toHaveBeenCalled();
    expect(goto).toHaveBeenCalledWith("/");
  });

  it("login page renders without redirect when anonymous", async () => {
    const LoginPage = (await import("./login/+page.svelte")).default;
    render(LoginPage);
    expect(goto).not.toHaveBeenCalled();
  });

  it("register page redirects to requested page when authenticated", async () => {
    setPageUrl("http://localhost/auth/register?redirect=/graph");
    vi.mocked(isAuthenticated).mockReturnValue(true);
    const RegisterPage = (await import("./register/+page.svelte")).default;
    render(RegisterPage);
    expect(goto).toHaveBeenCalledWith("/graph");
  });

  it("forgot password page renders", async () => {
    const ForgotPage = (await import("./forgot-password/+page.svelte")).default;
    const { container } = render(ForgotPage);
    expect(container.textContent?.length).toBeGreaterThan(0);
  });

  it("reset password page shows the form when token is present", async () => {
    setPageUrl("http://localhost/auth/reset-password?token=abc123");
    const ResetPage = (await import("./reset-password/+page.svelte")).default;
    const { container } = render(ResetPage);
    expect(container.textContent).toContain("New Password");
  });

  it("reset password page shows the missing token state when token is absent", async () => {
    const ResetPage = (await import("./reset-password/+page.svelte")).default;
    const { container } = render(ResetPage);
    expect(container.textContent).toContain("token not found");
  });

  it("yandex callback handles missing parameters", async () => {
    const YandexPage = (await import("./yandex/callback/+page.svelte")).default;
    render(YandexPage);
    await waitFor(() => {
      expect(screen.getByText(/Missing required authorization parameters/)).toBeInTheDocument();
    });
  });

  it("yandex callback handles successful authentication", async () => {
    setPageUrl("http://localhost/auth/yandex/callback?code=123&state=xyz");
    vi.mocked(handleYandexCallback).mockResolvedValue(true);
    const YandexPage = (await import("./yandex/callback/+page.svelte")).default;
    render(YandexPage);
    await waitFor(() => {
      expect(handleYandexCallback).toHaveBeenCalledWith("123", "xyz");
      expect(goto).toHaveBeenCalledWith("/");
    });
  });

  it("yandex callback handles failed authentication", async () => {
    setPageUrl("http://localhost/auth/yandex/callback?code=123&state=xyz");
    vi.mocked(handleYandexCallback).mockResolvedValue(false);
    vi.mocked(error).mockReturnValue("Yandex error");
    const YandexPage = (await import("./yandex/callback/+page.svelte")).default;
    render(YandexPage);
    await waitFor(() => {
      expect(screen.getByText(/Yandex error/)).toBeInTheDocument();
    });
  });
});
