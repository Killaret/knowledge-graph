const API_BASE = '/api/v1'

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
  const res = await fetch(`${API_BASE}/achievements`, { credentials: 'same-origin' })
  if (!res.ok) throw new Error('Failed to fetch achievements')
  const data = await res.json()
  return data.achievements ?? []
}

export async function fetchUserAchievements(): Promise<Achievement[]> {
  const res = await fetch(`${API_BASE}/users/me/achievements`, { credentials: 'same-origin' })
  if (!res.ok) throw new Error('Failed to fetch user achievements')
  const data = await res.json()
  return data.achievements ?? []
}

export async function markAchievementSeen(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/users/me/achievements/${id}/mark-seen`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'same-origin'
  })
  if (!res.ok) throw new Error('Failed to mark achievement seen')
}
