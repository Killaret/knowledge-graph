import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, cleanup } from "@testing-library/svelte";
import CockpitHUD from "./CockpitHUD.svelte";
import { cockpitStore } from "$features/cosmic-cockpit";

vi.mock("$features/cosmic-cockpit", () => ({
  cockpitStore: {
    syncing: false,
    lastSyncAt: null,
    fps: 60,
    firstPerson: false,
    toggleFirstPerson: vi.fn(),
  },
}));

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

  it("renders default cluster placeholder", () => {
    render(CockpitHUD);

    const clusterValue = screen.getByTestId("hud-cluster").querySelector(".hud-value");
    expect(clusterValue).toHaveTextContent(/Unknown|unknown/);
  });

  it("renders different health color bands", () => {
    const { rerender } = render(CockpitHUD, { props: { health: 30 } });
    const healthFill = () => screen.getByTestId("hud-health").querySelector(".health-fill") as HTMLElement;

    expect(healthFill().style.background).toContain("#f87171");

    rerender({ health: 60 });
    expect(healthFill().style.background).toContain("#facc15");

    rerender({ health: 90 });
    expect(healthFill().style.background).toContain("#2dd4bf");
  });

  it("shows sync ago when recently synced", () => {
    (cockpitStore as any).lastSyncAt = Date.now() - 5000;
    (cockpitStore as any).syncing = false;
    render(CockpitHUD);
    const syncValue = screen.getByTestId("hud-sync").querySelector(".hud-value");
    expect(syncValue).toHaveTextContent(/5|ago/);
  });

  it("shows syncing state", () => {
    (cockpitStore as any).lastSyncAt = null;
    (cockpitStore as any).syncing = true;
    render(CockpitHUD);
    expect(screen.getByTestId("hud-sync").querySelector(".hud-value")).toHaveTextContent(/syncing/i);
  });

  it("displays -- for non-finite fps", () => {
    (cockpitStore as any).fps = NaN;
    render(CockpitHUD);
    expect(screen.getByTestId("hud-fps").querySelector(".hud-value")).toHaveTextContent("--");
  });

  it("shows exit first person when active", () => {
    (cockpitStore as any).firstPerson = true;
    render(CockpitHUD);
    expect(screen.getByTestId("first-person-toggle")).toHaveTextContent(/exit/i);
  });
});
