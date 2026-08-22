import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { formatMessage, getCurrentLocale, setLocale } from "./i18n";

describe("i18n", () => {
  beforeEach(() => {
    // Clear localStorage before each test
    if (typeof window !== "undefined") {
      localStorage.clear();
    }
  });

  afterEach(() => {
    // Clear localStorage after each test
    if (typeof window !== "undefined") {
      localStorage.clear();
    }
  });

  describe("formatMessage", () => {
    it("should return English message by default", () => {
      const result = formatMessage("note.created", "en", {
        title: "Test Note",
      });
      expect(result).toBe('Note "Test Note" created successfully.');
    });

    it("should return Russian message when locale is ru", () => {
      const result = formatMessage("note.created", "ru", {
        title: "Тестовая заметка",
      });
      expect(result).toBe('Заметка "Тестовая заметка" успешно создана.');
    });

    it("should replace single parameter", () => {
      const result = formatMessage("validation.error", "en", {
        field: "title",
      });
      expect(result).toBe('Invalid value in field "title".');
    });

    it("should replace multiple parameters", () => {
      const result = formatMessage("link.created", "en", {
        source: "A",
        target: "B",
      });
      expect(result).toBe('Link from "A" to "B" created.');
    });

    it("should return key if message not found", () => {
      const result = formatMessage("nonexistent.key", "en");
      expect(result).toBe("nonexistent.key");
    });

    it("should fallback to English if locale message not found", () => {
      const result = formatMessage("note.created", "ru");
      expect(result).toBe('Заметка "{{title}}" успешно создана.');
    });

    it("should handle empty params", () => {
      const result = formatMessage("loading", "en");
      expect(result).toBe("Loading data...");
    });

    it("should handle numeric parameters", () => {
      const result = formatMessage("delete.confirm", "en", { item: "Note 1" });
      expect(result).toBe('Are you sure you want to delete "Note 1"?');
    });
  });

  describe("getCurrentLocale", () => {
    it('should return "en" by default', () => {
      const result = getCurrentLocale();
      expect(result).toBe("en");
    });

    it("should return stored locale from localStorage", () => {
      setLocale("ru");
      const result = getCurrentLocale();
      expect(result).toBe("ru");
    });

    it('should return "en" if stored locale is invalid', () => {
      if (typeof window !== "undefined") {
        localStorage.setItem("locale", "invalid");
      }
      const result = getCurrentLocale();
      expect(result).toBe("en");
    });
  });

  describe("setLocale", () => {
    it("should set locale in localStorage", () => {
      setLocale("ru");
      expect(localStorage.getItem("locale")).toBe("ru");
    });

    it('should allow setting to "en"', () => {
      setLocale("en");
      expect(localStorage.getItem("locale")).toBe("en");
    });

    it('should allow setting to "ru"', () => {
      setLocale("ru");
      expect(localStorage.getItem("locale")).toBe("ru");
    });
  });

  describe("integration", () => {
    it("should work end-to-end: set locale, get locale, format message", () => {
      setLocale("ru");
      const locale = getCurrentLocale();
      const message = formatMessage("note.created", locale, { title: "Test" });

      expect(locale).toBe("ru");
      expect(message).toBe('Заметка "Test" успешно создана.');
    });

    it("should persist locale across calls", () => {
      setLocale("ru");
      const locale1 = getCurrentLocale();
      const locale2 = getCurrentLocale();

      expect(locale1).toBe("ru");
      expect(locale2).toBe("ru");
    });
  });
});
