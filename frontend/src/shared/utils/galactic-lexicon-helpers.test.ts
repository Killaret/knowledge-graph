import { describe, it, expect } from "vitest";
import { GalacticLexicon } from "./galactic-lexicon";

const argMap: Record<string, unknown[]> = {
  noteCreated: ["My Note"],
  noteUpdated: ["My Note"],
  linkCreated: ["Source", "Target"],
  achievementUnlocked: ["Explorer"],
  validation: ["title"],
  achievementProgress: [3, 10],
  streakActive: [7],
  newFeature: ["Dark mode"],
  deleteConfirm: ["My Note"],
};

describe("GalacticLexicon helpers", () => {
  for (const [category, methods] of Object.entries(GalacticLexicon)) {
    for (const [key, fn] of Object.entries(methods as Record<string, unknown>)) {
      it(`${category}.${key} returns a localized string`, () => {
        const args = argMap[key] ?? [];
        const result = (fn as (...args: unknown[]) => string)(...args);
        expect(typeof result).toBe("string");
        expect(result.length).toBeGreaterThan(0);
      });
    }
  }
});
