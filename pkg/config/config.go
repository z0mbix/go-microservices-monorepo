package config

import (
	"cmp"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port        int
	LogLevel    string
	Environment string
}

type Option func(*Config) error

// WithDefaultPort sets the default port, but allows override via environment variable
func WithDefaultPort(port int) Option {
	return func(c *Config) error {
		c.Port = port
		portFromEnvVariable := os.Getenv("APP_PORT")
		if portFromEnvVariable != "" {
			parsedPort, err := strconv.Atoi(portFromEnvVariable)
			if err != nil {
				return fmt.Errorf("invalid port in APP_PORT environment variable: %w", err)
			}
			c.Port = parsedPort
		}
		return nil
	}
}

// New creates a new Config with the provided options
func New(opts ...Option) (*Config, error) {
	cfg := &Config{
		Environment: cmp.Or(os.Getenv("APP_ENV"), "local"),
		LogLevel:    cmp.Or(os.Getenv("APP_LOG_LEVEL"), "info"),
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
