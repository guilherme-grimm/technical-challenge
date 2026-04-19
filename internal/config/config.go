package config

import (
	"errors"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// carry of all configs
type Config struct {
	URI             string
	Database        string
	Collection      string
	Addr            string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	LogLevel        string
}

func New() (*Config, error) {
	_ = godotenv.Load()

	if os.Getenv("MONGO_URI") == "" {
		return nil, errors.New("MONGO_URI is required")
	}

	shutdownTimeout := getDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	readTimeout := getDuration("READ_TIMEOUT", 15*time.Second)
	writeTimeout := getDuration("WRITE_TIMEOUT", 15*time.Second)
	idleTimeout := getDuration("IDLE_TIMEOUT", 60*time.Second)

	cfg := &Config{
		URI:             os.Getenv("MONGO_URI"),
		Database:        getEnv("MONGO_DATABASE", "devices"),
		Collection:      getEnv("MONGO_COLLECTION", "devices"),
		Addr:            getEnv("ADDR", ":8080"),
		ShutdownTimeout: shutdownTimeout,
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		IdleTimeout:     idleTimeout,
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
	return cfg, nil
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
