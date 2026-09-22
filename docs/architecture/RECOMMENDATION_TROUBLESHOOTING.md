# Recommendation System Troubleshooting

## Problem: Recommendations Not Updating

### Checks

1. **Is worker running?** Check logs:
   ```bash
   docker logs knowledge-graph-worker
   # or
   journalctl -u knowledge-graph-worker -f
   ```

2. **Are tasks in queue?**
   ```bash
   redis-cli LLEN asynq:{default}
   redis-cli ZCARD asynq:scheduled
   ```

3. **Are migrations applied?**
   ```bash
   psql -d knowledge_base -c "\dt note_recommendations"
   ```

4. **Check task errors:**
   ```bash
   redis-cli LLEN asynq:{default}:failed
   ```

### Solutions

**Worker not running:**
```bash
# Start worker
cd backend && go run cmd/worker/main.go
# or via systemd
systemctl start knowledge-graph-worker
```

**Queue empty but no updates:**
- Check event handlers are triggering tasks
- Verify `GetAffectedNotes` returns correct IDs

## Problem: Queue Overflow

### Symptoms
- Queue length > `ASYNQ_QUEUE_MAX_LEN`
- Tasks being rejected
- Increasing latency

### Solutions

1. **Increase queue limit:**
   ```bash
   export ASYNQ_QUEUE_MAX_LEN=50000
   ```

2. **Add rate limiting in CLI:**
   ```bash
   ./bin/recommendation-cli --batch-delay=60
   ```

3. **Increase task delay (more deduplication):**
   ```bash
   export RECOMMENDATION_TASK_DELAY_SECONDS=10
   ```

4. **Scale workers:**
   ```bash
   # Increase concurrency
   export ASYNQ_CONCURRENCY=20
   
   # Or run multiple worker instances
   ```

5. **Clear stuck tasks:**
   ```bash
   redis-cli DEL asynq:{default}
   # Warning: removes all pending tasks
   ```

## Problem: Stale Recommendations

### Symptoms
- `X-Recommendations-Stale: true` header present
- Recommendations don't reflect recent changes
- Freshness check fails

### Causes

1. **Worker overloaded** — increase `ASYNQ_CONCURRENCY`
2. **Tasks failing** — check error logs
3. **Redis unavailable** — verify connection
4. **Database locks** — check for long-running transactions

### Diagnostics

```bash
# Check worker processing time in logs
grep "Finished refresh" /var/log/worker.log | tail -20

# Check for slow queries
psql -d knowledge_base -c "SELECT * FROM pg_stat_activity WHERE state = 'active';"

# Check Redis connection
redis-cli ping
```

### Solutions

**High processing time (>100ms):**
- Optimize `GetNeighborsBatch` query
- Add index on `note_links` table
- Reduce `RECOMMENDATION_TOP_N`

**Failing tasks:**
```bash
# View failed tasks
redis-cli LRANGE asynq:{default}:failed 0 -1

# Retry specific task (requires asynqmon or custom script)
```

**Database locks:**
```sql
-- Find blocking queries
SELECT * FROM pg_locks WHERE NOT granted;

-- Kill long-running transaction
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'idle in transaction' 
AND now() - query_start > interval '5 minutes';
```

## Problem: Redis Fallback Not Working

### Checks

```bash
# Test Redis connection
redis-cli ping

# Check if fallback keys exist
redis-cli --scan --pattern "recommendations:*" | head

# Check TTL on keys
redis-cli TTL recommendations:{note_id}
```

### Solutions

**Redis connection failed:**
```bash
# Check Redis service
systemctl status redis

# Verify connection string
echo $REDIS_URL
```

**Fallback disabled:**
```bash
export RECOMMENDATION_FALLBACK_ENABLED=true
export RECOMMENDATION_FALLBACK_SEMANTIC_ENABLED=true
```

## Problem: Initial Population Taking Too Long

### Symptoms
- CLI running for hours
- Queue keeps growing
- Memory usage increasing

### Solutions

1. **Reduce batch size:**
   ```bash
   ./bin/recommendation-cli --batch-size=50
   ```

2. **Increase delay between batches:**
   ```bash
   ./bin/recommendation-cli --batch-delay=120
   ```

3. **Run during off-peak hours** to reduce contention

4. **Monitor progress:**
   ```bash
   # Check how many notes have recommendations
   psql -d knowledge_base -c "SELECT COUNT(DISTINCT note_id) FROM note_recommendations;"
   
   # Compare with total notes
   psql -d knowledge_base -c "SELECT COUNT(*) FROM notes;"
   ```

## Performance Tuning

### Optimal Settings by Database Size

| Notes Count | Concurrency | Batch Delay | Queue Max |
|-------------|-------------|-------------|-----------|
| < 100       | 5           | 2s          | 1000      |
| 100-1000    | 10          | 5s          | 5000      |
| 1000-5000   | 15          | 10s         | 10000     |
| > 5000      | 20          | 30s         | 50000     |

### Monitoring Dashboard Commands

```bash
# Queue health
watch -n 2 'redis-cli LLEN asynq:{default}'

# Worker throughput
tail -f /var/log/worker.log | grep "Finished refresh" | wc -l

# Database size
psql -d knowledge_base -c "SELECT pg_size_pretty(pg_total_relation_size('note_recommendations'));"
```

## Emergency Procedures

### Clear All Recommendations and Rebuild

```bash
# 1. Stop worker
systemctl stop knowledge-graph-worker

# 2. Clear recommendations table
psql -d knowledge_base -c "TRUNCATE note_recommendations;"

# 3. Clear Redis cache
redis-cli --scan --pattern "recommendations:*" | xargs redis-cli del

# 4. Clear Asynq queues
redis-cli DEL asynq:{default} asynq:scheduled asynq:processed asynq:failed

# 5. Restart worker
systemctl start knowledge-graph-worker

# 6. Run initial population
./bin/recommendation-cli --batch-delay=60
```

### Disable Recommendations Temporarily

```bash
# Set to return empty results immediately
export RECOMMENDATION_FALLBACK_ENABLED=false
export RECOMMENDATION_FALLBACK_SEMANTIC_ENABLED=false

# Restart API service
systemctl restart knowledge-graph-api
```
