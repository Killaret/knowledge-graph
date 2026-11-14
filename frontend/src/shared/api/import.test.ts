import { describe, it, expect, beforeEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "../../../vitest-setup";
import {
  createBookmarkletNote,
  previewBookmarks,
  createBookmarksImport,
  getImportStatus,
  createSmartImport,
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

  // SMART-CREATE-1 blocker-1: the wire format must stay a raw
  // application/octet-stream body. Reverting to FormData/multipart restores
  // the SvelteKit CSRF 403 in production — this assertion turns red there.
  it("submits a document as raw octet-stream with metadata in the query", async () => {
    const seen: { contentType: string | null; query: string; body: string }[] = [];
    server.use(
      http.post("http://localhost:8080/api/v1/import/document", async ({ request }) => {
        seen.push({
          contentType: request.headers.get("content-type"),
          query: new URL(request.url).search,
          body: await request.text(),
        });
        return HttpResponse.json({ data: { task_id: "doc-1" } });
      })
    );

    const result = await createSmartImport({ type: "TEXT", content: "hello doc" });
    expect(result).toEqual({ task_id: "doc-1" });
    expect(seen).toHaveLength(1);
    expect(seen[0].contentType).toBe("application/octet-stream");
    expect(seen[0].contentType).not.toContain("multipart/");
    expect(seen[0].query).toContain("type=TEXT");
    expect(seen[0].body).toBe("hello doc");
  });

  it("submits a file with filename and content_type in the query", async () => {
    const seen: { contentType: string | null; params: URLSearchParams; body: string }[] = [];
    server.use(
      http.post("http://localhost:8080/api/v1/import/document", async ({ request }) => {
        seen.push({
          contentType: request.headers.get("content-type"),
          params: new URL(request.url).searchParams,
          body: await request.text(),
        });
        return HttpResponse.json({ data: { task_id: "doc-2" } });
      })
    );

    const file = new File(["pdf-bytes"], "doc.pdf", { type: "application/pdf" });
    const result = await createSmartImport({ type: "FILE", file });
    expect(result).toEqual({ task_id: "doc-2" });
    expect(seen).toHaveLength(1);
    expect(seen[0].contentType).toBe("application/octet-stream");
    expect(seen[0].params.get("type")).toBe("FILE");
    expect(seen[0].params.get("filename")).toBe("doc.pdf");
    expect(seen[0].params.get("content_type")).toBe("application/pdf");
    expect(seen[0].body).toBe("pdf-bytes");
  });
});
