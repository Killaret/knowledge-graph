/**
 * Simple FPS/performance monitor for canvas-based features.
 *
 * Can be used to throttle expensive updates or toggle quality presets.
 * Keeps only `requestAnimationFrame` / `performance` usage; no feature imports.
 */
export interface PerformanceMonitor {
  /** Average FPS over the last sample window. */
  fps: number;
  /** Estimated frame duration in milliseconds. */
  frameTimeMs: number;
  /** Call once per animation frame with a high-resolution timestamp. */
  tick(timestamp: number): void;
  /** Reset collected samples. */
  reset(): void;
}

const SAMPLE_COUNT = 30;

export function createPerformanceMonitor(): PerformanceMonitor {
  let samples: number[] = [];
  let lastTimestamp = 0;
  let averageFps = 60;
  let averageFrameTime = 16.67;

  function tick(timestamp: number) {
    if (lastTimestamp === 0) {
      lastTimestamp = timestamp;
      return;
    }

    const delta = timestamp - lastTimestamp;
    lastTimestamp = timestamp;

    if (delta <= 0) return;

    samples.push(delta);
    if (samples.length > SAMPLE_COUNT) {
      samples.shift();
    }

    const avgDelta = samples.reduce((sum, d) => sum + d, 0) / samples.length;
    averageFrameTime = avgDelta;
    averageFps = 1000 / avgDelta;
  }

  function reset() {
    samples = [];
    lastTimestamp = 0;
    averageFps = 60;
    averageFrameTime = 16.67;
  }

  return {
    get fps() {
      return averageFps;
    },
    get frameTimeMs() {
      return averageFrameTime;
    },
    tick,
    reset,
  };
}
