.PHONY: seed db-up db-down dev test

# Load environment variables
include .env
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

.DEFAULT_GOAL := dev
