import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, cleanup, waitFor } from "@testing-library/svelte";
import SmartCreateModal from "./SmartCreateModal.svelte";
import { createSmartImport, getImportStatus } from "$shared/api/import";

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

vi.mock("$shared/api/import", () => ({
  createSmartImport: vi.fn(),
  getImportStatus: vi.fn(),
}));

const doneStatus = {
  task_id: "task-1",
  status: "done" as const,
  progress: { total: 0, processed: 0, created: 0, skipped: 0, failed: 0 },
  notes: [
    { id: "n1", title: "Alpha" },
    { id: "n2", title: "Beta" },
  ],
};

describe("SmartCreateModal", () => {
  beforeEach(() => {
    cleanup();
    vi.clearAllMocks();
    vi.mocked(createSmartImport).mockResolvedValue({ task_id: "task-1" });
    vi.mocked(getImportStatus).mockResolvedValue({
      task_id: "task-1",
      status: "pending",
      progress: { total: 0, processed: 0, created: 0, skipped: 0, failed: 0 },
    });
  });

  it("shows three source tiles and the selected source field", () => {
    render(SmartCreateModal, { props: { open: true } });

    expect(screen.getByTestId("smart-create-tile-text")).toBeInTheDocument();
    expect(screen.getByTestId("smart-create-tile-url")).toBeInTheDocument();
    expect(screen.getByTestId("smart-create-tile-file")).toBeInTheDocument();
    // No title or type fields — decision 5.
    expect(screen.queryByTestId("smart-create-text")).toBeInTheDocument();
  });

  it("switches the source field with the tile", async () => {
    render(SmartCreateModal, { props: { open: true } });

    await fireEvent.click(screen.getByTestId("smart-create-tile-url"));
    expect(screen.getByTestId("smart-create-url")).toBeInTheDocument();

    await fireEvent.click(screen.getByTestId("smart-create-tile-file"));
    expect(screen.getByTestId("smart-create-file")).toBeInTheDocument();
  });

  it("submits text and shows the queued state", async () => {
    render(SmartCreateModal, { props: { open: true } });

    await fireEvent.input(screen.getByTestId("smart-create-text"), {
      target: { value: "three paragraphs" },
    });
    await fireEvent.click(screen.getByTestId("smart-create-submit"));

    expect(createSmartImport).toHaveBeenCalledWith({ type: "TEXT", content: "three paragraphs" });
    await waitFor(() => expect(screen.getByTestId("smart-create-pending")).toBeInTheDocument());
  });

  it("renders the created notes when the task completes", async () => {
    vi.mocked(getImportStatus).mockResolvedValue(doneStatus);
    render(SmartCreateModal, { props: { open: true } });

    await fireEvent.input(screen.getByTestId("smart-create-text"), { target: { value: "t" } });
    await fireEvent.click(screen.getByTestId("smart-create-submit"));

    await waitFor(() => expect(screen.getByTestId("smart-create-done")).toBeInTheDocument());
    expect(screen.getByText("Alpha")).toBeInTheDocument();
    expect(screen.getByText("Beta")).toBeInTheDocument();
    expect(screen.getByTestId("smart-create-open-graph")).toBeInTheDocument();
  });

  it("shows the raw-fragment line for a PARTIAL result", async () => {
    vi.mocked(getImportStatus).mockResolvedValue({ ...doneStatus, raw_chunks_count: 3 });
    render(SmartCreateModal, { props: { open: true } });

    await fireEvent.input(screen.getByTestId("smart-create-text"), { target: { value: "t" } });
    await fireEvent.click(screen.getByTestId("smart-create-submit"));

    await waitFor(() => expect(screen.getByTestId("smart-create-raw")).toBeInTheDocument());
    expect(screen.getByTestId("smart-create-raw")).toHaveTextContent("3");
  });

  it("shows handler errors on FAILED and the timeout text on reason=timeout", async () => {
    vi.mocked(getImportStatus).mockResolvedValue({
      task_id: "task-1",
      status: "failed",
      progress: { total: 0, processed: 0, created: 0, skipped: 0, failed: 0 },
      errors: ["parse failed"],
    });
    const first = render(SmartCreateModal, { props: { open: true } });

    await fireEvent.input(screen.getByTestId("smart-create-text"), { target: { value: "t" } });
    await fireEvent.click(screen.getByTestId("smart-create-submit"));

    await waitFor(() => expect(screen.getByTestId("smart-create-failed")).toBeInTheDocument());
    expect(screen.getByText("parse failed")).toBeInTheDocument();
    first.unmount();

    vi.mocked(getImportStatus).mockResolvedValue({
      task_id: "task-2",
      status: "failed",
      progress: { total: 0, processed: 0, created: 0, skipped: 0, failed: 0 },
      reason: "timeout",
    });
    render(SmartCreateModal, { props: { open: true } });
    await fireEvent.input(screen.getByTestId("smart-create-text"), { target: { value: "t" } });
    await fireEvent.click(screen.getByTestId("smart-create-submit"));

    await waitFor(() => expect(screen.getByTestId("smart-create-failed")).toBeInTheDocument());
    expect(screen.getByText("The handler did not answer")).toBeInTheDocument();
  });

  it("collapse hides the window while the task stays queued", async () => {
    const onClose = vi.fn();
    render(SmartCreateModal, { props: { open: true, onClose } });

    await fireEvent.input(screen.getByTestId("smart-create-text"), { target: { value: "t" } });
    await fireEvent.click(screen.getByTestId("smart-create-submit"));
    await waitFor(() => expect(screen.getByTestId("smart-create-pending")).toBeInTheDocument());

    await fireEvent.click(screen.getByTestId("smart-create-collapse"));
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(onClose).toHaveBeenCalled();
  });
});
