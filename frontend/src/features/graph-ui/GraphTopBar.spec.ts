import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/svelte";
import GraphTopBar from "./GraphTopBar.svelte";
import { graphStore } from "$shared/stores/graph.svelte";
import { LinkType } from "$entities";

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
      props: { isAuthenticated: true, currentView: "graph", layoutProvider: "d3", onToggleLayoutProvider },
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
    const firstType = LinkType.ALL_TYPES[0];
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
    expect(graphStore.hiddenLinkTypes).toHaveLength(LinkType.ALL_TYPES.length);

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
