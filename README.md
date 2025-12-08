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

### Linting

To run the linter locally, use the following command:

```bash
go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...
```

### Testing

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```
