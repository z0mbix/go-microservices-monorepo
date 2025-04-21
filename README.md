# Monorepo Services

This is a sample monorepo with multiple Go microservices for testing purposes.

## Services

- **billing**: Handles billing processing
- **order**: Manages customer orders
- **shipping**: Manages shipment tracking
- **user**: Handles user authentication and profile management

## Running Services

Each service can be configured using environment variables:

```bash
# Run the order service on it's default port with default log level (info)
go run services/order/main.go

# Run the billing service with custom configuration
APP_PORT=3000 APP_LOG_LEVEL=debug go run services/billing/main.go
```

## Environment Variables

All services support the following environment variables:

| Variable        | Description                                  | Default  |
|-----------------|----------------------------------------------|----------|
| APP_PORT        | HTTP server port                             | none     |
| APP_LOG_LEVEL   | Logging level (debug, info, warn, error)     | info     |
| APP_ENV         | Environment (development, production etc.)   | local    |
