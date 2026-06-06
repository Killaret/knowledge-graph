package db

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestConnectionLeak tests that pgxpool properly returns connections to the pool
// after use, preventing connection leaks under high load.
func TestConnectionLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping connection leak test in short mode")
	}

	ctx := context.Background()
	postgresURL := "postgres://postgres:postgres@localhost:5432/knowledge_base?sslmode=disable"

	// Configure pgxpool for testing
	config, err := pgxpool.ParseConfig(postgresURL)
	if err != nil {
		t.Fatalf("Failed to parse postgres config: %v", err)
	}

	// Use conservative settings for testing
	config.MaxConns = 5
	config.MinConns = 1
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create connection pool: %v", err)
	}
	defer pool.Close()

	// Get initial pool statistics
	initialStats := pool.Stat()
	t.Logf("Initial pool stats: TotalConns=%d, IdleConns=%d, MaxConns=%d",
		initialStats.TotalConns(), initialStats.IdleConns(), initialStats.MaxConns())

	// Simulate high load with 50 concurrent database operations
	const numGoroutines = 50
	var wg sync.WaitGroup
	startCh := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			<-startCh

			// Perform database operation
			queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			// Query notes to get some real data
			rows, err := pool.Query(queryCtx, "SELECT id, title FROM notes LIMIT 10")
			if err != nil {
				t.Logf("Goroutine %d: Query failed: %v", id, err)
				return
			}
			defer rows.Close()

			// Consume results
			for rows.Next() {
				var id string
				var title string
				if err := rows.Scan(&id, &title); err != nil {
					t.Logf("Goroutine %d: Scan failed: %v", id, err)
					return
				}
			}

			if err := rows.Err(); err != nil {
				t.Logf("Goroutine %d: Rows error: %v", id, err)
			}
		}(i)
	}

	// Start all goroutines simultaneously
	close(startCh)

	// Wait for all operations to complete
	wg.Wait()

	// Give some time for connections to be returned to pool
	time.Sleep(1 * time.Second)

	// Get final pool statistics
	finalStats := pool.Stat()
	t.Logf("Final pool stats: TotalConns=%d, IdleConns=%d, MaxConns=%d, AcquireCount=%d",
		finalStats.TotalConns(), finalStats.IdleConns(), finalStats.MaxConns(), finalStats.AcquireCount())

	// Check for connection leaks
	// After all operations, connections should be released back to pool
	if finalStats.TotalConns() > finalStats.MaxConns() {
		t.Errorf("Too many connections: %d (max: %d)", finalStats.TotalConns(), finalStats.MaxConns())
	}

	// Most connections should be idle after operations complete
	if finalStats.IdleConns() == 0 && finalStats.TotalConns() > 0 {
		t.Errorf("No idle connections after operations: likely connection leak")
	}

	t.Log("Connection leak test passed successfully")
}

// TestPoolStress tests pgxpool behavior under stress
func TestPoolStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ctx := context.Background()
	postgresURL := "postgres://postgres:postgres@localhost:5432/knowledge_base?sslmode=disable"

	config, err := pgxpool.ParseConfig(postgresURL)
	if err != nil {
		t.Fatalf("Failed to parse postgres config: %v", err)
	}

	config.MaxConns = 3 // Very small pool for stress testing
	config.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create connection pool: %v", err)
	}
	defer pool.Close()

	// Test 1: Rapid connection acquisition and release
	t.Run("RapidAcquisition", func(t *testing.T) {
		const iterations = 500
		start := time.Now()

		for i := 0; i < iterations; i++ {
			queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			rows, err := pool.Query(queryCtx, "SELECT 1")
			cancel()

			if err != nil {
				t.Logf("Iteration %d failed: %v", i, err)
				continue
			}

			rows.Close()
		}

		duration := time.Since(start)
		t.Logf("Rapid acquisition: %d iterations in %v (%.2f ops/sec)",
			iterations, duration, float64(iterations)/duration.Seconds())

		stats := pool.Stat()
		t.Logf("After rapid acquisition: TotalConns=%d, IdleConns=%d",
			stats.TotalConns(), stats.IdleConns())
	})

	// Test 2: Concurrent queries with limited pool
	t.Run("ConcurrentLimitedPool", func(t *testing.T) {
		const numGoroutines = 20
		const maxConns = 3
		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
				defer cancel()

				rows, err := pool.Query(queryCtx, "SELECT pg_sleep(0.05)")
				if err != nil {
					t.Logf("Goroutine %d failed: %v", id, err)
					return
				}
				rows.Close()

				mu.Lock()
				successCount++
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		t.Logf("Successfully completed %d/%d concurrent queries with pool size %d",
			successCount, numGoroutines, maxConns)

		stats := pool.Stat()
		t.Logf("After concurrent queries: TotalConns=%d, IdleConns=%d, AcquireCount=%d",
			stats.TotalConns(), stats.IdleConns(), stats.AcquireCount())

		// Pool should handle concurrent requests gracefully
		if successCount < numGoroutines/2 {
			t.Errorf("Too many failures: %d/%d successful", successCount, numGoroutines)
		}
	})

	t.Log("Pool stress test passed")
}