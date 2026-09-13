import { describe, it, expect, beforeEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "../../../vitest-setup";
import {
  getGraphData,
  getFullGraphData,
  getGraphDelta,
  getCachedGraph,
  getFreshGraph,
  normalizeNode,
  normalizeLink,
  type GraphData,
  type GraphNode,
  type GraphLink,
} from "./graph";

describe("graph API", () => {
  beforeEach(() => {
    // Treat tests as authenticated so getFullGraphData targets /v1/graph/full
    (window as { __SKIP_AUTH__?: boolean }).__SKIP_AUTH__ = true;
    // Reset any default handlers
    server.resetHandlers();
  });

  describe("getGraphData", () => {
    it("should return graph data for note", async () => {
      const mockGraphData: GraphData = {
        nodes: [
          {
            id: "1",
            title: "Center Node",
            type: "star",
            x: 0,
            y: 0,
            z: 0,
            size: 10,
          },
          {
            id: "2",
            title: "Related Node 1",
            type: "planet",
            x: 10,
            y: 10,
            z: 0,
            size: 5,
          },
          {
            id: "3",
            title: "Related Node 2",
            type: "moon",
            x: -10,
            y: 10,
            z: 0,
            size: 3,
          },
        ],
        links: [
          { source: "1", target: "2", weight: 0.8, link_type: "reference" },
          { source: "1", target: "3", weight: 0.6, link_type: "related" },
        ],
      };

      server.use(
        http.get("http://localhost:9091/api/v1/graph/note/1", () => {
          return HttpResponse.json({ data: mockGraphData });
        })
      );

      const result = await getGraphData("1", 2);

      expect(result.nodes).toHaveLength(3);
      expect(result.links).toHaveLength(2);
      expect(result.nodes[0].title).toBe("Center Node");
    });
  });

  describe("getFullGraphData", () => {
    it("should return full graph data", async () => {
      const mockGraphData: GraphData = {
        nodes: [
          { id: "1", title: "Node 1", type: "star" },
          { id: "2", title: "Node 2", type: "planet" },
        ],
        links: [{ source: "1", target: "2", weight: 1.0, link_type: "reference" }],
      };

      server.use(
        http.get("http://localhost:9091/api/v1/graph/full", () => {
          return HttpResponse.json({ data: mockGraphData });
        })
      );

      const result = await getFullGraphData(50);

      expect(result.nodes).toHaveLength(2);
    });

    it("should handle large graphs", async () => {
      const manyNodes: GraphNode[] = Array.from({ length: 100 }, (_, i) => ({
        id: String(i),
        title: `Node ${i}`,
        type: i % 3 === 0 ? "star" : "planet",
      }));

      const manyLinks: GraphLink[] = Array.from({ length: 99 }, (_, i) => ({
        source: String(i),
        target: String(i + 1),
        weight: 0.5,
        link_type: "reference",
      }));

      server.use(
        http.get("http://localhost:9091/api/v1/graph/full", () =>
          HttpResponse.json({ data: { nodes: manyNodes, links: manyLinks } })
        )
      );

      const result = await getFullGraphData(100);

      expect(result.nodes).toHaveLength(100);
      expect(result.links).toHaveLength(99);
    });
  });

  describe("edge cases", () => {
    it("should handle empty graph", async () => {
      server.use(
        http.get("http://localhost:9091/api/v1/graph/note/999", () =>
          HttpResponse.json({ nodes: [], links: [] })
        )
      );

      const result = await getGraphData("999", 1);

      expect(result.nodes).toHaveLength(0);
      expect(result.links).toHaveLength(0);
    });
  });

  describe("error handling", () => {
    it("should handle network errors for getGraphData", async () => {
      server.use(http.get("http://localhost:9091/api/v1/graph/note/1", () => HttpResponse.error()));

      await expect(getGraphData("1")).rejects.toThrow();
    });

    it("should handle HTTP 404 errors for getGraphData", async () => {
      server.use(
        http.get("http://localhost:9091/api/v1/graph/note/999", () =>
          HttpResponse.json({ error: "Not found" }, { status: 404 })
        )
      );

      await expect(getGraphData("999")).rejects.toThrow("Граф не найден");
    });

    it("should handle HTTP 500 errors for getGraphData", async () => {
      server.use(
        http.get("http://localhost:9091/api/v1/graph/note/1", () =>
          HttpResponse.json({ error: "Server error" }, { status: 500 })
        )
      );

      await expect(getGraphData("1")).rejects.toThrow();
    });

    it("should handle 400 for invalid depth parameter", async () => {
      server.use(
        http.get("http://localhost:9091/api/v1/graph/note/1", () =>
          HttpResponse.json({ error: "Invalid depth parameter" }, { status: 400 })
        )
      );

      await expect(getGraphData("1", -1)).rejects.toThrow();
    });

    it("should handle network errors for getFullGraphData", async () => {
      server.use(http.get("http://localhost:9091/api/v1/graph/full", () => HttpResponse.error()));

      await expect(getFullGraphData()).rejects.toThrow();
    });

    it("should handle HTTP 500 errors for getFullGraphData", async () => {
      server.use(
        http.get("http://localhost:9091/api/v1/graph/full", () =>
          HttpResponse.json({ error: "Server error" }, { status: 500 })
        )
      );

      await expect(getFullGraphData(1000)).rejects.toThrow();
    });

    it("should handle 503 when graph service is unavailable", async () => {
      server.use(
        http.get("http://localhost:9091/api/v1/graph/note/1", () =>
          HttpResponse.json({ error: "Service Unavailable" }, { status: 503 })
        )
      );

      await expect(getGraphData("1")).rejects.toThrow();
    });

    it("should handle timeout errors", async () => {
      server.use(
        http.get("http://localhost:9091/api/v1/graph/note/1", () =>
          HttpResponse.json({ error: "Request timeout" }, { status: 504 })
        )
      );

      await expect(getGraphData("1")).rejects.toThrow();
    });
  });

  describe("getFullGraphData without limit parameter", () => {
    it("should use default limit when called without parameter", async () => {
      const mockGraphData: GraphData = {
        nodes: [
          { id: "1", title: "Node 1", type: "star" },
          { id: "2", title: "Node 2", type: "planet" },
        ],
        links: [{ source: "1", target: "2", weight: 1.0, link_type: "reference" }],
      };

      server.use(
        http.get("http://localhost:9091/api/v1/graph/full", ({ request }) => {
          const url = new URL(request.url);
          const limit = url.searchParams.get("limit");
          // Default limit from config is expected
          expect(limit).toBeTruthy();
          return HttpResponse.json({ data: mockGraphData });
        })
      );

      const result = await getFullGraphData();

      expect(result.nodes).toHaveLength(2);
    });
  });
});

describe("normalizeNode and normalizeLink", () => {
  it("normalizes a node with missing fields", () => {
    const node = normalizeNode({ id: "1", title: "A" });
    expect(node).toEqual({
      id: "1",
      title: "A",
      type: "unknown",
    });
  });

  it("normalizes a node with all fields", () => {
    const node = normalizeNode({
      id: "1",
      title: "A",
      type: "star",
      x: 1,
      y: 2,
      z: 3,
      size: 5,
    });
    expect(node).toEqual({
      id: "1",
      title: "A",
      type: "star",
      x: 1,
      y: 2,
      z: 3,
      size: 5,
    });
  });

  it("normalizes a link with defaults", () => {
    const link = normalizeLink({ source: "a", target: "b" });
    expect(link).toEqual({
      id: undefined,
      source: "a",
      target: "b",
      weight: 0.5,
      link_type: "related",
      source_type: "user",
      last_weight_update: undefined,
    });
  });
});

describe("getGraphData normalization", () => {
  it("accepts a raw GraphData body without a wrapper", async () => {
    const data: GraphData = {
      nodes: [{ id: "1", title: "A" }],
      links: [{ source: "1", target: "1" }],
    };
    server.use(
      http.get("http://localhost:9091/api/v1/graph/note/1", () => HttpResponse.json(data))
    );

    const result = await getGraphData("1");
    expect(result.nodes).toHaveLength(1);
  });

  it("returns an empty graph for unrecognized object responses", async () => {
    server.use(
      http.get("http://localhost:9091/api/v1/graph/note/1", () =>
        HttpResponse.json({ weird: true })
      )
    );

    const result = await getGraphData("1");
    expect(result.nodes).toHaveLength(0);
    expect(result.links).toHaveLength(0);
  });
});

describe("getGraphDelta", () => {
  it("returns the parsed delta with current hash", async () => {
    server.use(
      http.get("http://localhost:9091/api/v1/graph/delta", () =>
        HttpResponse.json({ added_nodes: [{ id: "1", title: "A" }], current_hash: "abc" })
      )
    );

    const result = await getGraphDelta("prev");
    expect(result.added_nodes).toHaveLength(1);
    expect(result.current_hash).toBe("abc");
  });

  it("returns an empty object for a non-object response", async () => {
    server.use(
      http.get("http://localhost:9091/api/v1/graph/delta", () => HttpResponse.json("bad"))
    );

    const result = await getGraphDelta("prev");
    expect(result).toEqual({});
  });
});

describe("getCachedGraph", () => {
  it("returns graph data from the cached endpoint", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/me/graph/cached", () =>
        HttpResponse.json({ data: { nodes: [{ id: "1", title: "A" }], links: [] } })
      )
    );

    const result = await getCachedGraph();
    expect(result?.nodes).toHaveLength(1);
  });

  it("returns null for a 204 response", async () => {
    server.use(
      http.get(
        "http://localhost:8080/api/v1/me/graph/cached",
        () => new HttpResponse(null, { status: 204 })
      )
    );

    const result = await getCachedGraph();
    expect(result).toBeNull();
  });

  it("returns null on a network error", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/me/graph/cached", () => HttpResponse.error())
    );

    const result = await getCachedGraph();
    expect(result).toBeNull();
  });
});

describe("getFreshGraph", () => {
  it("returns the fresh graph response", async () => {
    const payload = {
      data: {
        fresh: { nodes: [{ id: "1", title: "A" }], links: [] },
        delta: { added_nodes: [{ id: "2", title: "B" }] },
      },
    };
    server.use(
      http.get("http://localhost:8080/api/v1/me/graph/fresh", () => HttpResponse.json(payload))
    );

    const result = await getFreshGraph();
    expect(result.fresh.nodes).toHaveLength(1);
    expect(result.delta?.added_nodes).toHaveLength(1);
  });

  it("falls back to an empty graph on an empty response", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/me/graph/fresh", () => HttpResponse.json({}))
    );

    const result = await getFreshGraph();
    expect(result.fresh.nodes).toHaveLength(0);
    expect(result.fresh.links).toHaveLength(0);
  });

  it("throws a localized error for a 500 response", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/me/graph/fresh", () =>
        HttpResponse.json({ error: "boom" }, { status: 500 })
      )
    );

    await expect(getFreshGraph()).rejects.toThrow();
  });
});

describe("getFullGraphData public endpoint", () => {
  it("calls the public endpoint when unauthenticated", async () => {
    (window as { __SKIP_AUTH__?: boolean }).__SKIP_AUTH__ = false;

    server.use(
      http.get("http://localhost:9091/api/v1/graph/public", () =>
        HttpResponse.json({ data: { nodes: [{ id: "1", title: "A" }], links: [] } })
      )
    );

    const result = await getFullGraphData(10);
    expect(result.nodes).toHaveLength(1);
  });
});

describe("graph API additional branches", () => {
  beforeEach(() => {
    (window as { __SKIP_AUTH__?: boolean }).__SKIP_AUTH__ = true;
    server.resetHandlers();
  });

  it("includes nocache in full graph query", async () => {
    server.use(
      http.get("http://localhost:9091/api/v1/graph/full", ({ request }) => {
        const url = new URL(request.url);
        const nocache = url.searchParams.get("nocache");
        expect(nocache).toBe("1");
        return HttpResponse.json({ data: { nodes: [{ id: "1" }], links: [] } });
      })
    );

    const result = await getFullGraphData(10, undefined, true);
    expect(result.nodes).toHaveLength(1);
  });

  it("falls back to backend on 429 for getGraphData", async () => {
    server.use(
      http.get("http://localhost:9091/api/v1/graph/note/1", () =>
        HttpResponse.json({ error: "Too Many Requests" }, { status: 429 })
      )
    );

    await expect(getGraphData("1")).rejects.toThrow();
  });

  it("falls back to backend on 408 for getFullGraphData", async () => {
    server.use(
      http.get("http://localhost:9091/api/v1/graph/full", () =>
        HttpResponse.json({ error: "Timeout" }, { status: 408 })
      )
    );

    await expect(getFullGraphData()).rejects.toThrow();
  });

  it("returns empty delta for non-object response", async () => {
    server.use(
      http.get("http://localhost:9091/api/v1/graph/delta", () => HttpResponse.json("not an object"))
    );

    const result = await getGraphDelta("prev");
    expect(result).toEqual({});
  });

  it("returns delta with current_hash", async () => {
    server.use(
      http.get("http://localhost:9091/api/v1/graph/delta", () =>
        HttpResponse.json({ added_nodes: [{ id: "1" }], current_hash: "new-hash" })
      )
    );

    const result = await getGraphDelta("prev");
    expect(result.current_hash).toBe("new-hash");
  });

  it("handles delta load errors", async () => {
    server.use(
      http.get("http://localhost:9091/api/v1/graph/delta", () =>
        HttpResponse.json({ error: "Server error" }, { status: 500 })
      )
    );

    await expect(getGraphDelta("prev")).rejects.toThrow();
  });

  it("returns null cached graph when body has no data", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/me/graph/cached", () =>
        HttpResponse.json({ meta: {} })
      )
    );

    const result = await getCachedGraph();
    expect(result).toBeNull();
  });

  it("returns empty fresh graph when response data is missing", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/me/graph/fresh", () => HttpResponse.json({}))
    );

    const result = await getFreshGraph();
    expect(result.fresh.nodes).toHaveLength(0);
    expect(result.fresh.links).toHaveLength(0);
  });
});
