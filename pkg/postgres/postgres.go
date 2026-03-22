package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

const ComponentName = "postgres"

// Component implements service.Component for PostgreSQL.
type Component struct {
	pool *pgxpool.Pool
	dsn  string
}

// NewComponent creates a postgres component with the given DSN.
// The connection is established when Init is called.
func NewComponent(dsn string) *Component {
	return &Component{dsn: dsn}
}

func (c *Component) Name() string { return ComponentName }

func (c *Component) Init(ctx context.Context) error {
	poolConfig, err := pgxpool.ParseConfig(c.dsn)
	if err != nil {
		return fmt.Errorf("parsing postgres config: %w", err)
	}

	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("connecting to postgres: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return fmt.Errorf("pinging postgres: %w", err)
	}

	c.pool = pool
	return nil
}

func (c *Component) Close(_ context.Context) error {
	if c.pool != nil {
		c.pool.Close()
	}
	return nil
}

// HealthCheck pings the database to verify connectivity.
func (c *Component) HealthCheck(ctx context.Context) error {
	if c.pool == nil {
		return fmt.Errorf("postgres pool not initialized")
	}
	return c.pool.Ping(ctx)
}

// Pool returns the underlying connection pool.
// Only valid after Init has been called.
func (c *Component) Pool() *pgxpool.Pool {
	return c.pool
}
