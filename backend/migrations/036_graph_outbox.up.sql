-- SYNC-1 stage A2 (owner decision 71): transactional outbox for graph events.
-- Every repository write leaves a row here inside the same transaction; the
-- relayer publishes unsent rows to Redis and marks them sent. Rows survive a
-- crash between commit and publish — the relay re-delivers them after restart
-- (at-least-once delivery; subscribers tolerate duplicates).
CREATE TABLE IF NOT EXISTS graph_outbox (
    id          BIGSERIAL PRIMARY KEY,
    event_type  TEXT        NOT NULL,
    entity_id   UUID        NOT NULL,
    user_id     UUID,
    payload     JSONB       NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at     TIMESTAMPTZ
);

-- The relay's hot path: oldest unsent rows first.
CREATE INDEX IF NOT EXISTS idx_graph_outbox_unsent ON graph_outbox (id) WHERE sent_at IS NULL;
-- Sent rows are purged after the retention window; this index feeds the purge.
CREATE INDEX IF NOT EXISTS idx_graph_outbox_sent_at ON graph_outbox (sent_at) WHERE sent_at IS NOT NULL;
