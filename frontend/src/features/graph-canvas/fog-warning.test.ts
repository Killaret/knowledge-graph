import { describe, it, expect } from "vitest";
import {
  createFogWarningState,
  updateFogWarning,
  DEFAULT_FOG_WARNING_HOLD_MS,
  DEFAULT_FOG_WARNING_REARM_MS,
} from "./fog-warning";

describe("fog-warning", () => {
  it("shows the danger warning when showWarning and isInteracting are both true", () => {
    let state = createFogWarningState(0);
    state = updateFogWarning(state, 100, true, true);
    expect(state.kind).toBe("danger");
    expect(state.dangerArmed).toBe(false);
  });

  it("does not show the danger warning when the user is not interacting", () => {
    let state = createFogWarningState(0);
    state = updateFogWarning(state, 100, true, false);
    expect(state.kind).toBe(null);
    expect(state.dangerArmed).toBe(true);
  });

  it("does not show the danger warning when showWarning is false", () => {
    let state = createFogWarningState(0);
    state = updateFogWarning(state, 100, false, true);
    expect(state.kind).toBe(null);
    expect(state.dangerArmed).toBe(true);
  });

  it("keeps the danger warning visible for the hold duration", () => {
    let state = createFogWarningState(0);
    state = updateFogWarning(state, 0, true, true);
    state = updateFogWarning(state, DEFAULT_FOG_WARNING_HOLD_MS - 1, true, true);
    expect(state.kind).toBe("danger");
  });

  it("hides the danger warning after the hold duration expires", () => {
    let state = createFogWarningState(0);
    state = updateFogWarning(state, 0, true, true);
    state = updateFogWarning(state, DEFAULT_FOG_WARNING_HOLD_MS, true, true);
    expect(state.kind).toBe(null);
  });

  it("does not re-trigger while the high-load condition persists", () => {
    let state = createFogWarningState(0);
    state = updateFogWarning(state, 0, true, true);
    const shownAt = state.shownAt;

    state = updateFogWarning(state, DEFAULT_FOG_WARNING_HOLD_MS + 1, true, true);
    expect(state.kind).toBe(null);
    expect(state.dangerArmed).toBe(false);

    state = updateFogWarning(
      state,
      DEFAULT_FOG_WARNING_HOLD_MS + 100,
      true,
      true
    );
    expect(state.kind).toBe(null);
    expect(state.dangerArmed).toBe(false);
    expect(state.shownAt).toBe(shownAt);
  });

  it("shows recovery toast after showWarning goes false for the rearm debounce", () => {
    let state = createFogWarningState(0);
    state = updateFogWarning(state, 0, true, true);
    state = updateFogWarning(state, DEFAULT_FOG_WARNING_HOLD_MS, true, true);
    expect(state.kind).toBe(null);

    // showWarning becomes false.
    const falseTime = DEFAULT_FOG_WARNING_HOLD_MS + 1;
    state = updateFogWarning(state, falseTime, false, true);
    expect(state.dangerShownInEpisode).toBe(true);
    expect(state.kind).toBe(null);

    // Wait for the rearm debounce; recovery arms and triggers.
    state = updateFogWarning(
      state,
      falseTime + DEFAULT_FOG_WARNING_REARM_MS,
      false,
      true
    );
    expect(state.kind).toBe("recovery");
    expect(state.dangerShownInEpisode).toBe(true);
  });

  it("hides the recovery toast after the hold duration and resets the episode", () => {
    let state = createFogWarningState(0);
    state = updateFogWarning(state, 0, true, true);
    const falseTime = DEFAULT_FOG_WARNING_HOLD_MS + 1;
    state = updateFogWarning(state, falseTime, false, true);
    state = updateFogWarning(
      state,
      falseTime + DEFAULT_FOG_WARNING_REARM_MS,
      false,
      true
    );
    expect(state.kind).toBe("recovery");

    state = updateFogWarning(
      state,
      falseTime + DEFAULT_FOG_WARNING_REARM_MS + DEFAULT_FOG_WARNING_HOLD_MS,
      false,
      true
    );
    expect(state.kind).toBe(null);
    expect(state.dangerShownInEpisode).toBe(false);
    expect(state.dangerArmed).toBe(true);
  });

  it("does not show recovery if showWarning only blips false", () => {
    let state = createFogWarningState(0);
    state = updateFogWarning(state, 0, true, true);
    state = updateFogWarning(state, DEFAULT_FOG_WARNING_HOLD_MS, true, true);

    // Brief false blip.
    const falseTime = DEFAULT_FOG_WARNING_HOLD_MS + 1;
    state = updateFogWarning(state, falseTime, false, true);

    // Back to true before the rearm debounce expires.
    state = updateFogWarning(
      state,
      falseTime + DEFAULT_FOG_WARNING_REARM_MS - 1,
      true,
      true
    );

    expect(state.kind).toBe(null);
    expect(state.recoveryArmed).toBe(false);
  });

  it("re-arms danger and can show a new warning after the episode completes", () => {
    let state = createFogWarningState(0);
    state = updateFogWarning(state, 0, true, true);
    const falseTime = DEFAULT_FOG_WARNING_HOLD_MS + 1;
    state = updateFogWarning(state, falseTime, false, true);
    state = updateFogWarning(
      state,
      falseTime + DEFAULT_FOG_WARNING_REARM_MS,
      false,
      true
    );
    state = updateFogWarning(
      state,
      falseTime + DEFAULT_FOG_WARNING_REARM_MS + DEFAULT_FOG_WARNING_HOLD_MS,
      false,
      true
    );
    expect(state.dangerArmed).toBe(true);

    state = updateFogWarning(
      state,
      falseTime +
        DEFAULT_FOG_WARNING_REARM_MS +
        DEFAULT_FOG_WARNING_HOLD_MS +
        1,
      true,
      true
    );
    expect(state.kind).toBe("danger");
  });
});
