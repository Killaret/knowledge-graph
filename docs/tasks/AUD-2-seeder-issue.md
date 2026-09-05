## Summary

`backend/cmd/seed` fails on every run. It cannot look up the `user` role, so the test user is never created.

Migration `029_remove_test_user` deletes the account that migration `019` used to insert, and the seeder is now the only thing that can recreate it. The test stack is therefore left with no account at the well-known zero UUID that `SKIP_AUTH` mode depends on.

## Steps to reproduce

Against a database with migrations applied:

```bash
APP_ENV=test \
SEED_TEST_USER_PASSWORD=any-password \
DATABASE_URL="postgresql://kb_user:kb_password@127.0.0.1:15434/knowledge_test?sslmode=disable" \
go run -C backend ./cmd/seed
```

## Actual

```
failed to look up 'user' role: sql: Scan error on column index 0, name "id":
converting driver.Value type string ("3916db86-7cd1-49b2-9aae-687e369fb1cf")
to a uint8: value out of range
exit status 1
```

Exit code 1, no user created.

## Expected

The seeder creates the test user with the `user` role and a fresh Argon2 hash of `SEED_TEST_USER_PASSWORD`, printing `Test user seeded successfully`.

## Cause

`backend/cmd/seed/main.go:56-58`

```go
var roleID uuid.UUID
if err := database.WithContext(ctx).Raw("SELECT id FROM user_roles WHERE name = 'user' LIMIT 1").Scan(&roleID).Error; err != nil {
```

`uuid.UUID` is `[16]byte`. GORM's `Scan` tries to fill it element by element from the driver's `string` value, and each element is a `uint8` that the text cannot fit. Execution never reaches the `roleID == uuid.Nil` guard on line 60, so that check is currently dead code.

Suggested fix: scan into a `string` and `uuid.Parse` it, or use `Pluck`, or `Row().Scan`. Keep the `uuid.Nil` guard — it becomes meaningful once reachable.

## Impact — measured, not assumed

Verified on the running test stack after migrations: `select count(*) from users` returns `0`.

Reads still work: in `SKIP_AUTH` mode `applyNoteScope` (`backend/internal/infrastructure/db/postgres/note_repo.go:40-44`) bypasses owner filtering entirely and never touches the `users` table.

Writes do not. `notes.creator_id` carries a foreign key to `users(id)` (`backend/migrations/016_add_auth_and_sharing.up.sql:17`), and `SKIP_AUTH` writes the zero UUID as the creator (`backend/internal/interfaces/api/middleware/skip_auth.go:31`). With no such row:

```
BEGIN;
INSERT INTO notes (id, title, content, creator_id, created_at, updated_at)
VALUES (gen_random_uuid(), 'fk probe', 'x', '00000000-0000-0000-0000-000000000000', NOW(), NOW());

ERROR:  insert or update on table "notes" violates foreign key constraint "notes_creator_id_fkey"
DETAIL:  Key (creator_id)=(00000000-0000-0000-0000-000000000000) is not present in table "users".
```

The same dependency exists for `user_settings.user_id`, `user_achievements.user_id`, `note_likes.user_id` and the sharing tables, all `NOT NULL REFERENCES users(id)`.

So any E2E or BDD scenario that creates a note, changes a setting or earns an achievement fails until the seeder works. Read-only scenarios pass, which makes the failure look intermittent rather than systematic.

Dev and personal stacks are unaffected beyond the removal of the legacy test account, which is the intended behaviour of migration `029`.

## Why CI did not catch it

`go test ./...` is green and honestly so — no test exercises the seeder against a database, so the failure is invisible to the suite. A regression test that brings up a database and runs the seeder should land with the fix; otherwise the next regression here passes unnoticed too.

## Design question worth separating

`SKIP_AUTH` is a test-only bypass, yet it depends on a specific row existing in the database. That coupling is what turned a seeder bug into a broken test stack.

`notes.creator_id` is nullable, and an insert with `creator_id = NULL` succeeds — so the bypass could plausibly avoid the persisted user altogether, or the fixture could be created by the test harness rather than by a production code path. Worth deciding deliberately rather than restoring the old arrangement by default.

The stale comment on `backend/internal/interfaces/api/middleware/skip_auth.go:55` — "Test user must exist in DB (migration 019)" — should be updated either way: migration `019` no longer provides that user.

## Minor, same area

`frontend/tests/auth-functional.spec.ts:30` looks for `text=test_user` — the login created by the now-removed migration `019`. The seeder and the shell seeders both create `testuser`. Worth checking on the next E2E run, since the locator sits in an `or` list and may pass on a different branch rather than fail loudly.

## Related

- Task specification: `docs/tasks/AUD-2-data-isolation.md`, acceptance criterion 4
- Review that found it: `docs/tasks/AUD-2-review-findings.md`
- Introduced in `889c33e`
