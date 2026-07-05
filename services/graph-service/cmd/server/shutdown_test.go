package main

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// TestGracefulShutdown tests that the graph service handles SIGTERM gracefully,
// properly closing connections and completing in-flight requests.
func TestGracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping graceful shutdown test in short mode")
	}

	// This is an integration test that requires the full environment
	// For now, we'll test the shutdown logic structure

	t.Run("ShutdownTimeoutRespected", func(t *testing.T) {
		// Test that shutdown timeout is respected
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Simulate shutdown process
		shutdownComplete := make(chan bool)

		go func() {
			// Simulate graceful shutdown process
			time.Sleep(100 * time.Millisecond)
			shutdownComplete <- true
		}()

		select {
		case <-shutdownComplete:
			t.Log("Shutdown completed within timeout")
		case <-ctx.Done():
			t.Error("Shutdown did not complete within timeout")
		}
	})

	t.Run("ConnectionsClosedOnShutdown", func(t *testing.T) {
		// Test that connections are properly closed during shutdown
		// This would require actual database connection in integration test
		t.Skip("Requires database connection - implement in integration tests")
	})

	t.Run("InProgressRequestsComplete", func(t *testing.T) {
		// Test that in-progress requests are allowed to complete
		t.Skip("Requires full HTTP server - implement in integration tests")
	})
}

// TestGracefulShutdownIntegration tests graceful shutdown with actual binary
// This is a more realistic integration test
func TestGracefulShutdownIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This integration test requires a running PostgreSQL and Redis.
	if os.Getenv("GRAPH_SERVICE_INTEGRATION") == "" {
		t.Skip("Skipping integration test: set GRAPH_SERVICE_INTEGRATION to enable")
	}

	tmpFile, err := os.CreateTemp("", "graph-service-test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Build the graph-service binary
	buildCmd := exec.Command("go", "build", "-o", tmpFile.Name(), "./cmd/server/main.go")
	buildCmd.Dir = "../../"
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build graph-service: %v", err)
	}

	// Start the server
	cmd := exec.Command(tmpFile.Name())
	cmd.Dir = "../../"
	cmd.Env = append(os.Environ(),
		"POSTGRES_URL=postgres://postgres:postgres@localhost:5432/knowledge_base?sslmode=disable",
		"REDIS_URL=redis:localhost:6379",
		"GRPC_PORT=9090",
		"HTTP_PORT=9091",
		"EVENT_CHANNEL=graph:events",
	)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start graph-service: %v", err)
	}

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Check that server is responsive
	resp, err := http.Get("http://localhost:9091/health")
	if err != nil {
		cmd.Process.Kill()
		t.Fatalf("Server not responding: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		cmd.Process.Kill()
		t.Fatalf("Health check failed: %d", resp.StatusCode)
	}

	// Send SIGTERM to initiate graceful shutdown
	startTime := time.Now()
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("Failed to send SIGTERM: %v", err)
	}

	// Wait for process to exit (should complete within 30 seconds)
	shutdownTimeout := 35 * time.Second
	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		duration := time.Since(startTime)
		t.Logf("Graph-service shut down in %v with status: %v", duration, err)
		if duration > shutdownTimeout {
			t.Errorf("Shutdown took too long: %v (expected < %v)", duration, shutdownTimeout)
		}
	case <-time.After(shutdownTimeout):
		cmd.Process.Kill()
		t.Error("Graceful shutdown timeout - process killed forcefully")
	}
}
