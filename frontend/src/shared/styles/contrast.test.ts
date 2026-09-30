import { describe, it, expect } from "vitest";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

// UI-QUICK-1: dim text must meet WCAG AA for normal text (>= 4.5:1).
// Regression guard: --carbon-text-dim was #5a5a6e = 2.92:1 on the dark frame.
const here = dirname(fileURLToPath(import.meta.url));
const css = readFileSync(join(here, "global.css"), "utf8");

function token(name: string): string {
  const m = css.match(new RegExp(`${name}:\\s*(#[0-9a-fA-F]{6})`));
  if (!m) throw new Error(`token ${name} not found in global.css`);
  return m[1];
}

function luminance(hex: string): number {
  const h = hex.replace("#", "");
  const [r, g, b] = [0, 2, 4].map((i) => parseInt(h.slice(i, i + 2), 16) / 255);
  const f = (c: number) => (c <= 0.04045 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4));
  return 0.2126 * f(r) + 0.7152 * f(g) + 0.0722 * f(b);
}

function contrast(fg: string, bg: string): number {
  let l1 = luminance(fg);
  let l2 = luminance(bg);
  if (l1 < l2) [l1, l2] = [l2, l1];
  return (l1 + 0.05) / (l2 + 0.05);
}

describe("carbon theme contrast", () => {
  it("--carbon-text-dim meets WCAG AA (>= 4.5:1) on the app background", () => {
    const ratio = contrast(token("--carbon-text-dim"), token("--carbon-graphite"));
    expect(ratio).toBeGreaterThanOrEqual(4.5);
  });

  it("--carbon-text and --carbon-text-muted keep AA headroom on the app background", () => {
    const bg = token("--carbon-graphite");
    expect(contrast(token("--carbon-text"), bg)).toBeGreaterThanOrEqual(4.5);
    expect(contrast(token("--carbon-text-muted"), bg)).toBeGreaterThanOrEqual(4.5);
  });
});
