import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/svelte";
import GraphTopBar from "./GraphTopBar.svelte";
import { graphStore } from "$shared/stores/graph.svelte";

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

vi.mock("$components/atoms/LangSwitcher.svelte", () => ({
  default: vi.fn(),
}));

const typeFilters = [
  { id: "all", label: "All", emoji: "🌌" },
  { id: "star", label: "Star", emoji: "⭐" },
];

const canvasController = {
  focusMode: false,
  fogEnabled: true,
  resetView: vi.fn(),
  openSearch: vi.fn(),
  toggleFocus: vi.fn(),
  toggleFog: vi.fn(),
};

describe("GraphTopBar — unified top bar", () => {
  afterEach(() => {
    graphStore.hiddenLinkTypes = [];
    cleanup();
  });

  it("renders the top bar and canvas controls", () => {
    render(GraphTopBar, {
      props: {
        isAuthenticated: false,
        currentView: "graph",
        searchQuery: "",
        selectedType: "all",
        typeFilters,
        nodeCount: 5,
        linkCount: 3,
        canvasController,
      },
    });

    expect(screen.getByTestId("graph-top-bar")).toBeInTheDocument();
    expect(screen.getByTestId("top-bar-search-input")).toBeInTheDocument();
    expect(screen.getByTestId("top-bar-reset")).toBeInTheDocument();
    expect(screen.getByTestId("top-bar-open-search")).toBeInTheDocument();
    expect(screen.getByTestId("top-bar-focus")).toBeInTheDocument();
    expect(screen.getByTestId("top-bar-fog")).toBeInTheDocument();
    expect(screen.getByTestId("view-toggle-graph")).toBeInTheDocument();
  });

  it("toggles fog via canvas controller", async () => {
    render(GraphTopBar, {
      props: {
        isAuthenticated: false,
        currentView: "graph",
        searchQuery: "",
        selectedType: "all",
        typeFilters,
        canvasController,
      },
    });

    const fogBtn = screen.getByTestId("top-bar-fog");
    expect(fogBtn).toHaveAttribute("aria-pressed", "true");

    await fireEvent.click(fogBtn);
    expect(canvasController.toggleFog).toHaveBeenCalled();
  });

  it("shows auth buttons for public and create button for authenticated", () => {
    const { unmount } = render(GraphTopBar, {
      props: {
        isAuthenticated: false,
        currentView: "graph",
        searchQuery: "",
        selectedType: "all",
        typeFilters,
        onSignIn: vi.fn(),
        onRegister: vi.fn(),
      },
    });

    expect(screen.getByTestId("top-bar-sign-in")).toBeInTheDocument();
    expect(screen.getByTestId("top-bar-register")).toBeInTheDocument();
    expect(screen.queryByTestId("create-note-button")).not.toBeInTheDocument();

    unmount();
    cleanup();

    render(GraphTopBar, {
      props: {
        isAuthenticated: true,
        currentView: "graph",
        searchQuery: "",
        selectedType: "all",
        typeFilters,
        onNoteCreate: vi.fn(),
      },
    });

    expect(screen.getByTestId("create-note-button")).toBeInTheDocument();
    expect(screen.queryByTestId("top-bar-sign-in")).not.toBeInTheDocument();
    expect(screen.queryByTestId("top-bar-register")).not.toBeInTheDocument();
  });

  it("emits view toggle and search callbacks", async () => {
    const onToggleView = vi.fn();
    const onSearch = vi.fn();

    render(GraphTopBar, {
      props: {
        isAuthenticated: false,
        currentView: "graph",
        searchQuery: "",
        selectedType: "all",
        typeFilters,
        onToggleView,
        onSearch,
      },
    });

    await fireEvent.click(screen.getByTestId("view-toggle-3d"));
    expect(onToggleView).toHaveBeenCalledWith("3d");

    const searchInput = screen.getByTestId("top-bar-search-input") as HTMLInputElement;
    await fireEvent.input(searchInput, { target: { value: "black hole" } });
    expect(onSearch).toHaveBeenCalledWith("black hole");
  });
});
