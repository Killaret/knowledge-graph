-- Migration 017: Create user_settings table
-- Created: 2024

-- Create user_settings table for flexible key-value storage
CREATE TABLE IF NOT EXISTS user_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key VARCHAR(100) NOT NULL,
    value JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, key)
);

-- Create index for fast lookups
CREATE INDEX IF NOT EXISTS idx_user_settings_user ON user_settings(user_id);
CREATE INDEX IF NOT EXISTS idx_user_settings_user_key ON user_settings(user_id, key);

-- Create trigger for updating updated_at
CREATE OR REPLACE FUNCTION update_user_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_user_settings_updated_at
    BEFORE UPDATE ON user_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_user_settings_updated_at();

-- Insert default settings for existing users
INSERT INTO user_settings (user_id, key, value)
SELECT id, 'galactic_mode', '{"value": false}'::jsonb
FROM users
WHERE deleted_at IS NULL
ON CONFLICT (user_id, key) DO NOTHING;

INSERT INTO user_settings (user_id, key, value)
SELECT id, 'show_achievement_notifications', '{"value": true}'::jsonb
FROM users
WHERE deleted_at IS NULL
ON CONFLICT (user_id, key) DO NOTHING;

INSERT INTO user_settings (user_id, key, value)
SELECT id, 'preferred_language', '{"value": "ru"}'::jsonb
FROM users
WHERE deleted_at IS NULL
ON CONFLICT (user_id, key) DO NOTHING;

COMMIT;
