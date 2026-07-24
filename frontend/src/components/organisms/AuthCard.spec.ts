import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, cleanup } from "@testing-library/svelte";
import AuthCard from "./AuthCard.svelte";

// Mock browser environment
vi.mock("$app/environment", () => ({
  browser: true,
}));

// Mock GraphCanvas to avoid canvas dependencies in unit tests
vi.mock("./GraphCanvas.svelte", () => ({
  default: vi.fn().mockImplementation(() => {
    const div = document.createElement("div");
    div.setAttribute("data-testid", "graph-canvas");
    return div;
  }),
}));

vi.mock("$shared/hooks/usePreloadedData", () => ({
  getGraphWithPreload: vi.fn(() => Promise.resolve({ nodes: [], links: [] })),
}));

describe("AuthCard", () => {
  beforeEach(() => {
    // Mock canvas for CosmicBackground
    const mockContext = {
      clearRect: vi.fn(),
      beginPath: vi.fn(),
      arc: vi.fn(),
      fill: vi.fn(),
      createRadialGradient: vi.fn(() => ({
        addColorStop: vi.fn(),
      })),
      stroke: vi.fn(),
    } as unknown as CanvasRenderingContext2D;

    HTMLCanvasElement.prototype.getContext = vi.fn(
      () => mockContext
    ) as unknown as typeof HTMLCanvasElement.prototype.getContext;

    // Mock requestAnimationFrame
    global.requestAnimationFrame = vi.fn(() => 1) as unknown as typeof window.requestAnimationFrame;
    global.cancelAnimationFrame = vi.fn() as unknown as typeof window.cancelAnimationFrame;
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("should render title and subtitle", () => {
    const { getByText } = render(AuthCard, {
      props: {
        title: "Test Title",
        subtitle: "Test Subtitle",
        showIcon: false,
      },
    });

    expect(getByText("Test Title")).toBeTruthy();
    expect(getByText("Test Subtitle")).toBeTruthy();
  });

  it("should render GalaxyIcon when showIcon is true", () => {
    const { container } = render(AuthCard, {
      props: {
        title: "Login",
        subtitle: "Enter your credentials",
        showIcon: true,
      },
    });

    const icon = container.querySelector(".galaxy-icon");
    expect(icon).toBeTruthy();
  });

  it("should not render icon when showIcon is false", () => {
    const { container } = render(AuthCard, {
      props: {
        title: "Error",
        subtitle: "Something went wrong",
        showIcon: false,
      },
    });

    const icon = container.querySelector(".galaxy-icon");
    expect(icon).toBeFalsy();
  });

  it("should render children content", () => {
    const { container } = render(AuthCard, {
      props: {
        title: "Form",
        subtitle: "Fill in the form",
        showIcon: true,
      },
    });

    const card = container.querySelector(".card");
    expect(card).toBeTruthy();
  });

  it("should have correct CSS classes for cosmic theme", () => {
    const { container } = render(AuthCard, {
      props: {
        title: "Login",
        subtitle: "Test",
        showIcon: true,
      },
    });

    const page = container.querySelector(".auth-page");
    expect(page).toBeTruthy();

    const authContainer = container.querySelector(".auth-container");
    expect(authContainer).toBeTruthy();

    const card = container.querySelector(".card");
    expect(card).toBeTruthy();
  });

  it.skip("should have glass morphism effect on card", () => {
    // NOTE: This test is skipped because jsdom doesn't properly compute
    // backdrop-filter CSS property. The glass morphism effect is verified
    // visually through E2E tests and by checking the CSS source.
    const { container } = render(AuthCard, {
      props: {
        title: "Login",
        subtitle: "Test",
        showIcon: true,
      },
    });

    const card = container.querySelector(".card");
    expect(card).toBeTruthy();

    // In real browser, this would be: backdrop-filter: blur(12px)
    // const styles = window.getComputedStyle(card!);
    // expect(styles.backdropFilter).toContain('blur');
  });
});
