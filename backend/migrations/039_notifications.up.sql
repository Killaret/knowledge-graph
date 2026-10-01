-- COMET-1 stage C: in-app notifications. Comet reminders land here first;
-- the table is intentionally generic (type + optional note link) so later
-- notification kinds reuse it.
CREATE TABLE IF NOT EXISTS notifications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    note_id uuid REFERENCES notes(id) ON DELETE CASCADE,
    type varchar(50) NOT NULL,
    title text NOT NULL,
    body text NOT NULL DEFAULT '',
    -- dedupe_key makes re-fired or duplicated asynq tasks idempotent
    dedupe_key text,
    created_at timestamptz NOT NULL DEFAULT now(),
    read_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_notifications_dedupe
    ON notifications(dedupe_key) WHERE dedupe_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_notifications_user_created
    ON notifications(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
    ON notifications(user_id) WHERE read_at IS NULL;
