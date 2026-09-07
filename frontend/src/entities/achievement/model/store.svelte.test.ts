import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { Achievement } from "./achievement";
import { state, startPolling, stopPolling, refreshNow, dismiss } from "./store.svelte";

// Use a fixed, positive interval for these tests so behavior doesn't depend
// on the real project config (which currently sets poll_interval_ms to 0 to
// disable polling entirely — see the dedicated test below for that case).
const ACHIEVEMENT_POLL_INTERVAL_MS = 5000;
vi.mock("$shared/config", async (importOriginal) => {
  const actual = await importOriginal<typeof import("$shared/config")>();
  return {
    ...actual,
    ACHIEVEMENT_POLL_INTERVAL_MS: 5000,
  };
});

vi.mock("$shared/stores/auth.svelte.js", () => ({
  isAuthenticated: vi.fn(() => true),
}));

vi.mock("../api/service", () => ({
  fetchUserAchievements: vi.fn(),
  markAchievementSeen: vi.fn(),
}));

const auth = await import("$shared/stores/auth.svelte.js");
const svc = await import("../api/service");

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

// Regression test: a poll interval of 0 previously caused setInterval(fn, 0)
// to be scheduled, which fires essentially continuously (browsers clamp to a
// few ms) instead of disabling polling — hammering the API hundreds of times
// per second. Verified via a real Playwright run against /graph: ~125
// requests/sec to /api/v1/users/me/achievements before the fix.
describe("achievements store with polling disabled (poll_interval_ms: 0)", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.doUnmock("$shared/config");
    vi.resetModules();
  });

  it("does not schedule a recurring timer when the configured interval is 0", async () => {
    vi.doMock("$shared/config", () => ({ ACHIEVEMENT_POLL_INTERVAL_MS: 0 }));
    vi.doMock("$shared/stores/auth.svelte.js", () => ({ isAuthenticated: vi.fn(() => true) }));
    vi.doMock("../api/service", () => ({
      fetchUserAchievements: vi.fn().mockResolvedValue([]),
      markAchievementSeen: vi.fn(),
    }));

    const store = await import("./store.svelte");
    const svc = await import("../api/service");

    store.startPolling();
    // Flush the initial refresh() call triggered synchronously by startPolling.
    await vi.runOnlyPendingTimersAsync();
    expect(svc.fetchUserAchievements).toHaveBeenCalledTimes(1);

    // Advancing time should not trigger any further calls since no interval
    // was scheduled (a setInterval(fn, 0) bug would fire many more here).
    await vi.advanceTimersByTimeAsync(10000);
    expect(svc.fetchUserAchievements).toHaveBeenCalledTimes(1);

    store.stopPolling();
  });
});
