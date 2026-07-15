import { writable } from 'svelte/store'
import type { Achievement } from '../services/achievements'
import * as svc from '../services/achievements'
import { ACHIEVEMENT_POLL_INTERVAL_MS } from '../config'
import { isAuthenticated } from './auth.svelte'

function createAchievementsStore() {
  const { subscribe, set } = writable<{ all: Achievement[]; new: Achievement[] }>({ all: [], new: [] })
  let timer: any = null
  let consecutiveErrors = 0
  const MAX_CONSECUTIVE_ERRORS = 5

  async function refresh() {
    if (!isAuthenticated()) return
    try {
      const list = await svc.fetchUserAchievements()
      consecutiveErrors = 0
      // new items: notification_seen === false
      const newOnes = list.filter(i => !i.notification_seen)
      set({ all: list, new: newOnes })
    } catch (e) {
      consecutiveErrors++
      console.error('achievements refresh failed', e)
      if (consecutiveErrors >= MAX_CONSECUTIVE_ERRORS) {
        console.warn('Stopping achievements polling after repeated failures')
        stopPolling()
      }
    }
  }

  function startPolling() {
    if (timer) return
    consecutiveErrors = 0
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
