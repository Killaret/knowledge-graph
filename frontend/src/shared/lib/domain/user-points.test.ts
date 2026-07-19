import { describe, it, expect } from "vitest";
import { UserPoints } from "./user-points";
import { Achievement } from "./achievement";

describe("UserPoints", () => {
  const makeAchievement = (overrides: Partial<{
      id: string;
      code: string;
      title: string;
      titleRu: string;
      description: string;
      descriptionRu: string;
      icon: string;
      points: number;
      earned: boolean;
      hidden: boolean;
      obtainedAt: string | null;
      notificationSeen: boolean;
    }> = {}) =>
    new Achievement({
      id: "1",
      code: "code",
      title: "Title",
      titleRu: "Title",
      description: "Description",
      descriptionRu: "Description",
      icon: "🏆",
      points: 10,
      earned: false,
      hidden: false,
      obtainedAt: null,
      notificationSeen: false,
      ...overrides,
    });

  it("computes total from earned achievements", () => {
    const up = new UserPoints([
      makeAchievement({ earned: true, points: 10 }),
      makeAchievement({ earned: true, points: 25 }),
      makeAchievement({ earned: false, points: 50 }),
    ]);
    expect(up.computedTotal).toBe(35);
    expect(up.total).toBe(35);
    expect(up.earnedCount).toBe(2);
    expect(up.totalCount).toBe(3);
    expect(up.isConsistent).toBe(true);
  });

  it("prefers reported total over computed total", () => {
    const up = new UserPoints(
      [makeAchievement({ earned: true, points: 10 })],
      100,
    );
    expect(up.total).toBe(100);
    expect(up.isConsistent).toBe(false);
  });

  it("considers a missing reported total consistent", () => {
    const up = new UserPoints([makeAchievement({ earned: true, points: 10 })]);
    expect(up.total).toBe(10);
    expect(up.isConsistent).toBe(true);
  });

  it("calculates level and points to next level", () => {
    const up = new UserPoints([], 49);
    expect(up.level).toBe(1);
    expect(up.pointsToNextLevel).toBe(1);

    const level2 = new UserPoints([], 50);
    expect(level2.level).toBe(2);
    expect(level2.pointsToNextLevel).toBe(100);
  });

  it("creates from API data", () => {
    const up = UserPoints.fromApi({
      achievements: [
        { id: "1", code: "a", title: "A", points: 10, earned: true },
        { id: "2", code: "b", title: "B", points: 25, earned: false },
      ],
      total_points: 10,
    });
    expect(up.total).toBe(10);
    expect(up.computedTotal).toBe(10);
    expect(up.isConsistent).toBe(true);
  });

  it("provides an empty instance", () => {
    const up = UserPoints.empty();
    expect(up.total).toBe(0);
    expect(up.earnedCount).toBe(0);
  });
});
