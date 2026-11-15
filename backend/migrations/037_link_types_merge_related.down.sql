-- Undo 037_link_types_merge_related.
--
-- Rows absorbed while live (was_deleted = false, including the survivor)
-- get their legacy type back and are revived; rows that were already
-- deleted keep their deleted_at but get their type and the
-- deleted_via_note_id marker back. The survivor keeps the merged (maximum)
-- weight — original per-row weights are not recoverable.

UPDATE links
SET link_type = metadata->>'link_type_before',
    deleted_at = NULL,
    deleted_via_note_id = NULL,
    metadata = metadata - 'merged_by' - 'link_type_before' - 'was_deleted' - 'deleted_via_note_id_before'
WHERE metadata->>'merged_by' = '037_link_types_merge'
  AND metadata->>'was_deleted' = 'false';

UPDATE links
SET link_type = metadata->>'link_type_before',
    deleted_via_note_id = (metadata->>'deleted_via_note_id_before')::uuid,
    metadata = metadata - 'merged_by' - 'link_type_before' - 'was_deleted' - 'deleted_via_note_id_before'
WHERE metadata->>'merged_by' = '037_link_types_merge'
  AND metadata->>'was_deleted' = 'true';

ALTER TABLE links ALTER COLUMN link_type SET DEFAULT 'reference';
