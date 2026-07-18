import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/svelte";
import SplashScreen from "./SplashScreen.svelte";

vi.mock("$app/environment", () => ({
  browser: true,
}));

describe("SplashScreen", () => {
  it("renders on first visit", () => {
    render(SplashScreen, {
      props: {},
    });

    expect(
      screen.getByRole("img", { name: /weltall protocol/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("Knowledge Graph")).toBeInTheDocument();
    expect(screen.getByText("Explore the cosmos of ideas")).toBeInTheDocument();
  });

  it("has correct accessibility attributes", () => {
    render(SplashScreen, {
      props: {},
    });

    const splash = screen.getByRole("img", {
      name: /weltall protocol - knowledge graph/i,
    });
    expect(splash).toBeInTheDocument();
  });
});
