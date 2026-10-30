-- Migration 019: Add test user for API testing
-- Created: 2026-05-05

-- Insert test user for SKIP_AUTH mode
-- This user is used when SKIP_AUTH=true for testing without authentication
-- Password: "test" (hashed with Argon2id)
-- ID is all zeros to match the SKIP_AUTH bypass user ID
INSERT INTO users (id, login, password_hash, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000000'::uuid,
    'test_user',
    '$argon2id$v=19$m=65536,t=3,p=4$vebiVgl5UEh0+ne0/eYhAg$jHdRBsuf4mGSnUUCJ04o3OMfkW1IGF3E4mHZFvjYjSE',
    NOW()
)
ON CONFLICT (id) DO NOTHING;
