import { describe, it, expect } from "vitest";
import { filterValidLinks } from "./graphUtils";

const nodes = [{ id: "a" }, { id: "b" }, { id: "c" }];

describe("filterValidLinks", () => {
  it("keeps links with string endpoints that exist in nodes", () => {
    const links = [{ source: "a", target: "b" }];
    expect(filterValidLinks(nodes, links)).toEqual(links);
  });

  it("keeps links with numeric index endpoints", () => {
    const links = [{ source: 0, target: 2 }];
    expect(filterValidLinks(nodes, links)).toEqual(links);
  });

  it("keeps links with object endpoints", () => {
    const links = [{ source: { id: "a" }, target: { id: "c" } }];
    expect(filterValidLinks(nodes, links)).toEqual(links);
  });

  it("drops links with unknown endpoints", () => {
    const links = [
      { source: "a", target: "z" },
      { source: 5, target: 0 },
      { source: { id: "a" }, target: { id: "unknown" } },
    ];
    expect(filterValidLinks(nodes, links)).toEqual([]);
  });
});
