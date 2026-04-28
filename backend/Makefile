.PHONY: seed db-up db-down dev test check-config check-migrations

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

.DEFAULT_GOAL := dev
