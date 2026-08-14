package postgres

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"

	"github.com/bdtfs/go-service-template/internal/model"
	"github.com/bdtfs/go-service-template/internal/pkg/storage"
)

const operationColumns = "external_id, type, status, user_id, amount, description, created_at"

// SaveOperation inserts op idempotently. A repeated external_id is a no-op
// (ON CONFLICT DO NOTHING); on success op.CreatedAt is populated.
func (s *Storage) SaveOperation(ctx context.Context, op *model.Operation) error {
	q, args, err := psql.Insert("operations").
		Columns("external_id", "type", "status", "user_id", "amount", "description").
		Values(op.ExternalID, op.Type, op.Status, op.UserID, op.Amount, op.Description).
		Suffix("ON CONFLICT (external_id) DO NOTHING RETURNING created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	err = s.trf.Transaction(ctx).QueryRow(ctx, q, args...).Scan(&op.CreatedAt)
	var pgErr *pgconn.PgError
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		// ON CONFLICT (external_id) DID NOTHING: the row already exists.
		return nil
	case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation:
		return storage.ErrEntityExists
	case err != nil:
		return fmt.Errorf("failed to execute query: %w", err)
	default:
		return nil
	}
}

// GetOperation returns the operation with the given external_id.
func (s *Storage) GetOperation(ctx context.Context, externalID string) (*model.Operation, error) {
	q, args, err := psql.Select(operationColumns).
		From("operations").
		Where(sq.Eq{"external_id": externalID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var op model.Operation
	err = pgxscan.Get(ctx, s.trf.Transaction(ctx), &op, q, args...)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, model.ErrOperationNotFound
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &op, nil
}

// GetOperationList returns operations matching the filter, newest first.
func (s *Storage) GetOperationList(ctx context.Context, f storage.OperationsFilter, offset, limit uint64) ([]model.Operation, error) {
	var and sq.And
	if f.UserID != nil {
		and = append(and, sq.Eq{"user_id": *f.UserID})
	}
	if f.Type != nil {
		and = append(and, sq.Eq{"type": *f.Type})
	}
	if f.DateFrom != nil {
		and = append(and, sq.GtOrEq{"created_at": *f.DateFrom})
	}
	if f.DateTo != nil {
		and = append(and, sq.Lt{"created_at": *f.DateTo})
	}

	q, args, err := psql.Select(operationColumns).
		From("operations").
		Where(and).
		OrderBy("created_at DESC").
		Offset(offset).
		Limit(limit).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var ops []model.Operation
	if err = pgxscan.Select(ctx, s.trf.Transaction(ctx), &ops, q, args...); err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return ops, nil
}

// UpdateOperationStatus sets a new status.
func (s *Storage) UpdateOperationStatus(ctx context.Context, externalID string, status model.OperationStatus) error {
	q, args, err := psql.Update("operations").
		Set("status", status).
		Where(sq.Eq{"external_id": externalID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	tag, err := s.trf.Transaction(ctx).Exec(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrOperationNotFound
	}

	return nil
}
