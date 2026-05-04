-- Migration 016: Rollback authentication and sharing tables

-- Drop views
DROP VIEW IF EXISTS note_access_view;
DROP VIEW IF EXISTS user_permissions_view;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS share_links;
DROP TABLE IF EXISTS note_shares;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;

-- Remove columns from existing tables
ALTER TABLE users DROP COLUMN IF EXISTS role_id;
ALTER TABLE users DROP COLUMN IF EXISTS email;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE notes DROP COLUMN IF EXISTS creator_id;
ALTER TABLE links DROP COLUMN IF EXISTS creator_id;

-- Recreate old share_links table (from migration 009)
CREATE TABLE IF NOT EXISTS share_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    shared_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMIT;
