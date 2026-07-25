import { describe, it, expect } from "vitest";
import {
  MessageFormatter,
  GalacticLexicon,
  createFormatter,
  getLexiconMessage,
  getMessageKeys,
} from "./galactic-lexicon";

describe("MessageFormatter", () => {
  describe("English standard (default)", () => {
    const formatter = new MessageFormatter(false);

    it("should return English success messages", () => {
      const message = formatter.success("noteCreated", "Test Note");
      expect(message).toContain("Test Note");
      expect(message).toContain("created");
      expect(message).not.toContain("Star");
    });

    it("should return English error messages", () => {
      const message = formatter.error("validation", "title");
      expect(message).toContain("title");
      expect(message).toContain("Invalid value");
      expect(message).not.toContain("anomaly");
    });

    it("should return English info messages", () => {
      const message = formatter.info("emptyGraph");
      expect(message).toContain("graph is empty");
      expect(message).toContain("starry sky");
    });

    it("should return English warning messages", () => {
      const message = formatter.warning("unsavedChanges");
      expect(message).toContain("unsaved changes");
      expect(message).not.toContain("ship log");
    });
  });

  describe("Russian standard", () => {
    const formatter = new MessageFormatter(false, "ru");

    it("should return Russian success messages", () => {
      const message = formatter.success("noteCreated", "Test Note");
      expect(message).toContain("Test Note");
      expect(message).toContain("создана");
      expect(message).not.toContain("Звезда");
    });

    it("should return Russian error messages", () => {
      const message = formatter.error("validation", "title");
      expect(message).toContain("title");
      expect(message).toContain("Неверное значение");
      expect(message).not.toContain("аномалию");
    });

    it("should return Russian info messages", () => {
      const message = formatter.info("emptyGraph");
      expect(message).toContain("Граф пуст");
      expect(message).toContain("звёздное небо");
    });

    it("should return Russian warning messages", () => {
      const message = formatter.warning("unsavedChanges");
      expect(message).toContain("несохраненные изменения");
      expect(message).not.toContain("бортовом журнале");
    });
  });

  describe("English galactic", () => {
    const formatter = new MessageFormatter(true);

    it("should return galactic English success messages", () => {
      const message = formatter.success("noteCreated", "Test Note");
      expect(message).toContain("Test Note");
      expect(message).toContain("Star");
      expect(message).toContain("ignited");
    });

    it("should return galactic English error messages", () => {
      const message = formatter.error("validation", "title");
      expect(message).toContain("title");
      expect(message).toContain("anomaly");
      expect(message).toContain("Sensors");
    });

    it("should return galactic English info messages", () => {
      const message = formatter.info("emptyGraph");
      expect(message).toContain("starry sky is empty");
      expect(message).toContain("universe");
    });

    it("should return galactic English warning messages", () => {
      const message = formatter.warning("unsavedChanges");
      expect(message).toContain("ship log");
      expect(message).toContain("entries");
    });
  });

  describe("Russian galactic", () => {
    const formatter = new MessageFormatter(true, "ru");

    it("should return galactic Russian success messages", () => {
      const message = formatter.success("noteCreated", "Test Note");
      expect(message).toContain("Test Note");
      expect(message).toContain("Звезда");
      expect(message).toContain("зажжена");
    });

    it("should return galactic Russian error messages", () => {
      const message = formatter.error("validation", "title");
      expect(message).toContain("title");
      expect(message).toContain("аномалию");
      expect(message).toMatch(/[Сс]енсоры/);
    });

    it("should return galactic Russian info messages", () => {
      const message = formatter.info("emptyGraph");
      expect(message).toContain("звёздное небо");
      expect(message).toContain("пусто");
    });

    it("should return galactic Russian warning messages", () => {
      const message = formatter.warning("unsavedChanges");
      expect(message).toContain("бортовом журнале");
      expect(message).toContain("записи");
    });
  });

  describe("mode and locale switching", () => {
    it("should switch between modes and locales", () => {
      const formatter = new MessageFormatter(false);

      let message = formatter.success("noteCreated", "Note");
      expect(message).toContain("created");

      formatter.setLocale("ru");
      message = formatter.success("noteCreated", "Note");
      expect(message).toContain("создана");

      formatter.setGalacticMode(true);
      message = formatter.success("noteCreated", "Note");
      expect(message).toContain("Звезда");

      formatter.setLocale("en");
      message = formatter.success("noteCreated", "Note");
      expect(message).toContain("Star");

      formatter.setGalacticMode(false);
      message = formatter.success("noteCreated", "Note");
      expect(message).toContain("created");
    });

    it("should report current mode and locale", () => {
      const formatter = new MessageFormatter(false, "ru");
      expect(formatter.isGalacticMode()).toBe(false);
      expect(formatter.getLocale()).toBe("ru");

      formatter.setGalacticMode(true);
      expect(formatter.isGalacticMode()).toBe(true);

      formatter.setLocale("en");
      expect(formatter.getLocale()).toBe("en");
    });
  });

  describe("createFormatter helper", () => {
    it("should create formatter with specified mode and locale", () => {
      const enStandard = createFormatter(false, "en");
      expect(enStandard.isGalacticMode()).toBe(false);
      expect(enStandard.getLocale()).toBe("en");
      expect(enStandard.success("noteCreated", "Note")).toContain("created");

      const ruGalactic = createFormatter(true, "ru");
      expect(ruGalactic.isGalacticMode()).toBe(true);
      expect(ruGalactic.getLocale()).toBe("ru");
      expect(ruGalactic.success("noteCreated", "Note")).toContain("Звезда");
    });
  });
});

describe("GalacticLexicon", () => {
  describe("success messages", () => {
    it("should return English standard note created message", () => {
      const message = GalacticLexicon.success.noteCreated("My Note", false, "en");
      expect(message).toContain("My Note");
      expect(message).toContain("created");
      expect(message).not.toContain("Star");
    });

    it("should return Russian standard note created message", () => {
      const message = GalacticLexicon.success.noteCreated("My Note", false, "ru");
      expect(message).toContain("My Note");
      expect(message).toContain("создана");
    });

    it("should return English galactic note created message", () => {
      const message = GalacticLexicon.success.noteCreated("My Note", true, "en");
      expect(message).toContain("My Note");
      expect(message).toContain("Star");
      expect(message).toContain("ignited");
    });

    it("should return Russian galactic note created message", () => {
      const message = GalacticLexicon.success.noteCreated("My Note", true, "ru");
      expect(message).toContain("My Note");
      expect(message).toContain("Звезда");
      expect(message).toContain("зажжена");
    });

    it("should return achievement unlocked message", () => {
      const enStandard = GalacticLexicon.success.achievementUnlocked("Explorer", false, "en");
      expect(enStandard).toContain("Explorer");
      expect(enStandard).toContain("Achievement");

      const ruGalactic = GalacticLexicon.success.achievementUnlocked("Explorer", true, "ru");
      expect(ruGalactic).toContain("Explorer");
      expect(ruGalactic).toContain("звезда");
    });
  });

  describe("error messages", () => {
    it("should return English standard validation error", () => {
      const message = GalacticLexicon.error.validation("email", false, "en");
      expect(message).toContain("email");
      expect(message).toContain("Invalid value");
    });

    it("should return Russian standard validation error", () => {
      const message = GalacticLexicon.error.validation("email", false, "ru");
      expect(message).toContain("email");
      expect(message).toContain("Неверное значение");
    });

    it("should return English galactic validation error", () => {
      const message = GalacticLexicon.error.validation("email", true, "en");
      expect(message).toContain("email");
      expect(message).toContain("anomaly");
    });

    it("should return Russian galactic validation error", () => {
      const message = GalacticLexicon.error.validation("email", true, "ru");
      expect(message).toContain("email");
      expect(message).toContain("аномалию");
    });

    it("should return unauthorized error", () => {
      const enStandard = GalacticLexicon.error.unauthorized(false, "en");
      expect(enStandard).toContain("Authorization required");

      const enGalactic = GalacticLexicon.error.unauthorized(true, "en");
      expect(enGalactic).toContain("star system");

      const ruStandard = GalacticLexicon.error.unauthorized(false, "ru");
      expect(ruStandard).toContain("Требуется авторизация");

      const ruGalactic = GalacticLexicon.error.unauthorized(true, "ru");
      expect(ruGalactic).toContain("Отказано");
      expect(ruGalactic).toContain("звёздной системе");
    });
  });

  describe("info messages", () => {
    it("should return empty graph message", () => {
      const enStandard = GalacticLexicon.info.emptyGraph(false, "en");
      expect(enStandard).toContain("graph is empty");

      const enGalactic = GalacticLexicon.info.emptyGraph(true, "en");
      expect(enGalactic).toContain("starry sky is empty");

      const ruStandard = GalacticLexicon.info.emptyGraph(false, "ru");
      expect(ruStandard).toContain("Граф пуст");

      const ruGalactic = GalacticLexicon.info.emptyGraph(true, "ru");
      expect(ruGalactic).toContain("звёздное небо");
    });

    it("should return streak message with days", () => {
      const enStandard = GalacticLexicon.info.streakActive(7, false, "en");
      expect(enStandard).toContain("7");
      expect(enStandard).toContain("days");

      const enGalactic = GalacticLexicon.info.streakActive(7, true, "en");
      expect(enGalactic).toContain("7");
      expect(enGalactic).toContain("journey");

      const ruStandard = GalacticLexicon.info.streakActive(7, false, "ru");
      expect(ruStandard).toContain("7");
      expect(ruStandard).toContain("дней");

      const ruGalactic = GalacticLexicon.info.streakActive(7, true, "ru");
      expect(ruGalactic).toContain("7");
      expect(ruGalactic).toContain("путешествие");
    });
  });

  describe("warning messages", () => {
    it("should return unsaved changes message", () => {
      const enStandard = GalacticLexicon.warning.unsavedChanges(false, "en");
      expect(enStandard).toContain("unsaved changes");

      const enGalactic = GalacticLexicon.warning.unsavedChanges(true, "en");
      expect(enGalactic).toContain("ship log");

      const ruStandard = GalacticLexicon.warning.unsavedChanges(false, "ru");
      expect(ruStandard).toContain("несохраненные изменения");

      const ruGalactic = GalacticLexicon.warning.unsavedChanges(true, "ru");
      expect(ruGalactic).toContain("бортовом журнале");
    });

    it("should return delete confirm message with item name", () => {
      const enStandard = GalacticLexicon.warning.deleteConfirm("My Note", false, "en");
      expect(enStandard).toContain("My Note");
      expect(enStandard).toContain("delete");

      const enGalactic = GalacticLexicon.warning.deleteConfirm("My Note", true, "en");
      expect(enGalactic).toContain("My Note");
      expect(enGalactic).toContain("black hole");

      const ruStandard = GalacticLexicon.warning.deleteConfirm("My Note", false, "ru");
      expect(ruStandard).toContain("My Note");
      expect(ruStandard).toContain("удалить");

      const ruGalactic = GalacticLexicon.warning.deleteConfirm("My Note", true, "ru");
      expect(ruGalactic).toContain("My Note");
      expect(ruGalactic).toContain("чёрную дыру");
    });
  });
});

describe("message consistency", () => {
  it("should have matching keys in both modes", () => {
    const formatter = new MessageFormatter(false, "en");

    const successKeys = [
      "noteCreated",
      "noteUpdated",
      "noteDeleted",
      "linkCreated",
      "settingsSaved",
      "achievementUnlocked",
      "shareCreated",
      "loginSuccess",
    ];

    successKeys.forEach((key) => {
      const technical = formatter.format("success", key, "test");
      expect(technical).toBeTruthy();

      formatter.setGalacticMode(true);
      const galactic = formatter.format("success", key, "test");
      expect(galactic).toBeTruthy();
      expect(galactic).not.toEqual(technical);

      formatter.setGalacticMode(false);
    });
  });
});

describe("getLexiconMessage compatibility function", () => {
  it("should return English standard messages", () => {
    const msg = getLexiconMessage("en", "standard", "success", "noteCreated");
    expect(msg).toBeTruthy();
    expect(msg).toContain("Note");
  });

  it("should return Russian standard messages", () => {
    const msg = getLexiconMessage("ru", "standard", "success", "noteCreated");
    expect(msg).toBeTruthy();
    expect(msg).toContain("Заметка");
  });

  it("should return English galactic messages when mode is galactic", () => {
    const msg = getLexiconMessage("en", "galactic", "success", "noteCreated");
    expect(msg).toBeTruthy();
    expect(msg).toContain("Star");
    expect(msg).toContain("ignited");
  });

  it("should return Russian galactic messages when mode is galactic", () => {
    const msg = getLexiconMessage("ru", "galactic", "success", "noteCreated");
    expect(msg).toBeTruthy();
    expect(msg).toContain("Звезда");
    expect(msg).toContain("зажжена");
  });

  it("should return messages for error category", () => {
    const msg = getLexiconMessage("ru", "standard", "error", "connectionExists");
    expect(msg).toBeTruthy();
  });

  it("should return messages for achievement category", () => {
    const ru = getLexiconMessage("ru", "galactic", "achievement", "unlocked", "Test Achievement");
    expect(ru).toContain("Test Achievement");
    expect(ru).toContain("⭐");

    const en = getLexiconMessage("en", "galactic", "achievement", "unlocked", "Test Achievement");
    expect(en).toContain("Test Achievement");
    expect(en).toContain("⭐");
  });

  it("should handle fallback for unknown keys", () => {
    const msg = getLexiconMessage("en", "standard", "success", "unknownKey");
    expect(msg).toBeTruthy();
    expect(msg).toContain("unknownKey");
  });

  it("should handle locale parameter", () => {
    const en = getLexiconMessage("en", "standard", "success", "noteCreated");
    expect(en).toContain("Note");

    const ru = getLexiconMessage("ru", "standard", "success", "noteCreated");
    expect(ru).toContain("Заметка");
  });
});

describe("getMessageKeys helper", () => {
  it("should return all message keys", () => {
    const keys = getMessageKeys();

    expect(keys).toHaveProperty("success");
    expect(keys).toHaveProperty("error");
    expect(keys).toHaveProperty("info");
    expect(keys).toHaveProperty("warning");

    expect(Array.isArray(keys.success)).toBe(true);
    expect(Array.isArray(keys.error)).toBe(true);
    expect(Array.isArray(keys.info)).toBe(true);
    expect(Array.isArray(keys.warning)).toBe(true);

    expect(keys.success.length).toBeGreaterThan(0);
    expect(keys.error.length).toBeGreaterThan(0);
    expect(keys.info.length).toBeGreaterThan(0);
    expect(keys.warning.length).toBeGreaterThan(0);
  });

  it("should include expected keys in success category", () => {
    const keys = getMessageKeys();
    expect(keys.success).toContain("noteCreated");
    expect(keys.success).toContain("noteUpdated");
    expect(keys.success).toContain("achievementUnlocked");
  });

  it("should include expected keys in error category", () => {
    const keys = getMessageKeys();
    expect(keys.error).toContain("validation");
    expect(keys.error).toContain("duplicateLink");
    expect(keys.error).toContain("unauthorized");
  });
});

describe("Comprehensive message coverage", () => {
  it("covers every GalacticLexicon wrapper and message function for all modes/locales", () => {
    const getArgs = (fn: (...args: any[]) => string) => {
      const paramCount = fn.length; // message args before the default useGalactic/locale
      if (paramCount >= 2) return ["Source", "Target"];
      if (paramCount === 1) return ["Value"];
      return [];
    };

    const results: string[] = [];
    for (const [category, methods] of Object.entries(GalacticLexicon)) {
      for (const [key, fn] of Object.entries(methods)) {
        const args = getArgs(fn as (...args: any[]) => string);
        for (const useGalactic of [false, true]) {
          for (const locale of ["en", "ru"]) {
            const message = (fn as (...args: any[]) => string)(...args, useGalactic, locale);
            expect(message).toBeTypeOf("string");
            results.push(`${category}.${key}:${message}`);
          }
        }
      }
    }
    expect(results.length).toBeGreaterThan(100);
  });
});
