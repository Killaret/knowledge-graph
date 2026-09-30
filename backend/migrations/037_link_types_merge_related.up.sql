-- LINK-TYPES-1: `related` becomes the single generic link type.
-- `reference` and `custom` merge into `related`; `dependency`, `parent`,
-- `child` are untouched (parent/child move out in ORIGIN-1).
--
-- A pair is UNORDERED (decision 53, same as SaveUserLink): one edge per pair
-- in any direction. For each pair with a live legacy row we keep exactly one
-- live merged row:
--   - survivor preference: live outranks deleted, user outranks gamma, then
--     weight — a user's link never becomes a model one, and a rejected
--     (deleted) link stays deleted;
--   - survivor weight = max(weight) over the LIVE merged rows only
--     (related/reference/custom in both directions) — never from a
--     `dependency` row or a deleted one;
--   - when a user row wins over live gamma rows, the heaviest gamma's model
--     weight moves into metadata.gamma — the same promotion as SaveUserLink.
--
-- Absorbed rows are stamped (merged_by, link_type_before, was_deleted) so
-- the down migration can undo the merge. Absorbed `related` rows are
-- retyped to `absorbed_related`: UNIQUE (source,target,link_type) covers
-- soft-deleted rows, so the key must be freed before the survivor is
-- converted; the real type survives in metadata. Absorbed rows deleted via
-- a note are detached from deleted_via_note_id — recorded as
-- deleted_via_note_id_before — so a note restore cannot resurrect a legacy
-- type or double the pair.
--
-- The second pass covers deleted legacy rows on pairs with nothing live:
-- they get the same stamp and detach, the row itself stays deleted.
--
-- "Pairs merged" is printed via RAISE NOTICE.

DO $$
DECLARE
    pair         RECORD;
    survivor     RECORD;
    best_gamma   RECORD;
    max_w        float;
    merged_count int := 0;
    absorbed     int;
BEGIN
    FOR pair IN
        SELECT LEAST(source_note_id, target_note_id) AS a,
               GREATEST(source_note_id, target_note_id) AS b
        FROM links
        WHERE link_type IN ('reference', 'custom')
          AND deleted_at IS NULL
        GROUP BY 1, 2
    LOOP
        -- Survivor among the merged rows of the pair, either direction:
        -- live first, then user over gamma, then weight.
        SELECT * INTO survivor
        FROM links
        WHERE link_type IN ('related', 'reference', 'custom')
          AND LEAST(source_note_id, target_note_id) = pair.a
          AND GREATEST(source_note_id, target_note_id) = pair.b
        ORDER BY (deleted_at IS NULL) DESC, (source_type = 'user') DESC, weight DESC, id
        LIMIT 1;

        -- Provenance of the heaviest live gamma row (other than the
        -- survivor) moves into metadata.gamma when a user row wins.
        SELECT * INTO best_gamma
        FROM links
        WHERE link_type IN ('related', 'reference', 'custom')
          AND deleted_at IS NULL
          AND source_type = 'gamma'
          AND id <> survivor.id
          AND LEAST(source_note_id, target_note_id) = pair.a
          AND GREATEST(source_note_id, target_note_id) = pair.b
        ORDER BY weight DESC, id
        LIMIT 1;

        -- Group maximum over the LIVE merged rows — computed before the
        -- absorb step soft-deletes them.
        SELECT max(weight) INTO max_w
        FROM links
        WHERE link_type IN ('related', 'reference', 'custom')
          AND deleted_at IS NULL
          AND LEAST(source_note_id, target_note_id) = pair.a
          AND GREATEST(source_note_id, target_note_id) = pair.b;

        -- Absorb every other merged row of the pair FIRST (frees the
        -- (source,target,'related') key before the survivor converts):
        -- live rows are soft-deleted, `related` rows are retyped to the
        -- sentinel `absorbed_related`, note-deleted rows are detached.
        UPDATE links
        SET deleted_at = COALESCE(deleted_at, now()),
            deleted_via_note_id = NULL,
            link_type = CASE WHEN link_type = 'related'
                             THEN 'absorbed_related' ELSE link_type END,
            metadata = COALESCE(metadata, '{}'::jsonb)
                || jsonb_build_object('merged_by', '037_link_types_merge',
                                      'link_type_before', link_type,
                                      'was_deleted', deleted_at IS NOT NULL)
                || CASE WHEN deleted_via_note_id IS NOT NULL
                        THEN jsonb_build_object('deleted_via_note_id_before', deleted_via_note_id)
                        ELSE '{}'::jsonb END
        WHERE id <> survivor.id
          AND link_type IN ('related', 'reference', 'custom')
          AND LEAST(source_note_id, target_note_id) = pair.a
          AND GREATEST(source_note_id, target_note_id) = pair.b;
        GET DIAGNOSTICS absorbed = ROW_COUNT;

        UPDATE links
        SET link_type = 'related',
            weight = max_w,
            deleted_at = NULL,
            deleted_via_note_id = NULL,
            metadata = COALESCE(metadata, '{}'::jsonb)
                || jsonb_build_object('merged_by', '037_link_types_merge',
                                      'link_type_before', survivor.link_type,
                                      'was_deleted', survivor.deleted_at IS NOT NULL)
                || CASE WHEN survivor.source_type = 'user' AND best_gamma.id IS NOT NULL
                        THEN jsonb_build_object('gamma', jsonb_build_object(
                                 'score', best_gamma.weight,
                                 'generated_at', best_gamma.created_at))
                        ELSE '{}'::jsonb END
        WHERE id = survivor.id;

        merged_count := merged_count + 1;
        RAISE NOTICE 'link-type merge: pair % <-> % collapsed into related (% absorbed)', pair.a, pair.b, absorbed;
    END LOOP;

    -- Second pass: deleted legacy rows on pairs with nothing live — the
    -- main loop never reached them. Detach the note-restore marker so a
    -- restore cannot resurrect a legacy type; the row stays deleted.
    UPDATE links
    SET deleted_via_note_id = NULL,
        metadata = COALESCE(metadata, '{}'::jsonb)
            || jsonb_build_object('merged_by', '037_link_types_merge',
                                  'link_type_before', link_type,
                                  'was_deleted', true)
            || CASE WHEN deleted_via_note_id IS NOT NULL
                    THEN jsonb_build_object('deleted_via_note_id_before', deleted_via_note_id)
                    ELSE '{}'::jsonb END
    WHERE link_type IN ('reference', 'custom')
      AND deleted_at IS NOT NULL
      AND COALESCE(metadata->>'merged_by', '') <> '037_link_types_merge';

    RAISE NOTICE 'link-type merge: % pairs merged', merged_count;
END $$;

ALTER TABLE links ALTER COLUMN link_type SET DEFAULT 'related';
