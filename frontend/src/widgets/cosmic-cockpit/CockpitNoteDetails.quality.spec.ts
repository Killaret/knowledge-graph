// NOTE-QUALITY-1: the quality row in the note panel — statuses, actions.
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, cleanup, waitFor, fireEvent } from "@testing-library/svelte";
import CockpitNoteDetails from "./CockpitNoteDetails.svelte";
import * as notesApi from "$shared/api/notes";
import * as linksApi from "$shared/api/links";
import * as qualityApi from "$shared/api/quality";

vi.mock("$shared/api/notes", () => ({ getNote: vi.fn() }));
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

const mockNote = {
  id: "n1",
  title: "Imported page",
  content: "## [Src](https://x.example)\n\nbody text",
  metadata: { source_url: "https://x.example" },
  type: "star",
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-06-01T00:00:00Z",
};

function record(
  over: Partial<qualityApi.QualityRecord["signals"]> = {},
  gates: string[] = [],
  verdict = "create"
) {
  return {
    signals: {
      words: 10,
      prose_words: 8,
      ends_with_sentence: true,
      truncated_by_import: false,
      unclosed_fence: false,
      headings: 1,
      coherence_min: null,
      coherence_median: null,
      sentences: 2,
      prose_share: 0.8,
      kind: "text",
      fragment_share: 0,
      max_block_words: 10,
      has_source_link: true,
      title_words: 3,
      title_generic: false,
      title_text_similarity: null,
      origin: "import",
      keywords: 0,
      links: 0,
      has_embedding: false,
      mojibake: false,
      ...over,
    },
    gates,
    verdict,
    reasons: gates,
    attempt: 1,
    needs_manual_review: false,
    computed_at: "2024-06-01T00:00:00Z",
    pipeline_version: "quality-v1",
    source_hash: "abc",
    trigger: "auto",
    can_refetch: false,
  } as qualityApi.QualityRecord;
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(notesApi.getNote).mockResolvedValue(mockNote as any);
  vi.mocked(linksApi.getNoteLinks).mockResolvedValue([]);
  vi.mocked(qualityApi.assessNoteQuality).mockResolvedValue({ enabled: true, enqueued: true });
});

afterEach(cleanup);

describe("Quality row", () => {
  it("is hidden when the feature is disabled", async () => {
    vi.mocked(qualityApi.getNoteQuality).mockResolvedValue({ enabled: false, quality: null });
    const { queryByTestId } = render(CockpitNoteDetails, { props: { nodeId: "n1" } });
    await waitFor(() => expect(notesApi.getNote).toHaveBeenCalled());
    await waitFor(() => expect(qualityApi.getNoteQuality).toHaveBeenCalled());
    expect(queryByTestId("quality-row")).toBeNull();
  });

  it("shows 'looks fine' for a clean verdict", async () => {
    vi.mocked(qualityApi.getNoteQuality).mockResolvedValue({ enabled: true, quality: record() });
    const { findByTestId } = render(CockpitNoteDetails, { props: { nodeId: "n1" } });
    const status = await findByTestId("quality-status");
    expect(status.textContent).toMatch(/looks fine|в порядке/);
  });

  it("shows 'link only' for a stub note", async () => {
    vi.mocked(qualityApi.getNoteQuality).mockResolvedValue({
      enabled: true,
      quality: record({ kind: "stub" }, ["stub"], "enrich"),
    });
    const { findByTestId } = render(CockpitNoteDetails, { props: { nodeId: "n1" } });
    const status = await findByTestId("quality-status");
    expect(status.textContent).toMatch(/link only|только ссылка/);
  });

  it("shows 'truncated' and the re-fetch button for truncated imports", async () => {
    const q = record({ truncated_by_import: true }, ["truncated"], "enrich");
    q.can_refetch = true;
    vi.mocked(qualityApi.getNoteQuality).mockResolvedValue({ enabled: true, quality: q });
    const { findByTestId } = render(CockpitNoteDetails, { props: { nodeId: "n1" } });
    const status = await findByTestId("quality-status");
    expect(status.textContent).toMatch(/truncated|обрезана/);
    expect(await findByTestId("quality-refetch")).toBeTruthy();
  });

  it("shows 'broken encoding' for mojibake", async () => {
    vi.mocked(qualityApi.getNoteQuality).mockResolvedValue({
      enabled: true,
      quality: record({ mojibake: true }, ["mojibake"], "manual"),
    });
    const { findByTestId } = render(CockpitNoteDetails, { props: { nodeId: "n1" } });
    const status = await findByTestId("quality-status");
    expect(status.textContent).toMatch(/broken encoding|битая кодировка/);
  });

  it("enqueues a manual assessment on 'Re-assess'", async () => {
    vi.mocked(qualityApi.getNoteQuality).mockResolvedValue({ enabled: true, quality: record() });
    const { findByTestId } = render(CockpitNoteDetails, { props: { nodeId: "n1" } });
    await fireEvent.click(await findByTestId("quality-improve"));
    await waitFor(() => expect(qualityApi.assessNoteQuality).toHaveBeenCalledWith("n1"));
  });

  it("previews a re-fetch without applying it", async () => {
    const q = record({ truncated_by_import: true }, ["truncated"], "enrich");
    q.can_refetch = true;
    vi.mocked(qualityApi.getNoteQuality).mockResolvedValue({ enabled: true, quality: q });
    vi.mocked(qualityApi.refetchPreview).mockResolvedValue({
      suggested_title: "Fresh title",
      title_candidates: ["Fresh title"],
      outline: [{ level: 2, text: "Section" }],
      length_runes: 5000,
      current_runes: 120,
    });

    const { findByTestId, queryByTestId } = render(CockpitNoteDetails, { props: { nodeId: "n1" } });
    await fireEvent.click(await findByTestId("quality-refetch"));

    const box = await findByTestId("refetch-preview");
    expect(box.textContent).toContain("Fresh title");
    expect(qualityApi.refetchApply).not.toHaveBeenCalled();

    // Cancel closes the preview without touching the note.
    await fireEvent.click(await findByTestId("refetch-cancel"));
    expect(queryByTestId("refetch-preview")).toBeNull();
  });

  it("applies a re-fetch and reloads note + quality", async () => {
    const q = record({ truncated_by_import: true }, ["truncated"], "enrich");
    q.can_refetch = true;
    vi.mocked(qualityApi.getNoteQuality).mockResolvedValue({ enabled: true, quality: q });
    vi.mocked(qualityApi.refetchPreview).mockResolvedValue({
      suggested_title: "Fresh title",
      length_runes: 5000,
      current_runes: 120,
    });
    vi.mocked(qualityApi.refetchApply).mockResolvedValue({
      title: "Fresh title",
      content_runes: 5000,
      restorable: true,
    });

    const { findByTestId } = render(CockpitNoteDetails, { props: { nodeId: "n1" } });
    await fireEvent.click(await findByTestId("quality-refetch"));
    await fireEvent.click(await findByTestId("refetch-apply"));

    await waitFor(() => expect(qualityApi.refetchApply).toHaveBeenCalledWith("n1", "Fresh title"));
    // note reloaded after apply
    await waitFor(() => expect(notesApi.getNote).toHaveBeenCalledTimes(2));
  });

  it("shows restore when metadata.previous_content is present", async () => {
    const noteWithPrev = {
      ...mockNote,
      metadata: {
        source_url: "https://x.example",
        previous_content: { content: "old", title: "old" },
      },
    };
    vi.mocked(notesApi.getNote).mockResolvedValue(noteWithPrev as any);
    vi.mocked(qualityApi.getNoteQuality).mockResolvedValue({ enabled: true, quality: record() });
    vi.mocked(qualityApi.refetchRestore).mockResolvedValue({ restored: true });

    const { findByTestId } = render(CockpitNoteDetails, { props: { nodeId: "n1" } });
    const btn = await findByTestId("quality-restore");
    await fireEvent.click(btn);
    await waitFor(() => expect(qualityApi.refetchRestore).toHaveBeenCalledWith("n1"));
  });
});
