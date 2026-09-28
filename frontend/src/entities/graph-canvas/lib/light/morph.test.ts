/**
 * GRAPH-LIGHT-1: the graph turns into the list and back.
 */
import { describe, expect, it } from "vitest";
import type { SimulationNode } from "../types";
import {
  createMorph,
  morphFinished,
  morphNoteOpacity,
  morphNodes,
  morphThreadFade,
  morphTowardList,
  noteTowardList,
} from "./morph";

/** Cards in list order; the list shows the band 0..300 of the canvas. */
const rows = [
  { id: "top", x: 40, y: 20 },
  { id: "middle", x: 40, y: 60 },
  { id: "below", x: 40, y: 900 },
];
const visible = { top: 0, bottom: 300 };

const nodes: SimulationNode[] = [
  { id: "top", title: "Top", type: "star", x: 300, y: 300 },
  { id: "middle", title: "Middle", type: "planet", x: 200, y: 200 },
  { id: "below", title: "Below the fold", type: "moon", x: 100, y: 100 },
  { id: "filtered", title: "Filtered out", type: "moon", x: 50, y: 50 },
];

describe("graph to list morph", () => {
  it("starts rows in list order, so the list fills in like a wave", () => {
    const morph = createMorph(rows, visible, "to-list", 0);
    expect(morph.rank.get("top")).toBe(0);
    expect(morph.rank.get("below")).toBe(1);
    const halfway = 0.5;
    expect(noteTowardList(morph, "top", halfway)).toBeGreaterThan(
      noteTowardList(morph, "below", halfway)
    );
  });

  it("moves the visible cards first on the way back too, then the ones beyond the edge", () => {
    const scrolled = [
      { id: "above", x: 40, y: -500 },
      { id: "first", x: 40, y: 100 },
      { id: "second", x: 40, y: 200 },
      { id: "below", x: 40, y: 900 },
    ];
    const back = createMorph(scrolled, visible, "to-graph", 0);
    expect(back.rank.get("first")).toBe(0);
    expect(back.rank.get("above")).toBeGreaterThan(back.rank.get("second")!);
    const early = 0.7;
    expect(noteTowardList(back, "first", early)).toBeLessThan(noteTowardList(back, "above", early));
    expect(noteTowardList(back, "first", 1)).toBe(1);
    expect(noteTowardList(back, "below", 0)).toBe(0);
  });

  it("runs toward the list one way and back toward the graph the other", () => {
    const there = createMorph(rows, visible, "to-list", 1000, 1000);
    const back = createMorph(rows, visible, "to-graph", 1000, 1000);
    expect(morphTowardList(there, 1000)).toBe(0);
    expect(morphTowardList(there, 2000)).toBe(1);
    expect(morphTowardList(back, 1000)).toBe(1);
    expect(morphTowardList(back, 2000)).toBe(0);
    expect(morphFinished(there, 1999)).toBe(false);
    expect(morphFinished(there, 2000)).toBe(true);
  });

  it("puts notes on their cards at the end and back in the graph at the start", () => {
    const transform = { x: 10, y: 20, k: 2 };
    const morph = createMorph(rows, visible, "to-list", 0);
    const atList = morphNodes(nodes, morph, 1, transform);
    const top = atList.find((n) => n.id === "top")!;
    expect(top.x! * transform.k + transform.x).toBeCloseTo(40);
    expect(top.y! * transform.k + transform.y).toBeCloseTo(20);
    const atGraph = morphNodes(nodes, morph, 0, transform);
    expect(atGraph.find((n) => n.id === "top")).toMatchObject({ x: 300, y: 300 });
  });

  it("takes a note whose card is off screen to the edge of the list and fades it there", () => {
    const morph = createMorph(rows, visible, "to-list", 0);
    const below = morphNodes(nodes, morph, 1, { x: 0, y: 0, k: 1 }).find((n) => n.id === "below")!;
    expect(below.y).toBeGreaterThan(visible.bottom);
    expect(below.y).toBeLessThan(visible.bottom + 100);
    expect(morphNoteOpacity(morph, "below", 0)).toBe(1);
    expect(morphNoteOpacity(morph, "below", 1)).toBe(0);
    expect(morphNoteOpacity(morph, "top", 1)).toBe(1);

    const scrolled = createMorph([{ id: "top", x: 40, y: -500 }], visible, "to-graph", 0);
    expect(scrolled.targets.get("top")!.y).toBeLessThan(visible.top);
  });

  it("fades notes without a card where they are and does not touch the simulation", () => {
    const morph = createMorph(rows, visible, "to-list", 0);
    const moved = morphNodes(nodes, morph, 1, { x: 0, y: 0, k: 1 });
    expect(moved.find((n) => n.id === "filtered")).toBe(nodes[3]);
    expect(morphNoteOpacity(morph, "filtered", 0.25)).toBeCloseTo(0.75);
    expect(nodes[0]).toMatchObject({ x: 300, y: 300 });
  });

  it("fades a note's threads as it leaves the graph", () => {
    const morph = createMorph(rows, visible, "to-list", 0);
    const early = morphThreadFade(morph, nodes, 0.3);
    expect(early.get("top")).toBeLessThan(early.get("below")!);
    const landed = morphThreadFade(morph, nodes, 1);
    expect([...landed.values()].every((v) => v === 0)).toBe(true);
    const start = morphThreadFade(morph, nodes, 0);
    expect([...start.values()].every((v) => v === 1)).toBe(true);
  });
});
