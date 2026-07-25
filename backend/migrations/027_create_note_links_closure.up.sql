CREATE MATERIALIZED VIEW note_links_closure AS
WITH RECURSIVE closure AS (
    SELECT
        l.source_note_id AS ancestor_id,
        l.target_note_id AS descendant_id,
        1 AS distance,
        l.weight AS path_weight,
        ARRAY[l.source_note_id, l.target_note_id]::uuid[] AS path
    FROM links l
    WHERE l.deleted_at IS NULL
    UNION ALL
    SELECT
        c.ancestor_id,
        l.target_note_id,
        c.distance + 1,
        c.path_weight * l.weight,
        c.path || l.target_note_id
    FROM closure c
    JOIN links l ON l.source_note_id = c.descendant_id
    WHERE l.deleted_at IS NULL
      AND c.distance < 10
      AND NOT l.target_note_id = ANY(c.path)
)
SELECT
    ancestor_id,
    descendant_id,
    MIN(distance) AS distance,
    MAX(path_weight) AS weight,
    (array_agg(path ORDER BY distance))[1] AS path
FROM closure
GROUP BY ancestor_id, descendant_id;

CREATE UNIQUE INDEX idx_note_links_closure_pk ON note_links_closure(ancestor_id, descendant_id);
CREATE INDEX idx_note_links_closure_descendant ON note_links_closure(descendant_id);
