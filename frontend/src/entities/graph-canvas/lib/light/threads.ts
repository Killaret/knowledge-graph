/**
 * GRAPH-LIGHT-1: links drawn as threads of light (decision 83).
 *
 * Threads are thin, slightly curved and blend from one end's colour to the
 * other's. Model links (source_type "gamma") are pale dots; `dependency` is
 * gold and, when its chain is highlighted, carries running dots from the first
 * task to the next (decision 74). Hovering a note lights up its threads and
 * almost hides the rest.
 */
import { isAutoLink } from "../link-renderers";
import { dependencyLinkKey, type DependencyChain } from "../dependency-chain";
import { seededRand } from "../renderer-utils";
import { getLinkEndpointId, type SimulationLink, type SimulationNode } from "../types";
import { halo, disc } from "./glyphs";
import { lightFrame } from "./style";
import { CYCLE_RED, GOLD, lightTypeColor, mixRgb, rgba, WHITE } from "./palette";

/** Opacity of threads outside the hovered neighbourhood. */
export const LIGHT_DIMMED_THREAD_ALPHA = 0.035;

const BASE_ALPHA = { auto: 0.15, dependency: 0.45, manual: 0.28 } as const;
const HOT_ALPHA = { auto: 0.55, dependency: 0.95, manual: 0.85 } as const;
const WIDTH_PX = { auto: 0.95, dependency: 1.35, manual: 1.05 } as const;

type ThreadKind = keyof typeof BASE_ALPHA;

function threadKind(link: SimulationLink): ThreadKind {
  if (isAutoLink(link)) return "auto";
  return link.link_type === "dependency" ? "dependency" : "manual";
}

/** Screen-pixel width converted to world units; lines grow gently with zoom. */
function widthFor(px: number): number {
  return (px * Math.pow(lightFrame.k, 0.45)) / lightFrame.k;
}

export interface LightThreadState {
  fadeOpacity: number;
  hoveredNodeId?: string | null;
  hoveredNeighborIds?: Set<string>;
  depChain?: DependencyChain | null;
  curveOffset?: number;
  isDuplicateHighlighted?: boolean;
}

export function lightThreadAlpha(link: SimulationLink, state: LightThreadState): number {
  const kind = threadKind(link);
  const sourceId = getLinkEndpointId(link.source);
  const targetId = getLinkEndpointId(link.target);
  const plain: number = BASE_ALPHA[kind];
  let focused = plain;
  if (state.hoveredNodeId) {
    const touches = sourceId === state.hoveredNodeId || targetId === state.hoveredNodeId;
    focused = touches ? HOT_ALPHA[kind] : LIGHT_DIMMED_THREAD_ALPHA;
  }
  if (state.depChain) {
    const depth = state.depChain.linkDepth.get(dependencyLinkKey(sourceId, targetId));
    focused =
      depth === undefined
        ? LIGHT_DIMMED_THREAD_ALPHA
        : HOT_ALPHA[kind] * Math.max(0.35, 1 - 0.12 * depth);
  }
  const mix = lightFrame.focusMix;
  const alpha = mix >= 1 ? focused : plain + (focused - plain) * mix;
  return alpha * state.fadeOpacity * threadLeft(sourceId, targetId);
}

/** During the list morph a thread fades with whichever end has gone further. */
function threadLeft(sourceId: string, targetId: string): number {
  const fade = lightFrame.threadFade;
  if (!fade) return 1;
  return Math.min(fade.get(sourceId) ?? 1, fade.get(targetId) ?? 1);
}

export function drawLightLink(
  ctx: CanvasRenderingContext2D,
  link: SimulationLink,
  source: SimulationNode,
  target: SimulationNode,
  state: LightThreadState
): void {
  if (source.x == null || source.y == null || target.x == null || target.y == null) return;
  const alpha = lightThreadAlpha(link, state);
  if (alpha <= 0.004) return;

  const kind = threadKind(link);
  const sourceId = getLinkEndpointId(link.source);
  const targetId = getLinkEndpointId(link.target);
  const chainKey = dependencyLinkKey(sourceId, targetId);
  const chainDepth = state.depChain?.linkDepth.get(chainKey);
  const isCycle = state.depChain?.cycleLinks.has(chainKey) ?? false;
  const lit =
    chainDepth !== undefined ||
    (!!state.hoveredNodeId &&
      (sourceId === state.hoveredNodeId || targetId === state.hoveredNodeId));

  const x1 = source.x;
  const y1 = source.y;
  const x2 = target.x;
  const y2 = target.y;
  const dx = x2 - x1;
  const dy = y2 - y1;
  const length = Math.hypot(dx, dy) || 1;
  const seed = link.id ?? `${sourceId}>${targetId}`;
  const bend =
    (seededRand(seed, 3) < 0.5 ? 1 : -1) * (0.06 + seededRand(seed, 4) * 0.09) * length +
    (state.curveOffset ?? 0);
  const cx = (x1 + x2) / 2 - (dy / length) * bend;
  const cy = (y1 + y2) / 2 + (dx / length) * bend;

  const previousComposite = ctx.globalCompositeOperation;
  ctx.globalCompositeOperation = "lighter";
  ctx.lineCap = "round";

  if (state.isDuplicateHighlighted) {
    ctx.strokeStyle = rgba(GOLD, Math.min(1, alpha * 2));
  } else if (isCycle) {
    ctx.strokeStyle = rgba(CYCLE_RED, alpha);
  } else if (kind === "dependency") {
    ctx.strokeStyle = rgba(GOLD, alpha);
  } else {
    const whiten = kind === "auto" ? 0.4 : 0.22;
    const gradient = ctx.createLinearGradient(x1, y1, x2, y2);
    gradient.addColorStop(0, rgba(mixRgb(lightTypeColor(source.type), WHITE, whiten), alpha));
    gradient.addColorStop(1, rgba(mixRgb(lightTypeColor(target.type), WHITE, whiten), alpha));
    ctx.strokeStyle = gradient;
  }
  ctx.lineWidth = widthFor(WIDTH_PX[kind] * (lit ? 1.3 : 1));
  ctx.setLineDash(kind === "auto" ? [0.5 / lightFrame.k, 4.2 / lightFrame.k] : []);
  ctx.beginPath();
  ctx.moveTo(x1, y1);
  ctx.quadraticCurveTo(cx, cy, x2, y2);
  ctx.stroke();
  ctx.setLineDash([]);

  if (chainDepth !== undefined && !lightFrame.stable) {
    const fade = Math.max(0.35, 1 - 0.12 * chainDepth);
    const seconds = lightFrame.time / 1000;
    const glowRadius = 5 / lightFrame.k;
    for (let i = 0; i < 3; i++) {
      const u = (((seconds * 0.5 - chainDepth * 0.34 + i / 3) % 1) + 1) % 1;
      const iu = 1 - u;
      const px = iu * iu * x1 + 2 * iu * u * cx + u * u * x2;
      const py = iu * iu * y1 + 2 * iu * u * cy + u * u * y2;
      halo(ctx, px, py, glowRadius, isCycle ? CYCLE_RED : GOLD, 0.55 * fade);
      disc(ctx, px, py, 1.3 / lightFrame.k, [255, 240, 205], 0.95 * fade);
    }
  }

  ctx.globalCompositeOperation = previousComposite;
}
