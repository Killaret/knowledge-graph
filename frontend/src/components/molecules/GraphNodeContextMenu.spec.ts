import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/svelte";
import GraphNodeContextMenu from "./GraphNodeContextMenu.svelte";

describe("GraphNodeContextMenu", () => {
  const node = { id: "n1", title: "Parent note", type: "star" };

  afterEach(() => {
    cleanup();
  });

  it("does not render when not visible", () => {
    render(GraphNodeContextMenu, {
      props: { x: 100, y: 100, visible: false, node, onClose: vi.fn(), onCreateChild: vi.fn() },
    });

    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });

  it("renders node title and menu items when visible", () => {
    render(GraphNodeContextMenu, {
      props: { x: 100, y: 100, visible: true, node, onClose: vi.fn(), onCreateChild: vi.fn() },
    });

    expect(screen.getByText("Parent note")).toBeInTheDocument();
    expect(screen.getByTestId("context-menu-create-child")).toBeInTheDocument();
  });

  it("calls onCreateChild and onClose when create child is clicked", async () => {
    const onCreateChild = vi.fn();
    const onClose = vi.fn();

    render(GraphNodeContextMenu, {
      props: { x: 100, y: 100, visible: true, node, onClose, onCreateChild },
    });

    const btn = screen.getByTestId("context-menu-create-child");
    await fireEvent.click(btn);

    expect(onCreateChild).toHaveBeenCalledOnce();
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("renders view details button only when onViewDetails is provided", () => {
    const { rerender } = render(GraphNodeContextMenu, {
      props: { x: 100, y: 100, visible: true, node, onClose: vi.fn(), onCreateChild: vi.fn() },
    });

    expect(screen.queryByTestId("context-menu-view-details")).not.toBeInTheDocument();

    rerender({ onViewDetails: vi.fn() });
    expect(screen.getByTestId("context-menu-view-details")).toBeInTheDocument();
  });

  it("calls onViewDetails and onClose when view details is clicked", async () => {
    const onViewDetails = vi.fn();
    const onClose = vi.fn();

    render(GraphNodeContextMenu, {
      props: {
        x: 100,
        y: 100,
        visible: true,
        node,
        onClose,
        onCreateChild: vi.fn(),
        onViewDetails,
      },
    });

    const btn = screen.getByTestId("context-menu-view-details");
    await fireEvent.click(btn);

    expect(onViewDetails).toHaveBeenCalledOnce();
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("closes on Escape key", async () => {
    const onClose = vi.fn();

    render(GraphNodeContextMenu, {
      props: { x: 100, y: 100, visible: true, node, onClose, onCreateChild: vi.fn() },
    });

    await fireEvent.keyDown(window, { key: "Escape" });

    await waitFor(() => {
      expect(onClose).toHaveBeenCalledOnce();
    });
  });

  it("closes on click outside the menu", async () => {
    const onClose = vi.fn();

    render(GraphNodeContextMenu, {
      props: { x: 100, y: 100, visible: true, node, onClose, onCreateChild: vi.fn() },
    });

    await fireEvent.click(document.body);

    await waitFor(() => {
      expect(onClose).toHaveBeenCalledOnce();
    });
  });
});
