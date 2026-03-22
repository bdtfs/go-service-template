# Go Service Template

A production-ready Go microservice template with hexagonal architecture, composable infrastructure modules, and built-in observability.

## Features

- **Hexagonal architecture** — clean separation of domain, ports, and adapters
- **Composable modules** — enable/disable infrastructure via `config.yaml`
- **Service types** — API (HTTP server), Consumer, Worker
- **Built-in observability** — structured logging, Prometheus metrics, health checks
- **HTTP middleware stack** — recovery, request-id, logging, metrics
- **YAML config** — with `${ENV_VAR:default}` interpolation
- **Graceful shutdown** — ordered component teardown on SIGINT/SIGTERM
- **Kubernetes ready** — `/healthz` and `/readyz` endpoints
- **Minimal dependencies** — stdlib where possible, no unnecessary frameworks

## Quick Start

```bash
# Clone and rename
git clone <this-repo> my-service
cd my-service

# Update module path
go mod edit -module github.com/yourorg/my-service
grep -rl 'github.com/bdtfs/go-service-template' --include='*.go' | xargs sed -i '' 's|github.com/bdtfs/go-service-template|github.com/yourorg/my-service|g'

# Configure
cp config.yaml config.yaml  # edit as needed

# Run
make run
```

## Architecture

```
cmd/service/main.go              Entry point — composes service
internal/
  config/                         YAML config loader
  domain/                         Business logic
  ports/
    inbound/                      Driving ports (handler interfaces)
    outbound/                     Driven ports (repository interfaces)
  adapters/
    inbound/                      HTTP handlers, consumers
    outbound/                     Database repos, API clients
pkg/
  service/                        Service builder + Component interface
  clog/                           Structured JSON logging
  metrics/                        Prometheus metrics + health checks
  middleware/                     HTTP middleware
  postgres/                       PostgreSQL component
  transactions/                   Transaction manager
```

### Service Composition

The service builder handles lifecycle, middleware, and graceful shutdown:

```go
func main() {
    cfg := config.Must(config.Load("config.yaml"))

    var opts []service.Option
    if cfg.Components.Postgres.Enabled {
        opts = append(opts, service.WithComponent(
            postgres.NewComponent(cfg.Components.Postgres.DSN),
        ))
    }

    svc := service.Must(service.New(cfg, opts...))

    svc.HandleFunc("GET /api/v1/items", itemHandler)

    svc.Run(context.Background())
}
```

### Service Types

Set `service.type` in `config.yaml`:

| Type       | Behavior                                           |
|------------|----------------------------------------------------|
| `api`      | Starts HTTP server, applies middleware              |
| `consumer` | Runs start functions for message consumption        |
| `worker`   | Runs background tasks via `WithStartFunc`           |

### Component Interface

Add infrastructure by implementing `service.Component`:

```go
type Component interface {
    Name() string
    Init(ctx context.Context) error
    Close(ctx context.Context) error
}
```

Components are initialized in registration order and closed in reverse order during shutdown.

## Configuration

`config.yaml` with environment variable interpolation:

```yaml
service:
  name: my-service
  type: api

server:
  port: ":8080"
  read_timeout: 5s
  write_timeout: 10s

log:
  level: ${LOG_LEVEL:info}
  format: json

metrics:
  enabled: true
  address: ":8081"
  namespace: ${METRICS_NAMESPACE:my-service}

components:
  postgres:
    enabled: ${POSTGRES_ENABLED:false}
    dsn: ${PG_DSN:postgresql://postgres:password@localhost:5432/mydb?sslmode=disable}
```

## Adding Components

Scaffold a new infrastructure component:

```bash
make add-component COMPONENT=redis
```

This creates `pkg/redis/redis.go` with the Component interface skeleton and updates `config.yaml`. Then wire it up in `main.go`.

## Endpoints

| Endpoint    | Port | Description              |
|-------------|------|--------------------------|
| `/healthz`  | 8081 | Kubernetes liveness      |
| `/readyz`   | 8081 | Kubernetes readiness     |
| `/metrics`  | 8081 | Prometheus metrics       |
| User routes | 8080 | Application HTTP server  |

## Development

```bash
make build           # Build binary
make run             # Build and run
make test            # Unit tests with race detector
make int-test        # Integration tests
make lint            # golangci-lint
make codegen         # Generate mocks
make dc-reup         # Restart docker-compose
```

## License

MIT
