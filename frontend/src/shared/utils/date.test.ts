import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { formatDate, formatDateTime, formatRelativeDate } from "./date";
import { setLocale } from "./i18n";

describe("date utils", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    localStorage.removeItem("locale");
  });

  describe("formatDate", () => {
    it("formats an ISO date string in the requested locale", () => {
      const result = formatDate("2026-07-21T10:00:00.000Z", undefined, "en-US");
      expect(result).toMatch(/2026/);
      expect(result).toMatch(/21/);
      expect(result).toMatch(/July/);
    });

    it("formats a Date object with custom options", () => {
      const result = formatDate(
        new Date("2026-07-21T10:00:00.000Z"),
        { month: "short", day: "2-digit" },
        "en-US"
      );
      expect(result).toMatch(/Jul/);
      expect(result).toMatch(/21/);
    });

    it("returns an invalid date message for unparseable input", () => {
      expect(formatDate("not-a-date", undefined, "ru-RU")).toBe("Некорректная дата");
      expect(formatDate("not-a-date", undefined, "en-US")).toBe("Invalid date");
    });

    it("falls back to String(input) on formatter errors", () => {
      const badInput = { toString: () => "bad-input" } as unknown as Date;
      expect(formatDate(badInput, undefined, "en-US")).toBe("bad-input");
    });

    // UI-QUICK-1: dates must follow the EN/RU interface switch, not a fixed
    // ru-RU default. The switch is the `locale` key in localStorage.
    it("follows the interface locale — English month under EN", () => {
      setLocale("en");
      expect(formatDate("2026-07-21T10:00:00.000Z")).toMatch(/July/);
    });

    it("follows the interface locale — Russian month under RU", () => {
      setLocale("ru");
      const result = formatDate("2026-07-21T10:00:00.000Z");
      expect(result).toMatch(/июл/);
      expect(result).not.toMatch(/July/);
    });
  });

  describe("formatDateTime", () => {
    it("formats a date with time", () => {
      const result = formatDateTime("2026-07-21T14:30:00.000Z", "en-US");
      expect(result).toMatch(/2026/);
      expect(result).toMatch(/:/);
    });

    it("follows the interface locale under RU", () => {
      setLocale("ru");
      expect(formatDateTime("2026-07-21T14:30:00.000Z")).toMatch(/июл/);
    });
  });

  describe("formatRelativeDate", () => {
    beforeEach(() => {
      vi.setSystemTime(new Date("2026-07-21T12:00:00.000Z"));
      setLocale("ru");
    });

    it('returns "сегодня" for the same day', () => {
      expect(formatRelativeDate("2026-07-21T10:00:00.000Z")).toBe("сегодня");
    });

    it('returns "вчера" for yesterday', () => {
      expect(formatRelativeDate("2026-07-20T10:00:00.000Z")).toBe("вчера");
    });

    it('returns "N дня назад" for recent dates', () => {
      expect(formatRelativeDate("2026-07-18T10:00:00.000Z")).toBe("3 дня назад");
    });

    it("returns English strings under EN", () => {
      setLocale("en");
      expect(formatRelativeDate("2026-07-21T10:00:00.000Z")).toBe("today");
      expect(formatRelativeDate("2026-07-20T10:00:00.000Z")).toBe("yesterday");
      expect(formatRelativeDate("2026-07-18T10:00:00.000Z")).toBe("3 days ago");
    });

    it("falls back to formatDate for older dates", () => {
      const result = formatRelativeDate("2026-07-10T10:00:00.000Z");
      expect(result).not.toBe("сегодня");
      expect(result).toMatch(/2026/);
    });
  });
});
