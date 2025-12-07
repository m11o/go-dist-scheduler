package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_WithDefaults(t *testing.T) {
	// Clear environment variables
	clearEnv(t)

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check database defaults
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "scheduler", cfg.Database.User)
	assert.Equal(t, "password", cfg.Database.Password)
	assert.Equal(t, "scheduler", cfg.Database.Name)
	assert.Equal(t, "disable", cfg.Database.SSLMode)

	// Check Redis defaults
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, "", cfg.Redis.Password)
	assert.Equal(t, 0, cfg.Redis.DB)
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	setEnv(t, map[string]string{
		"DB_HOST":       "db.example.com",
		"DB_PORT":       "5433",
		"DB_USER":       "testuser",
		"DB_PASSWORD":   "testpass",
		"DB_NAME":       "testdb",
		"DB_SSLMODE":    "require",
		"REDIS_HOST":    "redis.example.com",
		"REDIS_PORT":    "6380",
		"REDIS_PASSWORD": "redispass",
		"REDIS_DB":      "1",
	})
	defer clearEnv(t)

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check database config
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "require", cfg.Database.SSLMode)

	// Check Redis config
	assert.Equal(t, "redis.example.com", cfg.Redis.Host)
	assert.Equal(t, 6380, cfg.Redis.Port)
	assert.Equal(t, "redispass", cfg.Redis.Password)
	assert.Equal(t, 1, cfg.Redis.DB)
}

func TestLoad_WithPartialEnvironmentVariables(t *testing.T) {
	// Set only some environment variables
	setEnv(t, map[string]string{
		"DB_HOST": "custom.db.com",
		"DB_PORT": "3306",
		"REDIS_HOST": "custom.redis.com",
	})
	defer clearEnv(t)

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check overridden values
	assert.Equal(t, "custom.db.com", cfg.Database.Host)
	assert.Equal(t, 3306, cfg.Database.Port)
	assert.Equal(t, "custom.redis.com", cfg.Redis.Host)

	// Check default values
	assert.Equal(t, "scheduler", cfg.Database.User)
	assert.Equal(t, 6379, cfg.Redis.Port)
}

func TestDatabaseConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "default config",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "scheduler",
				Password: "password",
				Name:     "scheduler",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=scheduler password=password dbname=scheduler sslmode=disable",
		},
		{
			name: "custom config",
			config: DatabaseConfig{
				Host:     "db.example.com",
				Port:     5433,
				User:     "admin",
				Password: "secret",
				Name:     "mydb",
				SSLMode:  "require",
			},
			expected: "host=db.example.com port=5433 user=admin password=secret dbname=mydb sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.DSN()
			assert.Equal(t, tt.expected, dsn)
		})
	}
}

func TestRedisConfig_Addr(t *testing.T) {
	tests := []struct {
		name     string
		config   RedisConfig
		expected string
	}{
		{
			name: "default config",
			config: RedisConfig{
				Host: "localhost",
				Port: 6379,
			},
			expected: "localhost:6379",
		},
		{
			name: "custom config",
			config: RedisConfig{
				Host: "redis.example.com",
				Port: 6380,
			},
			expected: "redis.example.com:6380",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := tt.config.Addr()
			assert.Equal(t, tt.expected, addr)
		})
	}
}

func TestLoad_InvalidPortValue(t *testing.T) {
	setEnv(t, map[string]string{
		"DB_PORT": "invalid",
	})
	defer clearEnv(t)

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load database config")
}

// Helper function to set environment variables for testing
func setEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	for key, value := range vars {
		err := os.Setenv(key, value)
		require.NoError(t, err)
	}
}

// Helper function to clear environment variables after testing
func clearEnv(t *testing.T) {
	t.Helper()
	envVars := []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
	}
	for _, key := range envVars {
		err := os.Unsetenv(key)
		require.NoError(t, err)
	}
}
