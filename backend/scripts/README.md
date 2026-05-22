Seed script for base achievements

Run the SQL seed against your development database:

```bash
psql "postgresql://<user>:<password>@<host>:<port>/<db>?sslmode=disable" -f scripts/seed_achievements.sql
```

Or, from the project root when `PGHOST`/`PGUSER`/`PGPASSWORD`/`PGDATABASE` are set:

```bash
psql -f backend/scripts/seed_achievements.sql
```

Note: Ensure migrations have been applied first (the achievements table exists).