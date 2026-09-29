/**
 * GRAPH-LIGHT-1, «Скорость»: parts of the light that look the same for every
 * note of a colour are painted once and then stamped with drawImage — scaled
 * to the size, faded with globalAlpha. A halo is one radial gradient, and
 * creating it anew for every note on every frame was the main cost of a
 * light frame.
 */
import { mixRgb, rgba, WHITE, type Rgb } from "./palette";

export type Sprite = HTMLCanvasElement | OffscreenCanvas;

/** Sprites are painted at this radius and scaled when stamped. */
export const SPRITE_RADIUS = 64;

const sprites = new Map<string, Sprite | null>();
const spriteColors = new WeakMap<object, Rgb>();

/** A scratch canvas for painting; null where there is no canvas at all. */
export function createCanvas(width: number, height: number): Sprite | null {
  if (typeof document !== "undefined") {
    const canvas = document.createElement("canvas");
    canvas.width = width;
    canvas.height = height;
    return canvas;
  }
  return typeof OffscreenCanvas !== "undefined" ? new OffscreenCanvas(width, height) : null;
}

function sprite(
  key: string,
  color: Rgb,
  paint: (ctx: CanvasRenderingContext2D, radius: number) => void
): Sprite | null {
  const known = sprites.get(key);
  if (known !== undefined) return known;
  const canvas = createCanvas(SPRITE_RADIUS * 2, SPRITE_RADIUS * 2);
  const ctx = canvas?.getContext("2d") as CanvasRenderingContext2D | null | undefined;
  let made: Sprite | null = null;
  if (canvas && ctx) {
    paint(ctx, SPRITE_RADIUS);
    spriteColors.set(canvas, color);
    made = canvas;
  }
  sprites.set(key, made);
  return made;
}

/** Soft round glow: full colour in the middle, gone at the rim. */
export function haloSprite(color: Rgb): Sprite | null {
  return sprite(`halo:${color.join(",")}`, color, (ctx, r) => paintHalo(ctx, r, r, r, color, 1));
}

/** The halo itself; stamped from a sprite, painted directly where no sprite can be made. */
export function paintHalo(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  radius: number,
  color: Rgb,
  alpha: number
): void {
  const g = ctx.createRadialGradient(x, y, 0, x, y, radius);
  g.addColorStop(0, rgba(color, alpha));
  g.addColorStop(0.28, rgba(color, alpha * 0.32));
  g.addColorStop(1, rgba(color, 0));
  ctx.fillStyle = g;
  ctx.beginPath();
  ctx.arc(x, y, radius, 0, Math.PI * 2);
  ctx.fill();
}

/** A lit sphere: bright upper left, the colour, a darker rim. */
export function sphereSprite(color: Rgb): Sprite | null {
  return sprite(`sphere:${color.join(",")}`, color, (ctx, r) => {
    const g = ctx.createRadialGradient(r * 0.65, r * 0.65, r * 0.1, r, r, r);
    g.addColorStop(0, rgba(mixRgb(color, WHITE, 0.55), 1));
    g.addColorStop(0.7, rgba(color, 0.95));
    g.addColorStop(1, rgba(mixRgb(color, [8, 10, 24], 0.45), 1));
    ctx.fillStyle = g;
    ctx.beginPath();
    ctx.arc(r, r, r, 0, Math.PI * 2);
    ctx.fill();
  });
}

/** The colour a sprite was painted with (tests and diagnostics). */
export function spriteColor(image: unknown): Rgb | undefined {
  return typeof image === "object" && image !== null ? spriteColors.get(image) : undefined;
}
