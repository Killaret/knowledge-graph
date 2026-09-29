/**
 * GRAPH-LIGHT-1: the light-style sky — deep ink, a soft vignette and a
 * starfield in three layers that shift with panning at different speeds,
 * which gives depth without 3D. Drawn in screen coordinates.
 *
 * «Скорость»: while the camera stands still the sky is kept in a layer of its
 * own and stamped; while it moves (or the stars twinkle to the next step) the
 * sky changes every frame, so it is painted straight into the frame — keeping
 * it would only add a copy. Twinkling changes the sky at most ten times a second.
 */
import { seededRand } from "../renderer-utils";
import { createCanvas } from "./sprites";

interface BackgroundStar {
  x: number;
  y: number;
  parallax: number;
  size: number;
  alpha: number;
  twinkle: number;
  phase: number;
}

const LAYER_PARALLAX = [0.04, 0.1, 0.2];
const LAYER_SIZE = [0.7, 1.0, 1.5];
const LAYER_ALPHA = [0.22, 0.34, 0.52];
/** One star per this many square pixels. */
const STAR_DENSITY = 3600;

let cachedStars: BackgroundStar[] = [];
let cachedFor = "";

function starsFor(width: number, height: number): BackgroundStar[] {
  const key = `${Math.round(width)}x${Math.round(height)}`;
  if (key === cachedFor) return cachedStars;
  const count = Math.round((width * height) / STAR_DENSITY);
  const stars: BackgroundStar[] = [];
  for (let i = 0; i < count; i++) {
    const roll = seededRand("sky", i * 7);
    const layer = roll < 0.62 ? 0 : roll < 0.9 ? 1 : 2;
    stars.push({
      x: seededRand("sky", i * 7 + 1),
      y: seededRand("sky", i * 7 + 2),
      parallax: LAYER_PARALLAX[layer],
      size: LAYER_SIZE[layer] * (0.7 + seededRand("sky", i * 7 + 3) * 0.6),
      alpha: LAYER_ALPHA[layer] * (0.5 + seededRand("sky", i * 7 + 4) * 0.5),
      twinkle: seededRand("sky", i * 7 + 5) < 0.12 ? 0.6 + seededRand("sky", i * 7 + 6) * 1.4 : 0,
      phase: seededRand("sky", i * 7 + 6) * Math.PI * 2,
    });
  }
  cachedStars = stars;
  cachedFor = key;
  return stars;
}

/** Twinkling repaints the sky at most this often. */
const TWINKLE_STEP_MS = 100;

interface SkyLayer {
  canvas: HTMLCanvasElement | OffscreenCanvas;
  ctx: CanvasRenderingContext2D;
  /** The sky the layer holds. */
  key: string;
  /** The sky of the previous frame. */
  seen: string;
}

/** One layer per device pixel ratio: the page and the frame cache draw at different ones. */
const layers = new Map<number, SkyLayer | null>();

function skyLayer(width: number, height: number, scale: number): SkyLayer | null {
  let layer = layers.get(scale);
  if (layer === undefined) {
    const canvas = createCanvas(1, 1);
    const ctx = canvas?.getContext("2d") as CanvasRenderingContext2D | null | undefined;
    layer = canvas && ctx ? { canvas, ctx, key: "", seen: "" } : null;
    layers.set(scale, layer);
  }
  if (!layer) return null;
  const w = Math.max(1, Math.round(width * scale));
  const h = Math.max(1, Math.round(height * scale));
  if (layer.canvas.width !== w || layer.canvas.height !== h) {
    layer.canvas.width = w;
    layer.canvas.height = h;
    layer.key = "";
  }
  return layer;
}

/** Device pixels per CSS pixel of the target, read from its current transform. */
function deviceScale(ctx: CanvasRenderingContext2D): number {
  const a = typeof ctx.getTransform === "function" ? ctx.getTransform()?.a : undefined;
  return a !== undefined && a > 0 && Number.isFinite(a) ? a : 1;
}

export function drawLightBackground(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  transform: { x: number; y: number },
  time: number,
  twinkle: boolean
): void {
  const scale = deviceScale(ctx);
  const layer = skyLayer(width, height, scale);
  if (!layer) {
    paintLightSky(ctx, width, height, transform, time, twinkle);
    return;
  }
  const step = twinkle ? Math.floor(time / TWINKLE_STEP_MS) : -1;
  const skyTime = step * TWINKLE_STEP_MS;
  const key = [width, height, Math.round(transform.x), Math.round(transform.y), step].join(":");
  if (layer.key !== key) {
    if (layer.seen !== key) {
      layer.seen = key;
      paintLightSky(ctx, width, height, transform, skyTime, twinkle);
      return;
    }
    if (typeof layer.ctx.setTransform === "function")
      layer.ctx.setTransform(scale, 0, 0, scale, 0, 0);
    paintLightSky(layer.ctx, width, height, transform, skyTime, twinkle);
    layer.key = key;
  }
  ctx.drawImage(layer.canvas, 0, 0, width, height);
}

/** Paint the sky itself; drawLightBackground caches it. */
export function paintLightSky(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  transform: { x: number; y: number },
  time: number,
  twinkle: boolean
): void {
  ctx.fillStyle = "#04060d";
  ctx.fillRect(0, 0, width, height);

  const glow = ctx.createRadialGradient(
    width * 0.5,
    height * 0.45,
    0,
    width * 0.5,
    height * 0.45,
    Math.max(width, height) * 0.75
  );
  glow.addColorStop(0, "rgba(22,30,58,0.55)");
  glow.addColorStop(0.55, "rgba(10,14,30,0.3)");
  glow.addColorStop(1, "rgba(0,0,0,0)");
  ctx.fillStyle = glow;
  ctx.fillRect(0, 0, width, height);

  const seconds = time / 1000;
  for (const star of starsFor(width, height)) {
    let x = (star.x * width + transform.x * star.parallax) % width;
    if (x < 0) x += width;
    let y = (star.y * height + transform.y * star.parallax) % height;
    if (y < 0) y += height;
    const alpha =
      star.twinkle && twinkle
        ? star.alpha * (0.65 + 0.35 * Math.sin(seconds * star.twinkle + star.phase))
        : star.alpha;
    ctx.fillStyle = `rgba(210,222,255,${alpha.toFixed(3)})`;
    ctx.fillRect(x, y, star.size, star.size);
  }
}
