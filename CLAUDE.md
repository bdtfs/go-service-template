# Go Service Template

## Project Structure

A layered Go microservice template following the Avito Wallet service conventions:
a thin `deps` seam decoupling domain code from infrastructure, per-aggregate use
cases, a squirrel-based postgres adapter, per-endpoint handler packages, and a
lazy DI container. Reusable infrastructure lives in `pkg/`; the private service
graph lives in `internal/`.

```
cmd/service/main.go            Entry point — signal context and di.Run only
internal/
  config/                       YAML config (infra) + app: section (domain tunables)
  deps/                         Log / Metrics / Transaction interfaces + stubs
  model/                        Domain types, statuses, db tags, behavior
  di/                           Lazy composition root (get[T] accessors)
  pkg/                          Private service packages:
    clients/                    Outbound adapters (error.go + <svc>/{client,dto,errors})
    storage/                    errors.go, filter.go, transaction.go
    storage/postgres/           SQL adapter (squirrel + scany over pgx)
    usecase/<aggregate>/        usecase.go + one file per operation + errors.go
  handler/                      Inbound adapters, one package per endpoint:
    validator/                  Composable request-validation rules
    <endpoint>/                 handler.go (+ handler_test.go)
db/migrations/                  SQL migrations
pkg/                            Reusable infra libs (importable by any service):
  service/                      Service builder & Component interface
  clog/                         Structured JSON logging (wraps slog)
  metrics/                      Prometheus registry + Series metric naming
  middleware/                   HTTP middleware (recovery, request-id, logging, metrics)
  postgres/                     PostgreSQL component (pgx/v4)
  transactions/                 Transaction manager pattern
```

The worked reference domain is **operations** (create + get). Follow it as the
pattern for new aggregates: add `model` types, a `storage/postgres` adapter, a
`usecase/<aggregate>` package, a `handler/<endpoint>` package, then wire them in
`di` and register the route in `di.Container.RegisterRoutes`.

## The layers

- **deps** — the only seam between `internal/` and the concrete libs. Interfaces
  (`Log`, `Metrics`, `TransactionFactory/Manager`) that `clog.CLog`,
  `metrics.Registry`, and `pkg/transactions` satisfy structurally, plus no-op
  stubs for unit tests. Domain code imports `deps`, never `pkg/clog` etc.
- **model** — pure domain types with `db:` tags and behavior (`IsTerminal`). No
  I/O.
- **storage** — `errors.go`/`filter.go` shared types; `postgres/` builds SQL with
  squirrel (`$`-placeholders) and scans with scany. Every query runs against the
  `TransactionFactory` so the same method works inside and outside a transaction.
- **clients** — outbound adapters. `clients/error.go` holds shared
  `ResponseError`/`ValidationError`; each `<svc>/` has `client.go`, `dto.go`,
  `errors.go`.
- **usecase** — orchestration. Interfaces (`storage`, `notifierClient`, ...) are
  declared **at the point of use** in `usecase.go` with `//go:generate mockery`.
  Multi-write flows are wrapped in `trm.Do`.
- **handler** — one package per endpoint. Decodes/validates the request, calls
  the use case, maps domain errors to HTTP status codes.
- **di** — constructs infrastructure, registers routes, and exposes one
  `get()`-guarded accessor per dependency. `cmd` only calls `di.Run`.

## Metrics: the Series pattern

Every layer instruments itself through a `pkg/metrics.Series` created in its
constructor and advanced with `WithOperation`:

```go
s := metrics.NewSeries(metrics.SeriesTypeUseCase, "operation")   // in New()
ctx, series := uc.s.WithOperation(ctx, "create_operation")       // per call
uc.mc.Inc(series.Error("save_operation"))                        // on failure
uc.mc.Inc(series.Success())                                      // on success
```

`series.Success()/Error()` return `(name, prometheus.Labels)` which pass straight
into `deps.Metrics.Inc`.

## Key Commands

```bash
make build          # Build binary to ./bin/service
make run            # Build and run (needs postgres — see make dc-reup)
make test           # Unit tests with race detector
make int-test       # Integration tests (build tag: integration, needs postgres)
make lint           # golangci-lint
make codegen        # Generate mocks (mockery)
make dc-reup        # Start postgres via docker-compose
```

## Conventions

- **Point-of-use interfaces** — a consumer declares the narrow interface it needs
  (`storage`, `useCase`, ...) and depends on that, not the concrete type.
- **deps seam** — `internal/` imports `deps`, never the concrete infra libs.
- **Context propagation** — thread `ctx` (and the `Series`) through every layer.
- **Idempotency & transactions** — writes go through `TransactionManager.Do`;
  inserts use `ON CONFLICT` where the key is caller-supplied.
- **Error mapping** — storage returns sentinels (`ErrEntityNotFound`); use cases
  translate to domain errors (`ErrOperationNotFound`); handlers translate to HTTP.
- **No comments restating code** — comments state constraints, not narration.
- **Thin entrypoints** — `cmd/<binary>` owns signals and calls `di.Run`; no config,
  transport, storage, or use-case construction is allowed there.
- `pkg/` must NOT import from `internal/` (except `pkg/service`, which reads
  `internal/config`).

## Testing

```bash
go test -race ./...                    # Unit tests
TEST_DATABASE_URL=postgresql://... go test -tags=integration -race ./...
```

Unit tests use the `deps` stubs and small hand-written fakes for the point-of-use
interfaces so `go test ./...` is green without a codegen step. In a real service,
prefer the mockery-generated mocks (`make codegen`) referenced by the
`//go:generate` directives. Storage is covered by integration tests
(`operation_int_test.go`, `//go:build integration`) against a real postgres.
`internal/architecture/structure_test.go` enforces entrypoint size, root hygiene,
per-operation handler tests, and dependency direction.
