.PHONY: help migrate-up migrate-down migrate-create migrate-force migrate-version build run test lint

# デフォルトターゲット
help:
	@echo "Available commands:"
	@echo "  make migrate-up       - Run all pending migrations (in Docker)"
	@echo "  make migrate-down     - Rollback the last migration (in Docker)"
	@echo "  make migrate-create   - Create a new migration file (usage: make migrate-create name=your_migration_name) (in Docker)"
	@echo "  make migrate-force    - Force set migration version (usage: make migrate-force version=1) (in Docker)"
	@echo "  make migrate-version  - Show current migration version (in Docker)"
	@echo "  make build           - Build the application"
	@echo "  make run             - Run the application"
	@echo "  make test            - Run tests"
	@echo "  make lint            - Run linter"

# マイグレーションを実行（Dockerコンテナ内で実行）
migrate-up:
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Please create .env file with DB_PASSWORD."; \
		echo "  cp .env.example .env"; \
		exit 1; \
	fi
	@echo "Running migrations in Docker..."
	@docker compose run --rm migrate -action up

# 最後のマイグレーションをロールバック（Dockerコンテナ内で実行）
migrate-down:
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Please create .env file with DB_PASSWORD."; \
		echo "  cp .env.example .env"; \
		exit 1; \
	fi
	@echo "Rolling back last migration in Docker..."
	@docker compose run --rm migrate -action down

# 新しいマイグレーションファイルを作成（Dockerコンテナ内で実行）
# タイムスタンプベース（YYYYMMDDhhmmss）でファイルが作成されます
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: migration name is required. Usage: make migrate-create name=your_migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration files in Docker for $(name)..."
	@docker compose run --rm migrate -action create -name $(name)

# マイグレーションバージョンを強制設定（Dockerコンテナ内で実行）
migrate-force:
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Please create .env file with DB_PASSWORD."; \
		echo "  cp .env.example .env"; \
		exit 1; \
	fi
	@if [ -z "$(version)" ]; then \
		echo "Error: version is required. Usage: make migrate-force version=1"; \
		exit 1; \
	fi
	@echo "Forcing migration version in Docker to $(version)..."
	@docker compose run --rm migrate -action force -version $(version)

# 現在のマイグレーションバージョンを表示（Dockerコンテナ内で実行）
migrate-version:
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Please create .env file with DB_PASSWORD."; \
		echo "  cp .env.example .env"; \
		exit 1; \
	fi
	@echo "Current migration version (from Docker):"
	@docker compose run --rm migrate -action version

# アプリケーションをビルド
build:
	@echo "Building application..."
	@go build -o scheduler ./cmd/scheduler

# アプリケーションを実行
run:
	@echo "Running application..."
	@go run ./cmd/scheduler/main.go

# テストを実行
test:
	@echo "Running tests..."
	@go test ./...

# リンターを実行
lint:
	@echo "Running linter..."
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...
