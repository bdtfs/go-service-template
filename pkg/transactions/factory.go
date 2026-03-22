package transactions

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PgTransactionFactory struct {
	pool *pgxpool.Pool
}

func NewPgTransactionFactory(pool *pgxpool.Pool) TransactionFactory {
	return &PgTransactionFactory{pool: pool}
}

func (f *PgTransactionFactory) Begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := f.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (f *PgTransactionFactory) Transaction(ctx context.Context) Transaction {
	tx, ok := ctx.Value(txKey).(Transaction)
	if !ok {
		return f.pool
	}
	return tx
}
