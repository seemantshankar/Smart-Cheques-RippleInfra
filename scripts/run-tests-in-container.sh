#!/bin/bash

# Test Runner Script for Smart Cheques Ripple Infrastructure
# This script runs tests inside the Docker container to access the local database

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to check if databases are running
check_databases() {
    if ! docker compose ps | grep -q "postgres.*Up"; then
        print_error "PostgreSQL is not running. Please start the databases first with 'make db-local-start'"
        exit 1
    fi
}

# Function to run tests in container
run_tests_in_container() {
    local test_path="$1"
    local test_args="$2"
    
    print_status "Running tests in container for: $test_path"
    
    # Create a temporary test container
    docker run --rm \
        --network smart-payment-network \
        -v "$(pwd):/app" \
        -w /app \
        -e TEST_DB_HOST=postgres \
        -e TEST_DB_PORT=5432 \
        -e TEST_DB_USER=user \
        -e TEST_DB_PASSWORD=password \
        -e TEST_DB_NAME=smart_payment \
        -e TEST_DB_SSLMODE=disable \
        -e TEST_REDIS_HOST=redis \
        -e TEST_REDIS_PORT=6379 \
        golang:1.23-alpine \
        sh -c "
            apk add --no-cache git postgresql-client &&
            go mod download &&
            go test -v $test_path $test_args
        "
}

# Function to run all tests
run_all_tests() {
    print_status "Running all tests in container..."
    
    # Run repository tests
    print_status "Running repository tests..."
    run_tests_in_container "./internal/repository/..." "-count=1"
    
    # Run integration tests
    print_status "Running integration tests..."
    run_tests_in_container "./test/integration/..." "-count=1"
    
    # Run service tests
    print_status "Running service tests..."
    run_tests_in_container "./internal/services/..." "-count=1"
    
    print_success "All tests completed!"
}

# Function to run specific test
run_specific_test() {
    local test_path="$1"
    local test_args="$2"
    
    if [ -z "$test_path" ]; then
        print_error "Please specify a test path"
        exit 1
    fi
    
    run_tests_in_container "$test_path" "$test_args"
}

# Function to show help
show_help() {
    echo "Test Runner Script"
    echo ""
    echo "Usage: $0 [COMMAND] [TEST_PATH] [TEST_ARGS]"
    echo ""
    echo "Commands:"
    echo "  all                    - Run all tests"
    echo "  specific <test_path>   - Run specific test(s)"
    echo "  help                   - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 all                                    # Run all tests"
    echo "  $0 specific ./internal/repository/...     # Run repository tests"
    echo "  $0 specific ./test/integration/...        # Run integration tests"
    echo "  $0 specific ./internal/services/...       # Run service tests"
    echo ""
    echo "Note: Tests run inside Docker container to access local databases"
}

# Main script logic
main() {
    case "${1:-help}" in
        all)
            check_docker
            check_databases
            run_all_tests
            ;;
        specific)
            check_docker
            check_databases
            run_specific_test "$2" "$3"
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
