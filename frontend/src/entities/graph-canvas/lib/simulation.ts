/**
 * D3-force simulation management for GraphCanvas
 */
import * as d3Force from "d3-force";
import { filterValidLinks } from "$shared/utils/graphUtils";
import { getLinkEndpointId } from "./types";
import type { SimulationNode, SimulationLink, SimulationState, TransformState } from "./types";

export type { SimulationNode, SimulationLink, SimulationState, TransformState };

function getLinkId(link: SimulationLink): string {
  if (link.id) return link.id;
  const sourceId = getLinkEndpointId(link.source);
  const targetId = getLinkEndpointId(link.target);
  return `${sourceId}-${targetId}-${link.link_type ?? "related"}`;
}

// Easing function for smooth fade animation
function easeOutCubic(t: number): number {
  return 1 - Math.pow(1 - t, 3);
}

function anyOpacityBelowOne(state: SimulationState): boolean {
  for (const value of state.nodeOpacity.values()) {
    if (value < 0.999) return true;
  }
  for (const value of state.linkOpacity.values()) {
    if (value < 0.999) return true;
  }
  return false;
}

// Initialize opacity maps, preserving existing values so existing nodes/links
// stay visible while new ones fade in and removed links fade out.
function initializeOpacityMaps(
  nodes: SimulationNode[],
  links: SimulationLink[],
  state: SimulationState
): void {
  const prevNodeOpacity = state.nodeOpacity;
  const prevLinkOpacity = state.linkOpacity;

  state.nodeOpacity = new Map();
  state.linkOpacity = new Map();

  // Preserve node opacity for existing nodes, start at 0 for new ones.
  nodes.forEach((node) => {
    state.nodeOpacity.set(node.id, prevNodeOpacity.get(node.id) ?? 0);
  });

  // Preserve link opacity for existing links, start at 0 for new ones.
  links.forEach((link) => {
    const linkId = getLinkId(link);
    state.linkOpacity.set(linkId, prevLinkOpacity.get(linkId) ?? 0);
  });
}

function startFadeAnimation(
  state: SimulationState,
  totalNodes: number,
  onStable?: () => void
): void {
  if (state.fadeAnimationId !== null) {
    cancelAnimationFrame(state.fadeAnimationId);
    state.fadeAnimationId = null;
  }

  const animationStartTime = performance.now();
  const WAVE_DURATION = 100; // ms for full wave animation (instant fade-in)

  const animateFade = (timestamp: number) => {
    if (!state.simulation) {
      state.fadeAnimationId = null;
      return;
    }

    const elapsedTime = timestamp - animationStartTime;
    const waveProgress = Math.min(elapsedTime / WAVE_DURATION, 1);

    // Update node opacity
    state.nodeOpacity.forEach((currentOpacity, nodeId) => {
      const targetOpacity = easeOutCubic(Math.min(waveProgress, 1));
      const newOpacity = currentOpacity + (targetOpacity - currentOpacity) * 0.3; // Faster interpolation
      state.nodeOpacity.set(nodeId, Math.min(Math.max(newOpacity, 0), 1));
    });

    // Update link opacity (links fade in after nodes)
    state.linkOpacity.forEach((currentOpacity, linkId) => {
      const targetOpacity = easeOutCubic(Math.min(waveProgress, 1));
      const newOpacity = currentOpacity + (targetOpacity - currentOpacity) * 0.3; // Faster interpolation
      state.linkOpacity.set(linkId, Math.min(Math.max(newOpacity, 0), 1));
    });

    // Fade out removed (dying) links
    const stillDying: SimulationLink[] = [];
    const stillDyingOpacity = new Map<string, number>();
    state.dyingLinks.forEach((link) => {
      const linkId = getLinkId(link);
      const currentOpacity = state.dyingLinkOpacity.get(linkId) ?? 1;
      const newOpacity = Math.max(0, currentOpacity - 0.05);
      if (newOpacity > 0.01) {
        stillDying.push(link);
        stillDyingOpacity.set(linkId, newOpacity);
      }
    });
    state.dyingLinks = stillDying;
    state.dyingLinkOpacity = stillDyingOpacity;

    if (waveProgress < 1 || anyOpacityBelowOne(state) || stillDying.length > 0) {
      state.fadeAnimationId = requestAnimationFrame(animateFade);
    } else {
      state.fadeAnimationId = null;
      // Stop simulation after fade-in to prevent jitter
      if (state.simulation) {
        state.simulation.stop();
      }
      state.isRunning = false;
      state.stable = true;
      onStable?.();
    }
  };

  state.fadeAnimationId = requestAnimationFrame(animateFade);
}

/**
 * Start the force-directed graph simulation
 */
export function startSimulation(
  nodes: SimulationNode[],
  links: SimulationLink[],
  width: number,
  height: number,
  state: SimulationState,
  transform: TransformState,
  onTick: () => void,
  onResetView: () => void,
  onStable?: () => void
): void {
  if (!d3Force) {
    return;
  }

  // Reset transform when starting new simulation
  transform.x = 0;
  transform.y = 0;
  transform.k = 1;

  // Distribute nodes in a circle instead of single point (prevents extreme coordinates)
  // Preserve fixed-position nodes (e.g. Knowledge Core) that already have x/y/fx/fy.
  const simulationNodes: SimulationNode[] = nodes.map((n, i) => {
    if (n.fx != null && n.fy != null && n.x != null && n.y != null) {
      return { ...n };
    }
    const angle = (i / nodes.length) * 2 * Math.PI;
    const radius = Math.min(width, height) * 0.45; // 45% of smaller dimension for less overlap
    return {
      ...n,
      x: width / 2 + Math.cos(angle) * radius,
      y: height / 2 + Math.sin(angle) * radius,
    };
  });

  if (import.meta.env.DEV) {
    console.log(
      "[simulation] Initial node coordinates:",
      simulationNodes.slice(0, 3).map((n) => ({ id: n.id, x: n.x, y: n.y }))
    );
  }

  // Filter links to only include those where both source and target nodes exist
  const validLinks = filterValidLinks(nodes, links);
  if (validLinks.length !== links.length && import.meta.env.DEV) {
    console.warn(`[GraphCanvas] Filtered out ${links.length - validLinks.length} orphan links`);
  }

  // Find links that disappeared since the last simulation and keep them for a fade-out.
  const newLinkIds = new Set(validLinks.map((l) => getLinkId(l)));
  const newlyDying = state.simLinks
    .filter((l) => !newLinkIds.has(getLinkId(l)))
    .map((l) => ({ ...l }));

  for (const link of newlyDying) {
    const linkId = getLinkId(link);
    if (!state.dyingLinkOpacity.has(linkId)) {
      state.dyingLinkOpacity.set(linkId, 1);
      state.dyingLinks = [...state.dyingLinks, link];
    }
  }

  const edges: SimulationLink[] = validLinks.map((l) => ({
    id: l.id,
    source: l.source,
    target: l.target,
    weight: l.weight ?? 1,
    link_type: l.link_type,
    source_type: l.source_type,
    last_weight_update: l.last_weight_update,
  }));

  state.simLinks = edges;

  // Stop existing simulation and fade animation if any
  if (state.simulation) {
    state.simulation.stop();
  }
  if (state.fadeAnimationId !== null) {
    cancelAnimationFrame(state.fadeAnimationId);
    state.fadeAnimationId = null;
  }

  // Initialize opacity maps for fade effect using filtered links
  initializeOpacityMaps(nodes, edges, state);

  const totalNodes = nodes.length;

  // Scale forces slightly with graph density so 50 and 150 nodes both look readable.
  const densityFactor = Math.max(1, Math.sqrt(totalNodes / 50));

  state.simulation = d3Force
    .forceSimulation<SimulationNode, SimulationLink>(simulationNodes)
    .force(
      "link",
      d3Force
        .forceLink<SimulationNode, SimulationLink>(edges)
        .id((d) => d.id)
        .distance(120 * densityFactor)
        .strength(0.35)
    )
    .force("charge", d3Force.forceManyBody().strength(-180 * densityFactor))
    .force("center", d3Force.forceCenter(width / 2, height / 2).strength(0.25))
    .force("collision", d3Force.forceCollide().radius(35 * densityFactor))
    .alphaDecay(0.05) // Slower cooling so the layout has time to spread out
    .on("tick", () => {
      onTick();
    })
    .on("end", () => {
      // Simulation ended - nodes are stable
      // No additional fade animation needed (already handled in startFadeAnimation)
      state.fadeAnimationId = null;
    });

  // Warmup: run simulation synchronously for initial positioning
  if (state.simulation) {
    for (let i = 0; i < 300; i++) {
      state.simulation.tick();
    }

    // Compute transform BEFORE first draw
    onResetView();

    // Then start the animation
    state.simulation.alpha(1).restart();
    startFadeAnimation(state, totalNodes, onStable);
  }
  state.isRunning = true;
  state.stable = false;
}

/**
 * Stop the simulation
 */
export function stopSimulation(state: SimulationState): void {
  if (state.simulation) {
    state.simulation.stop();
    state.isRunning = false;
  }
  // Stop fade animation if running
  if (state.fadeAnimationId !== null) {
    cancelAnimationFrame(state.fadeAnimationId);
    state.fadeAnimationId = null;
  }
}

/**
 * Restart the simulation with alpha heat
 */
export function restartSimulation(state: SimulationState): void {
  if (state.simulation) {
    state.simulation.alpha(1).restart();
    state.isRunning = true;
  }
}

/**
 * Get current simulation nodes
 */
export function getSimulationNodes(state: SimulationState): SimulationNode[] {
  return state.simulation?.nodes() || [];
}

/**
 * Clear simulation state
 */
export function clearSimulation(state: SimulationState): void {
  if (state.simulation) {
    state.simulation.stop();
    state.simulation = null;
  }
  // Stop fade animation if running
  if (state.fadeAnimationId !== null) {
    cancelAnimationFrame(state.fadeAnimationId);
    state.fadeAnimationId = null;
  }
  state.simLinks = [];
  state.nodeOpacity = new Map();
  state.linkOpacity = new Map();
  state.dyingLinks = [];
  state.dyingLinkOpacity = new Map();
  state.isRunning = false;
  state.stable = false;
}
