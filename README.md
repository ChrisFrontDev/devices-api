# Devices API

A production-ready REST API for managing devices, built with Go following Clean Architecture and Domain-Driven Design principles.

## Features

- **CRUD Operations** - Create, read, update, and delete devices
- **Device States** - Active, In-Use, and Inactive state management
- **Business Rules** - Devices in-use cannot change name/brand
- **Filtering** - List devices by brand or state
- **Pagination** - Efficient data retrieval with limit/offset
- **PostgreSQL** - Production-grade database with connection pooling
- **Docker Ready** - Containerized with distroless images for security
- **CI/CD Pipeline** - Automated testing and security scanning
- **Health Checks** - Built-in endpoint for monitoring

## Quick Start

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 16+ (if running locally)

### 1. Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd devices-api

# Copy environment template
cp env.sample .env

# Update .env with your configuration
# At minimum, change POSTGRES_PASSWORD
```

### 2. Start with Docker (Recommended)

```bash
# Start all services (API + PostgreSQL)
docker-compose up -d

# Run database migrations
make migrate-up

# Check health
curl http://localhost:8080/health
```

### 3. Test the API

```bash
# Create a device
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPhone 15",
    "brand": "Apple",
    "state": "active"
  }'

# List devices
curl http://localhost:8080/api/v1/devices

# Get device by ID
curl http://localhost:8080/api/v1/devices/{id}

# Update device
curl -X PUT http://localhost:8080/api/v1/devices/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPhone 15 Pro",
    "brand": "Apple",
    "state": "in-use"
  }'

# Delete device
curl -X DELETE http://localhost:8080/api/v1/devices/{id}
```

## Project Structure

```
devices-api/
├── cmd/
│   └── api/              # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── domain/           # Business entities and rules
│   ├── service/          # Business logic
│   ├── repository/       # Data access layer
│   └── handler/
│       └── http/         # HTTP handlers (REST API)
├── pkg/
│   ├── database/         # Database utilities
│   └── pb/               # Protocol buffers (future gRPC)
├── migrations/           # Database migrations
├── Dockerfile            # Container image definition
├── docker-compose.yml    # Local development setup
└── Makefile             # Development automation
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/api/v1/devices` | Create device |
| `GET` | `/api/v1/devices` | List all devices |
| `GET` | `/api/v1/devices?brand=Apple` | Filter by brand |
| `GET` | `/api/v1/devices?state=active` | Filter by state |
| `GET` | `/api/v1/devices/{id}` | Get device by ID |
| `PUT` | `/api/v1/devices/{id}` | Full update |
| `PATCH` | `/api/v1/devices/{id}` | Partial update |
| `DELETE` | `/api/v1/devices/{id}` | Delete device |

## Development

### Local Development (without Docker)

```bash
# Load environment variables
set -a; source .env; set +a

# Install dependencies
go mod download

# Run tests
make test

# Run with coverage
make test-coverage

# Run the application
go run cmd/api/main.go
```

### Make Commands

```bash
make help              # Show all available commands
make test              # Run tests
make test-coverage     # Run tests with coverage report
make lint              # Run linter (golangci-lint)
make docker-up         # Start all services
make docker-down       # Stop all services
make db-up             # Start only PostgreSQL
make migrate-up        # Run migrations
make migrate-down      # Rollback migrations
```

### Running Tests

```bash
# Unit tests
go test ./internal/...

# With coverage
go test -v -race -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out

# Or use Makefile
make test-coverage
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HTTP_PORT` | HTTP server port | `8080` |
| `SERVER_GRPC_PORT` | gRPC server port (future) | `9090` |
| `DATABASE_URL` | PostgreSQL connection string | **required** |
| `POSTGRES_HOST` | Database host | `localhost` |
| `POSTGRES_PORT` | Database port | `5432` |
| `POSTGRES_USER` | Database user | `user` |
| `POSTGRES_PASSWORD` | Database password | **required** |
| `POSTGRES_DB` | Database name | `devices` |

See `env.sample` for complete configuration examples.

## Business Rules

1. **Device States**: Only `active`, `in-use`, or `inactive` are valid
2. **Update Restrictions**: Devices in `in-use` state cannot change name or brand
3. **State Transitions**: State changes are always allowed, regardless of current state
4. **Validation**: All fields (name, brand, state) are required

## Architecture

This project follows **Clean Architecture** principles:

- **Domain Layer** (`internal/domain/`) - Business entities, rules, and interfaces
- **Service Layer** (`internal/service/`) - Application business logic
- **Repository Layer** (`internal/repository/`) - Data persistence
- **Handler Layer** (`internal/handler/`) - API endpoints and request/response handling

### Design Principles

- **DRY** - Don't Repeat Yourself
- **KISS** - Keep It Simple, Stupid
- **YAGNI** - You Aren't Gonna Need It
- **TDA** - Tell, Don't Ask
- **Dependency Inversion** - High-level modules don't depend on low-level modules

## Security

### Best Practices

- ✅ Distroless Docker images (minimal attack surface)
- ✅ Non-root user in containers (UID 65532)
- ✅ Parameterized SQL queries (SQL injection protection)
- ✅ Environment-based secrets management
- ✅ Security scanning in CI (Trivy + Gosec)
- ✅ Dependency vulnerability checks

### Important Notes

- **Never commit `.env` files** - They contain sensitive credentials
- **Use `sslmode=require`** in production - Local dev uses `sslmode=disable`
- **Rotate passwords** regularly and use strong, unique passwords
- **Review** `SECURITY.md` for detailed security guidelines

## CI/CD

GitHub Actions pipeline includes:

- **Linting** - golangci-lint
- **Testing** - Unit tests with coverage
- **Security Scanning** - Trivy (containers) + Gosec (code)
- **Docker Build** - Multi-stage builds with caching
- **Dependency Review** - Blocks vulnerable dependencies

## Contributing

1. Follow **Test-Driven Development (TDD)** - Write tests first
2. Use **Conventional Commits** - `feat:`, `fix:`, `chore:`, etc.
3. Keep commits **granular** - One logical change per commit
4. Run tests before committing - `make test`
5. Ensure linting passes - `make lint`


## Support

For issues, questions, or contributions, please open an issue or pull request.
