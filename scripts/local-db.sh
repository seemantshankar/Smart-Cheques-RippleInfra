#!/bin/bash

# Local Database Management Script for Smart Cheques Ripple Infrastructure
# This script helps manage local PostgreSQL, MongoDB, and Redis instances

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.yml"
MIGRATIONS_DIR="migrations"
DB_NAME="smart_payment"
DB_USER="user"
DB_PASSWORD="password"
DB_HOST="localhost"
DB_PORT="5432"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
}

# Function to check if required tools are installed
check_requirements() {
    print_status "Checking requirements..."
    
    if ! docker compose version &> /dev/null; then
        print_error "docker compose is not available. Please install Docker Desktop first."
        exit 1
    fi
    
    if ! command -v psql &> /dev/null; then
        print_warning "psql (PostgreSQL client) is not installed. Some features may not work."
    fi
    
    print_success "Requirements check completed"
}

# Function to start local databases
start_db() {
    print_status "Starting local databases..."
    check_docker
    
    if docker compose -f "$COMPOSE_FILE" ps | grep -q "Up"; then
        print_warning "Some services are already running. Stopping them first..."
        docker compose -f "$COMPOSE_FILE" down
    fi
    
    docker compose -f "$COMPOSE_FILE" up -d postgres mongodb redis
    
    print_status "Waiting for databases to be ready..."
    sleep 10
    
    # Wait for PostgreSQL to be ready
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if docker compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U "$DB_USER" -d "$DB_NAME" > /dev/null 2>&1; then
            print_success "PostgreSQL is ready!"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            print_error "PostgreSQL failed to start within expected time"
            exit 1
        fi
        
        print_status "Waiting for PostgreSQL... (attempt $attempt/$max_attempts)"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_success "Local databases started successfully!"
    print_status "PostgreSQL: localhost:5432"
    print_status "MongoDB: localhost:27017"
    print_status "Redis: localhost:6379"
}

# Function to stop local databases
stop_db() {
    print_status "Stopping local databases..."
    docker compose -f "$COMPOSE_FILE" down
    print_success "Local databases stopped successfully!"
}

# Function to reset local databases (stop, remove volumes, start fresh)
reset_db() {
    print_status "Resetting local databases..."
    print_warning "This will delete all data in your local databases!"
    
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Reset cancelled"
        return
    fi
    
    stop_db
    
    print_status "Removing database volumes..."
    docker compose -f "$COMPOSE_FILE" down -v
    
    print_status "Starting fresh databases..."
    start_db
    
    print_success "Database reset completed successfully!"
}

# Function to run database migrations
run_migrations() {
    print_status "Running database migrations..."
    
    if ! docker-compose -f "$COMPOSE_FILE" ps | grep -q "postgres.*Up"; then
        print_error "PostgreSQL is not running. Please start the databases first."
        exit 1
    fi
    
    # Check if migrations directory exists
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        print_error "Migrations directory not found: $MIGRATIONS_DIR"
        exit 1
    fi
    
    # Run migrations using the db-migrate binary
    if [ -f "bin/db-migrate" ]; then
        print_status "Using db-migrate binary..."
        ./bin/db-migrate
    else
        print_warning "db-migrate binary not found. Building it first..."
        go build -o bin/db-migrate cmd/db-migrate/main.go
        ./bin/db-migrate
    fi
    
    print_success "Migrations completed successfully!"
}

# Function to check database status
status_db() {
    print_status "Checking database status..."
    
    if docker compose -f "$COMPOSE_FILE" ps | grep -q "Up"; then
        print_success "Databases are running:"
        docker compose -f "$COMPOSE_FILE" ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
    else
        print_warning "No databases are currently running"
    fi
    
    # Check if we can connect to PostgreSQL
    if command -v psql &> /dev/null; then
        if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT version();" > /dev/null 2>&1; then
            print_success "PostgreSQL connection test: OK"
        else
            print_warning "PostgreSQL connection test: Failed"
        fi
    fi
}

# Function to show database logs
logs_db() {
    print_status "Showing database logs..."
    docker compose -f "$COMPOSE_FILE" logs -f postgres mongodb redis
}

# Function to connect to PostgreSQL
connect_psql() {
    print_status "Connecting to PostgreSQL..."
    
    if ! command -v psql &> /dev/null; then
        print_error "psql is not installed. Please install PostgreSQL client tools."
        exit 1
    fi
    
    if ! docker-compose -f "$COMPOSE_FILE" ps | grep -q "postgres.*Up"; then
        print_error "PostgreSQL is not running. Please start the databases first."
        exit 1
    fi
    
    print_status "Connecting to PostgreSQL as $DB_USER..."
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"
}

# Function to show help
show_help() {
    echo "Local Database Management Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  start       Start local databases (PostgreSQL, MongoDB, Redis)"
    echo "  stop        Stop local databases"
    echo "  restart     Restart local databases"
    echo "  reset       Reset databases (delete all data and start fresh)"
    echo "  migrate     Run database migrations"
    echo "  status      Show database status"
    echo "  logs        Show database logs"
    echo "  connect     Connect to PostgreSQL using psql"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 start     # Start all databases"
    echo "  $0 migrate   # Run migrations"
    echo "  $0 status    # Check status"
    echo ""
    echo "Environment variables:"
    echo "  COMPOSE_FILE     Docker Compose file (default: docker-compose.yml)"
    echo "  MIGRATIONS_DIR   Migrations directory (default: migrations)"
}

# Main script logic
main() {
    case "${1:-help}" in
        start)
            check_requirements
            start_db
            ;;
        stop)
            check_requirements
            stop_db
            ;;
        restart)
            check_requirements
            stop_db
            sleep 2
            start_db
            ;;
        reset)
            check_requirements
            reset_db
            ;;
        migrate)
            check_requirements
            run_migrations
            ;;
        status)
            check_requirements
            status_db
            ;;
        logs)
            check_requirements
            logs_db
            ;;
        connect)
            check_requirements
            connect_psql
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
