// Helpers that turn preview rows into the import request payload.
import type { ImportItem, ImportPreviewItem } from "$shared/api/import";

/**
 * A preview row is importable when it is new and either fetched cleanly or
 * the user opted in to keep just its title (`title_only`).
 */
export function isImportable(item: ImportPreviewItem): boolean {
  return item.is_new && (!item.error || !!item.title_only);
}

/**
 * Map preview rows to the create-import payload. Title-only rows go with an
 * empty text so the backend creates the note from title + source URL alone.
 */
export function toImportItems(items: ImportPreviewItem[]): ImportItem[] {
  return items.filter(isImportable).map(({ title, url, text, type, title_only }) => ({
    title,
    url,
    text: title_only ? "" : text,
    type: type || "asteroid",
  }));
}
