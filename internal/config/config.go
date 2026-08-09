package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DBName     string
	DBUser     string
	DBPassword string
	DBPort     string
	DBHost     string
	DBSSLMode  string

	HTTPAddr              string
	GRPCAddr              string
	CredentialServiceAddr string
}

func Load() (*Config, error) {
	cfg := &Config{
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		HTTPAddr:              getEnv("HTTP_ADDR", ":8080"),
		GRPCAddr:              getEnv("GRPC_ADDR", ":9090"),
		CredentialServiceAddr: getEnv("CREDENTIAL_SERVICE_ADDR", "localhost:9090"),
	}

	var missing []string
	if cfg.DBName == "" {
		missing = append(missing, "DB_NAME")
	}
	if cfg.DBUser == "" {
		missing = append(missing, "DB_USER")
	}
	if cfg.DBPassword == "" {
		missing = append(missing, "DB_PASSWORD")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
