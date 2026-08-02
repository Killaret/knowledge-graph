import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/svelte";
import FloatingAuthPanel from "./FloatingAuthPanel.svelte";

// Mock navigation
vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

// Mock auth store
vi.mock("$shared/stores/auth.svelte", () => ({
  login: vi.fn(),
  loginWithApiKey: vi.fn(),
  register: vi.fn(),
  isLoading: vi.fn().mockReturnValue(false),
  error: vi.fn().mockReturnValue(null),
  isAuthenticated: vi.fn().mockReturnValue(false),
}));

describe("FloatingAuthPanel", () => {
  const baseProps = {
    open: true,
    initialTab: "login" as const,
    onClose: vi.fn(),
    onSuccess: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it("does not render when closed", () => {
    render(FloatingAuthPanel, {
      props: { ...baseProps, open: false },
    });

    expect(screen.queryByTestId("floating-auth-panel")).not.toBeInTheDocument();
  });

  it("renders login tab by default", () => {
    render(FloatingAuthPanel, { props: baseProps });

    const panel = screen.getByTestId("floating-auth-panel");
    expect(panel).toBeInTheDocument();

    const loginTab = screen.getByTestId("floating-auth-tab-login");
    expect(loginTab).toHaveAttribute("aria-selected", "true");

    const registerTab = screen.getByTestId("floating-auth-tab-register");
    expect(registerTab).toHaveAttribute("aria-selected", "false");
  });

  it("renders register tab when initialTab is register", () => {
    render(FloatingAuthPanel, {
      props: { ...baseProps, initialTab: "register" },
    });

    expect(
      screen.getByTestId("floating-auth-tab-register")
    ).toHaveAttribute("aria-selected", "true");
    expect(screen.getByTestId("floating-auth-tab-login")).toHaveAttribute(
      "aria-selected",
      "false"
    );
  });

  it("switches to register tab on click and back", async () => {
    render(FloatingAuthPanel, { props: baseProps });

    const registerTab = screen.getByTestId("floating-auth-tab-register");
    await fireEvent.click(registerTab);

    expect(registerTab).toHaveAttribute("aria-selected", "true");
    expect(screen.getByTestId("floating-auth-tab-login")).toHaveAttribute(
      "aria-selected",
      "false"
    );

    const loginTab = screen.getByTestId("floating-auth-tab-login");
    await fireEvent.click(loginTab);

    expect(loginTab).toHaveAttribute("aria-selected", "true");
    expect(screen.getByTestId("floating-auth-tab-register")).toHaveAttribute(
      "aria-selected",
      "false"
    );
  });

  it("calls onClose when close button is clicked", async () => {
    render(FloatingAuthPanel, { props: baseProps });

    const closeBtn = screen.getByTestId("floating-auth-close");
    await fireEvent.click(closeBtn);

    expect(baseProps.onClose).toHaveBeenCalledTimes(1);
  });

  it("calls onClose when Escape is pressed", async () => {
    render(FloatingAuthPanel, { props: baseProps });

    await fireEvent.keyDown(window, { key: "Escape" });

    expect(baseProps.onClose).toHaveBeenCalledTimes(1);
  });

  it("provides a draggable header", async () => {
    render(FloatingAuthPanel, { props: baseProps });

    const header = screen.getByTestId("floating-auth-drag-handle");
    expect(header).toBeInTheDocument();
    expect(header).toHaveAttribute("aria-label", "Drag to move");
  });

  it("resets to initial tab when reopened", async () => {
    const { rerender } = render(FloatingAuthPanel, {
      props: { ...baseProps, open: false },
    });

    // Open with register tab
    rerender({ ...baseProps, open: true, initialTab: "register" });

    const registerTab = await screen.findByTestId(
      "floating-auth-tab-register"
    );
    expect(registerTab).toHaveAttribute("aria-selected", "true");
  });
});
