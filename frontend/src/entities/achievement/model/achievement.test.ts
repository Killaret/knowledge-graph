import { describe, it, expect } from "vitest";
import { Achievement } from "./achievement";

describe("Achievement", () => {
  it("normalizes the public API shape", () => {
    const a = Achievement.fromApi({
      id: "a1",
      code: "first",
      title: "Explorer",
      description: "Create your first note",
      icon: "⭐",
      points: 10,
      earned: true,
      is_hidden: false,
    });

    expect(a.id).toBe("a1");
    expect(a.title).toBe("Explorer");
    expect(a.description).toBe("Create your first note");
    expect(a.icon).toBe("⭐");
    expect(a.points).toBe(10);
    expect(a.isUnlocked()).toBe(true);
    expect(a.isNew()).toBe(true);
  });

  it("normalizes the legacy localized API shape", () => {
    const a = Achievement.fromApi({
      code: "legacy",
      name_en: "Legacy",
      name_ru: "Наследие",
      description_en: "English desc",
      description_ru: "Русское описание",
      icon_emoji: "🚀",
      unlocked_at: "2024-01-01T00:00:00Z",
      notification_seen: false,
    });

    expect(a.title).toBe("Legacy");
    expect(a.titleRu).toBe("Наследие");
    expect(a.getTitle("ru")).toBe("Наследие");
    expect(a.icon).toBe("🚀");
    expect(a.isUnlocked()).toBe(true);
    expect(a.obtainedAt).toBe("2024-01-01T00:00:00Z");
  });

  it("detects obtained_at as unlocked", () => {
    const a = Achievement.fromApi({
      code: "obtained",
      title: "Obtained",
      obtained_at: "2024-02-01T00:00:00Z",
    });
    expect(a.isUnlocked()).toBe(true);
    expect(a.isNew()).toBe(true);
  });

  it("marks an achievement as seen", () => {
    const a = Achievement.fromApi({
      code: "new",
      title: "New",
      earned: true,
      notification_seen: false,
    });
    expect(a.isNew()).toBe(true);
    const seen = a.markSeen();
    expect(seen.isNew()).toBe(false);
    expect(seen.notificationSeen).toBe(true);
  });

  it("falls back to defaults for missing fields", () => {
    const a = Achievement.fromApi({ code: "minimal" });
    expect(a.title).toBe("minimal");
    expect(a.icon).toBe("🏆");
    expect(a.points).toBe(0);
    expect(a.earned).toBe(false);
  });
});
