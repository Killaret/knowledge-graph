import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, cleanup, waitFor, fireEvent } from "@testing-library/svelte";
import CockpitNoteDetails from "./CockpitNoteDetails.svelte";
import * as notesApi from "$shared/api/notes";
import * as linksApi from "$shared/api/links";
import { goto } from "$app/navigation";

vi.mock("$shared/api/notes", () => ({
  getNote: vi.fn(),
}));

vi.mock("$shared/api/links", () => ({
  getNoteLinks: vi.fn(),
  deleteAllNoteLinks: vi.fn(),
  updateLink: vi.fn(),
  deleteLink: vi.fn(),
}));

const mockNote = {
  id: "n1",
  title: "Alpha Centauri",
  content: "A triple star system.",
  metadata: { tags: ["inbox", "star"] },
  type: "star",
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-06-01T00:00:00Z",
};

const mockLinks = [
  {
    id: "l1",
    source_note_id: "n1",
    target_note_id: "n2",
    link_type: "related",
    weight: 0.5,
    source_type: "gamma",
    last_weight_update: new Date(Date.now() - 1000 * 60 * 5).toISOString(),
  },
  {
    id: "l2",
    source_note_id: "n1",
    target_note_id: "n3",
    link_type: "parent",
    weight: 0.9,
    source_type: "manual",
    last_weight_update: undefined,
  },
];

function setDefaultMocks() {
  vi.mocked(notesApi.getNote).mockResolvedValue(mockNote as any);
  vi.mocked(linksApi.getNoteLinks).mockResolvedValue(mockLinks as any);
  vi.mocked(linksApi.deleteAllNoteLinks).mockResolvedValue(undefined as any);
  vi.mocked(linksApi.updateLink).mockResolvedValue(undefined as any);
  vi.mocked(linksApi.deleteLink).mockResolvedValue(undefined as any);
}

describe("CockpitNoteDetails Component", () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    setDefaultMocks();
  });

  afterEach(() => {
    cleanup();
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it("loads and displays note details", async () => {
    const { getByText, getByTestId } = render(CockpitNoteDetails, {
      props: { nodeId: "n1" },
    });

    await waitFor(() => {
      expect(getByTestId("cockpit-note-details")).toBeInTheDocument();
    });

    expect(notesApi.getNote).toHaveBeenCalledWith("n1");
    expect(linksApi.getNoteLinks).toHaveBeenCalledWith("n1");
    expect(getByText("Alpha Centauri")).toBeInTheDocument();
    expect(getByText("A triple star system.")).toBeInTheDocument();
  });

  it("calls onClose callback", async () => {
    const onClose = vi.fn();
    const { getByTitle } = render(CockpitNoteDetails, {
      props: { nodeId: "n1", onClose },
    });

    await waitFor(() => {
      expect(getByTitle("Close details")).toBeInTheDocument();
    });

    fireEvent.click(getByTitle("Close details"));
    expect(onClose).toHaveBeenCalled();
  });

  it("calls onEdit and onDelete callbacks", async () => {
    const onEdit = vi.fn();
    const onDelete = vi.fn();
    const { getByRole } = render(CockpitNoteDetails, {
      props: { nodeId: "n1", onEdit, onDelete },
    });

    await waitFor(() => {
      expect(getByRole("button", { name: /Edit note/i })).toBeInTheDocument();
    });

    fireEvent.click(getByRole("button", { name: /Edit note/i }));
    expect(onEdit).toHaveBeenCalledWith("n1");

    fireEvent.click(getByRole("button", { name: /Delete note/i }));
    expect(onDelete).toHaveBeenCalledWith("n1");
  });

  it("calls onCreateChildNote callback", async () => {
    const onCreateChildNote = vi.fn();
    const { getByTestId } = render(CockpitNoteDetails, {
      props: { nodeId: "n1", onCreateChildNote },
    });

    await waitFor(() => {
      expect(getByTestId("note-details-create-child")).toBeInTheDocument();
    });

    fireEvent.click(getByTestId("note-details-create-child"));
    expect(onCreateChildNote).toHaveBeenCalledWith(expect.objectContaining({ id: "n1" }));
  });

  it("displays error state when note fails to load", async () => {
    vi.mocked(notesApi.getNote).mockRejectedValue(new Error("Network error"));

    const { getByText } = render(CockpitNoteDetails, {
      props: { nodeId: "n1" },
    });

    await waitFor(() => {
      expect(getByText(/Failed to load note/i)).toBeInTheDocument();
    });
  });

  it("deletes all links after confirmation", async () => {
    const { getByRole, getByTestId } = render(CockpitNoteDetails, {
      props: { nodeId: "n1" },
    });

    await waitFor(() => {
      expect(getByRole("button", { name: /Delete all/i })).toBeInTheDocument();
    });

    fireEvent.click(getByRole("button", { name: /Delete all/i }));

    await waitFor(() => {
      expect(getByTestId("confirm-modal-confirm")).toBeInTheDocument();
    });

    fireEvent.click(getByTestId("confirm-modal-confirm"));

    await waitFor(() => {
      expect(linksApi.deleteAllNoteLinks).toHaveBeenCalledWith("n1");
    });
  });

  it("edits a link and saves changes", async () => {
    const { getAllByRole, getByRole } = render(CockpitNoteDetails, {
      props: { nodeId: "n1" },
    });

    await waitFor(() => {
      expect(getAllByRole("button", { name: /Edit link/i }).length).toBeGreaterThan(0);
    });

    fireEvent.click(getAllByRole("button", { name: /Edit link/i })[0]);

    await waitFor(() => {
      expect(getByRole("button", { name: /Save/i })).toBeInTheDocument();
    });

    fireEvent.click(getByRole("button", { name: /Save/i }));

    await waitFor(() => {
      expect(linksApi.updateLink).toHaveBeenCalledWith("l1", expect.any(Object));
    });
  });

  it("deletes a single link", async () => {
    const { getAllByRole } = render(CockpitNoteDetails, {
      props: { nodeId: "n1" },
    });

    await waitFor(() => {
      expect(getAllByRole("button", { name: /Delete link/i }).length).toBeGreaterThan(0);
    });

    fireEvent.click(getAllByRole("button", { name: /Delete link/i })[0]);

    await waitFor(() => {
      expect(linksApi.deleteLink).toHaveBeenCalledWith("l1");
    });
  });

  it("navigates to full note page", async () => {
    const { getByText } = render(CockpitNoteDetails, {
      props: { nodeId: "n1" },
    });

    await waitFor(() => {
      expect(getByText(/View full page/i)).toBeInTheDocument();
    });

    fireEvent.click(getByText(/View full page/i));

    expect(goto).toHaveBeenCalledWith("/notes/n1");
  });

  it("cancels delete-all confirmation", async () => {
    const { getByRole, getByTestId, queryByTestId } = render(CockpitNoteDetails, {
      props: { nodeId: "n1" },
    });

    await waitFor(() => {
      expect(getByRole("button", { name: /Delete all/i })).toBeInTheDocument();
    });

    fireEvent.click(getByRole("button", { name: /Delete all/i }));

    await waitFor(() => {
      expect(getByTestId("confirm-modal-cancel")).toBeInTheDocument();
    });

    fireEvent.click(getByTestId("confirm-modal-cancel"));

    await waitFor(() => {
      expect(queryByTestId("confirm-modal-confirm")).not.toBeInTheDocument();
    });
    expect(linksApi.deleteAllNoteLinks).not.toHaveBeenCalled();
  });

  it("cancels link edit", async () => {
    const { getAllByRole, getByRole, queryByRole } = render(CockpitNoteDetails, {
      props: { nodeId: "n1" },
    });

    await waitFor(() => {
      expect(getAllByRole("button", { name: /Edit link/i }).length).toBeGreaterThan(0);
    });

    fireEvent.click(getAllByRole("button", { name: /Edit link/i })[0]);

    await waitFor(() => {
      expect(getByRole("button", { name: /Cancel/i })).toBeInTheDocument();
    });

    fireEvent.click(getByRole("button", { name: /Cancel/i }));

    await waitFor(() => {
      expect(queryByRole("button", { name: /Cancel/i })).not.toBeInTheDocument();
    });
  });

  it("shows link update error and reloads links after success", async () => {
    vi.mocked(linksApi.updateLink)
      .mockRejectedValueOnce(new Error("update failed"))
      .mockResolvedValueOnce(undefined as any);

    const { getAllByRole, getByRole, getByText } = render(CockpitNoteDetails, {
      props: { nodeId: "n1" },
    });

    await waitFor(() => {
      expect(getAllByRole("button", { name: /Edit link/i }).length).toBeGreaterThan(0);
    });

    fireEvent.click(getAllByRole("button", { name: /Edit link/i })[0]);

    await waitFor(() => {
      expect(getByRole("button", { name: /Save/i })).toBeInTheDocument();
    });

    fireEvent.click(getByRole("button", { name: /Save/i }));

    await waitFor(() => {
      expect(getByText(/Failed to update link/i)).toBeInTheDocument();
    });

    fireEvent.click(getByRole("button", { name: /Save/i }));

    await waitFor(() => {
      expect(linksApi.getNoteLinks).toHaveBeenCalledTimes(2);
    });
  });

  it("shows note without tags and a link without last update", async () => {
    vi.mocked(notesApi.getNote).mockResolvedValue({
      ...mockNote,
      metadata: {},
    } as any);
    vi.mocked(linksApi.getNoteLinks).mockResolvedValue([
      { ...mockLinks[0], last_weight_update: undefined },
    ] as any);

    const { getByText, queryByText } = render(CockpitNoteDetails, {
      props: { nodeId: "n1" },
    });

    await waitFor(() => {
      expect(getByText("Alpha Centauri")).toBeInTheDocument();
    });

    expect(queryByText("#inbox")).not.toBeInTheDocument();
    expect(getByText(/Links \(1\)/i)).toBeInTheDocument();
  });
});
