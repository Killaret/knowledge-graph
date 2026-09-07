/**
 * Internationalization (i18n) helper
 * Provides message formatting with language support
 */

import { messages } from "./i18n/messages";

export type Locale = "en" | "ru";
export type MessageParams = Record<string, string | number>;

/**
 * Format a message with parameters
 * @param key - Message key
 * @param locale - Language locale (default: 'en')
 * @param params - Parameters to replace in message (optional)
 * @returns Formatted message
 */
export function formatMessage(key: string, locale: Locale = "en", params?: MessageParams): string {
  const message = messages[locale]?.[key] || messages.en[key] || key;

  if (!params) {
    return message;
  }

  // Replace {{param}} placeholders
  return Object.entries(params).reduce((result, [paramKey, paramValue]) => {
    return result.replace(new RegExp(`{{${paramKey}}}`, "g"), String(paramValue));
  }, message);
}

/**
 * Get current locale from localStorage or default
 * @returns Current locale
 */
export function getCurrentLocale(): Locale {
  if (typeof window === "undefined") return "en";

  const stored = localStorage.getItem("locale");
  if (stored === "en" || stored === "ru") {
    return stored;
  }

  return "en";
}

/**
 * Set current locale in localStorage
 * @param locale - Locale to set
 */
export function setLocale(locale: Locale): void {
  if (typeof window !== "undefined") {
    localStorage.setItem("locale", locale);
  }
}
