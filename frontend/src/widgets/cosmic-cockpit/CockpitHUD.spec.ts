import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/svelte";
import CockpitHUD from "./CockpitHUD.svelte";

describe("CockpitHUD", () => {
  afterEach(() => {
    cleanup();
  });

  it("renders the cluster value with a static (non-animated) gradient-text style", () => {
    render(CockpitHUD, { props: { cluster: "Deep Space" } });

    const clusterValue = screen.getByTestId("hud-cluster").querySelector(".hud-value");
    expect(clusterValue).toHaveClass("cockpit-gradient-text");
    expect(clusterValue).toHaveClass("cockpit-gradient-text--static");
  });

  it("keeps numeric metrics as plain readable text, not gradient text", () => {
    render(CockpitHUD, { props: { nodeCount: 42, linkCount: 7 } });

    const nodeValue = screen.getByTestId("hud-node-count").querySelector(".hud-value");
    const linkValue = screen.getByTestId("hud-link-count").querySelector(".hud-value");

    expect(nodeValue).not.toHaveClass("cockpit-gradient-text");
    expect(linkValue).not.toHaveClass("cockpit-gradient-text");
    expect(nodeValue).toHaveTextContent("42");
    expect(linkValue).toHaveTextContent("7");
  });

  it("toggles first-person label via the HUD button", () => {
    render(CockpitHUD);

    expect(screen.getByTestId("first-person-toggle")).toBeInTheDocument();
  });
});
