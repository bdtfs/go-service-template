package create_operation_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bdtfs/go-service-template/internal/deps"
	"github.com/bdtfs/go-service-template/internal/handler/create_operation"
	"github.com/bdtfs/go-service-template/internal/model"
	"github.com/bdtfs/go-service-template/internal/pkg/usecase/operation"
)

type fakeUseCase struct {
	out model.Operation
	err error
}

func (f fakeUseCase) CreateOperation(context.Context, operation.CreateOperationIn) (model.Operation, error) {
	return f.out, f.err
}

func TestHandle_Created(t *testing.T) {
	h := create_operation.New(deps.NewLogStub(), deps.NewMetricsStub(), fakeUseCase{
		out: model.Operation{ExternalID: "ext-1", Status: model.StatusInProgress},
	})

	body := `{"type":"payment","user_id":1,"amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/operations", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Handle(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "ext-1", resp["external_id"])
	require.Equal(t, "in_progress", resp["status"])
}

func TestHandle_ValidationError(t *testing.T) {
	h := create_operation.New(deps.NewLogStub(), deps.NewMetricsStub(), fakeUseCase{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/operations", strings.NewReader(`{"type":"payment","user_id":0,"amount":100}`))
	rec := httptest.NewRecorder()

	h.Handle(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandle_BadJSON(t *testing.T) {
	h := create_operation.New(deps.NewLogStub(), deps.NewMetricsStub(), fakeUseCase{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/operations", strings.NewReader(`{not json`))
	rec := httptest.NewRecorder()

	h.Handle(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
