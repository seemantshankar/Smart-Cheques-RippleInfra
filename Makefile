.PHONY: help build build-docker test clean up down logs deps db-migrate db-seed db-clear

# Default target
help:
	@echo "Smart Payment Infrastructure - Development Commands"
	@echo ""
	@echo "Available commands:"
	@echo "  build       - Build all Go binaries"
	@echo "  build-docker - Build all Docker images"
	@echo "  up        - Start all services"
	@echo "  down      - Stop all services"
	@echo "  logs      - View logs from all services"
	@echo "  test      - Run all tests"
	@echo "  clean     - Clean up Docker resources"
	@echo "  deps      - Download Go dependencies"
	@echo "  health    - Check health of all services"
	@echo "  db-migrate - Run database migrations"
	@echo "  db-seed   - Seed development data"
	@echo "  db-clear  - Clear development data"

# Build all Go binaries
build:
	@echo "Building Go binaries..."
	@mkdir -p bin
	@go build -o bin/api-gateway ./cmd/api-gateway
	@go build -o bin/identity-service ./cmd/identity-service
	@go build -o bin/orchestration-service ./cmd/orchestration-service
	@go build -o bin/xrpl-service ./cmd/xrpl-service
	@go build -o bin/asset-gateway ./cmd/asset-gateway
	@go build -o bin/db-migrate ./cmd/db-migrate
	@echo "All binaries built successfully!"

# Build all Docker images
build-docker:
	docker-compose build

# Start all services
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# Run tests
test:
	go test -v ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@echo "Note: Ensure PostgreSQL is running on localhost:5432"
	go test -v ./test/integration/...

# Run unit tests only
test-unit:
	go test -v ./internal/... ./pkg/...

# Clean up Docker resources
clean:
	docker-compose down -v
	docker system prune -f
	rm -rf bin/

# Download dependencies
deps:
	go mod download
	go mod tidy

# Check health of all services
health:
	@echo "Checking service health..."
	@curl -s http://localhost:8000/health || echo "API Gateway: DOWN"
	@curl -s http://localhost:8001/health || echo "Identity Service: DOWN"
	@curl -s http://localhost:8002/health || echo "Orchestration Service: DOWN"
	@curl -s http://localhost:8003/health || echo "XRPL Service: DOWN"
	@curl -s http://localhost:8004/health || echo "Asset Gateway: DOWN"

# Development helpers
dev-up: deps build-docker up
	@echo "Development environment started!"
	@echo "Services available at:"
	@echo "  API Gateway:        http://localhost:8000"
	@echo "  Identity Service:   http://localhost:8001"
	@echo "  Orchestration:      http://localhost:8002"
	@echo "  XRPL Service:       http://localhost:8003"
	@echo "  Asset Gateway:      http://localhost:8004"

# Quick restart of a specific service
restart-%:
	docker-compose restart $*

# Build and restart a specific service
rebuild-%:
	docker-compose build $*
	docker-compose up -d $*

# Database management commands
db-migrate:
	@echo "Running database migrations..."
	@go run ./cmd/db-migrate -action=up

db-seed:
	@echo "Seeding development data..."
	@go run ./cmd/db-migrate -action=seed

db-clear:
	@echo "Clearing development data..."
	@go run ./cmd/db-migrate -action=clear

db-version:
	@echo "Checking migration version..."
	@go run ./cmd/db-migrate -action=version

db-rollback:
	@echo "Rolling back migrations..."
	@go run ./cmd/db-migrate -action=down