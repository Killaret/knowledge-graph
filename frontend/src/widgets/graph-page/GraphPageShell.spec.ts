import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/svelte";
import { isAuthenticated } from "$shared/stores/auth.svelte";
import { authState } from "$shared/stores/auth-session.svelte";
import { graphView } from "$shared/stores/graph-view.svelte";
import GraphPageShellTestWrapper from "./GraphPageShellTestWrapper.svelte";

vi.mock("$shared/stores/auth.svelte", async () => {
  const actual = await vi.importActual<typeof import("$shared/stores/auth.svelte")>(
    "$shared/stores/auth.svelte"
  );
  return { ...actual, isAuthenticated: vi.fn(() => false) };
});

describe("GraphPageShell", () => {
  beforeEach(() => {
    authState.currentUser = null;
    authState.accessToken = null;
    authState.apiKey = null;
    graphView.clear();
    vi.mocked(isAuthenticated).mockReturnValue(false);
  });

  it("renders with notes and computes type counts", async () => {
    const { container } = render(GraphPageShellTestWrapper, {
      props: {
        notes: [
          { id: "n1", title: "Star note", type: "star" },
          { id: "n2", title: "Planet note", type: "planet" },
        ],
      },
    });

    expect(container.querySelector('[data-testid="shell-children"]')).toBeInTheDocument();
    expect(screen.getByText(/All/)).toBeInTheDocument();
  });

  it("falls back to nodes when notes are missing", () => {
    render(GraphPageShellTestWrapper, {
      props: {
        nodes: [{ id: "n3", title: "Node", type: "star" }],
      },
    });
    expect(screen.getByTestId("type-dropdown-toggle")).toBeInTheDocument();
  });

  it("shows sign in and register when not authenticated", () => {
    render(GraphPageShellTestWrapper, {
      props: {
        onSignIn: vi.fn(),
        onRegister: vi.fn(),
      },
    });
    expect(screen.getByTestId("top-bar-sign-in")).toBeInTheDocument();
    expect(screen.getByTestId("top-bar-register")).toBeInTheDocument();
    expect(screen.queryByTestId("menu-import")).not.toBeInTheDocument();
  });

  it("shows authenticated controls when user is signed in", () => {
    vi.mocked(isAuthenticated).mockReturnValue(true);
    render(GraphPageShellTestWrapper, {
      props: {
        onImport: vi.fn(),
        onExport: vi.fn(),
        onNoteCreate: vi.fn(),
        onNoteDelete: vi.fn(),
        onNoteEdit: vi.fn(),
        onCreateChildNote: vi.fn(),
        onToggleFullGraph: vi.fn(),
        showFullGraph: true,
      },
    });
    expect(screen.getByTestId("menu-import")).toBeInTheDocument();
    expect(screen.getByTestId("menu-export")).toBeInTheDocument();
    expect(screen.queryByTestId("top-bar-sign-in")).not.toBeInTheDocument();
  });

  it("PUB-2 renders personal/community toggle and selects personal for authenticated", () => {
    authState.accessToken = "tok1";
    vi.mocked(isAuthenticated).mockReturnValue(true);
    render(GraphPageShellTestWrapper, {});

    const personal = screen.getByTestId("graph-view-personal");
    const community = screen.getByTestId("graph-view-community");
    expect(personal).toBeInTheDocument();
    expect(community).toBeInTheDocument();
    expect(personal.getAttribute("aria-pressed")).toBe("true");
    expect(community.getAttribute("aria-pressed")).toBe("false");
  });

  it("PUB-2 clicking community toggles the aria-pressed state", async () => {
    authState.accessToken = "tok1";
    vi.mocked(isAuthenticated).mockReturnValue(true);
    render(GraphPageShellTestWrapper, {});

    const community = screen.getByTestId("graph-view-community");
    await fireEvent.click(community);

    expect(community.getAttribute("aria-pressed")).toBe("true");
    expect(screen.getByTestId("graph-view-personal").getAttribute("aria-pressed")).toBe("false");
  });

  it("PUB-2 does not render graph-view-toggle for anonymous users even with stored personal", () => {
    localStorage.setItem("graph-view-mode", "personal");
    graphView.clear();
    graphView.restore();
    render(GraphPageShellTestWrapper, {});

    expect(screen.queryByTestId("graph-view-toggle")).not.toBeInTheDocument();
    expect(screen.queryByTestId("graph-view-personal")).not.toBeInTheDocument();
    expect(screen.queryByTestId("graph-view-community")).not.toBeInTheDocument();
  });
});
