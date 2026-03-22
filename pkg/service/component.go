package service

import "context"

// Component represents an infrastructure component (database, cache, queue, etc.)
// that participates in the service lifecycle.
type Component interface {
	// Name returns the component identifier (e.g., "postgres", "redis").
	Name() string
	// Init initializes the component. Called during service startup.
	Init(ctx context.Context) error
	// Close shuts down the component. Called during service shutdown.
	Close(ctx context.Context) error
}

// HealthChecker is optionally implemented by components that support health checks.
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}
