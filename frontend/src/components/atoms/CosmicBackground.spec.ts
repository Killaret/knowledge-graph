import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, cleanup } from "@testing-library/svelte";
import CosmicBackground from "./CosmicBackground.svelte";

// Mock browser environment
vi.mock("$app/environment", () => ({
  browser: true,
}));

describe("CosmicBackground", () => {
  beforeEach(() => {
    // Mock canvas context with minimal implementation
    const mockContext = {
      clearRect: vi.fn(),
      beginPath: vi.fn(),
      arc: vi.fn(),
      fill: vi.fn(),
      createRadialGradient: vi.fn(() => ({
        addColorStop: vi.fn(),
      })),
      createLinearGradient: vi.fn(() => ({
        addColorStop: vi.fn(),
      })),
      stroke: vi.fn(),
    } as unknown as CanvasRenderingContext2D;

    HTMLCanvasElement.prototype.getContext = vi.fn(() => mockContext) as any;

    // Mock requestAnimationFrame
    global.requestAnimationFrame = vi.fn((callback: FrameRequestCallback) => {
      return window.setTimeout(() => callback(performance.now()), 16) as unknown as number;
    });
    global.cancelAnimationFrame = vi.fn((id: number) => window.clearTimeout(id));

    // Mock window dimensions
    Object.defineProperty(window, "innerWidth", {
      value: 1920,
      writable: true,
    });
    Object.defineProperty(window, "innerHeight", {
      value: 1080,
      writable: true,
    });
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("should render canvas element", () => {
    const { container } = render(CosmicBackground);
    const canvas = container.querySelector("canvas");
    expect(canvas).toBeTruthy();
    expect(canvas).toHaveClass("cosmic-background");
  });

  it.skip("should have correct CSS positioning", () => {
    // NOTE: This test is skipped because jsdom doesn't properly compute
    // all CSS styles. The positioning is verified through E2E tests.
    const { container } = render(CosmicBackground);
    const canvas = container.querySelector("canvas");

    expect(canvas).toBeTruthy();
    expect(canvas).toHaveClass("cosmic-background");

    // In real browser:
    // const styles = window.getComputedStyle(canvas!);
    // expect(styles.position).toBe('fixed');
    // expect(styles.zIndex).toBe('0');
    // expect(styles.pointerEvents).toBe('none');
  });

  it("should be hidden from screen readers", () => {
    const { container } = render(CosmicBackground);
    const canvas = container.querySelector("canvas");
    expect(canvas).toHaveAttribute("aria-hidden", "true");
  });

  it("should handle window resize", () => {
    const { container } = render(CosmicBackground);
    const canvas = container.querySelector("canvas") as HTMLCanvasElement;

    // Simulate resize
    window.innerWidth = 800;
    window.innerHeight = 600;
    window.dispatchEvent(new Event("resize"));

    // Canvas should be responsive
    expect(canvas).toBeTruthy();
  });

  it("should handle mouse movement for parallax", () => {
    const { container } = render(CosmicBackground);

    // Simulate mouse movement
    const mouseEvent = new MouseEvent("mousemove", {
      clientX: 500,
      clientY: 300,
    });
    window.dispatchEvent(mouseEvent);

    // Component should still be rendered
    const canvas = container.querySelector("canvas");
    expect(canvas).toBeTruthy();
  });
});
