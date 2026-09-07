-- Migration 018: Rollback achievements system

DROP TABLE IF EXISTS user_achievements;
DROP TABLE IF EXISTS achievements;

COMMIT;
