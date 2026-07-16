import api from '$shared/api/client';

export type Achievement = {
  id: string
  code: string
  name_ru?: string
  name_en?: string
  description_ru?: string
  description_en?: string
  icon_emoji?: string
  unlocked_at?: string | null
  notification_seen?: boolean
}

export async function fetchAllAchievements(): Promise<Achievement[]> {
  const data = await api.get('v1/achievements').json<{ achievements: Achievement[] }>();
  return data.achievements ?? []
}

export async function fetchUserAchievements(): Promise<Achievement[]> {
  const data = await api.get('v1/users/me/achievements').json<{ achievements: Achievement[] }>();
  return data.achievements ?? []
}

export async function markAchievementSeen(id: string): Promise<void> {
  await api.post(`v1/users/me/achievements/${id}/mark-seen`);
}
