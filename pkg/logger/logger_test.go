package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("valid log levels", func(t *testing.T) {
		levels := []string{"debug", "info", "warn", "error"}
		for _, level := range levels {
			t.Run(level, func(t *testing.T) {
				logger, err := New(level)
				if err != nil {
					t.Fatalf("New(%q) returned unexpected error: %v", level, err)
				}
				if logger == nil {
					t.Fatalf("New(%q) returned nil logger", level)
				}
			})
		}
	})

	t.Run("case insensitivity", func(t *testing.T) {
		variants := []string{"DEBUG", "Debug", "dEbUg", "debug"}
		for _, variant := range variants {
			logger, err := New(variant)
			if err != nil {
				t.Fatalf("New(%q) returned unexpected error: %v", variant, err)
			}
			if logger == nil {
				t.Fatalf("New(%q) returned nil logger", variant)
			}
		}
	})

	t.Run("invalid log level", func(t *testing.T) {
		invalidLevels := []string{"", "trace", "critical", "unknown"}
		for _, level := range invalidLevels {
			_, err := New(level)
			if err == nil {
				t.Fatalf("New(%q) should have returned an error", level)
			}
			if !strings.Contains(err.Error(), "invalid log level") {
				t.Fatalf("New(%q) returned unexpected error message: %v", level, err)
			}
		}
	})
}

func TestLogOutput(t *testing.T) {
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	tests := []struct {
		name           string
		level          string
		logFunc        func(logger *slog.Logger)
		expectedLevel  string
		shouldContain  string
		shouldNotMatch bool
	}{
		{
			name:           "debug level includes debug messages",
			level:          "debug",
			logFunc:        func(logger *slog.Logger) { logger.Debug("test debug message") },
			expectedLevel:  "DEBUG",
			shouldContain:  "test debug message",
			shouldNotMatch: false,
		},
		{
			name:           "info level includes info messages",
			level:          "info",
			logFunc:        func(logger *slog.Logger) { logger.Info("test info message") },
			expectedLevel:  "INFO",
			shouldContain:  "test info message",
			shouldNotMatch: false,
		},
		{
			name:           "warn level includes warn messages",
			level:          "warn",
			logFunc:        func(logger *slog.Logger) { logger.Warn("test warn message") },
			expectedLevel:  "WARN",
			shouldContain:  "test warn message",
			shouldNotMatch: false,
		},
		{
			name:           "error level includes error messages",
			level:          "error",
			logFunc:        func(logger *slog.Logger) { logger.Error("test error message") },
			expectedLevel:  "ERROR",
			shouldContain:  "test error message",
			shouldNotMatch: false,
		},
		{
			name:           "info level excludes debug messages",
			level:          "info",
			logFunc:        func(logger *slog.Logger) { logger.Debug("should not appear") },
			shouldContain:  "should not appear",
			shouldNotMatch: true,
		},
		{
			name:           "warn level excludes info messages",
			level:          "warn",
			logFunc:        func(logger *slog.Logger) { logger.Info("should not appear") },
			shouldContain:  "should not appear",
			shouldNotMatch: true,
		},
		{
			name:           "error level excludes warn messages",
			level:          "error",
			logFunc:        func(logger *slog.Logger) { logger.Warn("should not appear") },
			shouldContain:  "should not appear",
			shouldNotMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			os.Stdout = w

			logger, err := New(tt.level)
			if err != nil {
				t.Fatalf("New(%q) returned unexpected error: %v", tt.level, err)
			}

			tt.logFunc(logger)

			w.Close()

			// Read the captured output
			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			if err != nil {
				t.Fatalf("failed to read captured output: %v", err)
			}

			output := buf.String()

			if tt.shouldNotMatch {
				if output != "" {
					t.Errorf("expected no output, got: %s", output)
				}
				return
			}

			if output == "" {
				t.Fatal("expected log output but got nothing")
			}

			var logData map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
				t.Fatalf("failed to parse JSON log output: %v\nOutput: %s", err, output)
			}

			if tt.expectedLevel != "" {
				if level, ok := logData["level"]; !ok || level != tt.expectedLevel {
					t.Errorf("expected level %q, got %v", tt.expectedLevel, level)
				}
			}

			if tt.shouldContain != "" {
				if msg, ok := logData["msg"]; !ok || msg != tt.shouldContain {
					t.Errorf("expected message to contain %q, got %v", tt.shouldContain, msg)
				}
			}
		})
	}
}

func TestLogAttributes(t *testing.T) {
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	logger, err := New("info")
	if err != nil {
		t.Fatalf("New returned unexpected error: %v", err)
	}

	logger.Info("test message with attributes",
		"string", "value",
		"int", 42,
		"bool", true,
		"float", 3.14,
	)

	w.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected log output but got nothing")
	}

	var logData map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("failed to parse JSON log output: %v\nOutput: %s", err, output)
	}

	expectedAttrs := map[string]interface{}{
		"string": "value",
		"int":    float64(42),
		"bool":   true,
		"float":  3.14,
	}

	for key, expectedValue := range expectedAttrs {
		value, ok := logData[key]
		if !ok {
			t.Errorf("expected attribute %q not found in log output", key)
			continue
		}

		if value != expectedValue {
			t.Errorf("attribute %q: expected %v (%T), got %v (%T)",
				key, expectedValue, expectedValue, value, value)
		}
	}
}

func TestDefaultLogger(t *testing.T) {
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	_, err = New("info")
	if err != nil {
		t.Fatalf("New returned unexpected error: %v", err)
	}

	slog.Info("default logger test")

	w.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected log output from default logger but got nothing")
	}

	var logData map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("failed to parse JSON log output: %v\nOutput: %s", err, output)
	}

	if msg, ok := logData["msg"]; !ok || msg != "default logger test" {
		t.Errorf("expected message \"default logger test\", got %v", msg)
	}
}
