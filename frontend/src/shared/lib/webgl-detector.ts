/**
 * Detects whether WebGL2 (or fallback WebGL) is available in the current browser.
 *
 * Keep this in `shared` so any layer can decide whether to render a 3D canvas
 * without importing a specific feature or renderer.
 */
export function isWebGLAvailable(): boolean {
  if (typeof document === "undefined") return false;

  const canvas = document.createElement("canvas");
  const gl =
    canvas.getContext("webgl2") ||
    canvas.getContext("webgl") ||
    canvas.getContext("experimental-webgl");

  if (gl && typeof WebGL2RenderingContext !== "undefined" && gl instanceof WebGL2RenderingContext) {
    return true;
  }
  if (gl && typeof WebGLRenderingContext !== "undefined" && gl instanceof WebGLRenderingContext) {
    return true;
  }
  return false;
}
