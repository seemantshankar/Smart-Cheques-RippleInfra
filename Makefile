.PHONY: help build build-docker test clean up down logs deps db-migrate db-seed db-clear lint lint-fix lint-script

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
	@echo "  test-container - Run all tests in container (with database access)"
	@echo "  test-container-specific - Run specific tests in container (set TEST_PATH)"
	@echo "  clean     - Clean up Docker resources"
	@echo "  deps      - Download Go dependencies"
	@echo "  health    - Check health of all services"
	@echo "  db-migrate - Run database migrations"
	@echo "  db-seed   - Seed development data"
	@echo "  db-clear  - Clear development data"
	@echo ""
	@echo "Local Database Commands:"
	@echo "  local-dev        - Quick local development setup"
	@echo "  db-local-start   - Start local databases only"
	@echo "  db-local-stop    - Stop local databases"
	@echo "  db-local-restart - Restart local databases"
	@echo "  db-local-reset   - Reset local databases (fresh start)"
	@echo "  db-local-migrate - Run migrations on local database"
	@echo "  db-local-migrate-manual - Run migrations manually (bypasses migration tool)"
	@echo "  db-local-status  - Check local database status"
	@echo "  db-local-logs    - Show local database logs"
	@echo "  db-local-connect - Connect to local PostgreSQL"

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
	docker compose build

# Start all services
up:
	docker compose up -d

# Stop all services
down:
	docker compose down

# View logs
logs:
	docker compose logs -f

# Run tests
test:
	go test -v ./...

# Run tests in container (with database access)
test-container:
	@echo "Running tests in container with database access..."
	@./scripts/run-tests-in-container.sh all

# Run specific tests in container
test-container-specific:
	@echo "Running specific tests in container..."
	@./scripts/run-tests-in-container.sh specific $(TEST_PATH)

# Run linters
lint:
	golangci-lint run

# Run linters and auto-fix issues
lint-fix:
	golangci-lint run --fix

# Run linters using the script
lint-script:
	./scripts/lint.sh

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
	docker compose restart $*

# Build and restart a specific service
rebuild-%:
	docker compose build $*
	docker compose up -d $*

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

# Local database management (using local-db.sh script)
db-local-start:
	@echo "Starting local databases..."
	@./scripts/local-db.sh start

db-local-stop:
	@echo "Stopping local databases..."
	@./scripts/local-db.sh stop

db-local-restart:
	@echo "Restarting local databases..."
	@./scripts/local-db.sh restart

db-local-reset:
	@echo "Resetting local databases..."
	@./scripts/local-db.sh reset

db-local-migrate:
	@echo "Running migrations on local database..."
	@./scripts/local-db.sh migrate

db-local-migrate-manual:
	@echo "Running migrations manually..."
	@./scripts/run-migrations-manually.sh run

db-local-status:
	@echo "Checking local database status..."
	@./scripts/local-db.sh status

db-local-logs:
	@echo "Showing local database logs..."
	@./scripts/local-db.sh logs

db-local-connect:
	@echo "Connecting to local PostgreSQL..."
	@./scripts/local-db.sh connect

# Quick local development setup
local-dev: db-local-start
	@echo "Local databases started!"
	@echo "Run 'make db-local-migrate' to apply migrations"
	@echo "Run 'make db-local-status' to check status"