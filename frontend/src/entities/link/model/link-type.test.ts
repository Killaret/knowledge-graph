import { describe, it, expect } from "vitest";
import { LinkType } from "./link-type";

describe("LinkType", () => {
  it("returns the correct static instances by type", () => {
    expect(LinkType.REFERENCE.type).toBe("reference");
    expect(LinkType.DEPENDENCY.type).toBe("dependency");
    expect(LinkType.RELATED.type).toBe("related");
    expect(LinkType.CUSTOM.type).toBe("custom");
    expect(LinkType.PARENT.type).toBe("parent");
    expect(LinkType.CHILD.type).toBe("child");
  });

  it("looks up a link type by string case-insensitively", () => {
    expect(LinkType.fromString("REFERENCE")).toBe(LinkType.REFERENCE);
    expect(LinkType.fromString(" Dependency ")).toBe(LinkType.DEPENDENCY);
    expect(LinkType.fromString(undefined)).toBe(LinkType.RELATED);
    expect(LinkType.fromString("unknown")).toBe(LinkType.RELATED);
  });

  it("exposes UI and creatable link types", () => {
    const uiTypes = LinkType.UI_TYPES.map((t) => t.type);
    expect(uiTypes).toContain("reference");
    expect(uiTypes).toContain("dependency");
    expect(uiTypes).toContain("related");
    expect(uiTypes).toContain("parent");
    expect(uiTypes).toContain("child");
    expect(uiTypes).not.toContain("custom");

    const creatableTypes = LinkType.CREATABLE_TYPES.map((t) => t.type);
    expect(creatableTypes).toContain("custom");
    expect(creatableTypes).toContain("reference");
    expect(creatableTypes).toContain("child");
  });

  it("exposes icon and description for each link type", () => {
    expect(LinkType.REFERENCE.icon).toBe("📖");
    expect(LinkType.REFERENCE.description.length).toBeGreaterThan(0);
    expect(LinkType.DEPENDENCY.icon).toBe("🔗");
    expect(LinkType.RELATED.example.length).toBeGreaterThan(0);
    expect(LinkType.CUSTOM.creatable).toBe(true);
  });

  it("computes color with weight and fade opacity", () => {
    const color = LinkType.REFERENCE.getColor(1, 1);
    expect(color).toMatch(/^rgba\(51, 102, 255, [\d.]+\)$/);
  });

  it("returns solid line for reference and dashed for dependency", () => {
    expect(LinkType.REFERENCE.getLineDash()).toEqual([]);
    expect(LinkType.DEPENDENCY.getLineDash()).toEqual([10, 3]);
  });

  it("switches related line dash based on weight", () => {
    expect(LinkType.RELATED.getLineDash(0.2)).toEqual([6, 4]);
    expect(LinkType.RELATED.getLineDash(0.5)).toEqual([]);
  });
});
