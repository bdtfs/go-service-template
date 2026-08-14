//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/bdtfs/go-service-template/internal/model"
	"github.com/bdtfs/go-service-template/internal/pkg/storage"
	"github.com/bdtfs/go-service-template/internal/pkg/storage/postgres"
)

const schema = `
CREATE TABLE IF NOT EXISTS operations (
    external_id TEXT PRIMARY KEY,
    type        TEXT        NOT NULL,
    status      TEXT        NOT NULL,
    user_id     BIGINT      NOT NULL,
    amount      BIGINT      NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);`

func dsn(t *testing.T) string {
	t.Helper()
	v := os.Getenv("TEST_DATABASE_URL")
	if v == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	return v
}

func setup(t *testing.T) (*postgres.Storage, *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	pool, err := pgxpool.Connect(ctx, dsn(t))
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, schema)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "TRUNCATE operations")
	require.NoError(t, err)

	return postgres.NewStorage(storage.NewTransactionFactory(pool)), pool
}

func TestStorage_SaveAndGetOperation(t *testing.T) {
	ctx := context.Background()
	st, _ := setup(t)

	op := &model.Operation{
		ExternalID:  "11111111-1111-1111-1111-111111111111",
		Type:        model.OperationTypePayment,
		Status:      model.StatusInProgress,
		UserID:      42,
		Amount:      1000,
		Description: "round trip",
	}

	require.NoError(t, st.SaveOperation(ctx, op))
	require.False(t, op.CreatedAt.IsZero())

	// SaveOperation is idempotent: a second save on the same external_id is a no-op.
	require.NoError(t, st.SaveOperation(ctx, op))

	got, err := st.GetOperation(ctx, op.ExternalID)
	require.NoError(t, err)
	require.Equal(t, op.ExternalID, got.ExternalID)
	require.Equal(t, op.Type, got.Type)
	require.Equal(t, op.Status, got.Status)
	require.Equal(t, op.UserID, got.UserID)
	require.Equal(t, op.Amount, got.Amount)
	require.Equal(t, op.Description, got.Description)
	require.True(t, got.CreatedAt.Equal(op.CreatedAt))
}

func TestStorage_GetOperation_NotFound(t *testing.T) {
	st, _ := setup(t)
	_, err := st.GetOperation(context.Background(), "missing")
	require.ErrorIs(t, err, model.ErrOperationNotFound)
}

func TestStorage_UpdateOperationStatus(t *testing.T) {
	ctx := context.Background()
	st, _ := setup(t)

	op := &model.Operation{
		ExternalID:  "22222222-2222-2222-2222-222222222222",
		Type:        model.OperationTypePayment,
		Status:      model.StatusInProgress,
		UserID:      7,
		Amount:      500,
		Description: "preserved on update",
	}
	require.NoError(t, st.SaveOperation(ctx, op))
	require.NoError(t, st.UpdateOperationStatus(ctx, op.ExternalID, model.StatusSuccess))

	got, err := st.GetOperation(ctx, op.ExternalID)
	require.NoError(t, err)
	require.Equal(t, model.StatusSuccess, got.Status)
	require.Equal(t, op.ExternalID, got.ExternalID)
	require.Equal(t, op.Type, got.Type)
	require.Equal(t, op.UserID, got.UserID)
	require.Equal(t, op.Amount, got.Amount)
	require.Equal(t, op.Description, got.Description)
	require.True(t, got.CreatedAt.Equal(op.CreatedAt))

	err = st.UpdateOperationStatus(ctx, "missing", model.StatusSuccess)
	require.ErrorIs(t, err, model.ErrOperationNotFound)
}

func TestStorage_GetOperationList_MapsRowsAndPreservesOrder(t *testing.T) {
	ctx := context.Background()
	st, pool := setup(t)
	userID := int64(42)
	opType := model.OperationTypePayment
	olderAt := time.Date(2026, time.August, 14, 8, 0, 0, 0, time.UTC)
	newerAt := olderAt.Add(time.Hour)

	operations := []*model.Operation{
		{
			ExternalID:  "33333333-3333-3333-3333-333333333333",
			Type:        opType,
			Status:      model.StatusInProgress,
			UserID:      userID,
			Amount:      300,
			Description: "older",
		},
		{
			ExternalID:  "44444444-4444-4444-4444-444444444444",
			Type:        opType,
			Status:      model.StatusSuccess,
			UserID:      userID,
			Amount:      400,
			Description: "newer",
		},
		{
			ExternalID:  "55555555-5555-5555-5555-555555555555",
			Type:        model.OperationTypeRefund,
			Status:      model.StatusFailed,
			UserID:      99,
			Amount:      500,
			Description: "filtered out",
		},
	}
	for _, op := range operations {
		require.NoError(t, st.SaveOperation(ctx, op))
	}
	_, err := pool.Exec(ctx, "UPDATE operations SET created_at = $1 WHERE external_id = $2", olderAt, operations[0].ExternalID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "UPDATE operations SET created_at = $1 WHERE external_id = $2", newerAt, operations[1].ExternalID)
	require.NoError(t, err)

	got, err := st.GetOperationList(ctx, storage.OperationsFilter{UserID: &userID, Type: &opType}, 0, 10)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, operations[1].ExternalID, got[0].ExternalID)
	require.Equal(t, operations[1].Type, got[0].Type)
	require.Equal(t, operations[1].Status, got[0].Status)
	require.Equal(t, operations[1].UserID, got[0].UserID)
	require.Equal(t, operations[1].Amount, got[0].Amount)
	require.Equal(t, operations[1].Description, got[0].Description)
	require.True(t, got[0].CreatedAt.Equal(newerAt))
	require.Equal(t, operations[0].ExternalID, got[1].ExternalID)
	require.Equal(t, operations[0].Description, got[1].Description)
	require.True(t, got[1].CreatedAt.Equal(olderAt))
}
