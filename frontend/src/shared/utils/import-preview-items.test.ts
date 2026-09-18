import { describe, it, expect } from "vitest";
import { isImportable, toImportItems } from "$shared/utils/import-preview-items";
import type { ImportPreviewItem } from "$shared/api/import";

const base: ImportPreviewItem = {
  title: "Example",
  url: "https://example.com",
  text: "fetched text",
  is_new: true,
};

describe("isImportable", () => {
  it("accepts a clean new item", () => {
    expect(isImportable(base)).toBe(true);
  });

  it("rejects duplicates", () => {
    expect(isImportable({ ...base, is_new: false })).toBe(false);
  });

  it("rejects an item whose fetch failed", () => {
    expect(isImportable({ ...base, error: "fetch failed" })).toBe(false);
  });

  it("accepts a failed item when the user opted into title-only import", () => {
    expect(isImportable({ ...base, error: "fetch failed", title_only: true })).toBe(true);
  });
});

describe("toImportItems", () => {
  it("maps importable items and defaults type to asteroid", () => {
    const items = toImportItems([base]);
    expect(items).toEqual([
      { title: "Example", url: "https://example.com", text: "fetched text", type: "asteroid" },
    ]);
  });

  it("drops failed and duplicate items", () => {
    const items = toImportItems([
      base,
      { ...base, url: "https://dup.example", is_new: false },
      { ...base, url: "https://err.example", error: "timeout" },
    ]);
    expect(items).toHaveLength(1);
    expect(items[0].url).toBe("https://example.com");
  });

  it("imports title-only rows with an empty text", () => {
    const items = toImportItems([
      { ...base, url: "https://err.example", text: "partial", error: "timeout", title_only: true },
    ]);
    expect(items).toHaveLength(1);
    expect(items[0].text).toBe("");
    expect(items[0].title).toBe("Example");
  });
});
