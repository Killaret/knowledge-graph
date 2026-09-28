/**
 * GRAPH-LIGHT-1: notes drawn as sources of light (decision 83).
 *
 * Each note is a core and a halo; halos add up (`lighter`), so dense places
 * glow. The kind of note reads from the shape of its light: a comet has a
 * tail, a nebula is a soft cloud, a black hole is a dark core in a ring, a
 * satellite has an orbit, dust is a scatter of specks, debris is a dim shard.
 * Colour comes from the type family (palette.ts), never from random
 * per-note variation. Unknown types and anomalies share one strange light.
 */
import { CelestialBody, type CelestialBodyDrawContext } from "$entities";
import { seededRand } from "../renderer-utils";
import { lightFrame } from "./style";
import { lightCoreRadius, lightTypeColor, mixRgb, rgba, WHITE, type Rgb } from "./palette";

const TAU = Math.PI * 2;
const DARK_CORE: Rgb = [3, 4, 9];
/** Comet tails share one direction, like a solar wind: up and to the left. */
const COMET_TAIL_DIRECTION: readonly [number, number] = [-0.72, -0.69];
/** Tail length for comets without a due date yet (COMET-1 brings the date). */
const COMET_DEFAULT_URGENCY = 0.55;

type LightGlyph = (
  ctx: CanvasRenderingContext2D,
  c: CelestialBodyDrawContext,
  color: Rgb,
  core: number
) => void;

export function halo(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  radius: number,
  color: Rgb,
  alpha: number
): void {
  if (alpha <= 0.003 || radius <= 0.2) return;
  const g = ctx.createRadialGradient(x, y, 0, x, y, radius);
  g.addColorStop(0, rgba(color, alpha));
  g.addColorStop(0.28, rgba(color, alpha * 0.32));
  g.addColorStop(1, rgba(color, 0));
  ctx.fillStyle = g;
  ctx.beginPath();
  ctx.arc(x, y, radius, 0, TAU);
  ctx.fill();
}

export function disc(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  radius: number,
  color: Rgb,
  alpha: number
): void {
  if (alpha <= 0.003 || radius <= 0) return;
  ctx.fillStyle = rgba(color, alpha);
  ctx.beginPath();
  ctx.arc(x, y, radius, 0, TAU);
  ctx.fill();
}

function withLighter(ctx: CanvasRenderingContext2D, draw: () => void): void {
  const previous = ctx.globalCompositeOperation;
  ctx.globalCompositeOperation = "lighter";
  try {
    draw();
  } finally {
    ctx.globalCompositeOperation = previous;
  }
}

/** Slow breathing of the halo; off in snapshot and focus modes. */
function breath(c: CelestialBodyDrawContext): number {
  if (lightFrame.stable || c.disableVariation || c.focusMode) return 1;
  const phase = seededRand(c.nodeId, 7) * TAU;
  return 1 + 0.08 * Math.sin((lightFrame.time / 1000) * 0.9 + phase);
}

/**
 * Lights keep a readable size at any zoom: the world-space core grows when
 * the view zooms out and shrinks when it zooms in, so on screen it changes
 * with the square root of the zoom instead of linearly.
 */
export function zoomCompensation(): number {
  return Math.min(1.8, Math.max(0.7, 1 / Math.sqrt(lightFrame.k)));
}

/** Hairline width that stays about one screen pixel at any zoom. */
function hairline(width: number): number {
  return width / lightFrame.k;
}

const galaxy: LightGlyph = (ctx, c, color, core) => {
  const b = breath(c);
  withLighter(ctx, () => {
    halo(ctx, c.x, c.y, core * 8, color, 0.2 * b);
    ctx.save();
    ctx.translate(c.x, c.y);
    ctx.rotate(seededRand(c.nodeId, 1) * TAU);
    ctx.scale(1, 0.34);
    halo(ctx, 0, 0, core * 5.2, mixRgb(color, WHITE, 0.25), 0.34);
    ctx.restore();
    halo(ctx, c.x, c.y, core * 2.4, mixRgb(color, WHITE, 0.45), 0.75);
    disc(ctx, c.x, c.y, core * 0.7, mixRgb(color, WHITE, 0.85), 1);
  });
};

const star: LightGlyph = (ctx, c, color, core) => {
  const b = breath(c);
  withLighter(ctx, () => {
    halo(ctx, c.x, c.y, core * 5.5, color, 0.26 * b);
    const spike = core * 4 * b;
    ctx.lineWidth = hairline(0.8);
    ctx.strokeStyle = rgba(mixRgb(color, WHITE, 0.55), 0.32);
    ctx.beginPath();
    ctx.moveTo(c.x - spike, c.y);
    ctx.lineTo(c.x + spike, c.y);
    ctx.moveTo(c.x, c.y - spike);
    ctx.lineTo(c.x, c.y + spike);
    ctx.stroke();
    halo(ctx, c.x, c.y, core * 1.9, mixRgb(color, WHITE, 0.5), 0.8);
    disc(ctx, c.x, c.y, core * 0.62, mixRgb(color, WHITE, 0.9), 1);
  });
};

const planet: LightGlyph = (ctx, c, color, core) => {
  withLighter(ctx, () => halo(ctx, c.x, c.y, core * 2.8, color, 0.22));
  const g = ctx.createRadialGradient(
    c.x - core * 0.35,
    c.y - core * 0.35,
    core * 0.1,
    c.x,
    c.y,
    core
  );
  g.addColorStop(0, rgba(mixRgb(color, WHITE, 0.55), 1));
  g.addColorStop(0.7, rgba(color, 0.95));
  g.addColorStop(1, rgba(mixRgb(color, [8, 10, 24], 0.45), 1));
  ctx.fillStyle = g;
  ctx.beginPath();
  ctx.arc(c.x, c.y, core, 0, TAU);
  ctx.fill();
};

const moon: LightGlyph = (ctx, c, color, core) => {
  withLighter(ctx, () => halo(ctx, c.x, c.y, core * 2.2, color, 0.14));
  disc(ctx, c.x, c.y, core * 0.85, mixRgb(color, [206, 214, 230], 0.6), 0.85);
};

/**
 * Asteroids are small but clearly visible: imported bookmarks become
 * asteroids by default (import/service.go), so on real data they are often
 * the most common kind of note.
 */
const asteroid: LightGlyph = (ctx, c, color, core) => {
  const turn = seededRand(c.nodeId, 1) * TAU;
  withLighter(ctx, () => halo(ctx, c.x, c.y, core * 2.6, color, 0.2));
  ctx.fillStyle = rgba(mixRgb(color, WHITE, 0.3), 0.95);
  ctx.beginPath();
  for (let i = 0; i < 6; i++) {
    const angle = turn + (i / 6) * TAU;
    const reach = core * (0.75 + seededRand(c.nodeId, 10 + i) * 0.45);
    const px = c.x + Math.cos(angle) * reach;
    const py = c.y + Math.sin(angle) * reach;
    if (i === 0) ctx.moveTo(px, py);
    else ctx.lineTo(px, py);
  }
  ctx.closePath();
  ctx.fill();
};

const nebula: LightGlyph = (ctx, c, color, core) => {
  const b = breath(c);
  withLighter(ctx, () => {
    for (let i = 0; i < 3; i++) {
      const ox = (seededRand(c.nodeId, 20 + i) - 0.5) * 1.8 * core * 2;
      const oy = (seededRand(c.nodeId, 30 + i) - 0.5) * 1.2 * core * 2;
      const size = 0.7 + seededRand(c.nodeId, 40 + i) * 0.6;
      halo(ctx, c.x + ox, c.y + oy, core * 6 * size, color, 0.22 * b);
    }
    disc(ctx, c.x, c.y, core * 0.7, mixRgb(color, WHITE, 0.8), 0.9);
  });
};

const comet: LightGlyph = (ctx, c, color, core) => {
  const u = COMET_DEFAULT_URGENCY;
  const b = breath(c);
  const [dx, dy] = COMET_TAIL_DIRECTION;
  withLighter(ctx, () => {
    const length = core * (3 + 12 * u);
    const width = core * 0.95;
    const tx = c.x + dx * length;
    const ty = c.y + dy * length;
    const g = ctx.createLinearGradient(c.x, c.y, tx, ty);
    g.addColorStop(0, rgba(mixRgb(color, WHITE, 0.4), 0.25 + 0.5 * u));
    g.addColorStop(1, rgba(color, 0));
    ctx.fillStyle = g;
    ctx.beginPath();
    ctx.moveTo(c.x - dy * width, c.y + dx * width);
    ctx.lineTo(tx, ty);
    ctx.lineTo(c.x + dy * width, c.y - dx * width);
    ctx.closePath();
    ctx.fill();
    halo(ctx, c.x, c.y, core * (2.2 + 2 * u), color, (0.2 + 0.35 * u) * b);
    disc(ctx, c.x, c.y, core * 0.6, mixRgb(color, WHITE, 0.85), 1);
  });
};

const satellite: LightGlyph = (ctx, c, color, core) => {
  withLighter(ctx, () => {
    halo(ctx, c.x, c.y, core * 2.2, color, 0.25);
    disc(ctx, c.x, c.y, core * 0.6, mixRgb(color, WHITE, 0.75), 1);
  });
  ctx.strokeStyle = rgba(mixRgb(color, WHITE, 0.3), 0.45);
  ctx.lineWidth = hairline(0.8);
  ctx.beginPath();
  ctx.ellipse(c.x, c.y, core * 2.6, core * 1.1, seededRand(c.nodeId, 1) * TAU, 0, TAU);
  ctx.stroke();
};

const blackhole: LightGlyph = (ctx, c, color, core) => {
  withLighter(ctx, () => {
    halo(ctx, c.x, c.y, core * 3.6, color, 0.12);
    ctx.strokeStyle = rgba(mixRgb(color, WHITE, 0.35), 0.85);
    ctx.lineWidth = hairline(1.3);
    ctx.beginPath();
    ctx.arc(c.x, c.y, core * 1.45, 0, TAU);
    ctx.stroke();
  });
  disc(ctx, c.x, c.y, core * 1.15, DARK_CORE, 0.95);
  ctx.strokeStyle = rgba(color, 0.35);
  ctx.lineWidth = hairline(0.6);
  ctx.beginPath();
  ctx.arc(c.x, c.y, core * 0.8, 0, TAU);
  ctx.stroke();
};

const dust: LightGlyph = (ctx, c, color, core) => {
  withLighter(ctx, () => {
    for (let i = 0; i < 4; i++) {
      const ox = (seededRand(c.nodeId, 50 + i) - 0.5) * 3.4 * core;
      const oy = (seededRand(c.nodeId, 60 + i) - 0.5) * 3.4 * core;
      const size = 0.6 + seededRand(c.nodeId, 70 + i) * 0.8;
      disc(ctx, c.x + ox, c.y + oy, core * 0.4 * size, mixRgb(color, WHITE, 0.45), 0.8);
    }
  });
};

const debris: LightGlyph = (ctx, c, color, core) => {
  const turn = seededRand(c.nodeId, 1) * TAU;
  ctx.fillStyle = rgba(color, 0.5);
  ctx.beginPath();
  ctx.moveTo(c.x + Math.cos(turn) * core * 1.2, c.y + Math.sin(turn) * core * 1.2);
  ctx.lineTo(c.x + Math.cos(turn + 2.3) * core, c.y + Math.sin(turn + 2.3) * core);
  ctx.lineTo(c.x + Math.cos(turn + 4.1) * core * 0.9, c.y + Math.sin(turn + 4.1) * core * 0.9);
  ctx.closePath();
  ctx.fill();
};

const technical: LightGlyph = (ctx, c, color, core) => {
  disc(ctx, c.x, c.y, core * 0.8, color, 0.7);
  ctx.strokeStyle = rgba(color, 0.45);
  ctx.lineWidth = hairline(0.8);
  ctx.beginPath();
  ctx.arc(c.x, c.y, core * 1.6, 0, TAU);
  ctx.stroke();
};

/** Unknown types and anomalies: a strange violet light in a broken ring. */
const anomaly: LightGlyph = (ctx, c, color, core) => {
  const b = breath(c);
  withLighter(ctx, () => {
    halo(ctx, c.x, c.y, core * 4.5, color, 0.24 * b);
    disc(ctx, c.x, c.y, core * 0.55, mixRgb(color, WHITE, 0.7), 0.95);
  });
  ctx.strokeStyle = rgba(color, 0.55);
  ctx.lineWidth = hairline(0.9);
  const turn = seededRand(c.nodeId, 1) * TAU;
  for (let i = 0; i < 3; i++) {
    const start = turn + (i * TAU) / 3;
    ctx.beginPath();
    ctx.arc(c.x, c.y, core * 1.7, start, start + TAU / 5);
    ctx.stroke();
  }
};

const LIGHT_GLYPHS: ReadonlyArray<[CelestialBody, LightGlyph]> = [
  [CelestialBody.GALAXY, galaxy],
  [CelestialBody.STAR, star],
  [CelestialBody.PLANET, planet],
  [CelestialBody.MOON, moon],
  [CelestialBody.ASTEROID, asteroid],
  [CelestialBody.NEBULA, nebula],
  [CelestialBody.COMET, comet],
  [CelestialBody.SATELLITE, satellite],
  [CelestialBody.BLACKHOLE, blackhole],
  [CelestialBody.DUST, dust],
  [CelestialBody.DEBRIS, debris],
  [CelestialBody.TECHNICAL, technical],
  [CelestialBody.UNKNOWN, anomaly],
  [CelestialBody.REALITY_RIFT, anomaly],
  [CelestialBody.CHROMATIC_MAW, anomaly],
  [CelestialBody.VOID_WHISPER, anomaly],
  [CelestialBody.COSMIC_ABOMINATION, anomaly],
];

/** Point every body at its light glyph. */
export function registerLightDrawers(): void {
  for (const [body, glyph] of LIGHT_GLYPHS) {
    const color = lightTypeColor(body.isAnomaly ? "unknown" : body.type);
    body.drawFunction = (ctx, c) =>
      glyph(ctx, c, color, lightCoreRadius(body.type, c.r) * zoomCompensation());
  }
}
