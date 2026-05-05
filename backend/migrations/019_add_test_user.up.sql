-- Migration 019: Add test user for API testing
-- Created: 2026-05-05

-- Insert test user for SKIP_AUTH mode
-- This user is used when SKIP_AUTH=true for testing without authentication
INSERT INTO users (id, login, password_hash, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000001'::uuid,
    'test_user',
    'test',
    NOW()
)
ON CONFLICT (id) DO NOTHING;
