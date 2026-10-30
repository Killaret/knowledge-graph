import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/svelte";
import LinkTypeSelector from "./LinkTypeSelector.svelte";

const testTypes = [
  {
    type: "reference",
    icon: "📖",
    label: "Reference",
    color: "#3366ff",
    description: "One note mentions another.",
    example: "See also",
  },
  {
    type: "dependency",
    icon: "🔗",
    label: "Dependency",
    color: "#ff6600",
    description: "Target depends on source.",
    example: "Requires",
  },
  {
    type: "custom",
    icon: "✨",
    label: "Custom",
    color: "#ff66ff",
    description: "You define it.",
    example: "My label",
  },
];

describe("LinkTypeSelector", () => {
  it("renders all type options", () => {
    render(LinkTypeSelector, { props: { types: testTypes } });

    testTypes.forEach((type) => {
      expect(screen.getByText(type.label)).toBeInTheDocument();
      expect(screen.getByText(type.icon)).toBeInTheDocument();
      expect(screen.getByText(type.description)).toBeInTheDocument();
    });
  });

  it("selects a type when clicked", async () => {
    const onSelect = vi.fn();
    const { rerender } = render(LinkTypeSelector, {
      props: { types: testTypes, selected: "reference", onSelect },
    });

    const dependencyBtn = screen.getByRole("radio", { name: /Dependency/i });
    await fireEvent.click(dependencyBtn);

    expect(onSelect).toHaveBeenCalledWith("dependency");

    rerender({ types: testTypes, selected: "dependency", onSelect });
    expect(dependencyBtn).toHaveAttribute("aria-checked", "true");
  });

  it("uses default selected from first type", () => {
    render(LinkTypeSelector, { props: { types: testTypes } });

    const firstBtn = screen.getByRole("radio", { name: /Reference/i });
    expect(firstBtn).toHaveAttribute("aria-checked", "true");
  });

  it("supports compact size without descriptions", () => {
    render(LinkTypeSelector, { props: { types: testTypes, size: "sm" } });

    expect(screen.queryByText(testTypes[0].description)).not.toBeInTheDocument();
    expect(screen.getByText(testTypes[0].label)).toBeInTheDocument();
  });

  it("is disabled when disabled prop is true", () => {
    render(LinkTypeSelector, { props: { types: testTypes, disabled: true } });

    const buttons = screen.getAllByRole("radio");
    buttons.forEach((btn) => {
      expect(btn).toBeDisabled();
    });
  });
});
