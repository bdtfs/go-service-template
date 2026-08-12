package storage

import (
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/bdtfs/go-service-template/internal/deps"
	"github.com/bdtfs/go-service-template/pkg/transactions"
)

// NewTransactionFactory returns a TransactionFactory backed by the pool. The
// factory hands storage adapters the ambient transaction when one is open (see
// TransactionManager.Do), or the pool otherwise.
func NewTransactionFactory(pool *pgxpool.Pool) deps.TransactionFactory {
	return transactions.NewPgTransactionFactory(pool)
}

// NewTransactionManager returns a TransactionManager backed by the pool. Wrap
// multi-write use-case steps in TransactionManager.Do to make them atomic.
func NewTransactionManager(pool *pgxpool.Pool) deps.TransactionManager {
	f, _ := transactions.NewPgTransactionFactory(pool).(*transactions.PgTransactionFactory)
	return transactions.NewPgTransactionManager(f)
}
