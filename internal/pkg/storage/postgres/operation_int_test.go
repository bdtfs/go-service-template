//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"

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

func setup(t *testing.T) *postgres.Storage {
	t.Helper()
	ctx := context.Background()

	pool, err := pgxpool.Connect(ctx, dsn(t))
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, schema)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "TRUNCATE operations")
	require.NoError(t, err)

	return postgres.NewStorage(storage.NewTransactionFactory(pool))
}

func TestStorage_SaveAndGetOperation(t *testing.T) {
	ctx := context.Background()
	st := setup(t)

	op := &model.Operation{
		ExternalID: "11111111-1111-1111-1111-111111111111",
		Type:       model.OperationTypePayment,
		Status:     model.StatusInProgress,
		UserID:     42,
		Amount:     1000,
	}

	require.NoError(t, st.SaveOperation(ctx, op))
	require.False(t, op.CreatedAt.IsZero())

	// SaveOperation is idempotent: a second save on the same external_id is a no-op.
	require.NoError(t, st.SaveOperation(ctx, op))

	got, err := st.GetOperation(ctx, op.ExternalID)
	require.NoError(t, err)
	require.Equal(t, op.ExternalID, got.ExternalID)
	require.Equal(t, model.StatusInProgress, got.Status)
}

func TestStorage_GetOperation_NotFound(t *testing.T) {
	st := setup(t)
	_, err := st.GetOperation(context.Background(), "missing")
	require.ErrorIs(t, err, model.ErrOperationNotFound)
}

func TestStorage_UpdateOperationStatus(t *testing.T) {
	ctx := context.Background()
	st := setup(t)

	op := &model.Operation{
		ExternalID: "22222222-2222-2222-2222-222222222222",
		Type:       model.OperationTypePayment,
		Status:     model.StatusInProgress,
		UserID:     7,
		Amount:     500,
	}
	require.NoError(t, st.SaveOperation(ctx, op))
	require.NoError(t, st.UpdateOperationStatus(ctx, op.ExternalID, model.StatusSuccess))

	got, err := st.GetOperation(ctx, op.ExternalID)
	require.NoError(t, err)
	require.Equal(t, model.StatusSuccess, got.Status)

	err = st.UpdateOperationStatus(ctx, "missing", model.StatusSuccess)
	require.ErrorIs(t, err, model.ErrOperationNotFound)
}
