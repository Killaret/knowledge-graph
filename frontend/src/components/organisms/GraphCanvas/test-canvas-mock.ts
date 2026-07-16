/**
 * Reusable canvas context mock for GraphCanvas tests
 */
import { vi } from 'vitest';

export function createMockCanvasContext() {
  const fillStyles: string[] = [];
  const strokeStyles: string[] = [];

  const ctx = {
    clearRect: vi.fn(),
    save: vi.fn(),
    restore: vi.fn(),
    translate: vi.fn(),
    scale: vi.fn(),
    rotate: vi.fn(),
    beginPath: vi.fn(),
    moveTo: vi.fn(),
    lineTo: vi.fn(),
    quadraticCurveTo: vi.fn(),
    stroke: vi.fn(),
    fill: vi.fn(function (this: typeof ctx) {
      fillStyles.push((this as any).fillStyle);
    }),
    closePath: vi.fn(),
    arc: vi.fn(),
    ellipse: vi.fn(),
    fillRect: vi.fn(),
    strokeRect: vi.fn(),
    setLineDash: vi.fn(),
    fillText: vi.fn(),
    measureText: vi.fn(() => ({ width: 50 })),
    createRadialGradient: vi.fn((x0, y0, r0, x1, y1, r1) => {
      const colorStops: Array<{ offset: number; color: string }> = [];
      const gradient = {
        addColorStop: vi.fn((offset: number, color: string) => {
          colorStops.push({ offset, color });
        }),
        getColorStops: () => colorStops
      };
      return gradient;
    }),
    createLinearGradient: vi.fn(() => ({
      addColorStop: vi.fn()
    })),
    roundRect: vi.fn(),
    set fillStyle(value: string) {
      (ctx as any)._fillStyle = value;
    },
    get fillStyle() {
      return (ctx as any)._fillStyle || '';
    },
    set strokeStyle(value: string) {
      (ctx as any)._strokeStyle = value;
      strokeStyles.push(value);
    },
    get strokeStyle() {
      return (ctx as any)._strokeStyle || '';
    },
    font: '',
    textAlign: 'center' as CanvasTextAlign,
    textBaseline: 'middle' as CanvasTextBaseline,
    lineWidth: 1,
    shadowBlur: 0,
    shadowColor: '',
    globalAlpha: 1,
    getFillStyles: () => fillStyles,
    getStrokeStyles: () => strokeStyles
  } as unknown as CanvasRenderingContext2D & { getFillStyles: () => string[]; getStrokeStyles: () => string[] };

  return ctx;
}

export function createMockCanvasContextWithoutFillSpy() {
  const ctx = createMockCanvasContext();
  (ctx as any).fill = vi.fn();
  return ctx;
}
