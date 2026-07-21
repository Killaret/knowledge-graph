# Backfill Recommendations

Bulk recalculation of recommendations for all notes in the database.

## Why is it needed?

The worker automatically updates recommendations only for **new or changed** notes. For existing notes, run the CLI manually.

## Quick start

```powershell
# Personal stack
docker compose -f docker-compose.personal.yml run --rm cli_personal ./cli

# Dev stack
docker compose run --rm cli ./cli
```

## Dry run (verify without creating tasks)

```powershell
docker compose -f docker-compose.personal.yml run --rm cli_personal ./cli --dry-run
```

## Progress monitoring

### 1. Worker logs

```powershell
docker compose -f docker-compose.personal.yml logs -f kg-worker-personal
```

Look for lines:

```
[RefreshService] Starting refresh recommendations for note {id}
[RefreshService] Successfully refreshed N recommendations for note {id}
```

### 2. Check in the database

```bash
docker compose -f docker-compose.personal.yml exec postgres_personal \
  psql -U personal -d knowledge_personal \
  -c "SELECT note_id, COUNT(*) as rec_count FROM note_recommendations GROUP BY note_id;"
```

### 3. Redis queue (asynqmon)

```bash
# Install asynqmon
go install github.com/hibiken/asynqmon@latest

# Run
asynqmon --redis-addr localhost:6380
```

## Configuration

Parameters from `knowledge-graph.config.json`:

```json
{
  "backend": {
    "recommendation": {
      "alpha": 0.5,
      "beta": 0.5,
      "gamma": 0.2,
      "depth": 3,
      "decay": 0.5,
      "top_n": 50,
      "keyword_similarity_method": "jaccard"
    }
  }
}
```

## Scoring formula

```
Final Score = α × Graph + β × Semantic + γ × Keyword
```

### Components

1. **Graph Score** (α = 0.5)
   - BFS traversal depth = 3
   - Decay factor = 0.5
   - Link types: reference, dependency, related

2. **Semantic Score** (β = 0.5)
   - Cosine similarity of 384-dimensional vectors
   - Model: `all-MiniLM-L6-v2`
   - From the `note_embeddings` table

3. **Keyword Score** (γ = 0.2)
   - Jaccard coefficient
   - Keywords from the NLP service
   - From the `note_keywords` table

## Troubleshooting

### Worker is not processing tasks

1. Check the queue:

   ```bash
   docker compose -f docker-compose.personal.yml exec redis_personal redis-cli LLEN asynq:{default}
   ```

2. Check worker logs:

   ```bash
   docker logs kg-worker-personal --tail 100
   ```

3. Restart the worker:

   ```bash
   docker compose -f docker-compose.personal.yml restart worker_personal
   ```

### Database connection errors

Check `DATABASE_URL`:

```bash
docker compose -f docker-compose.personal.yml exec backend_personal env | grep DATABASE_URL
```

### Empty recommendations

Possible causes:

- No links between notes → create links through the UI
- No embeddings → check the NLP service
- No keywords → check the NLP service

## Automation

Add to `crontab` for periodic recalculation:

```bash
# Once a week at 03:00
0 3 * * 0 docker compose -f /path/to/docker-compose.personal.yml run --rm cli_personal ./cli
```

## PowerShell helper script

Use `scripts/backfill-recommendations.ps1`:

```powershell
# Dry run
.\scripts\backfill-recommendations.ps1 -DryRun

# Real run
.\scripts\backfill-recommendations.ps1
```
