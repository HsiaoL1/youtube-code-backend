package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// AppConfig stores the runtime config shared by every service.
type AppConfig struct {
	ServiceName string
	Port        int
	Env         string

	// Database
	DatabaseDSN string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// JWT
	JWTSecret             string
	JWTAccessTokenExpiry  time.Duration
	JWTRefreshTokenExpiry time.Duration
}

// FromEnv loads service config from environment variables.
func FromEnv(defaultServiceName string, defaultPort int) (AppConfig, error) {
	cfg := AppConfig{
		ServiceName:           getEnv("SERVICE_NAME", defaultServiceName),
		Port:                  defaultPort,
		Env:                   getEnv("ENV", "development"),
		DatabaseDSN:           getEnv("DB_DSN", "host=localhost user=youtube password=youtube dbname=youtube port=5432 sslmode=disable"),
		RedisAddr:             getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:         getEnv("REDIS_PASSWORD", ""),
		RedisDB:               getEnvInt("REDIS_DB", 0),
		JWTSecret:             getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTAccessTokenExpiry:  getEnvDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshTokenExpiry: getEnvDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
	}

	if raw := os.Getenv("PORT"); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil {
			return AppConfig{}, fmt.Errorf("invalid PORT value %q: %w", raw, err)
		}
		cfg.Port = port
	}

	if cfg.Port <= 0 || cfg.Port > 65535 {
		return AppConfig{}, fmt.Errorf("port %d out of range", cfg.Port)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
