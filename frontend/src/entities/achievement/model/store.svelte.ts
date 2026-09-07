import { Achievement } from "./achievement";
import * as svc from "../api/service";
import { ACHIEVEMENT_POLL_INTERVAL_MS } from "$shared/config";
import { isAuthenticated } from "$shared/stores/auth.svelte.js";

interface AchievementsState {
  all: Achievement[];
  new: Achievement[];
}

export const state = $state<AchievementsState>({ all: [], new: [] });

let timer: ReturnType<typeof setInterval> | null = null;
let consecutiveErrors = 0;
const MAX_CONSECUTIVE_ERRORS = 5;

async function refresh() {
  if (!isAuthenticated()) return;
  try {
    const list = await svc.fetchUserAchievements();
    consecutiveErrors = 0;
    const newOnes = list.filter((a) => a.isNew());
    state.all = list;
    state.new = newOnes;
  } catch (e) {
    consecutiveErrors++;
    if (import.meta.env.DEV) {
      console.error("achievements refresh failed", e);
    }
    if (consecutiveErrors >= MAX_CONSECUTIVE_ERRORS) {
      if (import.meta.env.DEV) {
        console.warn("Stopping achievements polling after repeated failures");
      }
      stopPolling();
    }
  }
}

export function startPolling() {
  if (timer) return;
  consecutiveErrors = 0;
  refresh();
  // A poll interval of 0 (or negative) means polling is disabled via config,
  // not "poll as fast as possible" — setInterval(fn, 0) would otherwise fire
  // essentially continuously (browsers clamp to a few ms), hammering the API.
  if (ACHIEVEMENT_POLL_INTERVAL_MS > 0) {
    timer = setInterval(refresh, ACHIEVEMENT_POLL_INTERVAL_MS);
  }
}

export function stopPolling() {
  if (timer) clearInterval(timer);
  timer = null;
}

export function refreshNow() {
  return refresh();
}

export async function dismiss(id: string) {
  try {
    await svc.markAchievementSeen(id);
    await refresh();
  } catch (e) {
    if (import.meta.env.DEV) {
      console.error(e);
    }
  }
}
