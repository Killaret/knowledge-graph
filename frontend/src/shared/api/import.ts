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

export interface ImportPreviewItem extends ImportItem {
  is_new: boolean;
  existing_note_id?: string;
  error?: string;
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

export interface ImportTaskStatus {
  task_id: string;
  status: "pending" | "processing" | "done" | "failed";
  progress: ImportTaskProgress;
}

/**
 * Create a note from a captured web page (bookmarklet flow).
 */
export async function createBookmarkletNote(data: BookmarkletPayload): Promise<BookmarkletResult> {
  return api.post("v1/import/bookmarklet", { json: data }).json<BookmarkletResult>();
}

/**
 * Build a preview for a batch of captured web pages.
 */
export async function previewBookmarks(items: ImportItem[]): Promise<ImportPreviewResponse> {
  return api.post("v1/import/bookmarks/preview", { json: { items } }).json<ImportPreviewResponse>();
}

/**
 * Start an async batch import of captured web pages.
 */
export async function createBookmarksImport(items: ImportItem[]): Promise<ImportTask> {
  return api.post("v1/import/bookmarks", { json: { items } }).json<ImportTask>();
}

/**
 * Get the current status of a batch import task.
 */
export async function getImportStatus(taskId: string): Promise<ImportTaskStatus> {
  return api.get(`v1/import/${taskId}/status`).json<ImportTaskStatus>();
}
