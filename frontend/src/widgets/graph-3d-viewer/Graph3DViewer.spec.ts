import { describe, it, expect, vi, type MockedFunction } from "vitest";
import { render, screen, waitFor } from "@testing-library/svelte";
import { isWebGLAvailable } from "$shared/lib/webgl-detector";
import Graph3DViewer from "./Graph3DViewer.svelte";

vi.mock("$shared/lib/webgl-detector", () => ({
  isWebGLAvailable: vi.fn(),
}));

vi.mock(
  "$features/graph-3d/ui/Graph3DScene.svelte",
  () => import("$features/graph-3d/ui/__mocks__/Graph3DScene.svelte")
);

describe("Graph3DViewer", () => {
  it("shows an error overlay when WebGL is not available", async () => {
    (isWebGLAvailable as MockedFunction<typeof isWebGLAvailable>).mockReturnValue(false);

    render(Graph3DViewer, { props: { nodes: [], links: [] } });

    expect(screen.getByTestId("graph-3d-error")).toBeInTheDocument();
  });

  it("lazy-loads Graph3DScene and renders the 3D container when WebGL is available", async () => {
    (isWebGLAvailable as MockedFunction<typeof isWebGLAvailable>).mockReturnValue(true);

    render(Graph3DViewer, {
      props: {
        nodes: [{ id: "1", title: "A", type: "star" }],
        links: [],
        centerNodeId: "1",
        selectedNodeId: "1",
      },
    });

    await waitFor(() => {
      expect(screen.queryByTestId("graph-3d-loading")).not.toBeInTheDocument();
    });

    expect(screen.getByTestId("graph-3d-viewer")).toBeInTheDocument();
    expect(screen.getByTestId("graph-3d-scene")).toBeInTheDocument();
  });
});
