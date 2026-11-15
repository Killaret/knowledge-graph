/**
 * GRAPH-LIGHT-1, decision 81: recommendations of the hovered note drawn as a
 * third kind of line — pale dashes of their own colour, denser for a closer
 * match, only while hovering. Notes already linked to the hovered one, and
 * notes missing from the graph (deleted, filtered out), are skipped.
 */
import type { SimulationNode } from "../types";
import { lightFrame } from "./style";
import { mixRgb, rgba, RECOMMENDATION, WHITE, lightTypeColor, type Rgb } from "./palette";

/** Dash and gap in world units: a closer match gets denser dashes. */
export function recommendationDash(score: number): [number, number] {
  const q = Math.min(1, Math.max(0, (score - 0.35) / 0.5));
  return [3.2 / lightFrame.k, (2.5 + (1 - q) * 9) / lightFrame.k];
}

/** Ids of the hovered note's recommendations that are drawn this frame. */
export function activeRecommendationIds(hoveredId: string | null): Set<string> {
  if (!hoveredId || lightFrame.recommendationsFor !== hoveredId) return new Set();
  return new Set(lightFrame.recommendations.map((r) => r.id).filter((id) => id !== hoveredId));
}

export function drawLightRecommendations(
  ctx: CanvasRenderingContext2D,
  hovered: SimulationNode,
  nodeMap: Map<string, SimulationNode>,
  neighborIds?: Set<string>
): number {
  if (lightFrame.recommendationsFor !== hovered.id || hovered.x == null || hovered.y == null) {
    return 0;
  }
  const alpha = 0.75 * lightFrame.focusMix;
  if (alpha <= 0.01) return 0;

  let drawn = 0;
  ctx.save();
  ctx.lineCap = "round";
  ctx.lineWidth = 1.1 / lightFrame.k;
  ctx.strokeStyle = rgba(RECOMMENDATION, alpha);
  for (const rec of lightFrame.recommendations) {
    const target = nodeMap.get(rec.id);
    if (!target || target.id === hovered.id || neighborIds?.has(target.id)) continue;
    if (target.x == null || target.y == null) continue;
    ctx.setLineDash(recommendationDash(rec.score));
    ctx.beginPath();
    ctx.moveTo(hovered.x, hovered.y);
    ctx.lineTo(target.x, target.y);
    ctx.stroke();
    drawn++;
  }
  ctx.setLineDash([]);
  ctx.restore();
  return drawn;
}

/** Thin ring around the selected note with a slow outward pulse. */
export function drawSelectionRing(
  ctx: CanvasRenderingContext2D,
  node: SimulationNode,
  coreRadius: number
): void {
  if (node.x == null || node.y == null) return;
  const k = lightFrame.k;
  const ring = Math.max(coreRadius * 2.3, 8 / k);
  const color: Rgb = mixRgb(lightTypeColor(node.type), WHITE, 0.5);
  ctx.save();
  ctx.lineWidth = 1.2 / k;
  ctx.strokeStyle = rgba(color, 0.7);
  ctx.beginPath();
  ctx.arc(node.x, node.y, ring, 0, Math.PI * 2);
  ctx.stroke();
  if (lightFrame.ambient && !lightFrame.stable) {
    const phase = ((lightFrame.time / 1000) * 0.55) % 1;
    ctx.strokeStyle = rgba(color, (1 - phase) * 0.7);
    ctx.beginPath();
    ctx.arc(node.x, node.y, ring + (phase * 16) / k, 0, Math.PI * 2);
    ctx.stroke();
  }
  ctx.restore();
}
