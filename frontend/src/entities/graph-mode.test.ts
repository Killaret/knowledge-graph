import { describe, it, expect } from "vitest";
import { GraphMode } from "./graph-mode";

describe("GraphMode", () => {
  it("defaults to normal", () => {
    const m = new GraphMode();
    expect(m.isNormal).toBe(true);
    expect(m.isFocus).toBe(false);
    expect(m.icon).toBe("⚡");
    expect(m.focusIcon).toBe("🔘");
  });

  it("provides focus styling", () => {
    const m = GraphMode.focus();
    expect(m.isFocus).toBe(true);
    expect(m.icon).toBe("👁");
    expect(m.focusIcon).toBe("🎯");
    expect(m.borderColor).toBe("rgba(139, 92, 246, 0.8)");
    expect(m.textColor).toBe("#a78bfa");
  });

  it("toggles mode", () => {
    const m = GraphMode.normal().toggle();
    expect(m.isFocus).toBe(true);
    expect(m.toggle().isNormal).toBe(true);
  });

  it("creates from boolean and string", () => {
    expect(GraphMode.fromFocus(true).isFocus).toBe(true);
    expect(GraphMode.fromFocus(false).mode).toBe("normal");
    expect(GraphMode.fromString("focus").isFocus).toBe(true);
    expect(GraphMode.fromString("normal").isNormal).toBe(true);
    expect(GraphMode.fromString("unknown").isNormal).toBe(true);
  });

  it("compares equality", () => {
    expect(GraphMode.focus().equals(GraphMode.focus())).toBe(true);
    expect(GraphMode.focus().equals(GraphMode.normal())).toBe(false);
  });
});
