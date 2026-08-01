// Helper to turn a ky HTTPError from simple `{ error: string }` backend
// responses (used by auth/user handlers) into a localized, user-facing
// message. Falls back to a generic server error message when the response
// body can't be parsed or doesn't match a known error.
import { HTTPError } from "ky";
import { formatMessage, type Locale } from "$shared/utils/i18n";

// Maps known backend error strings (from gin.H{"error": "..."}) to i18n keys.
const KNOWN_ERROR_KEYS: Record<string, string> = {
  "email already in use": "profile.emailTaken",
  "old password is required": "profile.oldPasswordRequired",
  "invalid old password": "profile.oldPasswordInvalid",
  "invalid password": "profile.deletePasswordInvalid",
};

/**
 * Extracts a localized, user-friendly message from an API error.
 * Handles ky HTTPError responses shaped as `{ error: string }` (used by the
 * auth/user endpoints) and falls back to a generic server error message.
 */
export async function getApiErrorMessage(error: unknown, locale: Locale): Promise<string> {
  if (error instanceof HTTPError) {
    try {
      const body = (await error.response.clone().json()) as { error?: string };
      const rawMessage = body?.error;
      if (rawMessage) {
        const key = KNOWN_ERROR_KEYS[rawMessage.toLowerCase()];
        if (key) {
          return formatMessage(key, locale);
        }
        // Gin's validator error for the "email" tag isn't user-friendly raw;
        // map it explicitly instead of leaking Go struct field names.
        if (rawMessage.includes("'email' tag") || rawMessage.includes("Email")) {
          return formatMessage("profile.emailInvalid", locale);
        }
        return rawMessage;
      }
    } catch {
      // Response body wasn't JSON or already consumed; fall through.
    }
  }

  return formatMessage("server.error", locale);
}
