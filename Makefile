.PHONY: help migrate migrate-dry lint test build run

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

migrate: ## Run database migrations
	@echo "Running database migrations..."
	@if ! command -v psqldef > /dev/null 2>&1; then \
		echo "psqldef is not installed. Installing..."; \
		go install github.com/sqldef/sqldef/cmd/psqldef@latest; \
	fi
	@PGPASSWORD=$(DB_PASSWORD) PATH="$(HOME)/go/bin:$(PATH)" psqldef -U $(DB_USER) -p $(DB_PORT) -h $(DB_HOST) $(DB_NAME) < db/migrations/schema.sql

migrate-dry: ## Run database migrations in dry-run mode
	@echo "Running database migrations (dry-run)..."
	@if ! command -v psqldef > /dev/null 2>&1; then \
		echo "psqldef is not installed. Installing..."; \
		go install github.com/sqldef/sqldef/cmd/psqldef@latest; \
	fi
	@PGPASSWORD=$(DB_PASSWORD) PATH="$(HOME)/go/bin:$(PATH)" psqldef -U $(DB_USER) -p $(DB_PORT) -h $(DB_HOST) $(DB_NAME) --dry-run < db/migrations/schema.sql

lint: ## Run golangci-lint
	@echo "Running linter..."
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...

test: ## Run tests
	@echo "Running tests..."
	@go test ./... -v

build: ## Build the scheduler binary
	@echo "Building scheduler..."
	@go build -o bin/scheduler ./cmd/scheduler

run: ## Run the scheduler
	@echo "Running scheduler..."
	@go run ./cmd/scheduler/main.go
