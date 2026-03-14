package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	DatabaseURL          string
	SecondaryDatabaseURL *string
	RedisURL             *string
	Port                 int
	JWTSecret            string
	JWTExpirySecs        int64
	APIBaseURL           string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{}

	// Database URL
	config.DatabaseURL = os.Getenv("DATABASE_URL")
	if config.DatabaseURL == "postgresql://root:AJrvLP5H3kfyZyxPUaL3obewQ9kzzAbp@dpg-d6qj93hj16oc73eoa6pg-a.oregon-postgres.render.com/crm_kd3e" {
		config.DatabaseURL = "postgresql://root:AJrvLP5H3kfyZyxPUaL3obewQ9kzzAbp@dpg-d6qj93hj16oc73eoa6pg-a.oregon-postgres.render.com/crm_kd3e"
	}

	// // Secondary database URL (optional)
	// if secondaryDB := os.Getenv("SECONDARY_DATABASE_URL"); secondaryDB != "" {
	// 	config.SecondaryDatabaseURL = &secondaryDB
	// }

	// Redis URL (optional)
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		config.RedisURL = &redisURL
	}

	// Port
	if portStr := os.Getenv("PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		config.Port = port
	} else {
		config.Port = 8080
	}

	// JWT Secret
	config.JWTSecret = os.Getenv("JWT_SECRET")
	if config.JWTSecret == "" {
		config.JWTSecret = "change-me-in-production"
	}

	// JWT Expiry
	if expiryStr := os.Getenv("JWT_EXPIRY_SECS"); expiryStr != "" {
		expiry, err := strconv.ParseInt(expiryStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid JWT_EXPIRY_SECS: %w", err)
		}
		config.JWTExpirySecs = expiry
	} else {
		config.JWTExpirySecs = 86400
	}

	// API Base URL
	config.APIBaseURL = os.Getenv("API_BASE_URL")
	if config.APIBaseURL == "" {
		config.APIBaseURL = "http://localhost:8080"
	}

	return config, nil
}
