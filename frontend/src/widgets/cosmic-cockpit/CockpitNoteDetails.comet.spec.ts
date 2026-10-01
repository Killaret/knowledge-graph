// COMET-1 stage E: settled comet gets an archive suggestion in the details
// panel — the app proposes moving to `debris`, never archives on its own.
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, cleanup, waitFor, fireEvent } from "@testing-library/svelte";
import CockpitNoteDetails from "./CockpitNoteDetails.svelte";
import * as notesApi from "$shared/api/notes";
import * as linksApi from "$shared/api/links";
import * as qualityApi from "$shared/api/quality";

vi.mock("$shared/api/notes", () => ({ getNote: vi.fn(), updateNote: vi.fn() }));
vi.mock("$shared/api/links", () => ({
  getNoteLinks: vi.fn(),
  deleteAllNoteLinks: vi.fn(),
  updateLink: vi.fn(),
  deleteLink: vi.fn(),
}));
vi.mock("$shared/api/quality", () => ({
  getNoteQuality: vi.fn(),
  assessNoteQuality: vi.fn(),
  refetchPreview: vi.fn(),
  refetchApply: vi.fn(),
  refetchRestore: vi.fn(),
}));
vi.mock("$app/navigation", () => ({ goto: vi.fn() }));

const baseComet = {
  id: "c1",
  title: "Visit",
  content: "",
  metadata: {},
  type: "comet",
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-06-01T00:00:00Z",
};

beforeEach(() => {
  vi.mocked(notesApi.getNote).mockReset();
  vi.mocked(notesApi.updateNote).mockReset();
  vi.mocked(notesApi.getNote).mockResolvedValue({ ...baseComet } as never);
  vi.mocked(linksApi.getNoteLinks).mockResolvedValue([]);
  vi.mocked(qualityApi.getNoteQuality).mockResolvedValue({
    enabled: false,
    quality: null,
  } as never);
});

afterEach(() => cleanup());

describe("CockpitNoteDetails — comet archive suggestion", () => {
  it("offers debris archive for a comet marked done", async () => {
    vi.mocked(notesApi.getNote).mockResolvedValueOnce({
      ...baseComet,
      due_at: "2030-01-01T00:00:00Z",
      done_at: "2026-01-01T00:00:00Z",
    } as never);

    render(CockpitNoteDetails, { props: { nodeId: "c1" } });

    await waitFor(() =>
      expect(document.querySelector('[data-testid="comet-archive-btn"]')).toBeTruthy()
    );
  });

  it("offers debris archive for a comet past its due date", async () => {
    vi.mocked(notesApi.getNote).mockResolvedValueOnce({
      ...baseComet,
      due_at: "2000-01-01T00:00:00Z",
    } as never);

    render(CockpitNoteDetails, { props: { nodeId: "c1" } });

    await waitFor(() =>
      expect(document.querySelector('[data-testid="comet-archive-suggestion"]')).toBeTruthy()
    );
  });

  it("hides the suggestion for a live dated comet and for non-comets", async () => {
    vi.mocked(notesApi.getNote).mockResolvedValueOnce({
      ...baseComet,
      due_at: "2099-01-01T00:00:00Z",
    } as never);
    const { unmount } = render(CockpitNoteDetails, { props: { nodeId: "c1" } });
    await waitFor(() => expect(vi.mocked(notesApi.getNote)).toHaveBeenCalled());
    await new Promise((r) => setTimeout(r, 10));
    expect(document.querySelector('[data-testid="comet-archive-suggestion"]')).toBeNull();
    unmount();

    vi.mocked(notesApi.getNote).mockResolvedValueOnce({
      ...baseComet,
      type: "star",
      due_at: "2000-01-01T00:00:00Z",
    } as never);
    render(CockpitNoteDetails, { props: { nodeId: "s1" } });
    await waitFor(() => expect(vi.mocked(notesApi.getNote)).toHaveBeenCalledTimes(2));
    await new Promise((r) => setTimeout(r, 10));
    expect(document.querySelector('[data-testid="comet-archive-suggestion"]')).toBeNull();
  });

  it("clicking archive converts the comet to debris", async () => {
    vi.mocked(notesApi.getNote).mockResolvedValueOnce({
      ...baseComet,
      due_at: "2000-01-01T00:00:00Z",
    } as never);
    vi.mocked(notesApi.updateNote).mockResolvedValueOnce({
      ...baseComet,
      type: "debris",
    } as never);

    render(CockpitNoteDetails, { props: { nodeId: "c1" } });

    await waitFor(() =>
      expect(document.querySelector('[data-testid="comet-archive-btn"]')).toBeTruthy()
    );
    await fireEvent.click(document.querySelector('[data-testid="comet-archive-btn"]') as Element);
    await waitFor(() => expect(notesApi.updateNote).toHaveBeenCalledWith("c1", { type: "debris" }));
  });
});
