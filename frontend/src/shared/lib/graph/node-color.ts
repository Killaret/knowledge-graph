import { CelestialBody } from "$shared/lib/domain";

/**
 * Get color for a node type
 */
export function getNodeColor(type: string | undefined): string {
  return CelestialBody.fromString(type).color;
}
