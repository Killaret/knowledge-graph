import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, cleanup, fireEvent, waitFor } from "@testing-library/svelte";
import AuthCard from "./AuthCard.svelte";

// Mock browser environment
vi.mock("$app/environment", () => ({
  browser: true,
}));

// Mock GraphCanvas to avoid canvas dependencies in unit tests
vi.mock("$widgets/graph-canvas/GraphCanvas.svelte", () => ({
  default: vi.fn().mockImplementation(() => {
    const div = document.createElement("div");
    div.setAttribute("data-testid", "graph-canvas");
    return div;
  }),
}));

vi.mock("$features/preload/hooks/usePreloadedData", () => ({
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

  it("should not render subtitle when omitted", () => {
    const { container } = render(AuthCard, {
      props: { title: "No Subtitle", showIcon: false },
    });

    expect(container.querySelector(".subtitle")).toBeFalsy();
  });

  it("should open WeltallProtocol when logo is clicked", async () => {
    const { container } = render(AuthCard, {
      props: { title: "Login", showIcon: true },
    });

    const logo = container.querySelector(".logo-button") as HTMLElement;
    expect(logo).toBeTruthy();

    await fireEvent.click(logo);
    expect(container.querySelector(".protocol-overlay")).toBeTruthy();
  });

  it("should handle graph background load failure", async () => {
    const { getGraphWithPreload } = await import("$features/preload/hooks/usePreloadedData");
    vi.mocked(getGraphWithPreload).mockRejectedValue(new Error("load failed"));

    const consoleWarn = vi.spyOn(console, "warn").mockImplementation(() => {});

    render(AuthCard, { props: { title: "Login", showIcon: true } });

    await waitFor(() => expect(getGraphWithPreload).toHaveBeenCalledWith(100));
    expect(consoleWarn).toHaveBeenCalled();

    consoleWarn.mockRestore();
  });
});
