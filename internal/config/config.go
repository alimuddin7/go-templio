// Package config provides environment-based configuration loading.
// All fields are strictly typed; zero global state.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the full application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Session  SessionConfig
}

// AppConfig holds HTTP server settings.
type AppConfig struct {
	Env     string
	Port    string
	Prefork bool
}

// DatabaseConfig holds DB connection settings.
type DatabaseConfig struct {
	Driver   string // "sqlite" | "postgres"
	DSN      string // used for sqlite
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// PostgresDSN builds the postgres connection string.
func (d DatabaseConfig) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

// JWTConfig holds JWT signing settings.
type JWTConfig struct {
	Secret      string
	ExpiryHours time.Duration
}

// SessionConfig holds session cookie settings.
type SessionConfig struct {
	Secret string
	MaxAge int
}

// Load reads .env (if present) then populates Config from os.Getenv.
// Returns an error if required fields are missing.
func Load(envFile string) (*Config, error) {
	// Best-effort: if the file doesn't exist we still fall through to env vars.
	_ = godotenv.Load(envFile)

	cfg := &Config{
		App: AppConfig{
			Env:     getEnvOrDefault("APP_ENV", "development"),
			Port:    getEnvOrDefault("APP_PORT", "3000"),
			Prefork: parseBool(os.Getenv("APP_PREFORK")),
		},
		Database: DatabaseConfig{
			Driver:   getEnvOrDefault("DB_DRIVER", "sqlite"),
			DSN:      getEnvOrDefault("DB_DSN", "./go-templio.db"),
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     getEnvOrDefault("DB_NAME", "go_templio"),
		},
		JWT: JWTConfig{
			Secret:      getEnvOrDefault("JWT_SECRET", "change-me"),
			ExpiryHours: time.Duration(parseInt(os.Getenv("JWT_EXPIRY_HOURS"), 24)) * time.Hour,
		},
		Session: SessionConfig{
			Secret: getEnvOrDefault("SESSION_SECRET", "change-me"),
			MaxAge: parseInt(os.Getenv("SESSION_MAX_AGE"), 86400),
		},
	}

	if cfg.JWT.Secret == "change-me" && cfg.App.Env == "production" {
		return nil, fmt.Errorf("config: JWT_SECRET must be set in production")
	}

	return cfg, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseBool(v string) bool {
	b, _ := strconv.ParseBool(v)
	return b
}

func parseInt(v string, fallback int) int {
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return fallback
}
