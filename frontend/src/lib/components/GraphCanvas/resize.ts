/**
 * Canvas resize utilities for GraphCanvas
 */
import type { ResizeState } from './types';

export type { ResizeState };

/**
 * Resize canvas to fit parent container or window
 */
export function resizeCanvas(
  canvas: HTMLCanvasElement,
  state: ResizeState
): void {
  const rect = canvas.parentElement?.getBoundingClientRect();
  if (rect && rect.width > 0 && rect.height > 0) {
    state.width = rect.width;
    state.height = rect.height;
    canvas.width = state.width;
    canvas.height = state.height;
  } else {
    // Fallback: use window size if parent not available
    state.width = window.innerWidth;
    state.height = window.innerHeight - 80; // Account for controls
    canvas.width = state.width;
    canvas.height = state.height;
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
    }
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
    }
  };
}
