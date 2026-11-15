import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/svelte";
import LinkTargetPicker from "./LinkTargetPicker.svelte";

describe("LinkTargetPicker", () => {
  const nodes = [
    { id: "src", title: "Source note", type: "star" },
    { id: "a", title: "Alpha", type: "planet" },
    { id: "b", title: "Beta galaxy", type: "galaxy" },
    { id: "kc", title: "Knowledge Core", type: "technical" },
  ];

  afterEach(() => {
    cleanup();
  });

  function renderPicker(extra: Record<string, unknown> = {}) {
    return render(LinkTargetPicker, {
      props: {
        x: 100,
        y: 100,
        visible: true,
        sourceId: "src",
        nodes,
        onSelect: vi.fn(),
        onClose: vi.fn(),
        ...extra,
      },
    });
  }

  it("does not render when not visible", () => {
    renderPicker({ visible: false });
    expect(screen.queryByTestId("link-target-picker")).not.toBeInTheDocument();
  });

  it("lists all notes except the source and the technical core", () => {
    renderPicker();
    const options = screen.getAllByTestId("link-target-option");
    const titles = options.map((o) => o.textContent?.trim());
    expect(titles).toEqual(["Alpha", "Beta galaxy"]);
  });

  it("filters candidates by substring, case-insensitive", async () => {
    renderPicker();
    await fireEvent.input(screen.getByTestId("link-target-search"), {
      target: { value: "alp" },
    });
    const options = screen.getAllByTestId("link-target-option");
    expect(options).toHaveLength(1);
    expect(options[0]).toHaveTextContent("Alpha");
  });

  it("shows the empty state when nothing matches", async () => {
    renderPicker();
    await fireEvent.input(screen.getByTestId("link-target-search"), {
      target: { value: "zzz" },
    });
    expect(screen.queryAllByTestId("link-target-option")).toHaveLength(0);
    expect(screen.getByTestId("link-target-empty")).toBeInTheDocument();
  });

  it("calls onSelect with the picked node and closes", async () => {
    const onSelect = vi.fn();
    const onClose = vi.fn();
    renderPicker({ onSelect, onClose });

    await fireEvent.click(screen.getAllByTestId("link-target-option")[0]);

    expect(onSelect).toHaveBeenCalledWith(nodes[1]);
    expect(onClose).toHaveBeenCalledOnce();
  });
});
