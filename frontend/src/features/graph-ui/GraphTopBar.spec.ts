import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, cleanup, waitFor } from "@testing-library/svelte";
import GraphTopBar from "./GraphTopBar.svelte";
import { graphStore } from "$shared/stores/graph.svelte";
import { isGraph3DEnabled } from "$shared/config/config";
import { getSmartCreateAvailability } from "$shared/api/import";
import { LinkType } from "$entities";

vi.mock("$shared/config/config", async (importOriginal) => ({
  ...(await importOriginal<typeof import("$shared/config/config")>()),
  isGraph3DEnabled: vi.fn(() => true),
}));

vi.mock("$shared/api/import", () => ({
  getSmartCreateAvailability: vi.fn(),
}));

describe("GraphTopBar", () => {
  const typeFilters = [
    { id: "all", label: "All", emoji: "🔎" },
    { id: "note", label: "Notes", emoji: "📝" },
    { id: "bookmark", label: "Bookmarks", emoji: "📑" },
  ];

  beforeEach(() => {
    cleanup();
    vi.clearAllMocks();
    graphStore.reset();
    vi.mocked(isGraph3DEnabled).mockReturnValue(true);
    vi.mocked(getSmartCreateAvailability).mockResolvedValue({ available: true });
  });

  it("renders authenticated view with stats and view toggles", () => {
    const onToggleView = vi.fn();
    render(GraphTopBar, {
      props: {
        isAuthenticated: true,
        currentView: "graph",
        nodeCount: 12,
        linkCount: 5,
        onToggleView,
      },
    });

    expect(screen.getByTestId("graph-stats")).toHaveTextContent("12");
    expect(screen.getByTestId("graph-stats")).toHaveTextContent("5");
    expect(screen.getByTestId("view-toggle-graph")).toBeInTheDocument();
    expect(screen.getByTestId("view-toggle-3d")).toBeInTheDocument();
    expect(screen.getByTestId("view-toggle-list")).toBeInTheDocument();
  });

  it("FREEZE-3D-1: hides the 3D toggle while frontend.graph.3d.enabled is false", () => {
    vi.mocked(isGraph3DEnabled).mockReturnValue(false);
    render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "graph", onToggleView: vi.fn() },
    });

    expect(screen.getByTestId("view-toggle-graph")).toBeInTheDocument();
    expect(screen.getByTestId("view-toggle-list")).toBeInTheDocument();
    expect(screen.queryByTestId("view-toggle-3d")).toBeNull();
  });

  it("FREEZE-3D-1: shows the 3D toggle when enabled", () => {
    vi.mocked(isGraph3DEnabled).mockReturnValue(true);
    render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "graph", onToggleView: vi.fn() },
    });

    expect(screen.getByTestId("view-toggle-3d")).toBeInTheDocument();
  });

  it("switches views and calls onToggleView", async () => {
    const onToggleView = vi.fn();
    render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "graph", onToggleView },
    });

    const listButton = screen.getByTestId("view-toggle-list");
    await fireEvent.click(listButton);

    expect(onToggleView).toHaveBeenCalledWith("list");
  });

  it("shows layout provider toggle only in 3d view", () => {
    const onToggleLayoutProvider = vi.fn();
    const { rerender } = render(GraphTopBar, {
      props: {
        isAuthenticated: true,
        currentView: "graph",
        layoutProvider: "d3",
        onToggleLayoutProvider,
      },
    });

    expect(screen.queryByTestId("layout-provider-d3")).not.toBeInTheDocument();

    rerender({ currentView: "3d", layoutProvider: "graph-service" });

    expect(screen.getByTestId("layout-provider-d3")).toBeInTheDocument();
    expect(screen.getByTestId("layout-provider-graph-service")).toBeInTheDocument();
  });

  it("toggles layout provider", async () => {
    const onToggleLayoutProvider = vi.fn();
    render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "3d", onToggleLayoutProvider },
    });

    const graphService = screen.getByTestId("layout-provider-graph-service");
    await fireEvent.click(graphService);

    expect(onToggleLayoutProvider).toHaveBeenCalledWith("graph-service");
  });

  it("calls onSearch when typing in search box", async () => {
    const onSearch = vi.fn();
    render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "graph", onSearch },
    });

    const input = screen.getByTestId("top-bar-search-input");
    await fireEvent.input(input, { target: { value: "hello" } });

    expect(onSearch).toHaveBeenCalledWith("hello");
  });

  it("opens type filter dropdown and selects a filter", async () => {
    const onFilter = vi.fn();
    render(GraphTopBar, {
      props: {
        isAuthenticated: true,
        currentView: "graph",
        typeFilters,
        selectedType: "all",
        onFilter,
      },
    });

    const toggle = screen.getByTestId("type-dropdown-toggle");
    await fireEvent.click(toggle);

    const noteButton = screen.getByTestId("filter-chip-note");
    await fireEvent.click(noteButton);

    expect(onFilter).toHaveBeenCalledWith("note");
  });

  it("displays type counts in dropdown", async () => {
    render(GraphTopBar, {
      props: {
        isAuthenticated: true,
        currentView: "graph",
        typeFilters,
        typeCounts: { all: 10, note: 6 },
      },
    });

    const toggle = screen.getByTestId("type-dropdown-toggle");
    await fireEvent.click(toggle);

    const allButton = screen.getByTestId("filter-chip-all");
    expect(allButton).toHaveTextContent("10");
  });

  it("opens link type dropdown and toggles a link type", async () => {
    const firstType = LinkType.UI_TYPES[0];
    render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "graph" },
    });

    const toggle = screen.getByTestId("link-dropdown-toggle");
    await fireEvent.click(toggle);

    const chip = screen.getByTestId(`link-type-chip-${firstType.type}`);
    await fireEvent.click(chip);

    expect(graphStore.hiddenLinkTypes).toContain(firstType.type);

    await fireEvent.click(chip);
    expect(graphStore.hiddenLinkTypes).not.toContain(firstType.type);
  });

  it("shows and hides all link types", async () => {
    render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "graph" },
    });

    const toggle = screen.getByTestId("link-dropdown-toggle");
    await fireEvent.click(toggle);

    const hideAll = screen.getByTestId("link-types-hide-all");
    const showAll = screen.getByTestId("link-types-show-all");

    await fireEvent.click(hideAll);
    expect(graphStore.hiddenLinkTypes).toHaveLength(LinkType.UI_TYPES.length);

    await fireEvent.click(showAll);
    expect(graphStore.hiddenLinkTypes).toHaveLength(0);
  });

  it("updates min link weight via slider", async () => {
    render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "graph" },
    });

    const toggle = screen.getByTestId("link-dropdown-toggle");
    await fireEvent.click(toggle);

    const slider = screen.getByTestId("top-bar-min-weight");
    await fireEvent.input(slider, { target: { value: "0.5" } });

    expect(graphStore.minLinkWeight).toBe(0.5);
  });

  it("renders canvas controller buttons and invokes callbacks", async () => {
    const controller = {
      focusMode: false,
      fogEnabled: false,
      resetView: vi.fn(),
      openSearch: vi.fn(),
      toggleFocus: vi.fn(),
      toggleFog: vi.fn(),
    };

    render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "graph", canvasController: controller },
    });

    await fireEvent.click(screen.getByTestId("top-bar-reset"));
    await fireEvent.click(screen.getByTestId("top-bar-open-search"));
    await fireEvent.click(screen.getByTestId("top-bar-focus"));
    await fireEvent.click(screen.getByTestId("top-bar-fog"));

    expect(controller.resetView).toHaveBeenCalled();
    expect(controller.openSearch).toHaveBeenCalled();
    expect(controller.toggleFocus).toHaveBeenCalled();
    expect(controller.toggleFog).toHaveBeenCalled();
  });

  it("shows create note button when authenticated", () => {
    const onNoteCreate = vi.fn();
    render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "graph", onNoteCreate },
    });

    const button = screen.getByTestId("create-note-button");
    expect(button).toBeInTheDocument();

    fireEvent.click(button);
    expect(onNoteCreate).toHaveBeenCalled();
  });

  // SMART-CREATE-1: the button sits left of «＋» and is grey while the Java
  // listener has not written its liveness key.
  it("renders the smart create button before the create note button", async () => {
    render(GraphTopBar, {
      props: {
        isAuthenticated: true,
        currentView: "graph",
        onNoteCreate: vi.fn(),
        onSmartCreate: vi.fn(),
      },
    });

    const smart = screen.getByTestId("smart-create-button");
    const create = screen.getByTestId("create-note-button");
    expect(smart.compareDocumentPosition(create) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it("calls onSmartCreate when the handler is available", async () => {
    const onSmartCreate = vi.fn();
    render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "graph", onSmartCreate },
    });

    const button = screen.getByTestId("smart-create-button");
    await waitFor(() => expect(button).toHaveAttribute("aria-disabled", "false"));

    await fireEvent.click(button);
    expect(onSmartCreate).toHaveBeenCalled();
  });

  it("is grey and inert while the document handler is unavailable", async () => {
    vi.mocked(getSmartCreateAvailability).mockResolvedValue({ available: false });
    const onSmartCreate = vi.fn();
    render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "graph", onSmartCreate },
    });

    const button = screen.getByTestId("smart-create-button");
    await waitFor(() => expect(button).toHaveAttribute("aria-disabled", "true"));
    expect(button).toHaveAttribute(
      "title",
      'Document handler is unavailable. A regular note can be created with the "+" button.'
    );

    await fireEvent.click(button);
    expect(onSmartCreate).not.toHaveBeenCalled();
  });

  it("hides the smart create button for anonymous users", () => {
    render(GraphTopBar, {
      props: { isAuthenticated: false, currentView: "graph", onSmartCreate: vi.fn() },
    });
    expect(screen.queryByTestId("smart-create-button")).toBeNull();
  });

  it("shows the waiting icon while a task runs and the badge when it finishes", () => {
    const { rerender } = render(GraphTopBar, {
      props: { isAuthenticated: true, currentView: "graph", onSmartCreate: vi.fn() },
    });

    expect(screen.queryByTestId("smart-create-waiting")).toBeNull();
    expect(screen.queryByTestId("smart-create-badge")).toBeNull();

    rerender({ smartCreatePending: true });
    expect(screen.getByTestId("smart-create-waiting")).toBeInTheDocument();

    rerender({ smartCreatePending: false, smartCreateUnread: true });
    const badge = screen.getByTestId("smart-create-badge");
    expect(badge).toHaveTextContent("1");
  });

  it("shows sign in and register buttons when not authenticated", () => {
    const onSignIn = vi.fn();
    const onRegister = vi.fn();
    render(GraphTopBar, {
      props: { isAuthenticated: false, currentView: "graph", onSignIn, onRegister },
    });

    const signIn = screen.getByTestId("top-bar-sign-in");
    const register = screen.getByTestId("top-bar-register");

    expect(signIn).toBeInTheDocument();
    expect(register).toBeInTheDocument();

    fireEvent.click(signIn);
    fireEvent.click(register);

    expect(onSignIn).toHaveBeenCalled();
    expect(onRegister).toHaveBeenCalled();
  });
});
