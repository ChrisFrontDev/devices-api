# Devices API

A production-ready REST API for managing devices, built with Go following Clean Architecture and Domain-Driven Design principles.

## Features

- **CRUD Operations** - Create, read, update, and delete devices
- **Device States** - Active, In-Use, and Inactive state management
- **Business Rules** - Devices in-use cannot change name/brand
- **Filtering** - List devices by brand or state
- **Pagination** - Efficient data retrieval with limit/offset
- **Swagger/OpenAPI** - Interactive API documentation at `/swagger/index.html`
- **PostgreSQL** - Production-grade database with connection pooling
- **Docker Ready** - Containerized with distroless images for security
- **CI/CD Pipeline** - Automated testing and security scanning
- **Health Checks** - Built-in endpoint for monitoring
- **Integration Tests** - 43 tests with real PostgreSQL via testcontainers

## Quick Start

### Prerequisites

**Required:**
- Go 1.23+
- Docker & Docker Compose
- Make (for Makefile commands)

**Optional (for specific tasks):**
- PostgreSQL 16+ (only if running without Docker)
- [golangci-lint](https://golangci-lint.run/usage/install/) (for linting: `make lint`)
- [swag](https://github.com/swaggo/swag) (for Swagger docs: `make swagger`)
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) (for formatting: `make fmt`)
  ```bash
  go install golang.org/x/tools/cmd/goimports@latest
  ```
- `jq` (for parsing JSON in API examples)
  ```bash
  # macOS
  brew install jq

  # Linux
  sudo apt-get install jq
  ```

**Verify your setup:**
```bash
# Check required tools
go version          # Should be 1.23+
docker --version    # Docker installed
docker-compose --version  # Docker Compose installed
make --version      # Make installed

# Check optional tools (if needed)
golangci-lint --version   # For linting
swag --version            # For Swagger docs
goimports --version       # For formatting
jq --version              # For JSON parsing
```

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

# Load environment variables (required for migrations)
set -a; source .env; set +a

# Run database migrations
make migrate-up

# Check health
curl http://localhost:8080/health

# View API documentation
open http://localhost:8080/swagger/index.html
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
│   ├── service/          # Business logic (+ unit tests)
│   ├── repository/       # Data access layer (+ integration tests)
│   ├── testhelper/       # Test utilities (testcontainers)
│   └── handler/
│       └── http/         # HTTP handlers (+ integration tests)
├── pkg/
│   ├── database/         # Database utilities
│   └── pb/               # Protocol buffers (future gRPC)
├── migrations/           # Database migrations
├── Dockerfile            # Container image definition
├── docker-compose.yml    # Local development setup
└── Makefile             # Development automation
```

## API Endpoints

### Interactive Documentation

**Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

The Swagger UI provides:
- Interactive API testing
- Request/response examples
- Schema definitions
- Try-it-out functionality

### REST Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/swagger/*` | Swagger UI documentation |
| `POST` | `/api/v1/devices` | Create device |
| `GET` | `/api/v1/devices` | List all devices |
| `GET` | `/api/v1/devices?brand=Apple` | Filter by brand |
| `GET` | `/api/v1/devices?state=active` | Filter by state |
| `GET` | `/api/v1/devices/{id}` | Get device by ID |
| `PUT` | `/api/v1/devices/{id}` | Full update |
| `PATCH` | `/api/v1/devices/{id}` | Partial update |
| `DELETE` | `/api/v1/devices/{id}` | Delete device |

## Development

### Swagger Documentation

This project uses [swaggo/swag](https://github.com/swaggo/swag) for OpenAPI documentation.

**Generate/Update Swagger docs:**

```bash
# Install swag CLI (if not already installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs
swag init -g cmd/api/main.go -o ./docs --parseDependency --parseInternal

# Or use the Makefile command
make swagger
```

**Add Swagger annotations:**

```go
// @Summary Create a new device
// @Description Create a new device with name and brand
// @Tags devices
// @Accept json
// @Produce json
// @Param device body dto.CreateDeviceRequest true "Device data"
// @Success 201 {object} dto.DeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /devices [post]
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
    // ...
}
```

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
make help                       # Show all available commands
make test                       # Run all tests (unit + integration)
make test-unit                  # Run unit tests only (fast)
make test-integration           # Run integration tests (requires Docker)
make test-coverage              # Generate coverage report
make test-integration-coverage  # Integration coverage report
make lint                       # Run linters (11 enabled: bugs, security, quality)
make fmt                        # Format code (gofmt + goimports)
make fmt-check                  # Check code formatting
make docker-up                  # Start all services
make docker-down                # Stop all services
make db-up                      # Start only PostgreSQL
make migrate-up                 # Run migrations
make migrate-down               # Rollback migrations
```

### Running Tests

This project has both **unit tests** and **integration tests** with real PostgreSQL databases.

```bash
# Run all tests (unit + integration)
make test

# Run only unit tests (fast)
make test-unit

# Run only integration tests (requires Docker)
make test-integration

# Generate coverage report
make test-coverage
```

#### Integration Tests

Integration tests use [testcontainers-go](https://golang.testcontainers.org/) to spin up real PostgreSQL containers automatically.

**Requirements:**
- Docker must be running
- No manual database setup needed

**What's tested:**
- ✅ **43 integration tests** covering all features
- ✅ Real database operations (CRUD, filtering, pagination)
- ✅ All REST API endpoints end-to-end
- ✅ Business rules enforcement
- ✅ Error handling and edge cases

**Test execution:**
```bash
# Integration tests only (~7 seconds)
make test-integration

# With coverage report
make test-integration-coverage
open coverage-integration.html
```

Testcontainers automatically handles:
- Starting PostgreSQL container
- Running migrations
- Cleaning up after tests
- No ports conflicts or manual cleanup needed

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

## Testing Strategy

### Test Coverage

- **Unit Tests**: 30 tests for business logic (mocked dependencies)
- **Integration Tests**: 43 tests with real PostgreSQL database
  - 22 repository tests (database operations)
  - 21 HTTP handler tests (end-to-end API)
- **Total**: 73 tests with ~92% code coverage

### Running Tests Locally

```bash
# Quick unit tests during development
make test-unit              # ~1 second

# Full integration tests before push
make test-integration       # ~7 seconds (includes Docker startup)

# All tests for CI/CD validation
make test                   # ~13 seconds
```

## CI/CD

GitHub Actions pipeline includes:

- **Linting** - golangci-lint
- **Testing** - Unit + Integration tests with coverage
- **Security Scanning** - Trivy (containers) + Gosec (code)
- **Docker Build** - Multi-stage builds with caching
- **Dependency Review** - Blocks vulnerable dependencies

## Contributing

1. Follow **Test-Driven Development (TDD)** - Write tests first
2. Use **Conventional Commits** - `feat:`, `fix:`, `chore:`, etc.
3. Keep commits **granular** - One logical change per commit
4. Format code before committing - `make fmt`
5. Run tests before committing - `make test`
6. Ensure linting passes - `make lint`


## Future Improvements

- [ ] gRPC support (protocol buffers already in `pkg/pb/`)
- [ ] Structured logging (zerolog/zap)
- [ ] Metrics and monitoring (Prometheus/Grafana)
- [ ] Kubernetes deployment
- [ ] Device history and audit logs
- [ ] Bulk operations (batch create/update/delete)
- [ ] Load testing setup

## Support

For issues, questions, or contributions, please open an issue or pull request.
