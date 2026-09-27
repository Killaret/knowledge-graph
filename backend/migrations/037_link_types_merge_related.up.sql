-- LINK-TYPES-1: `related` becomes the single generic link type.
-- `reference` and `custom` merge into `related`; `dependency`, `parent`,
-- `child` are untouched (parent/child move out in ORIGIN-1).
--
-- The (source_note_id, target_note_id, link_type) UNIQUE constraint covers
-- soft-deleted rows, so a plain UPDATE would collide whenever a pair already
-- has a `related` row in any state. For each such pair we keep exactly one
-- `related` row with the group's maximum weight and soft-delete the absorbed
-- rows — stamped in metadata so the down migration can undo the merge.
--
-- "Pairs merged" is printed via RAISE NOTICE.

DO $$
DECLARE
    pair        RECORD;
    survivor_id uuid;
BEGIN
    FOR pair IN
        SELECT source_note_id AS s, target_note_id AS t
        FROM links
        WHERE link_type IN ('reference', 'custom') AND deleted_at IS NULL
        GROUP BY source_note_id, target_note_id
    LOOP
        -- An existing `related` row (live or soft-deleted) wins the merge.
        SELECT id INTO survivor_id
        FROM links
        WHERE source_note_id = pair.s
          AND target_note_id = pair.t
          AND link_type = 'related'
        ORDER BY (deleted_at IS NULL) DESC, weight DESC, id
        LIMIT 1;

        IF FOUND THEN
            -- One related link stays with the largest weight of the group;
            -- a soft-deleted survivor is revived so the pair keeps its link.
            UPDATE links
            SET weight = (SELECT max(weight) FROM links
                          WHERE source_note_id = pair.s AND target_note_id = pair.t),
                deleted_at = NULL,
                deleted_via_note_id = NULL
            WHERE id = survivor_id;

            UPDATE links
            SET deleted_at = now(),
                metadata = COALESCE(metadata, '{}'::jsonb) ||
                           jsonb_build_object('merged_by', '037_link_types_merge',
                                              'link_type_before', link_type)
            WHERE source_note_id = pair.s
              AND target_note_id = pair.t
              AND link_type IN ('reference', 'custom')
              AND deleted_at IS NULL;

            RAISE NOTICE 'link-type merge: pair % -> % collapsed into related', pair.s, pair.t;
        ELSE
            -- No related row: convert the heaviest live row, soft-delete the
            -- rest of the pair (only one row per (pair, type) can exist).
            SELECT id INTO survivor_id
            FROM links
            WHERE source_note_id = pair.s
              AND target_note_id = pair.t
              AND link_type IN ('reference', 'custom')
              AND deleted_at IS NULL
            ORDER BY weight DESC, id
            LIMIT 1;

            -- The converted row keeps its original type in metadata too, so
            -- the down migration can restore it (RHS sees pre-update values).
            UPDATE links
            SET link_type = 'related',
                metadata = COALESCE(metadata, '{}'::jsonb) ||
                           jsonb_build_object('merged_by', '037_link_types_merge',
                                              'link_type_before', link_type)
            WHERE id = survivor_id;

            UPDATE links
            SET deleted_at = now(),
                metadata = COALESCE(metadata, '{}'::jsonb) ||
                           jsonb_build_object('merged_by', '037_link_types_merge',
                                              'link_type_before', link_type)
            WHERE source_note_id = pair.s
              AND target_note_id = pair.t
              AND link_type IN ('reference', 'custom')
              AND deleted_at IS NULL;
        END IF;
    END LOOP;
END $$;

ALTER TABLE links ALTER COLUMN link_type SET DEFAULT 'related';
