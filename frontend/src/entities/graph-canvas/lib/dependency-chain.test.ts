import { describe, it, expect } from "vitest";
import {
  computeDependencyChain,
  computeDependencyCycleNodes,
  dependencyLinkKey,
  chainDepthOpacity,
  DEPENDENCY_HIGHLIGHT_DEPTH_DEFAULT,
  type DependencyLinkLike,
} from "./dependency-chain";

function dep(source: string, target: string): DependencyLinkLike {
  return { source, target, link_type: "dependency" };
}

function rel(source: string, target: string): DependencyLinkLike {
  return { source, target, link_type: "related" };
}

describe("computeDependencyChain", () => {
  it("returns null when the hovered node has no dependency links", () => {
    const links = [dep("a", "b"), rel("x", "y")];
    expect(computeDependencyChain("x", links)).toBeNull();
    expect(computeDependencyChain(null, links)).toBeNull();
  });

  it("ignores non-dependency links", () => {
    const links = [rel("a", "b")];
    expect(computeDependencyChain("a", links)).toBeNull();
  });

  it("walks both directions from the hovered node", () => {
    // a → b → c → d ; hover c
    const links = [dep("a", "b"), dep("b", "c"), dep("c", "d")];
    const chain = computeDependencyChain("c", links);
    expect(chain).not.toBeNull();
    expect(chain!.nodeDepth.get("c")).toBe(0);
    expect(chain!.nodeDepth.get("b")).toBe(1);
    expect(chain!.nodeDepth.get("a")).toBe(2);
    expect(chain!.nodeDepth.get("d")).toBe(1);
    expect(chain!.linkDepth.get(dependencyLinkKey("a", "b"))).toBe(2);
    expect(chain!.linkDepth.get(dependencyLinkKey("b", "c"))).toBe(1);
    expect(chain!.linkDepth.get(dependencyLinkKey("c", "d"))).toBe(1);
    expect(chain!.hasCycle).toBe(false);
  });

  it("covers branching chains", () => {
    // a → b ; a → e ; b → c ; e → c (diamond merge), hover b
    const links = [dep("a", "b"), dep("a", "e"), dep("b", "c"), dep("e", "c")];
    const chain = computeDependencyChain("b", links)!;
    expect(chain.nodeDepth.get("a")).toBe(1);
    expect(chain.nodeDepth.get("e")).toBe(2);
    expect(chain.nodeDepth.get("c")).toBe(1);
    expect(chain.hasCycle).toBe(false);
  });

  it("detects a cycle and marks its links", () => {
    // a → b → c → a ; hover b
    const links = [dep("a", "b"), dep("b", "c"), dep("c", "a")];
    const chain = computeDependencyChain("b", links)!;
    expect(chain.hasCycle).toBe(true);
    expect(chain.cycleLinks.size).toBe(3);
    expect(chain.cycleLinks.has(dependencyLinkKey("c", "a"))).toBe(true);
  });

  it("marks only the cyclic part in a mixed chain", () => {
    // cycle a→b→a plus tail b→c; hover c
    const links = [dep("a", "b"), dep("b", "a"), dep("b", "c")];
    const chain = computeDependencyChain("c", links)!;
    expect(chain.hasCycle).toBe(true);
    expect(chain.cycleLinks.has(dependencyLinkKey("a", "b"))).toBe(true);
    expect(chain.cycleLinks.has(dependencyLinkKey("b", "a"))).toBe(true);
    expect(chain.cycleLinks.has(dependencyLinkKey("b", "c"))).toBe(false);
  });

  it("treats a self-loop as a cycle", () => {
    const links = [dep("a", "a"), dep("a", "b")];
    const chain = computeDependencyChain("b", links)!;
    expect(chain.hasCycle).toBe(true);
    expect(chain.cycleLinks.has(dependencyLinkKey("a", "a"))).toBe(true);
  });

  it("bounds traversal by maxDepth", () => {
    // a0 → a1 → … → a14, hover a0
    const links: DependencyLinkLike[] = [];
    for (let i = 0; i < 14; i++) links.push(dep(`a${i}`, `a${i + 1}`));

    const chain = computeDependencyChain("a0", links, DEPENDENCY_HIGHLIGHT_DEPTH_DEFAULT)!;
    expect(chain.nodeDepth.get("a10")).toBe(10);
    // Depth 11 and beyond are cut off — the mutation "no depth limit" must fail.
    expect(chain.nodeDepth.has("a11")).toBe(false);
    expect(chain.nodeDepth.has("a14")).toBe(false);
  });

  it("respects a custom maxDepth", () => {
    const links = [dep("a", "b"), dep("b", "c"), dep("c", "d")];
    const chain = computeDependencyChain("a", links, 1)!;
    expect(chain.nodeDepth.has("b")).toBe(true);
    expect(chain.nodeDepth.has("c")).toBe(false);
  });

  it("caps a cycle traversal at the depth bound", () => {
    // Cycle a→b→c→d→a, hover a, depth 1: b (successor) and d (predecessor)
    // are in, c is beyond the bound — the cycle edge c→d is not in the chain,
    // so no cycle is reported at this depth.
    const links = [dep("a", "b"), dep("b", "c"), dep("c", "d"), dep("d", "a")];
    const chain = computeDependencyChain("a", links, 1)!;
    expect(chain.nodeDepth.has("b")).toBe(true);
    expect(chain.nodeDepth.has("d")).toBe(true);
    expect(chain.nodeDepth.has("c")).toBe(false);
    // Edges a→b and d→a lie on the cycle even though the loop itself is not
    // fully inside the depth bound — they are flagged so they render red.
    expect(chain.hasCycle).toBe(true);
    expect(chain.cycleLinks.has(dependencyLinkKey("b", "c"))).toBe(false);
    expect(chain.cycleLinks.has(dependencyLinkKey("a", "b"))).toBe(true);

    const deep = computeDependencyChain("a", links, 2)!;
    expect(deep.hasCycle).toBe(true);
  });
});

describe("computeDependencyCycleNodes", () => {
  it("returns an empty set for acyclic graphs", () => {
    expect(computeDependencyCycleNodes([dep("a", "b"), dep("b", "c")]).size).toBe(0);
  });

  it("finds nodes on a cycle, including self-loops", () => {
    const cyclic = computeDependencyCycleNodes([
      dep("a", "b"),
      dep("b", "c"),
      dep("c", "a"),
      dep("d", "d"),
      dep("e", "f"),
    ]);
    expect(cyclic.has("a")).toBe(true);
    expect(cyclic.has("b")).toBe(true);
    expect(cyclic.has("c")).toBe(true);
    expect(cyclic.has("d")).toBe(true);
    expect(cyclic.has("e")).toBe(false);
  });
});

describe("chainDepthOpacity", () => {
  it("decays with distance but never drops below the floor", () => {
    expect(chainDepthOpacity(0)).toBe(1);
    expect(chainDepthOpacity(2)).toBeLessThan(chainDepthOpacity(1));
    expect(chainDepthOpacity(100)).toBe(0.35);
  });
});
