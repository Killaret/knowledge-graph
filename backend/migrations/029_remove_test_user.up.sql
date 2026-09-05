-- Migration 029: Remove test user previously inserted by migration 019
-- Created: 2026-09-05
--
-- Test data must be created by the seeder (backend/cmd/seed) when APP_ENV=test,
-- not by schema migrations. This migration removes the legacy test user so the
-- seeder can recreate it with a fresh Argon2 hash and the correct role.
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000000';
