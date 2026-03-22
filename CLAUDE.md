# Go Service Template

## Project Structure

This is a hexagonal architecture Go microservice template with a composable module system.

```
cmd/service/main.go          Entry point — composes service from config
internal/
  config/                     YAML config with env var interpolation
  domain/                     Business logic (entities, value objects)
  ports/inbound/              Driving ports (interfaces the outside world calls)
  ports/outbound/             Driven ports (interfaces we call on the outside)
  adapters/inbound/           HTTP handlers, consumers, gRPC servers
  adapters/outbound/          Repository implementations, API clients
pkg/
  service/                    Service builder & Component interface
  clog/                       Structured JSON logging (wraps slog)
  metrics/                    Prometheus metrics, series, health checks
  middleware/                  HTTP middleware (recovery, request-id, logging, metrics)
  postgres/                   PostgreSQL component (pgx/v4)
  transactions/               Transaction manager pattern
```

## Key Commands

```bash
make build          # Build binary to ./bin/service
make run            # Build and run
make test           # Run all tests with race detector
make int-test       # Run integration tests (build tag: integration)
make lint           # Run golangci-lint
make codegen        # Generate mocks and code
make dc-reup        # Docker compose restart
make add-component COMPONENT=redis   # Scaffold a new component
```

## Architecture

### Service Composition

Services are composed in `main.go` via the service builder:

```go
cfg := config.Must(config.Load("config.yaml"))
svc := service.Must(service.New(cfg,
    service.WithComponent(postgres.NewComponent(cfg.Components.Postgres.DSN)),
))
svc.Run(context.Background())
```

### Component Interface

Infrastructure components implement `service.Component`:

```go
type Component interface {
    Name() string
    Init(ctx context.Context) error
    Close(ctx context.Context) error
}
```

### Config

YAML config at `config.yaml` with `${ENV_VAR:default}` interpolation.
Service type (`api`, `consumer`, `worker`) determines runtime behavior.

## Conventions

- **No panics** in library code — return errors
- **Interfaces in ports/**, implementations in adapters/
- **Context propagation** — pass ctx through all layers
- Config uses YAML with env var interpolation, not raw env parsing
- Tests use `clog.NewCLogStub()` and `metrics.NewRegistryStub()`
- The `pkg/` packages must NOT import from `internal/` (except `pkg/service` imports `internal/config`)

## Testing

```bash
go test -race ./...                    # Unit tests
go test -tags=integration -race ./...  # Integration tests
```

Use stubs for unit tests:
- `clog.NewCLogStub()` — no-op logger
- `metrics.NewRegistryStub()` — no-op metrics
- `transactions.NewTrmStub()` — pass-through transaction manager
