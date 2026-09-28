/**
 * GRAPH-LIGHT-1: camera moves for the 2D graph — a smooth fly-to a note and
 * back. The camera is kept as a world-space centre plus zoom, so a flight
 * stays correct when the canvas is resized mid-way (a side panel opening).
 */
export interface Camera {
  cx: number;
  cy: number;
  k: number;
}

export interface ScreenTransform {
  x: number;
  y: number;
  k: number;
}

export function cameraFromTransform(t: ScreenTransform, width: number, height: number): Camera {
  return { cx: (width / 2 - t.x) / t.k, cy: (height / 2 - t.y) / t.k, k: t.k };
}

export function applyCamera(t: ScreenTransform, cam: Camera, width: number, height: number): void {
  t.k = cam.k;
  t.x = width / 2 - cam.cx * cam.k;
  t.y = height / 2 - cam.cy * cam.k;
}

export function easeInOutCubic(u: number): number {
  return u < 0.5 ? 4 * u * u * u : 1 - Math.pow(-2 * u + 2, 3) / 2;
}

/** Position along the flight at progress u in [0, 1]; zoom moves in log space. */
export function interpolateCamera(from: Camera, to: Camera, u: number): Camera {
  const e = easeInOutCubic(Math.min(1, Math.max(0, u)));
  return {
    cx: from.cx + (to.cx - from.cx) * e,
    cy: from.cy + (to.cy - from.cy) * e,
    k: Math.exp(Math.log(from.k) + (Math.log(to.k) - Math.log(from.k)) * e),
  };
}

export function sameTransform(a: ScreenTransform, b: ScreenTransform, epsilon = 0.5): boolean {
  return (
    Math.abs(a.x - b.x) <= epsilon &&
    Math.abs(a.y - b.y) <= epsilon &&
    Math.abs(a.k - b.k) <= epsilon / 1000
  );
}

export interface CameraFlight {
  /** Fly to a camera; remembers where the first flight started. */
  flyTo(to: Camera, duration?: number): void;
  /**
   * Put the camera on `to` at once — for when something else carries the
   * motion (the list morph). Remembers the start like `flyTo`, so `flyBack`
   * still returns there.
   */
  jumpTo(to: Camera): void;
  /**
   * Fly back to where the flights started — only while still flying or when
   * the user has not moved the camera since landing. Forgets the start either way.
   */
  flyBack(duration?: number): void;
  /** Advance the flight to `now`; returns true while the camera is moving. */
  step(now: number): boolean;
  isFlying(): boolean;
}

/**
 * Camera flights over a mutable screen transform. A user zoom or pan in the
 * middle of a flight wins: the flight stops and leaves the camera to the user.
 */
export function createCameraFlight(
  transform: ScreenTransform,
  size: () => { width: number; height: number },
  now: () => number,
  reducedMotion: () => boolean
): CameraFlight {
  let tween: { from: Camera; to: Camera; t0: number; dur: number } | null = null;
  let lastApplied: ScreenTransform | null = null;
  let origin: Camera | null = null;
  let landed: ScreenTransform | null = null;

  const snapshot = (): ScreenTransform => ({ x: transform.x, y: transform.y, k: transform.k });

  function start(to: Camera, dur: number) {
    landed = null;
    const { width, height } = size();
    if (reducedMotion() || dur <= 0) {
      applyCamera(transform, to, width, height);
      tween = null;
      landed = snapshot();
      return;
    }
    tween = { from: cameraFromTransform(transform, width, height), to, t0: now(), dur };
    lastApplied = snapshot();
  }

  function rememberOrigin() {
    if (origin) return;
    const { width, height } = size();
    origin = cameraFromTransform(transform, width, height);
  }

  return {
    flyTo(to, duration = 650) {
      rememberOrigin();
      start(to, duration);
    },
    jumpTo(to) {
      rememberOrigin();
      start(to, 0);
    },
    flyBack(duration = 650) {
      const back = origin;
      const moving =
        tween !== null && lastApplied !== null && sameTransform(transform, lastApplied);
      const untouched = landed !== null && sameTransform(transform, landed);
      origin = null;
      landed = null;
      tween = null;
      if (back && (moving || untouched)) start(back, duration);
    },
    step(time) {
      if (!tween) return false;
      if (lastApplied && !sameTransform(transform, lastApplied)) {
        tween = null;
        return false;
      }
      const { width, height } = size();
      const u = (time - tween.t0) / tween.dur;
      applyCamera(transform, interpolateCamera(tween.from, tween.to, u), width, height);
      lastApplied = snapshot();
      if (u >= 1) {
        tween = null;
        landed = snapshot();
      }
      return true;
    },
    isFlying() {
      return tween !== null;
    },
  };
}
