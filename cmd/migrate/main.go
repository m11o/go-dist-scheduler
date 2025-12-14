package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/yourname/go-dist-scheduler/internal/config"
)

func main() {
	var (
		action  string
		version int
		name    string
	)

	flag.StringVar(&action, "action", "", "Migration action: up, down, version, force, create")
	flag.IntVar(&version, "version", 0, "Migration version (for force command)")
	flag.StringVar(&name, "name", "", "Migration name (for create command)")
	flag.Parse()

	if action == "" {
		log.Fatal("Error: -action flag is required (up, down, version, force, create)")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Build database URL
	// For migrations inside Docker, we use "postgres" as the host (service name in docker-compose)
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "postgres" // Default to docker-compose service name
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		dbHost,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	migrationsPath := "file://db/migrations"

	switch action {
	case "create":
		if name == "" {
			log.Fatal("Error: -name flag is required for create action")
		}
		if err := createMigration(name); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}

	case "up":
		if err := runMigration(dbURL, migrationsPath, "up", 0); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := runMigration(dbURL, migrationsPath, "down", 1); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("Migration rolled back successfully")

	case "version":
		if err := showVersion(dbURL, migrationsPath); err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}

	case "force":
		if version == 0 {
			log.Fatal("Error: -version flag is required for force action")
		}
		if err := runMigration(dbURL, migrationsPath, "force", version); err != nil {
			log.Fatalf("Force migration failed: %v", err)
		}
		fmt.Printf("Forced migration version to: %d\n", version)

	default:
		log.Fatalf("Unknown action: %s", action)
	}
}

func runMigration(dbURL, migrationsPath, action string, steps int) error {
	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	switch action {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	case "down":
		if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	case "force":
		if err := m.Force(steps); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown migration action: %s", action)
	}

	return nil
}

func showVersion(dbURL, migrationsPath string) error {
	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("No migrations applied yet")
			return nil
		}
		return err
	}

	fmt.Printf("Current migration version: %d\n", version)
	if dirty {
		fmt.Println("Warning: Database is in a dirty state")
	}

	return nil
}

func createMigration(name string) error {
	timestamp := time.Now().UTC().Format("20060102150405")
	upFile := fmt.Sprintf("db/migrations/%s_%s.up.sql", timestamp, name)
	downFile := fmt.Sprintf("db/migrations/%s_%s.down.sql", timestamp, name)

	// Create up migration file
	if err := os.WriteFile(upFile, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	// Create down migration file
	if err := os.WriteFile(downFile, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	fmt.Printf("Created migration files:\n")
	fmt.Printf("  %s\n", upFile)
	fmt.Printf("  %s\n", downFile)

	return nil
}
