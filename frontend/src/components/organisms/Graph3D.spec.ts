import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/svelte";
import Graph3D from "./Graph3D.svelte";

const nodes = [
  { id: "1", title: "Star A", type: "star" },
  { id: "2", title: "Planet B", type: "planet" },
];

const links = [{ source: "1", target: "2", weight: 0.8, link_type: "related" }];

describe("Graph3D", () => {
  it("renders the 3D graph container", async () => {
    render(Graph3D, { props: { nodes, links } });
    expect(screen.getByTestId("graph-3d-container")).toBeInTheDocument();
  });

  it("is accessible by role", async () => {
    render(Graph3D, { props: { nodes, links } });
    expect(screen.getByRole("img")).toBeInTheDocument();
  });
});
