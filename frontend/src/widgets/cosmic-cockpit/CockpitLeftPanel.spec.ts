import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, fireEvent, cleanup, waitFor } from "@testing-library/svelte";
import CockpitLeftPanel from "./CockpitLeftPanel.svelte";
import type { User } from "$shared/types";

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

const authMock = vi.hoisted(() => ({
  isAuthenticated: vi.fn<() => boolean>(() => false),
  currentUser: vi.fn<() => User | null>(() => null),
  logout: vi.fn(),
}));

vi.mock("$shared/stores/auth.svelte", () => authMock);

afterEach(() => {
  cleanup();
  authMock.isAuthenticated.mockReset().mockReturnValue(false);
  authMock.currentUser.mockReset().mockReturnValue(null);
});

describe("CockpitLeftPanel", () => {
  it("renders navigation links", () => {
    const { container } = render(CockpitLeftPanel);
    const links = container.querySelectorAll(".nav-link");
    expect(links.length).toBeGreaterThan(0);
    expect(links[0].getAttribute("href")).toBe("/");
  });

  it("navigates when a nav link is clicked", async () => {
    const { goto } = await import("$app/navigation");
    const { container } = render(CockpitLeftPanel);
    const graphLink = container.querySelector('.nav-link[href="/graph"]') as HTMLElement;
    expect(graphLink).toBeTruthy();

    await fireEvent.click(graphLink);
    await waitFor(() => expect(goto).toHaveBeenCalledWith("/graph"));
  });

  it("does not show the user badge when not authenticated", () => {
    render(CockpitLeftPanel);
    expect(screen.queryByTestId("user-badge")).toBeNull();
  });

  it("shows the user badge when authenticated", () => {
    authMock.isAuthenticated.mockReturnValue(true);
    authMock.currentUser.mockReturnValue({ email: "cosmonaut@example.com" } as User);

    render(CockpitLeftPanel);

    expect(screen.getByTestId("user-badge")).toBeTruthy();
    expect(screen.getByText("cosmonaut@example.com")).toBeTruthy();
  });

  it("renders the full graph toggle when onToggleFullGraph is provided", () => {
    render(CockpitLeftPanel, {
      props: {
        onToggleFullGraph: vi.fn(),
        showFullGraph: true,
      },
    });

    expect(screen.getByTestId("full-graph-toggle")).toBeTruthy();
  });

  it("does not render the full graph toggle when onToggleFullGraph is missing", () => {
    render(CockpitLeftPanel);
    expect(screen.queryByTestId("full-graph-toggle")).toBeNull();
  });

  it("toggles full graph state via the checkbox", async () => {
    const onToggleFullGraph = vi.fn();
    render(CockpitLeftPanel, {
      props: {
        onToggleFullGraph,
        showFullGraph: false,
      },
    });

    const checkbox = screen.getByTestId("full-graph-toggle") as HTMLInputElement;
    expect(checkbox.checked).toBe(false);

    await fireEvent.click(checkbox);
    expect(onToggleFullGraph).toHaveBeenCalledWith(true);
  });

  it("expands and renders the note list", async () => {
    const onNoteSelect = vi.fn();
    const { container } = render(CockpitLeftPanel, {
      props: {
        notes: [
          { id: "1", title: "Alpha Note" },
          { id: "2", title: "Beta Note" },
        ],
        onNoteSelect,
      },
    });

    const toggle = container.querySelector("#note-list-heading") as HTMLElement;
    await fireEvent.click(toggle);

    const items = screen.getAllByTestId("cockpit-note-tree-item");
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent("Alpha Note");

    await fireEvent.click(items[0]);
    expect(onNoteSelect).toHaveBeenCalledWith("1");
  });

  it("renders import and export buttons when callbacks are provided", () => {
    render(CockpitLeftPanel, {
      props: { onImport: vi.fn(), onExport: vi.fn() },
    });

    expect(screen.getByTestId("menu-import")).toBeTruthy();
    expect(screen.getByTestId("menu-export")).toBeTruthy();
  });

  it("does not render import/export buttons when callbacks are missing", () => {
    render(CockpitLeftPanel);
    expect(screen.queryByTestId("menu-import")).toBeNull();
    expect(screen.queryByTestId("menu-export")).toBeNull();
  });

  it("does not render removed graph controls, type filter, or link type chips", () => {
    const { container } = render(CockpitLeftPanel, {
      props: {
        onToggleFullGraph: vi.fn(),
      },
    });

    expect(container.querySelector(".graph-controls")).toBeNull();
    expect(screen.queryByTestId("cockpit-type-filter")).toBeNull();
    expect(screen.queryByTestId("link-type-chip-parent")).toBeNull();
    expect(screen.queryByTestId("link-types-show-all")).toBeNull();
    expect(screen.queryByTestId("link-types-hide-all")).toBeNull();
    expect(screen.queryByTestId("cockpit-min-weight")).toBeNull();
  });
});
