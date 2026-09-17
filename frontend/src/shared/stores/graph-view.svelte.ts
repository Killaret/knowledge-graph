import { browser } from "$app/environment";
import { accessToken, currentUser, getApiKey, isAuthenticated } from "./auth-session.svelte";

export type GraphViewMode = "personal" | "community";
const STORAGE_KEY = "graph-view-mode";

function readPreference(): GraphViewMode | null {
  if (!browser) return null;
  try {
    const value = localStorage.getItem(STORAGE_KEY);
    return value === "personal" || value === "community" ? value : null;
  } catch {
    return null;
  }
}

let preference = $state<GraphViewMode | null>(readPreference());
let revision = $state(0);

export const graphView = {
  get mode(): GraphViewMode {
    return isAuthenticated() ? (preference ?? "personal") : "community";
  },
  set mode(value: GraphViewMode) {
    if (!isAuthenticated() || (value !== "personal" && value !== "community")) return;
    if (preference === value) return;
    preference = value;
    revision += 1;
    if (browser) {
      try {
        localStorage.setItem(STORAGE_KEY, value);
      } catch {
        return;
      }
    }
  },
  get scopeKey(): string {
    const identity = isAuthenticated()
      ? (currentUser()?.id ?? accessToken() ?? getApiKey() ?? "test-session")
      : "anonymous";
    return `${this.mode}:${identity}:${revision}`;
  },
  restore(): void {
    preference = readPreference();
    revision += 1;
  },
  clear(): void {
    preference = null;
    revision += 1;
  },
};
