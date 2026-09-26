/**
 * Утилиты для форматирования дат
 */

import { formatMessage, getCurrentLocale } from "./i18n";

const DEFAULT_OPTIONS: Intl.DateTimeFormatOptions = {
  day: "numeric",
  month: "long",
  year: "numeric",
};

/**
 * Locale for Intl formatters follows the UI language switch (UI-QUICK-1):
 * "ru" → "ru-RU", "en" → "en-US". Callers may still pass an explicit locale.
 */
function currentIntlLocale(): string {
  return getCurrentLocale() === "ru" ? "ru-RU" : "en-US";
}

/**
 * Форматирует дату в локализованную строку
 * @param dateString - ISO строка даты или Date объект
 * @param options - Опции форматирования Intl.DateTimeFormat
 * @param locale - Локаль (по умолчанию — язык интерфейса)
 * @returns Отформатированная строка даты
 */
export function formatDate(
  dateString: string | Date,
  options?: Intl.DateTimeFormatOptions,
  locale?: string
): string {
  try {
    const date = typeof dateString === "string" ? new Date(dateString) : dateString;

    if (isNaN(date.getTime())) {
      const msgLocale = locale ? (locale.startsWith("ru") ? "ru" : "en") : getCurrentLocale();
      return formatMessage("time.invalidDate", msgLocale);
    }

    const formatter = new Intl.DateTimeFormat(
      locale ?? currentIntlLocale(),
      options ?? DEFAULT_OPTIONS
    );
    return formatter.format(date);
  } catch (error) {
    if (import.meta.env.DEV) {
      console.error("Error formatting date:", error);
    }
    return String(dateString);
  }
}

/**
 * Форматирует дату с временем
 * @param dateString - ISO строка даты или Date объект
 * @param locale - Локаль (по умолчанию — язык интерфейса)
 * @returns Отформатированная строка даты и времени
 */
export function formatDateTime(dateString: string | Date, locale?: string): string {
  const options: Intl.DateTimeFormatOptions = {
    day: "numeric",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  };

  return formatDate(dateString, options, locale);
}

/**
 * Форматирует относительную дату (сегодня, вчера, и т.д.)
 * @param dateString - ISO строка даты или Date объект
 * @param locale - Локаль (по умолчанию — язык интерфейса)
 * @returns Относительная строка даты
 */
export function formatRelativeDate(dateString: string | Date, locale?: string): string {
  const date = typeof dateString === "string" ? new Date(dateString) : dateString;
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  const messageLocale = locale ? (locale.startsWith("ru") ? "ru" : "en") : getCurrentLocale();

  if (diffDays === 0) {
    return formatMessage("time.today", messageLocale);
  } else if (diffDays === 1) {
    return formatMessage("time.yesterday", messageLocale);
  } else if (diffDays < 7) {
    return formatMessage("time.daysAgoLong", messageLocale, { count: diffDays });
  } else {
    return formatDate(dateString, undefined, locale);
  }
}
