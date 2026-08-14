package operation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bdtfs/go-service-template/internal/model"
	"github.com/bdtfs/go-service-template/internal/pkg/usecase/operation"
)

func TestGetOperation_MapsDomainNotFound(t *testing.T) {
	uc := newUseCase(&fakeStorage{getErr: model.ErrOperationNotFound}, &fakeNotifier{})

	_, err := uc.GetOperation(context.Background(), "missing")

	require.ErrorIs(t, err, operation.ErrOperationNotFound)
}

func TestGetOperation_PreservesUnexpectedFailure(t *testing.T) {
	storageErr := errors.New("storage unavailable")
	uc := newUseCase(&fakeStorage{getErr: storageErr}, &fakeNotifier{})

	_, err := uc.GetOperation(context.Background(), "ext-1")

	require.ErrorIs(t, err, storageErr)
}
