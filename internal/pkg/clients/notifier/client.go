package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bdtfs/go-service-template/internal/deps"
	"github.com/bdtfs/go-service-template/internal/model"
	"github.com/bdtfs/go-service-template/internal/pkg/clients"
	"github.com/bdtfs/go-service-template/pkg/metrics"
)

// Client is an outbound adapter for the notification service. It is a worked
// example of the client layer: JSON over net/http, per-operation metrics via a
// Series, and typed errors the use case can branch on.
type Client struct {
	mc      deps.Metrics
	s       metrics.Series
	http    *http.Client
	baseURL string
}

func New(mc deps.Metrics, httpClient *http.Client, baseURL string) *Client {
	return &Client{
		mc:      mc,
		s:       metrics.NewSeries(metrics.SeriesTypeClient, "notifier"),
		http:    httpClient,
		baseURL: baseURL,
	}
}

// NotifyOperationCreated posts an operation-created event. A non-2xx response
// yields clients.ResponseError; a transport failure yields ErrUnavailable.
func (c *Client) NotifyOperationCreated(ctx context.Context, event model.OperationCreated) error {
	ctx, series := c.s.WithOperation(ctx, "operation_created")

	body, err := json.Marshal(operationCreatedRequest{
		ExternalID: event.ExternalID,
		UserID:     event.UserID,
		Amount:     event.Amount,
	})
	if err != nil {
		c.mc.Inc(series.Error("marshal"))
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/events/operation-created", bytes.NewReader(body))
	if err != nil {
		c.mc.Inc(series.Error("build_request"))
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		c.mc.Inc(series.Error("transport"))
		return fmt.Errorf("%w: %s", ErrUnavailable, err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.mc.Inc(series.Error(fmt.Sprintf("status_%d", resp.StatusCode)))
		return clients.NewResponseError(resp.StatusCode)
	}

	c.mc.Inc(series.Success())
	return nil
}
