package db

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// TestConnectionLeak tests that connections are properly returned to the pool
// after use, preventing connection leaks under high load.
func TestConnectionLeak(t *testing.T) {
	// Skip if database is not available for testing
	if testing.Short() {
		t.Skip("Skipping connection leak test in short mode")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("Skipping connection leak test because DATABASE_URL is not set")
	}

	// Initialize database connection
	database, err := Connect(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		sqlDB, _ := database.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Get initial pool statistics
	initialStats := sqlDB.Stats()
	t.Logf("Initial pool stats: Open=%d, InUse=%d, Idle=%d",
		initialStats.OpenConnections, initialStats.InUse, initialStats.Idle)

	// Simulate high load with 100 concurrent database operations
	const numGoroutines = 100
	var wg sync.WaitGroup

	// Channel to control goroutine start
	startCh := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Wait for all goroutines to be ready
			<-startCh

			// Perform database operation
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Simulate a query
			rows, err := sqlDB.QueryContext(ctx, "SELECT 1")
			if err != nil {
				t.Logf("Goroutine %d: Query failed: %v", id, err)
				return
			}
			defer rows.Close()

			// Consume results
			for rows.Next() {
				var result int
				if err := rows.Scan(&result); err != nil {
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
	finalStats := sqlDB.Stats()
	t.Logf("Final pool stats: Open=%d, InUse=%d, Idle=%d",
		finalStats.OpenConnections, finalStats.InUse, finalStats.Idle)

	// Check for connection leaks
	// After all operations, InUse should be 0
	if finalStats.InUse > 0 {
		t.Errorf("Connection leak detected: %d connections still in use after all operations completed", finalStats.InUse)
	}

	// OpenConnections should be reasonable (close to MaxIdleConns after settling)
	// Allow some tolerance for connection cleanup
	if finalStats.OpenConnections > initialStats.MaxOpenConnections+5 {
		t.Errorf("Too many open connections: %d (expected ~%d)", finalStats.OpenConnections, initialStats.MaxOpenConnections)
	}

	// Wait duration should be reasonable for the load
	if finalStats.WaitCount > 10 {
		t.Logf("Warning: High wait count: %d (possible contention)", finalStats.WaitCount)
	}

	t.Log("Connection leak test passed successfully")
}

// TestConnectionPoolStress tests connection pool behavior under stress
func TestConnectionPoolStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("Skipping connection pool stress test because DATABASE_URL is not set")
	}

	database, err := Connect(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		sqlDB, _ := database.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Test 1: Rapid connection acquisition and release
	t.Run("RapidAcquisition", func(t *testing.T) {
		const iterations = 1000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			rows, err := sqlDB.QueryContext(ctx, "SELECT 1")
			cancel()

			if err != nil {
				t.Logf("Iteration %d failed: %v", i, err)
				continue
			}

			rows.Close()
		}

		duration := time.Since(start)
		t.Logf("Rapid acquisition test: %d iterations in %v (%.2f ops/sec)",
			iterations, duration, float64(iterations)/duration.Seconds())

		// Verify connections are released
		stats := sqlDB.Stats()
		if stats.InUse > 0 {
			t.Errorf("Connections not released: %d still in use", stats.InUse)
		}
	})

	// Test 2: Concurrent long-running queries
	t.Run("ConcurrentLongQueries", func(t *testing.T) {
		const numGoroutines = 20
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				// Simulate a slightly longer query
				rows, err := sqlDB.QueryContext(ctx, "SELECT pg_sleep(0.1)")
				if err != nil {
					t.Logf("Goroutine %d: Long query failed: %v", id, err)
					return
				}
				rows.Close()
			}(i)
		}

		wg.Wait()

		stats := sqlDB.Stats()
		t.Logf("After long queries: Open=%d, InUse=%d, Idle=%d",
			stats.OpenConnections, stats.InUse, stats.Idle)

		if stats.InUse > 0 {
			t.Errorf("Long queries didn't release connections: %d still in use", stats.InUse)
		}
	})

	t.Log("Connection pool stress test passed")
}
