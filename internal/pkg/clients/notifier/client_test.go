package notifier_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bdtfs/go-service-template/internal/deps"
	"github.com/bdtfs/go-service-template/internal/model"
	"github.com/bdtfs/go-service-template/internal/pkg/clients"
	"github.com/bdtfs/go-service-template/internal/pkg/clients/notifier"
)

func TestClient_NotifyOperationCreated_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/events/operation-created" {
			t.Errorf("path = %s, want /events/operation-created", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		expected := map[string]any{
			"external_id": "x",
			"user_id":     float64(1),
			"amount":      float64(10),
		}
		if !reflect.DeepEqual(body, expected) {
			t.Errorf("body = %#v, want %#v", body, expected)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := notifier.New(deps.NewMetricsStub(), srv.Client(), srv.URL)
	err := c.NotifyOperationCreated(context.Background(), model.OperationCreated{ExternalID: "x", UserID: 1, Amount: 10})
	require.NoError(t, err)
}

func TestClient_NotifyOperationCreated_ResponseError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := notifier.New(deps.NewMetricsStub(), srv.Client(), srv.URL)
	err := c.NotifyOperationCreated(context.Background(), model.OperationCreated{ExternalID: "x"})

	var respErr clients.ResponseError
	require.ErrorAs(t, err, &respErr)
	require.Equal(t, http.StatusInternalServerError, respErr.Code())
}
