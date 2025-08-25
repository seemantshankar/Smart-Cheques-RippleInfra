# Smart Payment Infrastructure

Enterprise B2B payment platform built on Ripple XRPL that eliminates trust deficits in cross-border transactions through AI-powered compliance monitoring, smart contract automation, and programmable payment instruments (Smart Cheques).

## Architecture

The platform consists of multiple microservices:

- **API Gateway** (`:8000`) - Main entry point for external requests
- **Identity Service** (`:8001`) - Enterprise KYB/KYC and user management
- **Orchestration Service** (`:8002`) - Contract analysis and milestone tracking
- **XRPL Service** (`:8003`) - XRPL blockchain integration and escrow management
- **Asset Gateway** (`:8004`) - Multi-asset bridging and treasury management

## Project Structure

```
├── cmd/                    # Main applications
│   ├── api-gateway/        # API Gateway service
│   ├── identity-service/   # Identity and access management
│   ├── orchestration-service/ # Contract orchestration
│   ├── xrpl-service/       # XRPL integration
│   └── asset-gateway/      # Asset management
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   └── models/            # Data models
├── pkg/                   # Public library code
│   ├── database/          # Database utilities
│   └── xrpl/             # XRPL client utilities
├── api/                   # API definitions
│   └── openapi.yaml       # OpenAPI specification
├── deployments/           # Deployment configurations
│   ├── docker/           # Docker files
│   └── sql/              # Database schemas
└── docker-compose.yml     # Local development environment
```

## Quick Start

### Prerequisites

- Go 1.21+ ✅ (Installed: Go 1.25.0)
- Docker and Docker Compose ✅ (Installed: Docker 28.3.2, Compose v2.39.1)
- Git ✅

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd smart-payment-infrastructure
   ```

2. **Start the development environment**
   ```bash
   docker-compose up -d
   ```

3. **Verify services are running**
   ```bash
   # Check all services
   docker-compose ps
   
   # Test health endpoints
   curl http://localhost:8000/health  # API Gateway
   curl http://localhost:8001/health  # Identity Service
   curl http://localhost:8002/health  # Orchestration Service
   curl http://localhost:8003/health  # XRPL Service
   curl http://localhost:8004/health  # Asset Gateway
   ```

4. **View logs**
   ```bash
   docker-compose logs -f [service-name]
   ```

### Development Commands

Use these Make commands for common development tasks:

```bash
# Build all Go binaries locally
make build

# Run tests
make test

# Download/update dependencies
make deps

# Start development environment with Docker
make dev-up

# Check service health
make health
```

### Messaging System

The platform includes a Redis-based messaging system for event-driven communication:

```bash
# Check messaging health
curl http://localhost:8000/health/messaging
curl http://localhost:8001/health/messaging
curl http://localhost:8002/health/messaging

# Monitor queue statistics
curl http://localhost:8000/admin/queue-stats

# Test event publishing
curl -X POST http://localhost:8000/admin/test-event \
  -H "Content-Type: application/json" \
  -d '{"event_type": "test.message", "data": {"message": "Hello!"}}'
```

**Event Types:**
- `enterprise.registered` - New enterprise registration
- `smart_cheque.created` - Smart Cheque creation
- `milestone.completed` - Milestone completion
- `payment.released` - Payment release
- `dispute.created` - Dispute initiation

### Development Workflow

1. **Make code changes** in the respective service directories
2. **Rebuild specific service**
   ```bash
   docker-compose build [service-name]
   docker-compose up -d [service-name]
   ```
3. **Build and test locally** (optional)
   ```bash
   make build  # Build Go binaries
   make test   # Run tests
   ```

### Database Setup

After starting the development environment, initialize the databases:

```bash
# Run database migrations
make db-migrate

# Seed development data
make db-seed

# Check migration version
make db-version
```

### Database Access

- **PostgreSQL**: `localhost:5432`
  - Database: `smart_payment`
  - Username: `user`
  - Password: `password`

- **MongoDB**: `localhost:27017`
  - Database: `smart_payment`
  - Username: `admin`
  - Password: `password`

- **Redis**: `localhost:6379`

### Database Management

```bash
# Migration commands
make db-migrate    # Run migrations
make db-rollback   # Rollback migrations
make db-version    # Check current version

# Data management
make db-seed       # Seed development data
make db-clear      # Clear all data
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_URL` | PostgreSQL connection string | `postgres://user:password@localhost:5432/smart_payment?sslmode=disable` |
| `MONGO_URL` | MongoDB connection string | `mongodb://admin:password@localhost:27017` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379` |
| `XRPL_NETWORK_URL` | XRPL network WebSocket URL | `wss://s.altnet.rippletest.net:51233` |
| `XRPL_TESTNET` | Use XRPL testnet | `true` |
| `ENV` | Environment (development/production) | `development` |

## API Documentation

The API is documented using OpenAPI 3.0 specification. View the documentation at:
- Local: `http://localhost:8000/docs` (when implemented)
- File: `api/openapi.yaml`

## Next Steps

This is the initial project structure. The next tasks will involve:

1. Setting up local development databases (Task 1.2)
2. Implementing basic message queuing infrastructure (Task 1.3)
3. Building XRPL integration foundation (Task 2.1)
4. Creating enterprise identity and access management (Task 4.1)

## Contributing

1. Follow Go best practices and conventions
2. Write tests for all new functionality
3. Update documentation when adding new features
4. Use conventional commit messages

## License

[License information to be added]