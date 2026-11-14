// API client for import operations
import { api } from "$shared/api/client";

export interface BookmarkletPayload {
  title: string;
  url: string;
  text: string;
  type?: string;
}

export interface BookmarkletResult {
  note_id: string;
  title: string;
  type: string;
}

export interface ImportItem {
  title: string;
  url: string;
  text?: string;
  type?: string;
}

export interface ImportOutlineEntry {
  level: number;
  text: string;
}

export interface ImportPreviewItem extends ImportItem {
  is_new: boolean;
  existing_note_id?: string;
  error?: string;
  /** Client-side opt-in: import this failed item with title only, no text. */
  title_only?: boolean;
  /** URL-HEADING-1 stage A: title candidates in priority order. */
  title_candidates?: string[];
  /** "rule" at stage A. */
  title_source?: string;
  /** Extracted page outline with normalized levels (2..6). */
  outline?: ImportOutlineEntry[];
  /** Number of noise elements removed from the content container. */
  noise_dropped?: number;
}

export interface ImportPreviewResponse {
  items: ImportPreviewItem[];
}

export interface ImportTask {
  task_id: string;
  message?: string;
}

export interface ImportTaskProgress {
  total: number;
  processed: number;
  created: number;
  skipped: number;
  failed: number;
}

export interface ImportTaskStatusNote {
  id: string;
  title: string;
}

export interface ImportTaskStatus {
  task_id: string;
  status: "pending" | "processing" | "done" | "failed";
  progress: ImportTaskProgress;
  /** SMART-CREATE-1: document tasks carry owner, result summary and notes. */
  user_id?: string;
  kind?: string;
  notes?: ImportTaskStatusNote[];
  raw_chunks_count?: number;
  errors?: string[];
  reason?: string;
}

export interface SmartCreateAvailability {
  available: boolean;
}

export type SmartImportType = "TEXT" | "URL" | "FILE";

export interface ImportOptions {
  default_type?: string;
  extract_content?: boolean;
}

export const MAX_IMPORT_BATCH_SIZE = 50;

/**
 * Create a note from a captured web page (bookmarklet flow).
 */
export async function createBookmarkletNote(data: BookmarkletPayload): Promise<BookmarkletResult> {
  return api.post("v1/import/bookmarklet", { json: data }).json<BookmarkletResult>();
}

/**
 * Build a preview for a batch of captured web pages.
 */
export async function previewBookmarks(
  items: ImportItem[],
  options?: ImportOptions
): Promise<ImportPreviewResponse> {
  const res = await api
    .post("v1/import/bookmarks/preview", {
      json: { items, options },
      // Content extraction fetches every URL in the batch — a 50-item batch
      // legitimately takes longer than the default 30s client timeout.
      timeout: 120000,
    })
    .json<{ data: ImportPreviewResponse }>();
  return res.data;
}

/**
 * Start an async batch import of captured web pages.
 */
export async function createBookmarksImport(
  items: ImportItem[],
  options?: ImportOptions
): Promise<ImportTask> {
  const res = await api
    .post("v1/import/bookmarks", { json: { items, options } })
    .json<{ data: ImportTask }>();
  return res.data;
}

/**
 * Get the current status of a batch import task.
 */
export async function getImportStatus(taskId: string): Promise<ImportTaskStatus> {
  const res = await api.get(`v1/import/${taskId}/status`).json<{ data: ImportTaskStatus }>();
  return res.data;
}

/**
 * SMART-CREATE-1: is the Java source-text-handler listener alive?
 */
export async function getSmartCreateAvailability(): Promise<SmartCreateAvailability> {
  const res = await api
    .get("v1/import/document/availability")
    .json<{ data: SmartCreateAvailability }>();
  return res.data;
}

/**
 * SMART-CREATE-1: send a document (text, URL or file) to the source-text
 * handler queue.
 *
 * The body is raw bytes (application/octet-stream) with the metadata in the
 * query string — deliberately NOT multipart/form-data. Form content types run
 * through SvelteKit's CSRF origin check, and the frontend container sets no
 * ORIGIN, so every multipart POST to /api/v1 died with 403 at the proxy
 * before reaching the backend. A raw body skips that check and survives every
 * deployment origin unchanged.
 */
export async function createSmartImport(input: {
  type: SmartImportType;
  content?: string;
  file?: File;
}): Promise<ImportTask> {
  const searchParams: Record<string, string> = { type: input.type };
  let body: BodyInit;
  if (input.type === "FILE") {
    if (!input.file) throw new Error("file is required for FILE");
    searchParams.filename = input.file.name;
    if (input.file.type) searchParams.content_type = input.file.type;
    body = input.file;
  } else {
    body = input.content ?? "";
  }
  const res = await api
    .post("v1/import/document", {
      searchParams,
      body,
      headers: { "content-type": "application/octet-stream" },
      timeout: 120000,
    })
    .json<{ data: ImportTask }>();
  return res.data;
}
