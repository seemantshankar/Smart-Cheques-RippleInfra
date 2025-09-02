#!/bin/bash

# Manual Migration Runner for Smart Cheques Ripple Infrastructure
# This script runs migrations by executing SQL files directly

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="smart_payment"
DB_USER="user"
DB_PASSWORD="password"
MIGRATIONS_DIR="migrations"

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

# Function to check if PostgreSQL is accessible
check_postgres() {
    if ! docker compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
        print_error "Cannot connect to PostgreSQL. Make sure the database is running."
        exit 1
    fi
    print_success "PostgreSQL connection verified"
}

# Function to run a single migration file
run_migration() {
    local file="$1"
    local direction="$2"
    
    print_status "Running migration: $file ($direction)"
    
    if [ ! -f "$file" ]; then
        print_error "Migration file not found: $file"
        return 1
    fi
    
    # Run the migration
    if docker compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -f "/tmp/$(basename "$file")" > /dev/null 2>&1; then
        print_success "Migration completed: $file"
        return 0
    else
        print_error "Migration failed: $file"
        return 1
    fi
}

# Function to run all migrations
run_all_migrations() {
    print_status "Starting manual migration process..."
    
    # Check PostgreSQL connection
    check_postgres
    
    # Create migration tracking table if it doesn't exist
    print_status "Creating migration tracking table..."
    docker compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -c "
        CREATE TABLE IF NOT EXISTS manual_migrations (
            id SERIAL PRIMARY KEY,
            filename VARCHAR(255) NOT NULL,
            direction VARCHAR(10) NOT NULL,
            executed_at TIMESTAMP DEFAULT NOW()
        );
    " > /dev/null 2>&1
    
    # Get all migration files in order
    local up_files=($(find "$MIGRATIONS_DIR" -name "*.up.sql" | sort))
    
    print_status "Found ${#up_files[@]} migration files to run"
    
    # Run each migration
    for file in "${up_files[@]}"; do
        local filename=$(basename "$file")
        
        # Check if migration was already run
        if docker compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT 1 FROM manual_migrations WHERE filename = '$filename' AND direction = 'up';" | grep -q 1; then
            print_warning "Migration already executed: $filename (skipping)"
            continue
        fi
        
        # Copy migration file to container
        docker compose cp "$file" postgres:/tmp/
        
        # Run the migration
        if run_migration "$file" "up"; then
            # Record successful migration
            docker compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -c "
                INSERT INTO manual_migrations (filename, direction) VALUES ('$filename', 'up');
            " > /dev/null 2>&1
        else
            print_error "Migration failed: $filename"
            print_error "Stopping migration process"
            exit 1
        fi
    done
    
    print_success "All migrations completed successfully!"
}

# Function to show migration status
show_status() {
    print_status "Migration status:"
    
    if ! docker compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1 FROM manual_migrations LIMIT 1;" > /dev/null 2>&1; then
        print_warning "No migration tracking table found"
        return
    fi
    
    docker compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT 
            filename,
            direction,
            executed_at
        FROM manual_migrations 
        ORDER BY executed_at;
    "
}

# Function to show help
show_help() {
    echo "Manual Migration Runner"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  run       - Run all pending migrations"
    echo "  status    - Show migration status"
    echo "  help      - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 run     # Run all migrations"
    echo "  $0 status  # Check status"
}

# Main script logic
main() {
    case "${1:-help}" in
        run)
            run_all_migrations
            ;;
        status)
            show_status
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
