.PHONY: seed db-up db-down dev test check-config check-migrations check-env-usage

# Load environment variables
-include .env
export

# Database commands
db-up:
	docker-compose up -d postgres

db-down:
	docker-compose down

# Seed test data (uses Python script from root)
seed:
	python ../scripts/seed_data.py

# Development

dev:
	go run cmd/server/main.go

# Run tests
test:
	go test ./...

# Build
build:
	go build -o bin/server cmd/server/main.go

build-worker:
	go build -o bin/worker cmd/worker/main.go

# Configuration validation
check-config:
	go run cmd/checkconfig/main.go

# Migration drift check - compares GORM models with database schema
check-migrations:
	@echo "Checking migration drift..."
	@echo "Note: This requires a running PostgreSQL instance"
	go run cmd/checkmigrations/main.go

# Check that all os.Getenv calls are only in config.go (with exceptions for tests)
check-env-usage:
	@echo "Checking for os.Getenv usage outside of config.go..."
	@# Find all os.Getenv calls in .go files except:
	@# - config.go itself
	@# - _test.go files
	@# - cmd/checkconfig/main.go (it sets up minimal env for validation)
	@# - cmd/ directories (CLI tools may need direct env access)
	@# - Any file with "test" in the path
	@found=$$(grep -r "os\.Getenv\|os\.GetEnv" --include="*.go" . | \
		grep -v "config/config.go" | \
		grep -v "_test.go" | \
		grep -v "cmd/checkconfig/main.go" | \
		grep -v "/cmd/" | \
		grep -v "/test" || true); \
	if [ -n "$$found" ]; then \
		echo "❌ ERROR: Found os.Getenv usage outside of config.go:"; \
		echo "$$found"; \
		echo ""; \
		echo "All environment variable access should go through internal/config package."; \
		echo "This ensures consistent loading priority: env > JSON config > hardcoded default."; \
		exit 1; \
	else \
		echo "✓ All os.Getenv usage is properly isolated in config.go"; \
	fi

.DEFAULT_GOAL := dev
