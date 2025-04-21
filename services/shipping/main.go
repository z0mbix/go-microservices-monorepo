package main

import (
	"github.com/z0mbix/go-microservices-monorepo/pkg/config"
	"github.com/z0mbix/go-microservices-monorepo/pkg/service"
	"github.com/z0mbix/go-microservices-monorepo/pkg/version"
)

const (
	serviceName = "shipping"
	servicePort = 8003
)

func main() {
	serviceVersion := version.Version()

	cfg, err := config.New(
		config.WithDefaultPort(servicePort),
	)
	if err != nil {
		panic(err)
	}

	svc, err := service.NewWithName(
		serviceName,
		service.WithEnvironment(cfg.Environment),
		service.WithPort(cfg.Port),
		service.WithLogLevel(cfg.LogLevel),
		service.WithVersion(serviceVersion),
	)
	if err != nil {
		panic(err)
	}

	err = svc.Run()
	if err != nil {
		panic(err)
	}
}
