package transactions

import (
	"context"
	"errors"
)

type PgTransactionManager struct {
	trf *PgTransactionFactory
}

func NewPgTransactionManager(trf *PgTransactionFactory) TransactionManager {
	return &PgTransactionManager{
		trf: trf,
	}
}

// Do function creates Tx for the given tx factory and stores it in the context
func (tm *PgTransactionManager) Do(ctx context.Context, fn func(context.Context) error) error {
	tx, err := tm.trf.Begin(ctx)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, txKey, tx)

	err = fn(ctx)
	if err != nil {
		return errors.Join(err, tx.Rollback(ctx))
	}

	return tx.Commit(ctx)
}
