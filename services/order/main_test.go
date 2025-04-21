package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/z0mbix/go-microservices-monorepo/pkg/config"
	"github.com/z0mbix/go-microservices-monorepo/pkg/service"
)

func TestServiceInitialization(t *testing.T) {
	// Save original environment to restore after tests
	originalPort := os.Getenv("APP_PORT")
	originalLogLevel := os.Getenv("APP_LOG_LEVEL")
	originalEnv := os.Getenv("APP_ENV")

	defer func() {
		os.Setenv("APP_PORT", originalPort)
		os.Setenv("APP_LOG_LEVEL", originalLogLevel)
		os.Setenv("APP_ENV", originalEnv)
	}()

	t.Run("service uses default port when env var not set", func(t *testing.T) {
		os.Unsetenv("APP_PORT")

		cfg, err := config.New(config.WithDefaultPort(servicePort))
		if err != nil {
			t.Fatalf("failed to create config: %v", err)
		}

		svc, err := service.NewWithName(
			serviceName,
			service.WithPort(cfg.Port),
		)
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		if svc.Port != servicePort {
			t.Errorf("expected port %d, got %d", servicePort, svc.Port)
		}
	})

	t.Run("service uses environment port when set", func(t *testing.T) {
		envPort := 9999
		os.Setenv("APP_PORT", strconv.Itoa(envPort))

		cfg, err := config.New(config.WithDefaultPort(servicePort))
		if err != nil {
			t.Fatalf("failed to create config: %v", err)
		}

		svc, err := service.NewWithName(
			serviceName,
			service.WithPort(cfg.Port),
		)
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		if svc.Port != envPort {
			t.Errorf("expected port %d, got %d", envPort, svc.Port)
		}
	})

	t.Run("service uses configured log level", func(t *testing.T) {
		os.Setenv("APP_LOG_LEVEL", "debug")

		cfg, err := config.New(config.WithDefaultPort(servicePort))
		if err != nil {
			t.Fatalf("failed to create config: %v", err)
		}

		svc, err := service.NewWithName(
			serviceName,
			service.WithLogLevel(cfg.LogLevel),
		)
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		if svc.LogLevel != "debug" {
			t.Errorf("expected log level 'debug', got %q", svc.LogLevel)
		}
	})

	t.Run("service uses configured environment", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")

		cfg, err := config.New(config.WithDefaultPort(servicePort))
		if err != nil {
			t.Fatalf("failed to create config: %v", err)
		}

		svc, err := service.NewWithName(
			serviceName,
			service.WithEnvironment(cfg.Environment),
		)
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		if svc.Environment != "production" {
			t.Errorf("expected environment 'production', got %q", svc.Environment)
		}
	})
}

func TestServiceEndpoints(t *testing.T) {
	// Create a minimal service for testing HTTP endpoints
	svc, err := service.NewWithName(
		serviceName,
		service.WithVersion("test-version"),
	)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			fmt.Fprintf(w, "%s service", svc.Name)
		case "/_ready":
			fmt.Fprintf(w, "%s service is ready", svc.Name)
		case "/_live":
			fmt.Fprintf(w, "%s service is alive", svc.Name)
		case "/_version":
			fmt.Fprintf(w, "%s", svc.Version)
		default:
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "root endpoint returns service name",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedBody:   "order service",
		},
		{
			name:           "ready endpoint returns readiness status",
			path:           "/_ready",
			expectedStatus: http.StatusOK,
			expectedBody:   "order service is ready",
		},
		{
			name:           "live endpoint returns liveness status",
			path:           "/_live",
			expectedStatus: http.StatusOK,
			expectedBody:   "order service is alive",
		},
		{
			name:           "version endpoint returns service version",
			path:           "/_version",
			expectedStatus: http.StatusOK,
			expectedBody:   "test-version",
		},
		{
			name:           "unknown endpoint returns 404",
			path:           "/unknown",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
	}

	client := testServer.Client()
	baseURL := testServer.URL

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(baseURL + tt.path)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Read response body
			buf := make([]byte, 1024)
			n, err := resp.Body.Read(buf)
			if err != nil && err.Error() != "EOF" {
				t.Fatalf("failed to read response body: %v", err)
			}

			body := string(buf[:n])
			if body != tt.expectedBody {
				t.Errorf("expected response body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

// TestIntegration performs an end-to-end test of the service configuration
func TestIntegration(t *testing.T) {
	// Skip in CI environments or when running short tests
	if testing.Short() || os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in short mode or CI environment")
	}

	// Save original environment to restore after tests
	originalPort := os.Getenv("APP_PORT")
	originalLogLevel := os.Getenv("APP_LOG_LEVEL")
	originalEnv := os.Getenv("APP_ENV")

	// Restore environment variables after test
	defer func() {
		os.Setenv("APP_PORT", originalPort)
		os.Setenv("APP_LOG_LEVEL", originalLogLevel)
		os.Setenv("APP_ENV", originalEnv)
	}()

	// Set up test environment
	testPort := 19999 // Use high port number to avoid conflicts
	testVersion := "integration-test"

	os.Setenv("APP_PORT", strconv.Itoa(testPort))
	os.Setenv("APP_ENV", "test")
	os.Setenv("APP_LOG_LEVEL", "debug")

	// Create configuration
	cfg, err := config.New(
		config.WithDefaultPort(servicePort), // This should be overridden by APP_PORT
	)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	if cfg.Port != testPort {
		t.Fatalf("expected port to be %d, got %d", testPort, cfg.Port)
	}

	// Create service
	svc, err := service.NewWithName(
		serviceName,
		service.WithEnvironment(cfg.Environment),
		service.WithPort(cfg.Port),
		service.WithLogLevel(cfg.LogLevel),
		service.WithVersion(testVersion),
	)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	// Verify service properties
	if svc.Port != testPort {
		t.Errorf("expected port %d, got %d", testPort, svc.Port)
	}
	if svc.Environment != "test" {
		t.Errorf("expected environment 'test', got %q", svc.Environment)
	}
	if svc.LogLevel != "debug" {
		t.Errorf("expected log level 'debug', got %q", svc.LogLevel)
	}
	if svc.Version != testVersion {
		t.Errorf("expected version %q, got %q", testVersion, svc.Version)
	}
}
