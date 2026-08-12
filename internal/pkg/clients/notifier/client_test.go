package notifier_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bdtfs/go-service-template/internal/deps"
	"github.com/bdtfs/go-service-template/internal/pkg/clients"
	"github.com/bdtfs/go-service-template/internal/pkg/clients/notifier"
)

func TestClient_NotifyOperationCreated_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/events/operation-created", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := notifier.New(deps.NewMetricsStub(), srv.Client(), srv.URL)
	err := c.NotifyOperationCreated(context.Background(), notifier.OperationCreatedIn{ExternalID: "x", UserID: 1, Amount: 10})
	require.NoError(t, err)
}

func TestClient_NotifyOperationCreated_ResponseError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := notifier.New(deps.NewMetricsStub(), srv.Client(), srv.URL)
	err := c.NotifyOperationCreated(context.Background(), notifier.OperationCreatedIn{ExternalID: "x"})

	var respErr clients.ResponseError
	require.ErrorAs(t, err, &respErr)
	require.Equal(t, http.StatusInternalServerError, respErr.Code())
}
