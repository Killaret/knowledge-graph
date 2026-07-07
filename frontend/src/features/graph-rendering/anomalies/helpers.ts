/**
 * Anomaly helpers
 */
import { graphConfig2D, anomalyConfig } from '$shared/config';
import { stringHash } from '$shared/lib/graph/helpers';

/**
 * Seeded random number generator for deterministic anomaly parameters
 */
export function seededRandom(seed: number): number {
  const x = Math.sin(seed) * 10000;
  return x - Math.floor(x);
}

/**
 * Generate deterministic parameters for anomaly visualization
 */
export interface AnomalyParams {
  crackCount: number;
  tentacleCount: number;
  particleCount: number;
  colorShift1: number;
  colorShift2: number;
  deformAmount: number;
  rotationOffset: number;
  seedBase: number;
}

export type AnomalyRenderer = (
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  params: AnomalyParams
) => void;

export function getAnomalyParams(nodeId: string): AnomalyParams {
  const hash = stringHash(nodeId);
  
  // Use different parts of hash for different parameters
  const hash1 = hash % 1000;
  const hash2 = (hash >> 10) % 1000;
  const hash3 = (hash >> 20) % 1000;
  const hash4 = (hash >> 25) % 1000;
  const hash5 = (hash >> 30) % 1000;
  
  const rr = anomalyConfig.reality_rift;
  const cm = anomalyConfig.chromatic_maw;
  const vw = anomalyConfig.void_whisper;
  const ca = anomalyConfig.cosmic_abomination;
  
  return {
    crackCount: ca.crack_count_min + Math.floor((hash1 / 1000) * (ca.crack_count_max - ca.crack_count_min)),
    tentacleCount: ca.tentacle_count_min + Math.floor((hash2 / 1000) * (ca.tentacle_count_max - ca.tentacle_count_min)),
    particleCount: ca.particle_count_min + Math.floor((hash3 / 1000) * (ca.particle_count_max - ca.particle_count_min)),
    colorShift1: (hash4 / 1000) * cm.hue_shift_range,
    colorShift2: (hash5 / 1000) * vw.hue_shift_range,
    deformAmount: rr.deform_amount_min + (hash1 / 1000) * (rr.deform_amount_max - rr.deform_amount_min),
    rotationOffset: (hash2 / 1000) * Math.PI * 2,
    seedBase: hash,
  };
}
