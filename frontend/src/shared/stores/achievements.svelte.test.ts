import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { Achievement } from "$shared/lib/domain";
import {
  state,
  startPolling,
  stopPolling,
  refreshNow,
  dismiss,
} from "./achievements.svelte";
import { ACHIEVEMENT_POLL_INTERVAL_MS } from "$shared/config";

vi.mock("./auth.svelte", () => ({
  isAuthenticated: vi.fn(() => true),
}));

vi.mock("$shared/services/achievements", () => ({
  fetchUserAchievements: vi.fn(),
  markAchievementSeen: vi.fn(),
}));

const auth = await import("./auth.svelte");
const svc = await import("$shared/services/achievements");

describe("achievements store", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.clearAllMocks();

    // Reset internal timer and consecutive errors without calling the API.
    vi.mocked(auth.isAuthenticated).mockReturnValue(false);
    startPolling();
    stopPolling();

    vi.mocked(auth.isAuthenticated).mockReturnValue(true);
    vi.mocked(svc.fetchUserAchievements).mockResolvedValue([]);
    vi.mocked(svc.markAchievementSeen).mockResolvedValue(undefined);

    // Reset state
    state.all = [];
    state.new = [];
  });

  afterEach(() => {
    stopPolling();
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it("refreshNow loads achievements and extracts new ones", async () => {
    const achievement = Achievement.fromApi({
      id: "a1",
      code: "first",
      title: "First",
      description: "",
      icon: "",
      points: 10,
      earned: true,
      unlocked_at: "2024-01-01T00:00:00Z",
      notification_seen: false,
    });

    vi.mocked(svc.fetchUserAchievements).mockResolvedValue([achievement]);

    await refreshNow();

    expect(state.all).toHaveLength(1);
    expect(state.new).toHaveLength(1);
    expect(state.new[0].id).toBe("a1");
  });

  it("does not refresh when user is not authenticated", async () => {
    vi.mocked(auth.isAuthenticated).mockReturnValue(false);
    await refreshNow();
    expect(svc.fetchUserAchievements).not.toHaveBeenCalled();
  });

  it("polls achievements on an interval", async () => {
    await startPolling();
    expect(svc.fetchUserAchievements).toHaveBeenCalledTimes(1);

    vi.advanceTimersByTime(ACHIEVEMENT_POLL_INTERVAL_MS);
    expect(svc.fetchUserAchievements).toHaveBeenCalledTimes(2);

    vi.advanceTimersByTime(ACHIEVEMENT_POLL_INTERVAL_MS);
    expect(svc.fetchUserAchievements).toHaveBeenCalledTimes(3);

    stopPolling();
  });

  it("dismiss marks an achievement seen and refreshes", async () => {
    const achievement = Achievement.fromApi({
      id: "a1",
      code: "first",
      title: "First",
      description: "",
      icon: "",
      points: 10,
      earned: true,
      unlocked_at: "2024-01-01T00:00:00Z",
      notification_seen: false,
    });

    vi.mocked(svc.fetchUserAchievements).mockResolvedValue([achievement]);
    vi.mocked(svc.markAchievementSeen).mockResolvedValue(undefined);

    await dismiss("a1");

    expect(svc.markAchievementSeen).toHaveBeenCalledWith("a1");
    expect(svc.fetchUserAchievements).toHaveBeenCalled();
  });
});
