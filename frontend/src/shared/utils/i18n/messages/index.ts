/**
 * I18n messages index
 * Merges per-domain message dictionaries into the final messages object.
 */

import { en as commonEn, ru as commonRu } from "./common";
import { en as authEn, ru as authRu } from "./auth";
import { en as notesEn, ru as notesRu } from "./notes";
import { en as graphEn, ru as graphRu } from "./graph";
import { en as importEn, ru as importRu } from "./import";
import { en as profileEn, ru as profileRu } from "./profile";
import { en as uiEn, ru as uiRu } from "./ui";

type Locale = "en" | "ru";

function mergeLocale(locale: Locale, objects: Record<string, string>[]): Record<string, string> {
  const result: Record<string, string> = {};
  const seen = new Set<string>();

  for (const obj of objects) {
    for (const [key, value] of Object.entries(obj)) {
      if (seen.has(key)) {
        throw new Error(`Duplicate i18n key "${key}" in locale "${locale}"`);
      }
      seen.add(key);
      result[key] = value;
    }
  }

  return result;
}

export const messages: Record<Locale, Record<string, string>> = {
  en: mergeLocale("en", [commonEn, authEn, notesEn, graphEn, importEn, profileEn, uiEn]),
  ru: mergeLocale("ru", [commonRu, authRu, notesRu, graphRu, importRu, profileRu, uiRu]),
};
