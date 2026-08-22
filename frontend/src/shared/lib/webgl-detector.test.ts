import { describe, it, expect, vi, beforeEach, beforeAll } from "vitest";
import { isWebGLAvailable } from "./webgl-detector";

describe("isWebGLAvailable", () => {
  beforeAll(() => {
    vi.stubGlobal("WebGLRenderingContext", class WebGLRenderingContext {});
    vi.stubGlobal("WebGL2RenderingContext", class WebGL2RenderingContext {});
  });

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("returns true when webgl2 context is available", () => {
    const { WebGL2RenderingContext: GL2 } = globalThis as unknown as {
      WebGL2RenderingContext: new () => WebGL2RenderingContext;
    };
    const mockContext = new GL2();
    const getContext = vi.fn().mockReturnValue(mockContext);
    vi.stubGlobal("document", {
      createElement: () => ({ getContext }),
    });

    expect(isWebGLAvailable()).toBe(true);
    expect(getContext).toHaveBeenCalledWith("webgl2");
  });

  it("falls back to webgl context when webgl2 is missing", () => {
    const { WebGLRenderingContext: GL1 } = globalThis as unknown as {
      WebGLRenderingContext: new () => WebGLRenderingContext;
    };
    const getContext = vi.fn().mockReturnValueOnce(null).mockReturnValueOnce(new GL1());
    vi.stubGlobal("document", {
      createElement: () => ({ getContext }),
    });

    expect(isWebGLAvailable()).toBe(true);
    expect(getContext).toHaveBeenCalledWith("webgl");
  });

  it("returns false when no WebGL context is available", () => {
    const getContext = vi.fn().mockReturnValue(null);
    vi.stubGlobal("document", {
      createElement: () => ({ getContext }),
    });

    expect(isWebGLAvailable()).toBe(false);
  });
});
