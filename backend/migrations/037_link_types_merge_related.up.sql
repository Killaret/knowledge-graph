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
--     `dependency` row or a deleted one — EXCEPT when a user row wins over
--     live gamma rows: then the survivor keeps the user's own weight and the
--     heaviest gamma's model weight moves into metadata.gamma — the same
--     promotion as SaveUserLink (A′ from review: the model's weight must not
--     become the user's);
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
-- The second pass covers deleted legacy rows the main loop never reached —
-- pairs with nothing live. A row deleted together with a note converts to
-- `related` and KEEPS deleted_via_note_id so "Restore" brings the pair's
-- only link back as the generic type (E′); when the (src,dst,'related') key
-- is already taken the row falls through to the final pass, which detaches
-- the marker — restoring the note then revives nothing extra — or when
-- another `related` row already sits on the unordered pair, so a restore
-- cannot double the edge. Rows deleted on their own (rejected) also get the
-- stamp and stay deleted.
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
            weight = CASE WHEN survivor.source_type = 'user' AND best_gamma.id IS NOT NULL
                          THEN survivor.weight
                          ELSE max_w END,
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

    -- Second pass, part 1: a legacy row deleted via a note on a pair with
    -- nothing live converts to `related` and keeps deleted_via_note_id, so
    -- restoring the note brings the pair's only link back as the generic
    -- type. The retype is skipped when another `related` row already exists
    -- on the unordered pair — a restore would double the edge — and only the
    -- best candidate per directed key converts: two same-direction legacy
    -- rows (reference + custom) retyped together would hit UNIQUE
    -- (source,target,link_type), which covers soft-deleted rows.
    UPDATE links l
    SET link_type = 'related',
        metadata = COALESCE(l.metadata, '{}'::jsonb)
            || jsonb_build_object('merged_by', '037_link_types_merge',
                                  'link_type_before', l.link_type,
                                  'was_deleted', true)
    WHERE l.link_type IN ('reference', 'custom')
      AND l.deleted_at IS NOT NULL
      AND l.deleted_via_note_id IS NOT NULL
      AND COALESCE(l.metadata->>'merged_by', '') <> '037_link_types_merge'
      AND NOT EXISTS (
          SELECT 1 FROM links o
          WHERE o.id <> l.id
            AND o.link_type = 'related'
            AND LEAST(o.source_note_id, o.target_note_id) = LEAST(l.source_note_id, l.target_note_id)
            AND GREATEST(o.source_note_id, o.target_note_id) = GREATEST(l.source_note_id, l.target_note_id)
      )
      AND NOT EXISTS (
          SELECT 1 FROM links p
          WHERE p.id <> l.id
            AND p.link_type IN ('reference', 'custom')
            AND p.deleted_at IS NOT NULL
            AND p.deleted_via_note_id IS NOT NULL
            AND p.source_note_id = l.source_note_id
            AND p.target_note_id = l.target_note_id
            AND COALESCE(p.metadata->>'merged_by', '') <> '037_link_types_merge'
            AND (p.weight, p.id) > (l.weight, l.id)
      );

    -- Second pass, part 2: every other deleted legacy row — standalone
    -- (rejected) ones and the collision leftovers above — detaches the
    -- note-restore marker so a restore cannot resurrect a legacy type or a
    -- second edge on the pair; the row itself stays deleted.
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
