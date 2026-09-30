/**
 * Incremental simulation updates (UI-LOAD-1): append nodes/links to a running
 * d3-force simulation without rebuilding it — the existing layout is kept and
 * new nodes grow out of their already-placed neighbors.
 */
import { getLinkEndpointId } from "./types";
import { getLinkId, startFadeAnimation } from "./simulation";
import type { SimulationLink, SimulationNode, SimulationState } from "./types";

interface ForceLinkLike {
  links(): SimulationLink[];
  links(links: SimulationLink[]): ForceLinkLike;
}

/**
 * Append nodes and links to the running simulation and gently reheat it.
 * Returns false when there is no live simulation to extend.
 *
 * Placement rule: a new node spawns next to an already-placed neighbor when it
 * has one in `newLinks`, otherwise near the canvas center — so the graph grows
 * outward from its own structure instead of teleporting in a ring.
 */
export function addNodesToSimulation(
  state: SimulationState,
  newNodes: SimulationNode[],
  newLinks: SimulationLink[],
  width: number,
  height: number,
  onStable?: () => void
): boolean {
  const sim = state.simulation;
  if (!sim || newNodes.length === 0) return false;

  const placed = new Map<string, SimulationNode>(sim.nodes().map((n) => [n.id, n]));

  const positioned: SimulationNode[] = newNodes.map((n, i) => {
    const anchorLink = newLinks.find(
      (l) => getLinkEndpointId(l.source) === n.id || getLinkEndpointId(l.target) === n.id
    );
    const anchorId = anchorLink
      ? getLinkEndpointId(
          getLinkEndpointId(anchorLink.source) === n.id ? anchorLink.target : anchorLink.source
        )
      : undefined;
    const anchor = anchorId ? placed.get(anchorId) : undefined;

    const angle = (i / Math.max(1, newNodes.length)) * 2 * Math.PI;
    const baseX = anchor?.x ?? width / 2;
    const baseY = anchor?.y ?? height / 2;
    return {
      ...n,
      x: baseX + Math.cos(angle) * 60,
      y: baseY + Math.sin(angle) * 60,
    };
  });

  sim.nodes([...sim.nodes(), ...positioned]);

  const edges: SimulationLink[] = newLinks.map((l) => ({
    id: l.id,
    source: l.source,
    target: l.target,
    weight: l.weight ?? 1,
    link_type: l.link_type,
    source_type: l.source_type,
    gamma_origin: l.gamma_origin,
    last_weight_update: l.last_weight_update,
  }));

  const linkForce = sim.force("link") as ForceLinkLike | undefined;
  if (edges.length > 0 && linkForce && typeof linkForce.links === "function") {
    linkForce.links([...linkForce.links(), ...edges]);
  }
  if (edges.length > 0) {
    state.simLinks = [...state.simLinks, ...edges];
  }

  // New arrivals start transparent and fade in via the same opacity maps the
  // renderer already reads.
  for (const n of positioned) state.nodeOpacity.set(n.id, 0);
  for (const e of edges) state.linkOpacity.set(getLinkId(e), 0);

  // The fade loop may already have finished — restart it so the new arrivals
  // actually appear instead of staying at opacity 0.
  if (state.fadeAnimationId === null) {
    startFadeAnimation(state, sim.nodes().length, onStable);
  }

  // Gentle reheat — the layout is perturbed, not restarted.
  sim.alpha(0.4).restart();
  state.isRunning = true;
  state.stable = false;
  return true;
}
