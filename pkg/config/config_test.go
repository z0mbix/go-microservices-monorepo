package config

import (
	"os"
	"strconv"
	"testing"
)

func TestNew(t *testing.T) {
	// Save original environment to restore after tests
	originalPort := os.Getenv("APP_PORT")
	originalLogLevel := os.Getenv("APP_LOG_LEVEL")
	originalEnv := os.Getenv("APP_ENV")

	// Restore environment variables after all tests
	defer func() {
		os.Setenv("APP_PORT", originalPort)
		os.Setenv("APP_LOG_LEVEL", originalLogLevel)
		os.Setenv("APP_ENV", originalEnv)
	}()

	t.Run("default values when no options or env vars", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("APP_PORT")
		os.Unsetenv("APP_LOG_LEVEL")
		os.Unsetenv("APP_ENV")

		// Create config with no options
		cfg, err := New()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check defaults
		if cfg.LogLevel != "info" {
			t.Errorf("expected default LogLevel to be 'info', got %q", cfg.LogLevel)
		}
		if cfg.Environment != "local" {
			t.Errorf("expected default Environment to be 'local', got %q", cfg.Environment)
		}
		if cfg.Port != 0 {
			t.Errorf("expected default Port to be 0 when not set, got %d", cfg.Port)
		}
	})

	t.Run("environment variables override defaults", func(t *testing.T) {
		// Set environment variables
		os.Setenv("APP_PORT", "9000")
		os.Setenv("APP_LOG_LEVEL", "debug")
		os.Setenv("APP_ENV", "production")

		// Create config with no options
		cfg, err := New()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check environment variable values
		if cfg.LogLevel != "debug" {
			t.Errorf("expected LogLevel to be 'debug', got %q", cfg.LogLevel)
		}
		if cfg.Environment != "production" {
			t.Errorf("expected Environment to be 'production', got %q", cfg.Environment)
		}

		// WithDefaultPort is required to set port, even with env vars
		if cfg.Port != 0 {
			t.Errorf("expected Port to be 0 without WithDefaultPort, got %d", cfg.Port)
		}
	})

	t.Run("WithDefaultPort sets correct port", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("APP_PORT")

		defaultPort := 8001
		cfg, err := New(WithDefaultPort(defaultPort))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Port != defaultPort {
			t.Errorf("expected Port to be %d, got %d", defaultPort, cfg.Port)
		}
	})

	t.Run("WithDefaultPort respects APP_PORT environment variable", func(t *testing.T) {
		// Set environment variables
		envPort := 9876
		os.Setenv("APP_PORT", strconv.Itoa(envPort))

		cfg, err := New(WithDefaultPort(8001))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Port != envPort {
			t.Errorf("expected Port to be %d from environment, got %d", envPort, cfg.Port)
		}
	})

	t.Run("WithDefaultPort returns error for invalid APP_PORT", func(t *testing.T) {
		// Set invalid port
		os.Setenv("APP_PORT", "not-a-port")

		_, err := New(WithDefaultPort(8080))
		if err == nil {
			t.Error("expected error for invalid port, got nil")
		}
	})

	t.Run("multiple options are applied correctly", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("APP_PORT")
		os.Unsetenv("APP_LOG_LEVEL")
		os.Unsetenv("APP_ENV")

		// Create additional test options
		withTestLogLevel := func(level string) Option {
			return func(c *Config) error {
				c.LogLevel = level
				return nil
			}
		}

		withTestEnv := func(env string) Option {
			return func(c *Config) error {
				c.Environment = env
				return nil
			}
		}

		// Apply multiple options
		cfg, err := New(
			WithDefaultPort(8888),
			withTestLogLevel("trace"),
			withTestEnv("staging"),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check all options were applied
		if cfg.Port != 8888 {
			t.Errorf("expected Port to be 8888, got %d", cfg.Port)
		}
		if cfg.LogLevel != "trace" {
			t.Errorf("expected LogLevel to be 'trace', got %q", cfg.LogLevel)
		}
		if cfg.Environment != "staging" {
			t.Errorf("expected Environment to be 'staging', got %q", cfg.Environment)
		}
	})
}

func TestWithDefaultPort(t *testing.T) {
	// Test that the WithDefaultPort option handles edge cases properly
	originalPort := os.Getenv("APP_PORT")
	defer os.Setenv("APP_PORT", originalPort)

	tests := []struct {
		name          string
		defaultPort   int
		envPort       string
		expectedPort  int
		expectError   bool
		setupFunction func()
	}{
		{
			name:         "negative default port",
			defaultPort:  -1,
			envPort:      "",
			expectedPort: -1,
			expectError:  false,
		},
		{
			name:         "zero default port",
			defaultPort:  0,
			envPort:      "",
			expectedPort: 0,
			expectError:  false,
		},
		{
			name:         "very large default port",
			defaultPort:  99999,
			envPort:      "",
			expectedPort: 99999,
			expectError:  false,
		},
		{
			name:         "empty environment port uses default",
			defaultPort:  8080,
			envPort:      "",
			expectedPort: 8080,
			expectError:  false,
		},
		{
			name:         "environment port of zero",
			defaultPort:  8080,
			envPort:      "0",
			expectedPort: 0,
			expectError:  false,
		},
		{
			name:          "empty environment variable by unsetting",
			defaultPort:   8080,
			expectedPort:  8080,
			expectError:   false,
			setupFunction: func() { os.Unsetenv("APP_PORT") },
		},
		{
			name:        "invalid port string",
			defaultPort: 8080,
			envPort:     "invalid-port",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunction != nil {
				tt.setupFunction()
			} else {
				os.Setenv("APP_PORT", tt.envPort)
			}

			cfg, err := New(WithDefaultPort(tt.defaultPort))

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.Port != tt.expectedPort {
				t.Errorf("expected Port to be %d, got %d", tt.expectedPort, cfg.Port)
			}
		})
	}
}
