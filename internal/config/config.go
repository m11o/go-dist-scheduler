package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config represents the application configuration.
type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
}

// DatabaseConfig represents database connection configuration.
type DatabaseConfig struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     int    `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"scheduler"`
	Password string `envconfig:"DB_PASSWORD" required:"true"`
	Name     string `envconfig:"DB_NAME" default:"scheduler"`
	SSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`
}

// RedisConfig represents Redis connection configuration.
type RedisConfig struct {
	Host     string `envconfig:"REDIS_HOST" default:"localhost"`
	Port     int    `envconfig:"REDIS_PORT" default:"6379"`
	Password string `envconfig:"REDIS_PASSWORD" default:""`
	DB       int    `envconfig:"REDIS_DB" default:"0"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg.Database); err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	if err := envconfig.Process("", &cfg.Redis); err != nil {
		return nil, fmt.Errorf("failed to load redis config: %w", err)
	}

	return &cfg, nil
}

// DSN returns the database connection string.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

// Addr returns the Redis server address.
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
