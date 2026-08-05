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

/**
 * Create a note from a captured web page (bookmarklet flow).
 */
export async function createBookmarkletNote(data: BookmarkletPayload): Promise<BookmarkletResult> {
  return api.post("v1/import/bookmarklet", { json: data }).json<BookmarkletResult>();
}
