import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/svelte";
import { goto } from "$app/navigation";
import { isAuthenticated, initAuth, currentUser, isInitialized } from "$shared/stores/auth.svelte";
import type { User } from "$shared/types";

vi.mock("$shared/stores/auth.svelte", async () => {
  const actual = await vi.importActual<typeof import("$shared/stores/auth.svelte")>(
    "$shared/stores/auth.svelte"
  );
  return {
    ...actual,
    isAuthenticated: vi.fn(() => true),
    initAuth: vi.fn(),
    currentUser: vi.fn(() => null),
    isInitialized: vi.fn(() => true),
  };
});

describe("Profile page", () => {
  beforeEach(() => {
    vi.mocked(goto).mockClear();
    vi.mocked(isAuthenticated).mockReturnValue(true);
    vi.mocked(currentUser).mockReturnValue(null);
    vi.mocked(isInitialized).mockReturnValue(true);
    vi.mocked(initAuth).mockClear();
  });

  it("redirects when not authenticated and shows loading", async () => {
    vi.mocked(isAuthenticated).mockReturnValue(false);
    const Page = (await import("./+page.svelte")).default;
    render(Page);
    await waitFor(() => {
      expect(screen.getByText(/Loading profile/i)).toBeInTheDocument();
    });
    expect(goto).toHaveBeenCalledWith("/auth/login?redirect=/profile");
  });

  it("renders the profile editor when authenticated", async () => {
    vi.mocked(currentUser).mockReturnValue({
      id: "user-1",
      login: "test",
      email: "test@example.com",
      role: "user",
      created_at: new Date().toISOString(),
    } as User);
    const Page = (await import("./+page.svelte")).default;
    render(Page);
    await waitFor(() => {
      expect(screen.getByTestId("profile-content")).toBeInTheDocument();
    });
    expect(goto).not.toHaveBeenCalled();
  });
});
