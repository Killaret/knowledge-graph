import { describe, it, expect } from "vitest";
import { FilterState } from "./filter-state";
import type { Note } from "$shared/api/notes";

function note(overrides: Partial<Note> = {}): Note {
  return {
    id: "1",
    title: "A",
    content: "",
    metadata: {},
    type: "star",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

const getNoteType = (n: Note) => n.type || "unknown";

describe("FilterState", () => {
  it("uses defaults", () => {
    const f = new FilterState();
    expect(f.selectedType).toBe("all");
    expect(f.sortBy).toBe("created");
    expect(f.currentView).toBe("graph");
    expect(f.isTypeActive).toBe(false);
    expect(f.isSearchActive).toBe(false);
  });

  it("filters by type", () => {
    const notes = [note({ id: "1", type: "star" }), note({ id: "2", type: "planet" })];
    const f = new FilterState({ selectedType: "star" });
    const filtered = f.filterNotes(notes, getNoteType);
    expect(filtered).toHaveLength(1);
    expect(filtered[0].id).toBe("1");
  });

  it("filters inbox by tag", () => {
    const notes = [
      note({ id: "1", metadata: { tags: ["#inbox"] } }),
      note({ id: "2", metadata: { tags: ["work"] } }),
    ];
    const f = new FilterState({ selectedType: "inbox" });
    const filtered = f.filterNotes(notes, getNoteType);
    expect(filtered).toHaveLength(1);
    expect(filtered[0].id).toBe("1");
  });

  it("searches title and content", () => {
    const notes = [
      note({ id: "1", title: "Cosmos" }),
      note({ id: "2", title: "Galaxy", content: "cosmic dust" }),
      note({ id: "3", title: "Other" }),
    ];
    const f = new FilterState({ searchQuery: "cos" });
    const filtered = f.filterNotes(notes, getNoteType);
    expect(filtered.map((n) => n.id)).toEqual(["1", "2"]);
  });

  it("sorts by type", () => {
    const notes = [note({ id: "1", type: "nebula" }), note({ id: "2", type: "asteroid" })];
    const f = new FilterState({ sortBy: "type" });
    const sorted = f.sortNotes(notes);
    expect(sorted[0].type).toBe("asteroid");
    expect(sorted[1].type).toBe("nebula");
  });

  it("filters graph data", () => {
    const graphData = {
      nodes: [
        { id: "1", title: "A" },
        { id: "2", title: "B" },
      ],
      links: [{ source: "1", target: "2" }],
    };
    const notes = [note({ id: "1", type: "star" }), note({ id: "2", type: "planet" })];
    const f = new FilterState({ selectedType: "star" });
    const result = f.filterGraphData(graphData, notes, getNoteType);
    expect(result.nodes).toHaveLength(1);
    expect(result.links).toHaveLength(0);
  });

  it("creates an updated copy", () => {
    const f = new FilterState({ selectedType: "all", searchQuery: "foo" });
    const next = f.with({ selectedType: "star" });
    expect(f.selectedType).toBe("all");
    expect(next.selectedType).toBe("star");
    expect(next.searchQuery.value).toBe("foo");
  });
});
