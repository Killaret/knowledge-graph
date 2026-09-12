import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/svelte";
import { writable } from "svelte/store";
import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { isAuthenticated, initAuth } from "$shared/stores/auth.svelte";
import { createBookmarkletNote } from "$shared/api/import";

vi.mock("$shared/api/import", () => ({
  createBookmarkletNote: vi.fn(),
}));

vi.mock("$shared/stores/auth.svelte", async () => {
  const actual = await vi.importActual<typeof import("$shared/stores/auth.svelte")>(
    "$shared/stores/auth.svelte"
  );
  return {
    ...actual,
    isAuthenticated: vi.fn(() => true),
    initAuth: vi.fn(),
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

describe("Import page", () => {
  beforeEach(() => {
    vi.mocked(goto).mockClear();
    vi.mocked(isAuthenticated).mockReturnValue(true);
    vi.mocked(initAuth).mockClear();
    vi.mocked(createBookmarkletNote).mockReset();
    setPageUrl("http://localhost/import");
  });

  it("redirects when not authenticated", async () => {
    vi.mocked(isAuthenticated).mockReturnValue(false);
    const Page = (await import("./+page.svelte")).default;
    render(Page);
    await waitFor(() => {
      expect(screen.getByText("Authorization required.")).toBeInTheDocument();
    });
    expect(goto).toHaveBeenCalledWith(expect.stringMatching(/\/auth\/login\?redirect=/));
  });

  it("shows error when title or url are missing", async () => {
    const Page = (await import("./+page.svelte")).default;
    render(Page);
    await waitFor(() => {
      expect(screen.getByText("No page data provided.")).toBeInTheDocument();
    });
  });

  it("creates a note and shows success", async () => {
    setPageUrl("http://localhost/import?title=Test&url=http://example.com&text=hello");
    vi.mocked(createBookmarkletNote).mockResolvedValue({
      note_id: "note-1",
      title: "Test",
      type: "star",
    });
    const Page = (await import("./+page.svelte")).default;
    render(Page);
    await waitFor(() => {
      expect(screen.getByText(/Note "Test" created successfully/)).toBeInTheDocument();
    });
    expect(createBookmarkletNote).toHaveBeenCalledWith({ title: "Test", url: "http://example.com", text: "hello" });
  });

  it("shows error when import fails", async () => {
    setPageUrl("http://localhost/import?title=Test&url=http://example.com");
    vi.mocked(createBookmarkletNote).mockRejectedValue(new Error("fail"));
    const Page = (await import("./+page.svelte")).default;
    render(Page);
    await waitFor(() => {
      expect(screen.getByText("Failed to create note from page.")).toBeInTheDocument();
    });
  });

  it("navigates to the created note", async () => {
    setPageUrl("http://localhost/import?title=Test&url=http://example.com");
    vi.mocked(createBookmarkletNote).mockResolvedValue({
      note_id: "note-1",
      title: "Test",
      type: "star",
    });
    const Page = (await import("./+page.svelte")).default;
    render(Page);
    const noteButton = await waitFor(() => screen.getByTestId("import-open-note"));
    await fireEvent.click(noteButton);
    expect(goto).toHaveBeenCalledWith("/notes/note-1");
  });
});
