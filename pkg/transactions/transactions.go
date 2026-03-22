package transactions

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type contextKey string

const txKey contextKey = "tx"

type Transaction interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type TransactionFactory interface {
	Transaction(ctx context.Context) Transaction
}

type TransactionManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
