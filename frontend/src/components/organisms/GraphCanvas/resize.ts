/**
 * Canvas resize utilities for GraphCanvas
 */
import type { ResizeState } from "./types";

export type { ResizeState };

/**
 * Resize canvas to fit parent container or window
 */
export function resizeCanvas(canvas: HTMLCanvasElement, state: ResizeState): void {
  const rect = canvas.parentElement?.getBoundingClientRect();
  if (rect && rect.width > 0 && rect.height > 0) {
    state.width = rect.width;
    state.height = rect.height;
    const dpr = Math.max(1, window.devicePixelRatio || 1);
    // Set CSS size
    canvas.style.width = `${Math.round(state.width)}px`;
    canvas.style.height = `${Math.round(state.height)}px`;
    // Set backing store size
    canvas.width = Math.round(state.width * dpr);
    canvas.height = Math.round(state.height * dpr);
    const ctx = canvas.getContext("2d");
    if (ctx) {
      // setTransform may not be available in test environments with mock canvas
      if (typeof ctx.setTransform === "function") {
        ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      } else {
        // Fallback for environments without setTransform support
        ctx.scale(dpr, dpr);
      }
    }
  } else {
    // Fallback: use window size if parent not available
    state.width = window.innerWidth;
    state.height = window.innerHeight - 80; // Account for controls
    const dpr = Math.max(1, window.devicePixelRatio || 1);
    canvas.style.width = `${Math.round(state.width)}px`;
    canvas.style.height = `${Math.round(state.height)}px`;
    canvas.width = Math.round(state.width * dpr);
    canvas.height = Math.round(state.height * dpr);
    const ctx = canvas.getContext("2d");
    if (ctx) {
      // setTransform may not be available in test environments with mock canvas
      if (typeof ctx.setTransform === "function") {
        ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      } else {
        // Fallback for environments without setTransform support
        ctx.scale(dpr, dpr);
      }
    }
  }
}

/**
 * Setup ResizeObserver for the canvas
 */
export function setupResizeObserver(
  canvas: HTMLCanvasElement,
  onResize: () => void
): { disconnect: () => void } {
  const resizeObserver = new ResizeObserver(() => {
    onResize();
  });

  if (canvas.parentElement) {
    resizeObserver.observe(canvas.parentElement);
  }

  return {
    disconnect: () => {
      resizeObserver.disconnect();
    },
  };
}

/**
 * Delayed resize to ensure container has settled dimensions
 */
export function scheduleDelayedResize(
  callback: () => void,
  delayMs: number = 100
): { clear: () => void } {
  const timer = setTimeout(callback, delayMs);

  return {
    clear: () => {
      clearTimeout(timer);
    },
  };
}
