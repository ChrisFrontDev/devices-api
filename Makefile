.PHONY: run build test lint proto clean help docker-build docker-run docker-up docker-down docker-logs db-up db-down migrate-up migrate-down

# Variables
APP_NAME := devices-api
CMD_PATH := ./cmd/api/main.go
BIN_DIR := ./bin
BUILD_OUTPUT := $(BIN_DIR)/$(APP_NAME)

# Database configuration (uses same variables as docker-compose)
# Override via environment variables if needed
POSTGRES_HOST ?= localhost
POSTGRES_PORT ?= 5432
POSTGRES_USER ?= user
POSTGRES_DB ?= devices
# POSTGRES_PASSWORD must be set via environment for security

# Targets

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''
	@echo 'Database configuration (via environment variables):'
	@echo '  POSTGRES_HOST     (default: localhost)'
	@echo '  POSTGRES_PORT     (default: 5432)'
	@echo '  POSTGRES_USER     (default: user)'
	@echo '  POSTGRES_PASSWORD (required for migrations)'
	@echo '  POSTGRES_DB       (default: devices)'

run: ## Run the application locally
	go run $(CMD_PATH)

build: ## Build the application binary
	@mkdir -p $(BIN_DIR)
	go build -o $(BUILD_OUTPUT) $(CMD_PATH)
	@echo "Binary built: $(BUILD_OUTPUT)"

test: ## Run all tests with race detector
	go test -v -race ./...

test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linters
	golangci-lint run

proto: ## Generate protobuf files using Buf
	buf generate

clean: ## Clean build artifacts
	rm -rf $(BIN_DIR)
	rm -rf pkg/pb/*.go
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -t $(APP_NAME) .

docker-run: ## Run Docker container
	docker run -p 8080:8080 -p 9090:9090 $(APP_NAME)

docker-up: ## Start all services with docker-compose
	docker-compose up -d

docker-down: ## Stop all services with docker-compose
	docker-compose down

docker-logs: ## Show logs from docker-compose
	docker-compose logs -f

db-up: ## Start PostgreSQL database only
	docker-compose up -d postgres

db-down: ## Stop PostgreSQL database only
	docker-compose stop postgres

migrate-up: ## Run database migrations up (requires POSTGRES_PASSWORD env var)
ifndef POSTGRES_PASSWORD
	@echo "Error: POSTGRES_PASSWORD environment variable is required"
	@echo "Usage: export POSTGRES_PASSWORD=yourpassword && make migrate-up"
	@exit 1
endif
	@echo "Applying migrations to $(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)..."
	@if [ ! -d "migrations" ]; then \
		echo "Error: migrations directory not found"; \
		exit 1; \
	fi
	@for file in migrations/*.up.sql; do \
		if [ -f "$$file" ]; then \
			echo "Running $$file..."; \
			PGPASSWORD=$(POSTGRES_PASSWORD) psql -h $(POSTGRES_HOST) -p $(POSTGRES_PORT) -U $(POSTGRES_USER) -d $(POSTGRES_DB) -f $$file || exit 1; \
		fi \
	done
	@echo "✓ Migrations completed successfully!"

migrate-down: ## Run database migrations down (requires POSTGRES_PASSWORD env var)
ifndef POSTGRES_PASSWORD
	@echo "Error: POSTGRES_PASSWORD environment variable is required"
	@echo "Usage: export POSTGRES_PASSWORD=yourpassword && make migrate-down"
	@exit 1
endif
	@echo "Reverting migrations from $(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)..."
	@if [ ! -d "migrations" ]; then \
		echo "Error: migrations directory not found"; \
		exit 1; \
	fi
	@for file in $$(ls migrations/*.down.sql 2>/dev/null | sort -r); do \
		if [ -f "$$file" ]; then \
			echo "Running $$file..."; \
			PGPASSWORD=$(POSTGRES_PASSWORD) psql -h $(POSTGRES_HOST) -p $(POSTGRES_PORT) -U $(POSTGRES_USER) -d $(POSTGRES_DB) -f $$file || exit 1; \
		fi \
	done
	@echo "✓ Migrations reverted successfully!"
