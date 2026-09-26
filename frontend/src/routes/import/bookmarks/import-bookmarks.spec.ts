import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/svelte";
import { CelestialBody } from "$entities/shared/model/celestial-body";

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

vi.mock("$shared/stores/auth.svelte.js", async () => {
  const actual = await vi.importActual<typeof import("$shared/stores/auth.svelte")>(
    "$shared/stores/auth.svelte"
  );
  return {
    ...actual,
    isAuthenticated: vi.fn(() => true),
    initAuth: vi.fn(async () => {}),
    skipAuthMode: vi.fn(() => false),
  };
});

vi.mock("$shared/api/import", () => ({
  previewBookmarks: vi.fn(async () => ({
    items: [
      {
        title: "Example",
        url: "https://example.com",
        text: "",
        type: "asteroid",
        is_new: true,
      },
    ],
  })),
  createBookmarksImport: vi.fn(),
  getImportStatus: vi.fn(),
  MAX_IMPORT_BATCH_SIZE: 50,
}));

// IMP-1: the type selector on /import/bookmarks must offer exactly
// CelestialBody.UI_TYPES — no `technical`, `unknown`, or anomaly types.
// A hardcoded list here (the original defect) must fail this test.
describe("Import bookmarks page — type selector", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("offers exactly the UI types and none of the system types", async () => {
    const Page = (await import("./+page.svelte")).default;
    render(Page);

    const textarea = await screen
      .findByLabelText(/paste|list|заклад|список/i)
      .catch(() => document.getElementById("import-list"));
    await fireEvent.input(textarea as HTMLElement, {
      target: { value: "Example | https://example.com" },
    });

    await fireEvent.click(screen.getByTestId("preview-import"));

    const select = (await waitFor(() =>
      document.querySelector("select.table-select")
    )) as HTMLSelectElement;
    expect(select).toBeTruthy();

    const options = Array.from(select.options).map((o) => o.value);
    const expected = CelestialBody.UI_TYPES.map((b) => b.type);

    expect(options.sort()).toEqual([...expected].sort());
    for (const banned of [
      "technical",
      "unknown",
      "reality_rift",
      "chromatic_maw",
      "void_whisper",
      "cosmic_abomination",
    ]) {
      expect(options).not.toContain(banned);
    }

    // IMP-3 §4: every option carries a description hint in its title.
    for (const option of Array.from(select.options)) {
      expect(option.getAttribute("title"), `option ${option.value} has no hint`).toBeTruthy();
    }
  });
});

// UI-QUICK-1: mass-import page quick fixes — extract checkbox on by default,
// real newline in the textarea placeholder, and a way back to the graph.
describe("Import bookmarks page — UI-QUICK-1", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("checks the extract-content toggle by default", async () => {
    const Page = (await import("./+page.svelte")).default;
    render(Page);

    const checkbox = (await waitFor(() =>
      document.querySelector(".extract-toggle input[type=checkbox]")
    )) as HTMLInputElement;
    expect(checkbox).toBeTruthy();
    expect(checkbox.checked).toBe(true);
  });

  it("uses a real newline in the list placeholder, not a literal backslash-n", async () => {
    const Page = (await import("./+page.svelte")).default;
    render(Page);

    const textarea = (await waitFor(() =>
      document.getElementById("import-list")
    )) as HTMLTextAreaElement;
    const placeholder = textarea.getAttribute("placeholder") ?? "";
    expect(placeholder).toContain("\n");
    expect(placeholder).not.toContain(String.raw`\n`);
  });

  it("links back to the graph from the tabs row", async () => {
    const Page = (await import("./+page.svelte")).default;
    render(Page);

    const back = await screen.findByTestId("back-to-graph");
    expect(back.getAttribute("href")).toBe("/graph");
  });
});
