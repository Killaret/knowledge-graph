import { describe, it, expect } from "vitest";
import {
  localInputToIso,
  isoToLocalInput,
  remindChoiceToSeconds,
  secondsToRemindChoice,
} from "./comet";

describe("comet utils", () => {
  it("converts a local datetime input value to an ISO instant", () => {
    const iso = localInputToIso("2026-03-01T12:30");
    expect(iso).toMatch(/^2026-03-01T\d{2}:30:00\.000Z$/);
  });

  it("round-trips local input through ISO and back", () => {
    const local = "2026-06-15T08:45";
    expect(isoToLocalInput(localInputToIso(local))).toBe(local);
  });

  it("returns null/empty for empty input", () => {
    expect(localInputToIso("")).toBeNull();
    expect(localInputToIso("not-a-date")).toBeNull();
    expect(isoToLocalInput(null)).toBe("");
    expect(isoToLocalInput(undefined)).toBe("");
    expect(isoToLocalInput("garbage")).toBe("");
  });

  it("maps reminder presets to seconds", () => {
    expect(remindChoiceToSeconds("none", 60)).toBeNull();
    expect(remindChoiceToSeconds("15m", 60)).toBe(900);
    expect(remindChoiceToSeconds("1h", 60)).toBe(3600);
    expect(remindChoiceToSeconds("1d", 60)).toBe(86400);
  });

  it("maps a custom reminder in minutes to seconds", () => {
    expect(remindChoiceToSeconds("custom", 45)).toBe(2700);
    expect(remindChoiceToSeconds("custom", 0)).toBeNull();
    expect(remindChoiceToSeconds("custom", -5)).toBeNull();
    expect(remindChoiceToSeconds("custom", 1.9)).toBe(60);
  });

  it("round-trips reminder seconds back to a choice", () => {
    expect(secondsToRemindChoice(null)).toEqual({ choice: "none", customMinutes: 60 });
    expect(secondsToRemindChoice(0)).toEqual({ choice: "none", customMinutes: 60 });
    expect(secondsToRemindChoice(3600)).toEqual({ choice: "1h", customMinutes: 60 });
    expect(secondsToRemindChoice(2700)).toEqual({ choice: "custom", customMinutes: 45 });
  });
});
