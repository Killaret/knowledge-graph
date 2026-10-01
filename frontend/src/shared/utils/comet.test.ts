import { describe, it, expect } from "vitest";
import {
  localInputToIso,
  isoToLocalInput,
  remindChoiceToSeconds,
  secondsToRemindChoice,
  cometUrgency,
  isCometSettled,
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

  // COMET-1 stage E: метафора «приближается».
  it("urgency is null for undated comets and 0 for settled ones", () => {
    const now = Date.parse("2026-10-01T12:00:00Z");
    expect(cometUrgency(null, null, now)).toBeNull();
    expect(cometUrgency(undefined, null, now)).toBeNull();
    expect(cometUrgency("2026-10-01T12:00:00Z", "2026-09-30T00:00:00Z", now)).toBe(0);
    expect(cometUrgency("2026-09-30T00:00:00Z", null, now)).toBe(0); // прошла
    expect(cometUrgency("garbage", null, now)).toBeNull();
  });

  it("urgency grows as the due date approaches", () => {
    const now = Date.parse("2026-10-01T12:00:00Z");
    const far = cometUrgency("2026-10-08T12:00:00Z", null, now); // ровно неделя
    const mid = cometUrgency("2026-10-04T12:00:00Z", null, now); // 3 дня
    const near = cometUrgency("2026-10-01T13:00:00Z", null, now); // час
    expect(far).toBe(0);
    expect(mid).toBeGreaterThan(far!);
    expect(near).toBeGreaterThan(mid!);
    expect(near).toBeLessThanOrEqual(1);
    // дальше недели — тоже 0
    expect(cometUrgency("2027-01-01T00:00:00Z", null, now)).toBe(0);
  });

  it("isCometSettled: done or past-due comets are settled", () => {
    expect(isCometSettled("2030-01-01T00:00:00Z", null)).toBe(false);
    expect(isCometSettled("2030-01-01T00:00:00Z", "2026-01-01T00:00:00Z")).toBe(true);
    expect(isCometSettled("2000-01-01T00:00:00Z", null)).toBe(true);
    expect(isCometSettled(null, null)).toBe(false);
  });
});
