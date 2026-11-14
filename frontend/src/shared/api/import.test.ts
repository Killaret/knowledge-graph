import { describe, it, expect, beforeEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "../../../vitest-setup";
import {
  createBookmarkletNote,
  previewBookmarks,
  createBookmarksImport,
  getImportStatus,
  type BookmarkletPayload,
  type ImportItem,
} from "./import";

describe("import API", () => {
  beforeEach(() => {
    server.resetHandlers();
  });

  it("creates a bookmarklet note", async () => {
    const payload: BookmarkletPayload = {
      title: "Title",
      url: "https://example.com",
      text: "Body",
      type: "star",
    };
    const response = { note_id: "n1", title: "Title", type: "star" };

    server.use(
      http.post("http://localhost:8080/api/v1/import/bookmarklet", () =>
        HttpResponse.json(response)
      )
    );

    const result = await createBookmarkletNote(payload);
    expect(result).toEqual(response);
  });

  it("previews a batch of bookmarks", async () => {
    const items: ImportItem[] = [{ title: "A", url: "https://a.com" }];
    const response = { items: [{ ...items[0], is_new: true }] };

    server.use(
      http.post("http://localhost:8080/api/v1/import/bookmarks/preview", () =>
        HttpResponse.json({ data: response })
      )
    );

    const result = await previewBookmarks(items, { default_type: "star" });
    expect(result).toEqual(response);
  });

  it("creates a batch bookmarks import", async () => {
    const items: ImportItem[] = [{ title: "A", url: "https://a.com" }];
    const response = { task_id: "t1", message: "started" };

    server.use(
      http.post("http://localhost:8080/api/v1/import/bookmarks", () =>
        HttpResponse.json({ data: response })
      )
    );

    const result = await createBookmarksImport(items);
    expect(result).toEqual(response);
  });

  it("gets import task status", async () => {
    const response = {
      task_id: "t1",
      status: "done" as const,
      progress: { total: 10, processed: 10, created: 10, skipped: 0, failed: 0 },
    };

    server.use(
      http.get("http://localhost:8080/api/v1/import/t1/status", () =>
        HttpResponse.json({ data: response })
      )
    );

    const result = await getImportStatus("t1");
    expect(result).toEqual(response);
  });
});
