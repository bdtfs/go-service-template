package deps

import (
	"context"

	"github.com/bdtfs/go-service-template/pkg/transactions"
)

// Transaction is the query surface a storage adapter runs against. It is either
// the ambient transaction (inside TransactionManager.Do) or the connection pool.
type Transaction = transactions.Transaction

// TransactionFactory hands out the current Transaction for a context.
type TransactionFactory interface {
	Transaction(ctx context.Context) Transaction
}

// TransactionManager runs fn inside a single database transaction, committing on
// success and rolling back on error.
type TransactionManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

// TrmStub runs fn without a real transaction. Use it in unit tests.
type TrmStub struct{}

func NewTrmStub() *TrmStub { return &TrmStub{} }

func (t *TrmStub) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
