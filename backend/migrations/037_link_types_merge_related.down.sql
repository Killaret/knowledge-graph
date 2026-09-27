-- Undo 037_link_types_merge_related.
--
-- Rows absorbed by the merge were soft-deleted with their original type kept
-- in metadata->>'link_type_before'; revive them. The survivor keeps the merged
-- (maximum) weight — original per-row weights are not recoverable.

UPDATE links
SET deleted_at = NULL,
    link_type = metadata->>'link_type_before',
    metadata = metadata - 'merged_by' - 'link_type_before'
WHERE metadata->>'merged_by' = '037_link_types_merge';

ALTER TABLE links ALTER COLUMN link_type SET DEFAULT 'reference';
