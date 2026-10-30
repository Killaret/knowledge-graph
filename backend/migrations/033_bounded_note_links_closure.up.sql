-- LINKS-1 rework: bound the closure recursion and stop enumerating every
-- simple path.
--
-- The 032 shape walked ALL simple paths up to depth 10 to fill the `path`
-- column. Once gamma links gave every note two outgoing edges the graph
-- became dense and the working set exploded (x4 per level: ~81k paths at
-- depth 5, ~309k at 6, tens of millions by 10) — REFRESH took ~9 minutes on
-- the default seed and ran synchronously per event.
--
-- Consumers read `distance <= depth` with depth <= 5. The `path` column was
-- needed by a single query (path between two notes) and is now computed on
-- demand in the graph service instead of being stored.
--
--   levels: one row per (ancestor, descendant, distance) via UNION dedup —
--           bounded by pairs x depth, immune to graph density;
--   paths:  extends walks only while they can still be shortest (the pairs
--           join prunes any extension landing past the pair's min distance).
--           A walk reaching a pair at its minimal distance is simple by
--           definition — a cycle would make it strictly longer — so no
--           per-row cycle check is needed.
--
-- Semantics change: `weight` is now MAX(path_weight) over shortest paths,
-- not over all paths up to depth 10; `path` is no longer stored.

DROP MATERIALIZED VIEW IF EXISTS note_links_closure;

CREATE MATERIALIZED VIEW note_links_closure AS
WITH RECURSIVE
levels AS (
    SELECT
        l.source_note_id AS ancestor_id,
        l.target_note_id AS descendant_id,
        1 AS distance
    FROM links l
    WHERE l.deleted_at IS NULL
    UNION
    SELECT
        c.ancestor_id,
        l.target_note_id,
        c.distance + 1
    FROM levels c
    JOIN links l ON l.source_note_id = c.descendant_id
    WHERE l.deleted_at IS NULL
      AND c.distance < 5
),
pairs AS (
    SELECT ancestor_id, descendant_id, MIN(distance) AS distance
    FROM levels
    GROUP BY ancestor_id, descendant_id
),
paths AS (
    SELECT
        l.source_note_id AS ancestor_id,
        l.target_note_id AS descendant_id,
        1 AS distance,
        l.weight AS path_weight
    FROM links l
    WHERE l.deleted_at IS NULL
    UNION ALL
    SELECT
        p.ancestor_id,
        l.target_note_id,
        p.distance + 1,
        p.path_weight * l.weight
    FROM paths p
    JOIN links l ON l.source_note_id = p.descendant_id
    JOIN pairs pr
      ON pr.ancestor_id = p.ancestor_id
     AND pr.descendant_id = l.target_note_id
    WHERE l.deleted_at IS NULL
      AND p.distance < pr.distance
)
SELECT
    p.ancestor_id,
    p.descendant_id,
    p.distance,
    MAX(p.path_weight) AS weight
FROM paths p
GROUP BY p.ancestor_id, p.descendant_id, p.distance;

CREATE UNIQUE INDEX idx_note_links_closure_pk ON note_links_closure(ancestor_id, descendant_id);
CREATE INDEX idx_note_links_closure_descendant ON note_links_closure(descendant_id);
