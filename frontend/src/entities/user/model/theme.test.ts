import { describe, it, expect } from "vitest";
import { Theme } from "./theme";

describe("Theme", () => {
  it("defaults to standard", () => {
    const t = new Theme();
    expect(t.mode).toBe("standard");
    expect(t.isGalactic).toBe(false);
    expect(t.isStandard).toBe(true);
  });

  it("detects galactic mode", () => {
    const t = Theme.galactic();
    expect(t.isGalactic).toBe(true);
    expect(t.useGalacticMode).toBe(true);
  });

  it("chooses values by mode", () => {
    expect(Theme.standard().choose("Note", "Star")).toBe("Note");
    expect(Theme.galactic().choose("Note", "Star")).toBe("Star");
  });

  it("labels by mode", () => {
    expect(Theme.standard().label("Share Note", "Open Portal")).toBe("Share Note");
    expect(Theme.galactic().label("Share Note", "Open Portal")).toBe("Open Portal");
  });

  it("transforms known labels in galactic mode", () => {
    const galacticMap = { Confirm: "Engage", Cancel: "Abort" };
    expect(Theme.galactic().transformLabel("Confirm", galacticMap)).toBe("Engage");
    expect(Theme.standard().transformLabel("Confirm", galacticMap)).toBe("Confirm");
    expect(Theme.galactic().transformLabel("Unknown", galacticMap)).toBe("Unknown");
  });

  it("can be created from boolean", () => {
    expect(Theme.fromBoolean(true).isGalactic).toBe(true);
    expect(Theme.fromBoolean(false).mode).toBe("standard");
  });

  it("can be created from string", () => {
    expect(Theme.fromString("galactic").isGalactic).toBe(true);
    expect(Theme.fromString("standard").mode).toBe("standard");
    expect(Theme.fromString("unknown").mode).toBe("standard");
  });

  it("compares equality", () => {
    expect(Theme.galactic().equals(Theme.galactic())).toBe(true);
    expect(Theme.galactic().equals(Theme.standard())).toBe(false);
  });
});
