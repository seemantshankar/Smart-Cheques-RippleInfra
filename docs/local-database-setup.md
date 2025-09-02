# Local Database Setup

This document explains how to set up and use local databases for development instead of Supabase.

## Overview

The Smart Cheques Ripple Infrastructure is configured to use local databases by default:
- **PostgreSQL**: Local instance for relational data
- **MongoDB**: Local instance for document storage
- **Redis**: Local instance for caching and message queuing

## Prerequisites

1. **Docker**: Make sure Docker is installed and running
2. **Docker Compose**: Should be available (either standalone or as `docker compose`)
3. **Go**: For building and running the application
4. **PostgreSQL Client** (optional): For direct database access

## Quick Start

### 1. Start Local Databases

```bash
# Start only the databases (PostgreSQL, MongoDB, Redis)
make db-local-start

# Or use the script directly
./scripts/local-db.sh start
```

### 2. Run Database Migrations

```bash
# Apply all database migrations
make db-local-migrate

# Or use the script directly
./scripts/local-db.sh migrate
```

### 3. Check Status

```bash
# Check if databases are running
make db-local-status

# Or use the script directly
./scripts/local-db.sh status
```

## Available Commands

### Makefile Targets

```bash
make local-dev          # Quick setup: start databases
make db-local-start     # Start local databases
make db-local-stop      # Stop local databases
make db-local-restart   # Restart local databases
make db-local-reset     # Reset databases (fresh start)
make db-local-migrate   # Run migrations
make db-local-status    # Check status
make db-local-logs      # Show logs
make db-local-connect   # Connect to PostgreSQL
```

### Direct Script Usage

```bash
./scripts/local-db.sh start      # Start databases
./scripts/local-db.sh stop       # Stop databases
./scripts/local-db.sh restart    # Restart databases
./scripts/local-db.sh reset      # Reset databases
./scripts/local-db.sh migrate    # Run migrations
./scripts/local-db.sh status     # Check status
./scripts/local-db.sh logs       # Show logs
./scripts/local-db.sh connect    # Connect to PostgreSQL
./scripts/local-db.sh help       # Show help
```

## Database Configuration

### Connection Details

- **PostgreSQL**:
  - Host: `localhost`
  - Port: `5432`
  - Database: `smart_payment`
  - Username: `user`
  - Password: `password`
  - Connection String: `postgres://user:password@localhost:5432/smart_payment?sslmode=disable`

- **MongoDB**:
  - Host: `localhost`
  - Port: `27017`
  - Database: `smart_payment`
  - Username: `admin`
  - Password: `password`
  - Connection String: `mongodb://admin:password@localhost:27017/smart_payment`

- **Redis**:
  - Host: `localhost`
  - Port: `6379`
  - No authentication required

### Environment Variables

The system automatically uses local database configurations. You can override these by setting environment variables:

```bash
# Load local environment configuration
source env.local

# Or set individual variables
export DATABASE_URL="postgres://user:password@localhost:5432/smart_payment?sslmode=disable"
export MONGO_URL="mongodb://admin:password@localhost:27017/smart_payment"
export REDIS_URL="localhost:6379"
```

## Development Workflow

### 1. Start Development Environment

```bash
# Start databases and run migrations
make local-dev
make db-local-migrate
```

### 2. Run Your Application

```bash
# Build and run specific services
make build
./bin/api-gateway
./bin/identity-service
# ... etc
```

### 3. Monitor and Debug

```bash
# Check database status
make db-local-status

# View database logs
make db-local-logs

# Connect to database for debugging
make db-local-connect
```

### 4. Reset When Needed

```bash
# Reset databases for fresh start
make db-local-reset
make db-local-migrate
```

## Troubleshooting

### Common Issues

1. **Port Already in Use**:
   ```bash
   # Check what's using the ports
   lsof -i :5432  # PostgreSQL
   lsof -i :27017 # MongoDB
   lsof -i :6379  # Redis
   ```

2. **Docker Not Running**:
   ```bash
   # Start Docker Desktop or Docker daemon
   # Then try again
   make db-local-start
   ```

3. **Migration Errors**:
   ```bash
   # Check database status first
   make db-local-status
   
   # If databases are running, try resetting
   make db-local-reset
   make db-local-migrate
   ```

4. **Permission Issues**:
   ```bash
   # Make sure the script is executable
   chmod +x scripts/local-db.sh
   ```

### Database Reset

If you encounter persistent issues, you can completely reset the databases:

```bash
# This will delete all data and start fresh
make db-local-reset
make db-local-migrate
```

## Production vs Development

- **Development**: Uses local databases with the configuration in this document
- **Production**: Will use Supabase or other cloud database services

The application automatically detects the environment and uses appropriate configurations. For local development, it defaults to local database instances.

## Next Steps

Once your local databases are running:

1. Run migrations: `make db-local-migrate`
2. Start your services: `make build && ./bin/api-gateway`
3. Test your endpoints
4. Use `make db-local-connect` to inspect data directly

## Support

If you encounter issues:

1. Check the troubleshooting section above
2. Verify Docker is running: `docker info`
3. Check database logs: `make db-local-logs`
4. Ensure ports are available: `lsof -i :5432`
