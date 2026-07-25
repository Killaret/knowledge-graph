import { describe, it, expect } from "vitest";
import { createPerformanceMonitor } from "./performance-monitor";

describe("createPerformanceMonitor", () => {
  it("tracks average frame time from tick samples", () => {
    const monitor = createPerformanceMonitor();
    monitor.tick(100);
    monitor.tick(110);
    monitor.tick(120);

    expect(monitor.frameTimeMs).toBeCloseTo(10, 1);
    expect(monitor.fps).toBeCloseTo(100, 1);
  });

  it("resets collected samples", () => {
    const monitor = createPerformanceMonitor();
    monitor.tick(100);
    monitor.tick(110);
    monitor.reset();
    expect(monitor.fps).toBe(60);
    expect(monitor.frameTimeMs).toBeCloseTo(16.67, 1);
  });

  it("ignores the first tick because there is no previous timestamp", () => {
    const monitor = createPerformanceMonitor();
    monitor.tick(100);
    expect(monitor.fps).toBe(60);
  });
});
