package service

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/z0mbix/go-microservices-monorepo/pkg/logger"
)

type Service struct {
	Environment string
	LogLevel    string
	Log         *slog.Logger
	Name        string
	Port        int
	Version     string
}

type Option func(*Service)

func WithEnvironment(env string) Option {
	return func(s *Service) {
		s.Environment = env
	}
}

func WithLogLevel(level string) Option {
	return func(s *Service) {
		s.LogLevel = level
		if s.Log != nil {
			newLogger, err := logger.New(level)
			if err == nil {
				s.Log = newLogger
			}
		}
	}
}

func WithPort(port int) Option {
	return func(c *Service) {
		c.Port = port
	}
}

func WithVersion(version string) Option {
	return func(s *Service) {
		s.Version = version
	}
}

func NewWithName(name string, opts ...Option) (*Service, error) {
	svc := &Service{
		LogLevel: "info",
		Name:     name,
		Port:     8000,
	}

	var err error
	svc.Log, err = logger.New(svc.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %w", err)
	}

	for _, opt := range opts {
		opt(svc)
	}

	return svc, nil
}

func (s *Service) Run() error {
	s.Log.Info("starting",
		"service", s.Name,
		"port", s.Port,
		"version", s.Version,
		"environment", s.Environment,
		"level", s.LogLevel,
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s service", s.Name)
	})

	http.HandleFunc("/_ready", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s service is ready", s.Name)
	})

	http.HandleFunc("/_live", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s service is alive", s.Name)
	})

	http.HandleFunc("/_version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", s.Version)
	})

	port := fmt.Sprintf(":%d", s.Port)

	return http.ListenAndServe(port, nil)
}
