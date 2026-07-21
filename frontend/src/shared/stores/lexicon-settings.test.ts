import { describe, it, expect, beforeEach } from "vitest";
import { mode, setLocale, setMode, getMessage } from "./lexicon-settings";

describe("lexicon-settings", () => {
  beforeEach(() => {
    mode.set("standard");
    localStorage.clear();
  });

  it("mode subscribe notifies immediately and on changes", () => {
    const values: string[] = [];
    const unsubscribe = mode.subscribe((m) => values.push(m));

    expect(values).toEqual(["standard"]);

    mode.set("galactic");
    expect(values).toEqual(["standard", "galactic"]);

    unsubscribe();
    mode.set("standard");
    expect(values).toEqual(["standard", "galactic"]);
  });

  it("mode update derives new value from current", () => {
    mode.set("standard");
    mode.update((current) =>
      current === "standard" ? "galactic" : "standard",
    );

    const values: string[] = [];
    mode.subscribe((m) => values.push(m))();
    expect(values).toEqual(["galactic"]);
  });

  it("setMode updates the mode", () => {
    setMode("galactic");

    const values: string[] = [];
    mode.subscribe((m) => values.push(m))();
    expect(values).toEqual(["galactic"]);
  });

  it("setLocale persists the locale", () => {
    setLocale("ru");
    expect(localStorage.getItem("locale")).toBe("ru");
  });

  it("getMessage returns a message for the current locale and mode", async () => {
    const message = await getMessage("success", "noteCreated", "Test");
    expect(message).toContain("Test");
    expect(message).toContain("created");
  });

  it("getMessage uses galactic mode when set", async () => {
    setMode("galactic");
    const message = await getMessage("success", "noteCreated", "Test");
    expect(message).toContain("Star");
  });
});
