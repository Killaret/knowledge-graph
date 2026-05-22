import { writable } from 'svelte/store'
import type { Achievement } from '../services/achievements'
import * as svc from '../services/achievements'
import { ACHIEVEMENT_POLL_INTERVAL_MS } from '../config'

function createAchievementsStore() {
  const { subscribe, set, update } = writable<{ all: Achievement[]; new: Achievement[] }>({ all: [], new: [] })
  let timer: any = null

  async function refresh() {
    try {
      const list = await svc.fetchUserAchievements()
      // new items: notification_seen === false
      const newOnes = list.filter(i => !i.notification_seen)
      set({ all: list, new: newOnes })
    } catch (e) {
      console.error('achievements refresh failed', e)
    }
  }

  function startPolling() {
    if (timer) return
    refresh()
    timer = setInterval(refresh, ACHIEVEMENT_POLL_INTERVAL_MS)
  }

  function stopPolling() {
    if (timer) clearInterval(timer)
    timer = null
  }

  function refreshNow() { return refresh() }

  async function dismiss(id: string) {
    try {
      await svc.markAchievementSeen(id)
      await refresh()
    } catch (e) { console.error(e) }
  }

  return { subscribe, startPolling, stopPolling, refreshNow, dismiss }
}

export const achievementsStore = createAchievementsStore()
