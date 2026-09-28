/**
 * GRAPH-LIGHT-1: the light-style sky — deep ink, a soft vignette and a
 * starfield in three layers that shift with panning at different speeds,
 * which gives depth without 3D. Drawn in screen coordinates.
 */
import { seededRand } from "../renderer-utils";

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

export function drawLightBackground(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  transform: { x: number; y: number },
  time: number,
  stable: boolean
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
      star.twinkle && !stable
        ? star.alpha * (0.65 + 0.35 * Math.sin(seconds * star.twinkle + star.phase))
        : star.alpha;
    ctx.fillStyle = `rgba(210,222,255,${alpha.toFixed(3)})`;
    ctx.fillRect(x, y, star.size, star.size);
  }
}
