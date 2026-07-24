import { describe, it, expect } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "../../../vitest-setup";
import { fetchAllAchievements, fetchUserAchievements, markAchievementSeen } from "./achievements";
import { Achievement } from "$entities";

const baseUrl = "http://localhost:8080/api";

describe("achievements service", () => {
  const apiData = {
    id: "a1",
    code: "first",
    title: "First Step",
    name_en: "First Step",
    description: "Create your first note",
    points: 10,
    earned: true,
    unlocked_at: "2024-01-01T00:00:00Z",
  };

  it("fetchAllAchievements maps API data to Achievement instances", async () => {
    server.use(
      http.get(`${baseUrl}/v1/achievements`, () => HttpResponse.json({ achievements: [apiData] }))
    );

    const result = await fetchAllAchievements();
    expect(result).toHaveLength(1);
    expect(result[0]).toBeInstanceOf(Achievement);
    expect(result[0].code).toBe("first");
  });

  it("fetchUserAchievements maps API data to Achievement instances", async () => {
    server.use(
      http.get(`${baseUrl}/v1/users/me/achievements`, () =>
        HttpResponse.json({ achievements: [apiData] })
      )
    );

    const result = await fetchUserAchievements();
    expect(result).toHaveLength(1);
    expect(result[0]).toBeInstanceOf(Achievement);
  });

  it("markAchievementSeen posts to the mark-seen endpoint", async () => {
    server.use(
      http.post(`${baseUrl}/v1/users/me/achievements/a1/mark-seen`, () =>
        HttpResponse.json({}, { status: 204 })
      )
    );

    await expect(markAchievementSeen("a1")).resolves.toBeUndefined();
  });
});
