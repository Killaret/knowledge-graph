export type FogWarningKind = "danger" | "recovery";

export interface FogWarningState {
  kind: FogWarningKind | null;
  dangerArmed: boolean;
  recoveryArmed: boolean;
  shownAt: number;
  lastFalseAt: number;
  lastShowWarning: boolean;
  dangerShownInEpisode: boolean;
}

export const DEFAULT_FOG_WARNING_HOLD_MS = 2000;
export const DEFAULT_FOG_WARNING_REARM_MS = 1000;

export function createFogWarningState(now: number): FogWarningState {
  return {
    kind: null,
    dangerArmed: true,
    recoveryArmed: false,
    shownAt: 0,
    lastFalseAt: now,
    lastShowWarning: false,
    dangerShownInEpisode: false,
  };
}

export function updateFogWarning(
  state: FogWarningState,
  now: number,
  showWarning: boolean,
  isInteracting: boolean,
  holdMs = DEFAULT_FOG_WARNING_HOLD_MS,
  rearmMs = DEFAULT_FOG_WARNING_REARM_MS
): FogWarningState {
  let {
    kind,
    dangerArmed,
    recoveryArmed,
    shownAt,
    lastFalseAt,
    lastShowWarning,
    dangerShownInEpisode,
  } = state;

  // Track the last moment showWarning became false so we can debounce re-arming.
  if (!showWarning && lastShowWarning) {
    lastFalseAt = now;
  }
  lastShowWarning = showWarning;

  const holdExpired = kind !== null && now - shownAt >= holdMs;

  if (holdExpired) {
    if (kind === "recovery") {
      // Recovery hold ended. Reset the episode and debounce re-arming the danger
      // toast so it does not fire on brief FPS blips.
      dangerShownInEpisode = false;
      if (!showWarning && now - lastFalseAt >= rearmMs) {
        dangerArmed = true;
      }
    }
    // If a danger hold ended while the load is still high, the banner just hides
    // and the episode continues until showWarning goes false and the recovery toast
    // is armed below.
    kind = null;
    shownAt = 0;
  }

  // Arm recovery when the banner is free, a danger was shown in this episode, and
  // the load has been continuously false for the rearm debounce period. This prevents
  // the recovery toast from flashing on a momentary FPS spike.
  if (
    kind === null &&
    !recoveryArmed &&
    dangerShownInEpisode &&
    !showWarning &&
    now - lastFalseAt >= rearmMs
  ) {
    recoveryArmed = true;
  }

  // Trigger the danger toast only when the low-FPS condition is active, the user is
  // actively interacting with the canvas, and the trigger is armed.
  if (showWarning && isInteracting && dangerArmed) {
    kind = "danger";
    dangerArmed = false;
    recoveryArmed = false;
    shownAt = now;
    dangerShownInEpisode = true;
  }

  // Trigger the recovery toast when the load has dropped and the user is interacting.
  if (!showWarning && isInteracting && recoveryArmed && dangerShownInEpisode) {
    kind = "recovery";
    recoveryArmed = false;
    shownAt = now;
  }

  // Debounced re-arm of the danger toast when there is no active episode.
  if (
    kind === null &&
    !dangerArmed &&
    !dangerShownInEpisode &&
    !showWarning &&
    now - lastFalseAt >= rearmMs
  ) {
    dangerArmed = true;
  }

  return {
    kind,
    dangerArmed,
    recoveryArmed,
    shownAt,
    lastFalseAt,
    lastShowWarning,
    dangerShownInEpisode,
  };
}
