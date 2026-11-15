// NOTE-QUALITY-1: quality assessment and re-fetch endpoints.
import { api } from "./client";

export interface QualitySignals {
  words: number;
  prose_words: number;
  ends_with_sentence: boolean;
  truncated_by_import: boolean;
  unclosed_fence: boolean;
  headings: number;
  coherence_min: number | null;
  coherence_median: number | null;
  sentences: number;
  prose_share: number;
  kind: "stub" | "collection" | "text";
  fragment_share: number;
  max_block_words: number;
  has_source_link: boolean;
  title_words: number;
  title_generic: boolean;
  title_text_similarity: number | null;
  origin: "import" | "manual";
  keywords: number;
  links: number;
  has_embedding: boolean;
  mojibake: boolean;
}

export interface QualityRecord {
  signals: QualitySignals;
  gates: string[];
  verdict: "create" | "enrich" | "manual";
  reasons: string[];
  attempt: number;
  needs_manual_review: boolean;
  computed_at: string;
  pipeline_version: string;
  source_hash: string;
  trigger: "auto" | "manual";
  can_refetch: boolean;
}

export interface QualityLogEntry {
  note_id: string;
  source_hash: string;
  trigger: "auto" | "manual";
  changed: boolean;
  record: QualityRecord;
}

export interface QualityResponse {
  enabled: boolean;
  quality: QualityRecord | null;
  log?: QualityLogEntry[];
}

export interface RefetchPreview {
  suggested_title: string;
  title_candidates?: string[];
  title_source?: string;
  outline?: { level: number; text: string }[];
  noise_dropped?: number;
  sections_dropped?: number;
  length_runes: number;
  current_runes: number;
}

export async function getNoteQuality(id: string): Promise<QualityResponse> {
  return api.get(`v1/notes/${id}/quality`).json<QualityResponse>();
}

export async function assessNoteQuality(
  id: string
): Promise<{ enabled: boolean; enqueued?: boolean }> {
  return api.post(`v1/notes/${id}/quality/assess`).json<{ enabled: boolean; enqueued?: boolean }>();
}

export async function refetchPreview(id: string): Promise<RefetchPreview> {
  return api.post(`v1/notes/${id}/refetch/preview`).json<RefetchPreview>();
}

export async function refetchApply(
  id: string,
  title?: string
): Promise<{ title: string; content_runes: number; restorable: boolean }> {
  return api
    .post(`v1/notes/${id}/refetch`, { json: title ? { title } : {} })
    .json<{ title: string; content_runes: number; restorable: boolean }>();
}

export async function refetchRestore(id: string): Promise<{ restored: boolean }> {
  return api.post(`v1/notes/${id}/refetch/restore`).json<{ restored: boolean }>();
}
