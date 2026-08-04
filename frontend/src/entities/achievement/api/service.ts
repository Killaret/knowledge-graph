import api from "$shared/api/client";
import { Achievement, type AchievementApiData } from "../model/achievement";

export type { AchievementApiData } from "../model/achievement";
export { Achievement } from "../model/achievement";

export async function fetchAllAchievements(): Promise<Achievement[]> {
  const data = await api.get("v1/achievements").json<{ achievements: AchievementApiData[] }>();
  return (data.achievements ?? []).map((a) => Achievement.fromApi(a));
}

export async function fetchUserAchievements(): Promise<Achievement[]> {
  const data = await api
    .get("v1/users/me/achievements")
    .json<{ achievements: AchievementApiData[] }>();
  return (data.achievements ?? []).map((a) => Achievement.fromApi(a));
}

export async function markAchievementSeen(id: string): Promise<void> {
  await api.post(`v1/users/me/achievements/${id}/mark-seen`);
}
