-- Migration 017: Rollback user_settings table

DROP TRIGGER IF EXISTS trigger_user_settings_updated_at ON user_settings;
DROP FUNCTION IF EXISTS update_user_settings_updated_at();

DROP TABLE IF EXISTS user_settings;

COMMIT;
