package get_operation_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bdtfs/go-service-template/internal/deps"
	"github.com/bdtfs/go-service-template/internal/handler/get_operation"
	"github.com/bdtfs/go-service-template/internal/model"
	"github.com/bdtfs/go-service-template/internal/pkg/usecase/operation"
)

type fakeUseCase struct {
	out model.Operation
	err error
}

func (f fakeUseCase) GetOperation(context.Context, string) (model.Operation, error) {
	return f.out, f.err
}

func TestHandle_OK(t *testing.T) {
	h := get_operation.New(deps.NewLogStub(), deps.NewMetricsStub(), fakeUseCase{out: model.Operation{
		ExternalID: "ext-1",
		Status:     model.StatusInProgress,
	}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/operations/ext-1", http.NoBody)
	req.SetPathValue("external_id", "ext-1")
	rec := httptest.NewRecorder()

	h.Handle(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"external_id":"ext-1","type":"","status":"in_progress","user_id":0,"amount":0,"description":""}`, rec.Body.String())
}

func TestHandle_NotFound(t *testing.T) {
	h := get_operation.New(deps.NewLogStub(), deps.NewMetricsStub(), fakeUseCase{err: operation.ErrOperationNotFound})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/operations/missing", http.NoBody)
	req.SetPathValue("external_id", "missing")
	rec := httptest.NewRecorder()

	h.Handle(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandle_MissingExternalID(t *testing.T) {
	h := get_operation.New(deps.NewLogStub(), deps.NewMetricsStub(), fakeUseCase{})
	rec := httptest.NewRecorder()

	h.Handle(rec, httptest.NewRequest(http.MethodGet, "/api/v1/operations/", http.NoBody))

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
