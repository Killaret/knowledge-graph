import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { transformRawGraph, buildNotesGraph, ensureNotesInGraph, loadGraph } from "./graphLoader";
import { authState } from "$shared/stores/auth-session.svelte";
import { graphView } from "$shared/stores/graph-view.svelte";
import type { Note } from "$shared/api/notes";

const privateNote: Note = {
  id: "private",
  title: "Private",
  content: "private text",
  metadata: {},
  created_at: "2026-09-14",
  updated_at: "2026-09-14",
  type: "star",
};

const mockGraph = {
  nodes: [{ id: "public", title: "Public", type: "star" }],
  links: [],
  hash: "public-hash",
};

const mockNotesApi = vi.hoisted(() => ({
  getNotes: vi.fn(),
  getNote: vi.fn(),
}));
vi.mock("$shared/api/notes", () => mockNotesApi);

const mockGraphApi = vi.hoisted(() => ({
  getFullGraphData: vi.fn(),
  getGraphData: vi.fn(),
  normalizeNode: vi.fn((n) => n),
  normalizeLink: vi.fn((l) => l),
}));
vi.mock("$shared/api/graph", () => mockGraphApi);

describe("graphLoader", () => {
  describe("transformRawGraph", () => {
    it("normalizes mixed-case node fields and defaults type to star", () => {
      const raw = {
        nodes: [
          { Id: "n1", Title: "Note 1", Type: "planet", created_at: "2024-01-01T00:00:00Z" },
          { ID: "n2", title: "Note 2", CreatedAt: "2024-02-01T00:00:00Z" },
          { id: "n3", Title: "Note 3", createdAt: "2024-03-01T00:00:00Z" },
        ],
        links: [],
      };

      const result = transformRawGraph(raw as unknown as Parameters<typeof transformRawGraph>[0]);

      expect(result.nodes).toHaveLength(3);
      expect(result.nodes[0]).toMatchObject({
        id: "n1",
        title: "Note 1",
        type: "planet",
        createdAt: "2024-01-01T00:00:00Z",
      });
      expect(result.nodes[1]).toMatchObject({
        id: "n2",
        title: "Note 2",
        type: "star",
        createdAt: "2024-02-01T00:00:00Z",
      });
      expect(result.nodes[2]).toMatchObject({
        id: "n3",
        title: "Note 3",
        type: "star",
        createdAt: "2024-03-01T00:00:00Z",
      });
    });

    it("normalizes link source/target fields", () => {
      const raw = {
        nodes: [{ id: "n1" }, { id: "n2" }],
        links: [
          { source_note_id: "n1", target_note_id: "n2", weight: 0.8, link_type: "related" },
          { source: "n2", target: "n1" },
        ],
      };

      const result = transformRawGraph(raw as unknown as Parameters<typeof transformRawGraph>[0]);

      expect(result.links).toHaveLength(2);
      expect(result.links[0]).toMatchObject({
        source: "n1",
        target: "n2",
        weight: 0.8,
        link_type: "related",
        source_type: "user",
      });
      expect(result.links[1]).toMatchObject({
        source: "n2",
        target: "n1",
        weight: 0.5,
        link_type: "related",
        source_type: "user",
      });
    });

    it("injects the knowledge core as a technical node when missing", () => {
      const raw = {
        nodes: [{ id: "n1", title: "Note 1" }],
        links: [],
      };
      const knowledgeCore: Note = {
        id: "00000000-0000-0000-0000-000000000001",
        title: "Knowledge Core",
        content: "Help",
        metadata: {},
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      const result = transformRawGraph(
        raw as unknown as Parameters<typeof transformRawGraph>[0],
        knowledgeCore
      );

      expect(result.nodes).toHaveLength(2);
      expect(result.nodes[1]).toMatchObject({
        id: knowledgeCore.id,
        title: knowledgeCore.title,
        type: "technical",
        createdAt: knowledgeCore.created_at,
      });
    });

    it("preserves the graph hash", () => {
      const raw = { nodes: [], links: [], hash: "abc123" };
      const result = transformRawGraph(raw as unknown as Parameters<typeof transformRawGraph>[0]);
      expect(result.hash).toBe("abc123");
    });
  });

  describe("buildNotesGraph", () => {
    it("creates a node for every note", () => {
      const notes: Note[] = [
        {
          id: "n1",
          title: "Note 1",
          content: "",
          metadata: {},
          type: "star",
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        },
        {
          id: "n2",
          title: "Note 2",
          content: "",
          metadata: {},
          created_at: "2024-02-01T00:00:00Z",
          updated_at: "2024-02-01T00:00:00Z",
        },
      ];

      const result = buildNotesGraph(notes);

      expect(result.nodes).toHaveLength(2);
      expect(result.nodes[0]).toMatchObject({ id: "n1", title: "Note 1", type: "star" });
      expect(result.nodes[1]).toMatchObject({ id: "n2", title: "Note 2", type: "unknown" });
      expect(result.links).toHaveLength(0);
    });
  });

  describe("ensureNotesInGraph", () => {
    it("adds missing note nodes to the graph", () => {
      const graph = {
        nodes: [{ id: "n1", title: "Note 1", type: "star" }],
        links: [],
      };
      const notes: Note[] = [
        {
          id: "n1",
          title: "Note 1",
          content: "",
          metadata: {},
          type: "star",
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        },
        {
          id: "n2",
          title: "Note 2",
          content: "",
          metadata: {},
          type: "planet",
          created_at: "2024-02-01T00:00:00Z",
          updated_at: "2024-02-01T00:00:00Z",
        },
      ];

      const result = ensureNotesInGraph(graph as typeof graph, notes);

      expect(result.nodes).toHaveLength(2);
      expect(result.nodes[1]).toMatchObject({ id: "n2", title: "Note 2", type: "planet" });
    });

    it("returns the same graph when all notes are already present", () => {
      const graph = {
        nodes: [{ id: "n1", title: "Note 1", type: "star" }],
        links: [],
      };
      const notes: Note[] = [
        {
          id: "n1",
          title: "Note 1",
          content: "",
          metadata: {},
          type: "star",
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        },
      ];

      const result = ensureNotesInGraph(graph as typeof graph, notes);

      expect(result).toBe(graph);
    });
  });

  describe("PUB-2 loadGraph", () => {
    beforeEach(() => {
      authState.accessToken = "pub2-test-token";
      authState.currentUser = null;
      graphView.mode = "community";
      mockGraphApi.getFullGraphData.mockReset();
      mockGraphApi.getGraphData.mockReset();
      mockNotesApi.getNotes.mockReset();
      mockNotesApi.getNote.mockReset();
    });

    afterEach(() => {
      authState.accessToken = null;
      graphView.clear();
    });

    it.each([true, false])("PUB-2 community never mixes private notes (full=%s)", async (full) => {
      mockGraphApi.getFullGraphData.mockResolvedValue(mockGraph);
      const result = await loadGraph(
        {
          full,
          includeKnowledgeCore: true,
          fallbackToNotes: true,
          ensureNotesInGraph: true,
        },
        [privateNote]
      );
      expect(result.graph.nodes.map((n) => n.id)).toEqual(["public"]);
      expect(result.notes.map((n) => n.id)).toEqual(["public"]);
      expect(result.knowledgeCore).toBeNull();
      expect(mockNotesApi.getNotes).not.toHaveBeenCalled();
      expect(mockNotesApi.getNote).not.toHaveBeenCalled();
      expect(mockGraphApi.getGraphData).not.toHaveBeenCalled();
      expect(mockGraphApi.getFullGraphData).toHaveBeenCalledWith(
        0,
        undefined,
        undefined,
        "community"
      );
    });

    it("PUB-2 empty community graph does not retain private notes", async () => {
      mockGraphApi.getFullGraphData.mockResolvedValue({ nodes: [], links: [], hash: "empty" });
      const result = await loadGraph(
        {
          full: true,
          includeKnowledgeCore: true,
          fallbackToNotes: true,
          ensureNotesInGraph: true,
        },
        [privateNote]
      );
      expect(result.graph.nodes).toHaveLength(0);
      expect(result.notes).toHaveLength(0);
      expect(result.knowledgeCore).toBeNull();
      expect(mockNotesApi.getNotes).not.toHaveBeenCalled();
      expect(mockNotesApi.getNote).not.toHaveBeenCalled();
    });

    it("PUB-2 personal mode keeps private notes and optional knowledge core", async () => {
      graphView.mode = "personal";
      mockNotesApi.getNotes.mockResolvedValue([privateNote]);
      mockGraphApi.getFullGraphData.mockResolvedValue({ nodes: [], links: [], hash: "personal" });
      const result = await loadGraph(
        {
          full: true,
          includeKnowledgeCore: false,
          fallbackToNotes: true,
          ensureNotesInGraph: true,
        },
        [privateNote]
      );
      expect(result.graph.nodes.map((n) => n.id)).toEqual(["private"]);
      expect(result.notes.map((n) => n.id)).toEqual(["private"]);
      expect(result.knowledgeCore).toBeNull();
    });
  });
});
