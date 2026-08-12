package postgres

import (
	sq "github.com/Masterminds/squirrel"

	"github.com/bdtfs/go-service-template/internal/deps"
)

// psql builds queries with PostgreSQL-style ($1, $2, ...) placeholders.
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Storage is the PostgreSQL adapter. It runs every query against the
// TransactionFactory so the same methods work inside and outside a transaction.
type Storage struct {
	trf deps.TransactionFactory
}

func NewStorage(trf deps.TransactionFactory) *Storage {
	return &Storage{trf: trf}
}
