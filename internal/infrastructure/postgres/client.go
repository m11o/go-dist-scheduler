package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// NewClient creates a new PostgreSQL database connection pool.
func NewClient(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
