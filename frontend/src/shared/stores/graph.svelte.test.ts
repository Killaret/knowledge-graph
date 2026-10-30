import { describe, it, expect, beforeEach } from "vitest";
import { graphStore } from "./graph.svelte";

describe("graphStore", () => {
  beforeEach(() => {
    graphStore.reset();
  });

  it("has default state", () => {
    expect(graphStore.selectedNodeId).toBeNull();
    expect(graphStore.searchQuery).toBe("");
    expect(graphStore.selectedType).toBe("all");
    expect(graphStore.currentView).toBe("graph");
    expect(graphStore.graphData).toEqual({ nodes: [], links: [] });
    expect(graphStore.hoveredNodeId).toBeNull();
    expect(graphStore.hiddenLinkTypes).toEqual([]);
    expect(graphStore.minLinkWeight).toBe(0);
  });

  it("selects and deselects a node", () => {
    graphStore.selectNode("n1");
    expect(graphStore.selectedNodeId).toBe("n1");
    graphStore.selectNode(null);
    expect(graphStore.selectedNodeId).toBeNull();
  });

  it("toggles link types", () => {
    graphStore.toggleLinkType("related");
    expect(graphStore.hiddenLinkTypes).toEqual(["related"]);

    graphStore.toggleLinkType("related");
    expect(graphStore.hiddenLinkTypes).toEqual([]);

    graphStore.toggleLinkType("parent");
    graphStore.toggleLinkType("related");
    expect(graphStore.hiddenLinkTypes).toEqual(["parent", "related"]);
  });

  it("sets view and search query", () => {
    graphStore.currentView = "3d";
    expect(graphStore.currentView).toBe("3d");

    graphStore.searchQuery = "black hole";
    expect(graphStore.searchQuery).toBe("black hole");
  });

  it("filters type and weight", () => {
    graphStore.selectedType = "star";
    expect(graphStore.selectedType).toBe("star");

    graphStore.minLinkWeight = 2;
    expect(graphStore.minLinkWeight).toBe(2);
  });

  it("updates graph data", () => {
    const data = {
      nodes: [{ id: "n1", title: "Note 1", type: "star" as const }],
      links: [{ source: "n1", target: "n1", weight: 1 }],
    };
    graphStore.graphData = data;
    expect(graphStore.graphData).toEqual(data);
  });

  it("tracks hovered node", () => {
    graphStore.hoveredNodeId = "n1";
    expect(graphStore.hoveredNodeId).toBe("n1");
    graphStore.hoveredNodeId = null;
    expect(graphStore.hoveredNodeId).toBeNull();
  });

  it("clears search", () => {
    graphStore.searchQuery = "query";
    graphStore.clearSearch();
    expect(graphStore.searchQuery).toBe("");
  });

  it("resets to default state", () => {
    graphStore.selectNode("n1");
    graphStore.searchQuery = "query";
    graphStore.selectedType = "planet";
    graphStore.currentView = "list";
    graphStore.hiddenLinkTypes = ["related"];
    graphStore.minLinkWeight = 3;

    graphStore.reset();

    expect(graphStore.selectedNodeId).toBeNull();
    expect(graphStore.searchQuery).toBe("");
    expect(graphStore.selectedType).toBe("all");
    expect(graphStore.currentView).toBe("graph");
    expect(graphStore.graphData).toEqual({ nodes: [], links: [] });
    expect(graphStore.hoveredNodeId).toBeNull();
    expect(graphStore.hiddenLinkTypes).toEqual([]);
    expect(graphStore.minLinkWeight).toBe(0);
  });
});
