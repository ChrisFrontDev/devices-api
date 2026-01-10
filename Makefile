.PHONY: run build test lint proto clean help

# Variables
APP_NAME := devices-api
CMD_PATH := ./cmd/api/main.go
BIN_DIR := ./bin
BUILD_OUTPUT := $(BIN_DIR)/$(APP_NAME)

# Targets

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Run the application
	go run $(CMD_PATH)

build: ## Build the application binary
	mkdir -p $(BIN_DIR)
	go build -o $(BUILD_OUTPUT) $(CMD_PATH)

test: ## Run tests
	go test -v -race ./...

lint: ## Run linters
	golangci-lint run

proto: ## Generate protobuf files using Buf
	buf generate

clean: ## Clean build artifacts
	rm -rf $(BIN_DIR)
	rm -rf pkg/pb/*.go

docker-build: ## Build Docker image
	docker build -t $(APP_NAME) .

docker-run: ## Run Docker container
	docker run -p 8080:8080 -p 9090:9090 $(APP_NAME)
