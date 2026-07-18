/**
 * Line dash pattern utility
 */

import { LinkType } from "$shared/lib/domain";

/**
 * Get line dash pattern based on link type and weight.
 * Kept as a thin adapter for callers that still pass a raw link type string.
 */
export function getLineDash(linkType?: string, weight?: number): number[] {
  return LinkType.fromString(linkType).getLineDash(weight);
}
