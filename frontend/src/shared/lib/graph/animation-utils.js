/**
 * Animation utilities for graph rendering
 */

/**
 * Get glow intensity based on time and node ID (pulsating effect)
 * @param {string} nodeId
 * @param {number} time
 * @param {number} nodeCount
 * @returns {number}
 */
export function getGlowIntensity(nodeId, time, nodeCount) {
  const PERFORMANCE_THRESHOLD_NODES = 100;
  
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

/**
 * String hash function for deterministic animations
 * @param {string} str
 * @returns {number}
 */
function stringHash(str) {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32bit integer
  }
  return Math.abs(hash);
}
