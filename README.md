# go dist scheduler

## Setup

### Prerequisites

- Go 1.25.5 or later
- Docker and Docker Compose

### Environment Configuration

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` to configure your local environment if needed. The default values work with the Docker Compose setup.
   
   **Important:** `DB_PASSWORD` must be set in the `.env` file for database connections to work. For development, the default value in `.env.example` is acceptable, but use a strong password in production.

### Running with Docker Compose

Start PostgreSQL and Redis services:

```bash
docker compose up -d
```

Stop services:

```bash
docker compose down
```

Stop services and remove volumes:

```bash
docker compose down -v
```

### Database Migrations

After starting the PostgreSQL service with Docker Compose, run the database migrations:

```bash
make migrate-up
```

To rollback the last migration:

```bash
make migrate-down
```

To check the current migration version:

```bash
make migrate-version
```

To create a new migration:

```bash
make migrate-create name=your_migration_name
```

See `make help` for all available migration commands.

### Configuration

The application configuration is managed through environment variables. See `.env.example` for available options.

The `internal/config` package provides configuration loading:

```go
import "github.com/m11o/go-dist-scheduler/internal/config"

cfg, err := config.Load()
if err != nil {
    // handle error
}

// Use configuration
dbDSN := cfg.Database.DSN()
redisAddr := cfg.Redis.Addr()
```

## Development

### Building and Running

Build the application:

```bash
make build
```

Run the application:

```bash
make run
```

Or use the Makefile commands:

```bash
make help
```

### Linting

To run the linter locally, use the following command:

```bash
make lint
```

Or directly:

```bash
go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...
```

### Testing

Run all tests:

```bash
make test
```

Or directly:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```
