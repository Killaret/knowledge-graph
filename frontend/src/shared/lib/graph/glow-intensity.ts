/**
 * Glow intensity utility for node pulsating effect
 */
import { graphConfig2D } from "$shared/config";
import { stringHash } from "./helpers";

/**
 * Get glow intensity based on time and node ID (pulsating effect).
 * Above the configured visual-fx threshold the glow is reduced to a
 * steady minimal value to keep the frame rate healthy on large graphs.
 */
export function getGlowIntensity(nodeId: string, time: number, nodeCount: number): number {
  const threshold = graphConfig2D.visual_fx_threshold ?? 500;
  if (nodeCount > threshold) {
    return 0.3; // Minimal steady glow
  }

  const hash = stringHash(nodeId);
  const phase = (hash % 1000) / 1000;
  const period = 2000 + (hash % 1000);
  const t = (time + phase * period) % period;
  const normalizedT = t / period;

  // Sine wave from 0.3 to 1.0 (absolute value to ensure positive)
  return 0.3 + 0.7 * Math.abs(Math.sin(normalizedT * Math.PI * 2));
}
