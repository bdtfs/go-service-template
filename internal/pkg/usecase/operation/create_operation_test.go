package operation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/bdtfs/go-service-template/internal/deps"
	"github.com/bdtfs/go-service-template/internal/model"
	"github.com/bdtfs/go-service-template/internal/pkg/clients/notifier"
	"github.com/bdtfs/go-service-template/internal/pkg/usecase/operation"
)

// The example tests use small hand-written fakes so `go test ./...` passes
// without a codegen step. In real services, prefer the mockery-generated mocks
// (`make codegen`) referenced by the //go:generate directives in usecase.go.

type fakeStorage struct {
	saved   *model.Operation
	saveErr error
}

func (f *fakeStorage) SaveOperation(_ context.Context, op *model.Operation) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = op
	return nil
}

func (f *fakeStorage) GetOperation(context.Context, string) (*model.Operation, error) {
	return f.saved, nil
}

func (f *fakeStorage) UpdateOperationStatus(context.Context, string, model.OperationStatus) error {
	return nil
}

type fakeNotifier struct {
	called bool
	err    error
}

func (f *fakeNotifier) NotifyOperationCreated(context.Context, notifier.OperationCreatedIn) error {
	f.called = true
	return f.err
}

func newUseCase(st *fakeStorage, nc *fakeNotifier) *operation.UseCase {
	return operation.New(
		deps.NewTrmStub(),
		deps.NewLogStub(),
		deps.NewMetricsStub(),
		st,
		nc,
		func() uuid.UUID { return uuid.Nil },
	)
}

func TestCreateOperation_Success(t *testing.T) {
	st, nc := &fakeStorage{}, &fakeNotifier{}
	uc := newUseCase(st, nc)

	op, err := uc.CreateOperation(context.Background(), operation.CreateOperationIn{
		ExternalID: "ext-1",
		Type:       model.OperationTypePayment,
		UserID:     10,
		Amount:     500,
	})

	require.NoError(t, err)
	require.Equal(t, model.StatusInProgress, op.Status)
	require.Equal(t, "ext-1", op.ExternalID)
	require.True(t, nc.called)
	require.NotNil(t, st.saved)
}

func TestCreateOperation_GeneratesExternalID(t *testing.T) {
	st, nc := &fakeStorage{}, &fakeNotifier{}
	uc := newUseCase(st, nc)

	op, err := uc.CreateOperation(context.Background(), operation.CreateOperationIn{
		Type:   model.OperationTypeRefund,
		UserID: 1,
		Amount: 1,
	})

	require.NoError(t, err)
	require.Equal(t, uuid.Nil.String(), op.ExternalID)
}

func TestCreateOperation_Validation(t *testing.T) {
	uc := newUseCase(&fakeStorage{}, &fakeNotifier{})

	_, err := uc.CreateOperation(context.Background(), operation.CreateOperationIn{
		Type:   "bogus",
		Amount: 100,
	})
	require.ErrorIs(t, err, operation.ErrInvalidType)

	_, err = uc.CreateOperation(context.Background(), operation.CreateOperationIn{
		Type:   model.OperationTypePayment,
		Amount: 0,
	})
	require.ErrorIs(t, err, operation.ErrInvalidAmount)
}

func TestCreateOperation_NotifierFailureRollsBack(t *testing.T) {
	st := &fakeStorage{}
	nc := &fakeNotifier{err: errors.New("boom")}
	uc := newUseCase(st, nc)

	_, err := uc.CreateOperation(context.Background(), operation.CreateOperationIn{
		ExternalID: "ext-2",
		Type:       model.OperationTypePayment,
		UserID:     1,
		Amount:     10,
	})
	require.Error(t, err)
	require.True(t, nc.called)
}
