.PHONY: help migrate-up migrate-down migrate-create migrate-force migrate-version build run test lint

# デフォルトターゲット
help:
	@echo "Available commands:"
	@echo "  make migrate-up       - Run all pending migrations"
	@echo "  make migrate-down     - Rollback the last migration"
	@echo "  make migrate-create   - Create a new migration file (usage: make migrate-create name=your_migration_name)"
	@echo "  make migrate-force    - Force set migration version (usage: make migrate-force version=1)"
	@echo "  make migrate-version  - Show current migration version"
	@echo "  make build           - Build the application"
	@echo "  make run             - Run the application"
	@echo "  make test            - Run tests"
	@echo "  make lint            - Run linter"

# データベース接続情報
# 環境変数から読み込まれます。設定されていない場合はデフォルト値を使用します。
# セキュリティ上、本番環境では環境変数を使用してください。
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= scheduler
DB_PASSWORD ?= password
DB_NAME ?= scheduler
DB_SSLMODE ?= disable
DATABASE_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# マイグレーションディレクトリ
MIGRATIONS_DIR := db/migrations

# マイグレーションを実行
migrate-up:
	@echo "Running migrations..."
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

# 最後のマイグレーションをロールバック
migrate-down:
	@echo "Rolling back last migration..."
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

# 新しいマイグレーションファイルを作成
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: migration name is required. Usage: make migrate-create name=your_migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration files for $(name)..."
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# マイグレーションバージョンを強制設定
migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "Error: version is required. Usage: make migrate-force version=1"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(version)..."
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(version)

# 現在のマイグレーションバージョンを表示
migrate-version:
	@echo "Current migration version:"
	@go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

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
