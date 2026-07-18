/**
 * Glow intensity utility for node pulsating effect
 */
import { stringHash } from "./helpers";

const PERFORMANCE_THRESHOLD_NODES = 100;

/**
 * Get glow intensity based on time and node ID (pulsating effect)
 */
export function getGlowIntensity(
  nodeId: string,
  time: number,
  nodeCount: number,
): number {
  if (nodeCount > PERFORMANCE_THRESHOLD_NODES) {
    return 0.3; // Minimal glow for stars only
  }

  const hash = stringHash(nodeId);
  const phase = (hash % 1000) / 1000;
  const period = 2000 + (hash % 1000);
  const t = (time + phase * period) % period;
  const normalizedT = t / period;

  // Sine wave from 0.3 to 1.0 (absolute value to ensure positive)
  return 0.3 + 0.7 * Math.abs(Math.sin(normalizedT * Math.PI * 2));
}
