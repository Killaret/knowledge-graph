import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { formatDate, formatDateTime, formatRelativeDate } from "./date";

describe("date utils", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
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
      expect(formatDate("not-a-date")).toBe("Некорректная дата");
    });

    it("falls back to String(input) on formatter errors", () => {
      const badInput = { toString: () => "bad-input" } as unknown as Date;
      expect(formatDate(badInput, undefined, "en-US")).toBe("bad-input");
    });
  });

  describe("formatDateTime", () => {
    it("formats a date with time", () => {
      const result = formatDateTime("2026-07-21T14:30:00.000Z", "en-US");
      expect(result).toMatch(/2026/);
      expect(result).toMatch(/:/);
    });
  });

  describe("formatRelativeDate", () => {
    beforeEach(() => {
      vi.setSystemTime(new Date("2026-07-21T12:00:00.000Z"));
    });

    it('returns "Сегодня" for the same day', () => {
      expect(formatRelativeDate("2026-07-21T10:00:00.000Z")).toBe("Сегодня");
    });

    it('returns "Вчера" for yesterday', () => {
      expect(formatRelativeDate("2026-07-20T10:00:00.000Z")).toBe("Вчера");
    });

    it('returns "N дня назад" for recent dates', () => {
      expect(formatRelativeDate("2026-07-18T10:00:00.000Z")).toBe("3 дня назад");
    });

    it("falls back to formatDate for older dates", () => {
      const result = formatRelativeDate("2026-07-10T10:00:00.000Z");
      expect(result).not.toBe("Сегодня");
      expect(result).toMatch(/2026/);
    });
  });
});
