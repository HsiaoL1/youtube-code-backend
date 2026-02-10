package config

import (
	"fmt"
	"os"
	"strconv"
)

// AppConfig stores the minimal runtime config shared by every service.
type AppConfig struct {
	ServiceName string
	Port        int
}

// FromEnv loads service config from environment variables.
func FromEnv(defaultServiceName string, defaultPort int) (AppConfig, error) {
	cfg := AppConfig{
		ServiceName: getEnv("SERVICE_NAME", defaultServiceName),
		Port:        defaultPort,
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
