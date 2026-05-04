-- Migration 016: Add authentication and sharing tables
-- Created: 2024

-- 1. Add deleted_at to users table for soft delete
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- 2. Add email column to users for password reset and notifications
ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT UNIQUE;
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE email IS NOT NULL;

-- 3. Add creator_id to notes table
ALTER TABLE notes ADD COLUMN IF NOT EXISTS creator_id UUID REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_notes_creator_id ON notes(creator_id);

-- 4. Add creator_id to links table
ALTER TABLE links ADD COLUMN IF NOT EXISTS creator_id UUID REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_links_creator_id ON links(creator_id);

-- 5. Create note_shares table for direct user-to-user sharing
CREATE TABLE IF NOT EXISTS note_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    shared_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission TEXT NOT NULL DEFAULT 'read' CHECK (permission IN ('read', 'write')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    UNIQUE(note_id, shared_with_user_id)
);

CREATE INDEX IF NOT EXISTS idx_note_shares_note_id ON note_shares(note_id);
CREATE INDEX IF NOT EXISTS idx_note_shares_shared_by ON note_shares(shared_by_user_id);
CREATE INDEX IF NOT EXISTS idx_note_shares_shared_with ON note_shares(shared_with_user_id);
CREATE INDEX IF NOT EXISTS idx_note_shares_expires ON note_shares(expires_at) WHERE expires_at IS NOT NULL;

-- 6. Create share_links table for link-based sharing (replaces/extends existing share_links)
-- Drop old share_links if exists and create new one with enhanced fields
DROP TABLE IF EXISTS share_links CASCADE;

CREATE TABLE share_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    shared_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    permission TEXT NOT NULL DEFAULT 'read' CHECK (permission IN ('read', 'write')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    max_uses INTEGER,
    uses_count INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_share_links_token ON share_links(token);
CREATE INDEX IF NOT EXISTS idx_share_links_note_id ON share_links(note_id);
CREATE INDEX IF NOT EXISTS idx_share_links_expires ON share_links(expires_at) WHERE expires_at IS NOT NULL;

-- 7. Create api_keys table for API key authentication
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    scopes TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);

-- 8. Create refresh_tokens table for token rotation
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    device_info TEXT,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    replaced_by_token UUID REFERENCES refresh_tokens(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at);

-- 9. Create user_roles table for role-based access control
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Insert default roles
INSERT INTO user_roles (name, description) VALUES
    ('admin', 'System administrator with full access'),
    ('user', 'Regular user with standard permissions'),
    ('guest', 'Guest user with limited read-only access')
ON CONFLICT (name) DO NOTHING;

-- 10. Create role_permissions table for fine-grained permissions
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES user_roles(id) ON DELETE CASCADE,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(role_id, resource, action)
);

-- Insert default permissions for admin role
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'notes', 'create' FROM user_roles r WHERE r.name = 'admin' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'notes', 'read' FROM user_roles r WHERE r.name = 'admin' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'notes', 'update' FROM user_roles r WHERE r.name = 'admin' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'notes', 'delete' FROM user_roles r WHERE r.name = 'admin' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'notes', 'share' FROM user_roles r WHERE r.name = 'admin' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'links', 'create' FROM user_roles r WHERE r.name = 'admin' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'links', 'read' FROM user_roles r WHERE r.name = 'admin' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'links', 'update' FROM user_roles r WHERE r.name = 'admin' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'links', 'delete' FROM user_roles r WHERE r.name = 'admin' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'users', 'manage' FROM user_roles r WHERE r.name = 'admin' ON CONFLICT DO NOTHING;

-- Insert default permissions for user role
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'notes', 'create' FROM user_roles r WHERE r.name = 'user' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'notes', 'read' FROM user_roles r WHERE r.name = 'user' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'notes', 'update' FROM user_roles r WHERE r.name = 'user' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'notes', 'delete' FROM user_roles r WHERE r.name = 'user' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'notes', 'share' FROM user_roles r WHERE r.name = 'user' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'links', 'create' FROM user_roles r WHERE r.name = 'user' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'links', 'read' FROM user_roles r WHERE r.name = 'user' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'links', 'update' FROM user_roles r WHERE r.name = 'user' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'links', 'delete' FROM user_roles r WHERE r.name = 'user' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'users', 'read_own' FROM user_roles r WHERE r.name = 'user' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'users', 'update_own' FROM user_roles r WHERE r.name = 'user' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'users', 'delete_own' FROM user_roles r WHERE r.name = 'user' ON CONFLICT DO NOTHING;

-- Insert default permissions for guest role
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'notes', 'read' FROM user_roles r WHERE r.name = 'guest' ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, resource, action)
SELECT r.id, 'links', 'read' FROM user_roles r WHERE r.name = 'guest' ON CONFLICT DO NOTHING;

-- 11. Add role_id to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id UUID REFERENCES user_roles(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);

-- Set default role for existing users
UPDATE users SET role_id = (SELECT id FROM user_roles WHERE name = 'user') WHERE role_id IS NULL;

-- 12. Create audit_log table for security events
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    resource_type TEXT,
    resource_id UUID,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log(created_at);

-- 13. Create function for automatic updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 14. Create views for easier querying
CREATE OR REPLACE VIEW user_permissions_view AS
SELECT 
    u.id as user_id,
    r.name as role_name,
    rp.resource,
    rp.action
FROM users u
JOIN user_roles r ON u.role_id = r.id
JOIN role_permissions rp ON r.id = rp.role_id
WHERE u.deleted_at IS NULL;

CREATE OR REPLACE VIEW note_access_view AS
SELECT 
    n.id as note_id,
    n.creator_id,
    ns.shared_with_user_id,
    COALESCE(ns.permission, 'owner') as permission
FROM notes n
LEFT JOIN note_shares ns ON n.id = ns.note_id
WHERE n.deleted_at IS NULL OR n.deleted_at IS NOT NULL;

-- 15. Grant permissions (if using row-level security, enable here)
-- Note: Row-level security can be enabled later if needed

COMMIT;
