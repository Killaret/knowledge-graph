import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, cleanup } from "@testing-library/svelte";
import CockpitPanel from "./CockpitPanel.svelte";
import { cockpitStore } from "$features/cosmic-cockpit";

vi.mock("$app/environment", () => ({
  browser: true,
}));

function resetPanels() {
  (["top", "bottom", "left", "right"] as const).forEach((position) => {
    cockpitStore.setPanel(position, { open: false, pinned: false, hovering: false });
  });
}

describe("CockpitPanel", () => {
  afterEach(() => {
    resetPanels();
    cleanup();
  });

  it.each(["right", "bottom", "left", "top"] as const)(
    "shows the pull-handle for %s when collapsed",
    (position) => {
      render(CockpitPanel, { props: { position, size: 300, title: "Test" } });

      expect(screen.getByTestId(`cockpit-handle-${position}`)).toBeInTheDocument();
    }
  );

  it.each(["right", "bottom", "left", "top"] as const)(
    "hides the pull-handle for %s once the panel is open",
    (position) => {
      cockpitStore.setPanel(position, { open: true });

      render(CockpitPanel, { props: { position, size: 300, title: "Test" } });

      expect(screen.queryByTestId(`cockpit-handle-${position}`)).not.toBeInTheDocument();
    }
  );

  it("renders the panel title with the animated gradient-text class", () => {
    render(CockpitPanel, { props: { position: "top", size: 64, title: "Navigation" } });

    const title = screen.getByText("Navigation");
    expect(title).toHaveClass("cockpit-gradient-text");
    expect(title.getAttribute("style")).toContain("--cockpit-text-delay");
  });

  it("uses a different gradient-text delay for different panel positions", () => {
    const { unmount } = render(CockpitPanel, {
      props: { position: "top", size: 64, title: "Top Title" },
    });
    const topStyle = screen.getByText("Top Title").getAttribute("style");
    unmount();
    cleanup();

    render(CockpitPanel, { props: { position: "left", size: 300, title: "Left Title" } });
    const leftStyle = screen.getByText("Left Title").getAttribute("style");

    expect(topStyle).not.toEqual(leftStyle);
  });
});
