.PHONY: seed db-up db-down dev test migrate-up migrate-down

# Load environment variables
include .env
export

# Database commands
db-up:
	docker-compose up -d postgres

db-down:
	docker-compose down

# Migration commands
migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

# Seed test data
seed:
	go run scripts/seed.go

# Development

dev:
	go run cmd/api/main.go

# Run tests
test:
	go test ./...

# Build
build:
	go build -o bin/api cmd/api/main.go

.DEFAULT_GOAL := dev
