# Go Service Template

A production-ready Go microservice template using a layered service architecture
(the Avito Wallet service conventions): a `deps` seam decoupling domain code from
infrastructure, per-aggregate use cases, a squirrel/scany postgres adapter,
per-endpoint handler packages, and a lazy DI container.

## Features

- **Strict dependency direction** — use cases own narrow ports over domain values; clients and storage are adapters wired only by `di`
- **Point-of-use interfaces** — each consumer declares the narrow interface it needs; mocks via mockery
- **Series metrics** — every layer instruments itself through a `pkg/metrics.Series`
- **Composable infra modules** — enable/disable components via `config.yaml`
- **Service types** — API (HTTP server), Consumer, Worker
- **Built-in observability** — structured logging, Prometheus metrics, health checks
- **HTTP middleware stack** — recovery, request-id, logging, metrics
- **YAML config** — infra sections plus an `app:` section for domain tunables
- **Graceful shutdown** — ordered component teardown on SIGINT/SIGTERM
- **Kubernetes ready** — `/healthz` and `/readyz` endpoints
- **Worked reference domain** — an `operations` aggregate demonstrating every layer end-to-end

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
cmd/service/main.go              Entry point — signal context and di.Run only
internal/
  config/                         YAML config (infra) + app: domain tunables
  deps/                           Log / Metrics / Transaction interfaces + stubs
  model/                          Domain types, statuses, events, errors, behavior
  di/                             Lazy composition root (get[T] accessors)
  pkg/
    clients/                      Outbound adapters (error.go + <svc>/{client,dto,errors})
    storage/                      errors.go, filter.go, transaction.go
    storage/postgres/             SQL adapter (squirrel + scany over pgx)
    usecase/<aggregate>/          usecase.go + one file per operation + errors.go
  handler/
    validator/                    Composable request-validation rules
    <endpoint>/                   handler.go (+ handler_test.go)
db/migrations/                   SQL migrations
pkg/                             Reusable infra libs (importable by any service):
  service/                        Service builder + Component interface
  clog/                           Structured JSON logging
  metrics/                        Prometheus registry + Series metric naming
  middleware/                     HTTP middleware
  postgres/                       PostgreSQL component
  transactions/                   Transaction manager
```

The `deps` package is the only seam between `internal/` and the concrete libs:
its interfaces are satisfied structurally by `clog.CLog`, `metrics.Registry`, and
`pkg/transactions`, so domain code never imports them directly.

Use cases never import `clients`, `storage`, `handler`, or `di`. Point-of-use
interfaces accept values from `model`; outbound clients translate those values
to private transport DTOs, and storage adapters translate infrastructure misses
to domain errors. PostgreSQL scans into private, operation-specific row structs
and explicitly maps them into tag-free domain models. The recursive architecture
test rejects any reversal or any struct tag in `internal/model`; committed clean
and db/database/json/yaml/gorm/bson fixtures keep the checker from regressing
into a no-op.

### Service Composition

The service builder handles lifecycle, middleware, and graceful shutdown. DI owns
configuration, concrete dependencies, and route registration; the entrypoint only owns
the process signal context:

```go
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := di.Run(ctx, "config.yaml"); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
```

### Adding an aggregate

Follow the `operations` reference domain: add `model` types → a `storage/postgres`
adapter → a `usecase/<aggregate>` package (declaring its `storage`/client
interfaces at the point of use) → a `handler/<endpoint>` package, then add the
accessors and route in `di`.

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

| Endpoint                              | Port | Description                     |
|---------------------------------------|------|---------------------------------|
| `POST /api/v1/operations`             | 8080 | Create an operation (reference) |
| `GET /api/v1/operations/{external_id}`| 8080 | Fetch an operation (reference)  |
| `/healthz`                            | 8081 | Kubernetes liveness             |
| `/readyz`                             | 8081 | Kubernetes readiness            |
| `/metrics`                            | 8081 | Prometheus metrics              |

The `operations` routes are the worked reference domain — replace them with your
own aggregates.

```bash
# with postgres up (make dc-reup) and the service running (make run):
curl -s -XPOST localhost:8080/api/v1/operations \
  -d '{"type":"payment","user_id":1,"amount":1000}'
# → {"external_id":"...","status":"in_progress"}
```

## Development

```bash
make build           # Build binary
make run             # Build and run
make test            # Unit tests with race detector
make int-test        # Integration tests
make lint            # pinned golangci-lint v2.11.0, matching CI
make codegen         # Generate mocks
make dc-reup         # Restart docker-compose
```

`internal/architecture/structure_test.go` also verifies that the Makefile and
GitHub Actions use the same exact golangci-lint release.

## License

MIT
