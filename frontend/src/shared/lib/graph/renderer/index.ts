/**
 * Graph renderer main exports
 */
import { graphConfig2D } from '$lib/config';
import type { SimulationNode, SimulationLink } from '$lib/components/GraphCanvas/types';

// Re-export types
export type { SimulationNode, SimulationLink };

// Re-export utils
export * from './utils';

// Re-export node renderers
export * from './nodes';

// Re-export anomaly renderers
export * from './anomalies';

// Performance thresholds
const PERFORMANCE_THRESHOLD_NODES = 100;
const PERFORMANCE_THRESHOLD_LINKS = 50;

/**
 * Draw all links with animation and hover effects
 */
export function drawAllLinks(
  ctx: CanvasRenderingContext2D,
  links: SimulationLink[],
  nodes: SimulationNode[],
  linkOpacity: Map<string, number>,
  animationTime: number,
  hoveredNodeId: string | null,
  highlightedLinkId: string | null
): void {
  // Implementation stays in renderer.ts for now
  // Will be migrated in next phase
}

/**
 * Draw a single node
 */
export function drawNode(
  ctx: CanvasRenderingContext2D,
  node: SimulationNode,
  angle: number,
  variation?: { sizeMultiplier: number; hueShift: number },
  nodeId?: string,
  nodeCount?: number,
  time?: number,
  disableVariation?: boolean
): void {
  // Implementation stays in renderer.ts for now
  // Will be migrated in next phase
}

/**
 * Draw node title
 */
export function drawNodeTitle(
  ctx: CanvasRenderingContext2D,
  node: SimulationNode,
  x: number,
  y: number,
  r: number
): void {
  // Implementation stays in renderer.ts for now
  // Will be migrated in next phase
}

/**
 * Draw all nodes
 */
export function drawAllNodes(
  ctx: CanvasRenderingContext2D,
  nodes: SimulationNode[],
  angles: Map<string, number>,
  enableShadows: boolean,
  nodeOpacity: Map<string, number>,
  disableVariation: boolean,
  animationTime: number,
  hoveredNodeId: string | null,
  particleSystem: any,
  focusMode: boolean,
  searchMatchIds: string[]
): void {
  // Implementation stays in renderer.ts for now
  // Will be migrated in next phase
}

/**
 * Draw unknown type node (dispatcher for anomalies)
 */
export function drawUnknown(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  nodeId: string,
  customRenderers?: Record<number, import('./anomalies').AnomalyRenderer>
): void {
  const { getAnomalyParams } = import('./anomalies');
  const params = getAnomalyParams(nodeId);
  const renderers = customRenderers ?? {
    0: import('./anomalies').drawRealityRift,
    1: import('./anomalies').drawChromaticMaw,
    2: import('./anomalies').drawVoidWhisper,
    3: import('./anomalies').drawCosmicAbomination,
  } as Record<number, import('./anomalies').AnomalyRenderer>;

  const hash = (globalThis as any).stringHash ? (globalThis as any).stringHash(nodeId) : 0;
  const anomalyType = hash % 4;
  const rendererFn = renderers[anomalyType] ?? renderers[0];
  rendererFn(ctx, x, y, r, params);
}

/**
 * Draw background
 */
export function drawBackground(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  particleSystem: any,
  animationTime: number,
  disableVariation: boolean
): void {
  // Implementation stays in renderer.ts for now
  // Will be migrated in next phase
}

/**
 * Draw animated link
 */
export function drawAnimatedLink(
  ctx: CanvasRenderingContext2D,
  source: { x: number; y: number },
  target: { x: number; y: number },
  weight: number,
  linkType: string,
  fadeOpacity: number,
  animationTime: number
): void {
  // Implementation stays in renderer.ts for now
  // Will be migrated in next phase
}

/**
 * Draw link
 */
export function drawLink(
  ctx: CanvasRenderingContext2D,
  source: { x: number; y: number },
  target: { x: number; y: number },
  weight: number,
  linkType: string,
  fadeOpacity: number
): void {
  // Implementation stays in renderer.ts for now
  // Will be migrated in next phase
}

/**
 * Main draw function
 */
export function draw(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  nodes: SimulationNode[],
  simLinks: SimulationLink[],
  angles: Map<string, number>,
  transform: { x: number; y: number; k: number },
  linkOpacity: Map<string, number>,
  nodeOpacity: Map<string, number>,
  animationTime: number,
  hoveredNodeId: string | null,
  highlightedLinkId: string | null,
  particleSystem: any,
  blackHole: any,
  ghostNode: any,
  gravitySystem: any,
  disableVariation: boolean,
  focusMode: boolean,
  searchMatchIds: string[]
): void {
  // Implementation stays in renderer.ts for now
  // Will be migrated in next phase
}

/**
 * Reset view to center the graph
 */
export function resetView(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  nodes: SimulationNode[],
  transform: { x: number; y: number; k: number }
): void {
  // Implementation stays in renderer.ts for now
  // Will be migrated in next phase
}
