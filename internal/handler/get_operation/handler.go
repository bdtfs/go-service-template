// Package get_operation handles the get-operation HTTP transport.
package get_operation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bdtfs/go-service-template/internal/deps"
	"github.com/bdtfs/go-service-template/internal/model"
	"github.com/bdtfs/go-service-template/internal/pkg/usecase/operation"
	"github.com/bdtfs/go-service-template/pkg/metrics"
)

//go:generate ../../../bin/mockery --name useCase
type useCase interface {
	GetOperation(ctx context.Context, externalID string) (model.Operation, error)
}

type Handler struct {
	log deps.Log
	mc  deps.Metrics
	s   metrics.Series
	uc  useCase
}

func New(log deps.Log, mc deps.Metrics, uc useCase) *Handler {
	return &Handler{
		log: log,
		mc:  mc,
		s:   metrics.NewSeries(metrics.SeriesTypeApiHandler, "get_operation"),
		uc:  uc,
	}
}

type response struct {
	ExternalID  string `json:"external_id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	UserID      int64  `json:"user_id"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, series := h.s.WithOperation(r.Context(), "handle")

	externalID := r.PathValue("external_id")
	if externalID == "" {
		h.mc.Inc(series.Error("validation"))
		writeError(w, http.StatusBadRequest, "external_id is required")
		return
	}

	op, err := h.uc.GetOperation(ctx, externalID)
	if err != nil {
		switch {
		case errors.Is(err, operation.ErrOperationNotFound):
			h.mc.Inc(series.Error("not_found"))
			writeError(w, http.StatusNotFound, "operation not found")
		default:
			h.mc.Inc(series.Error("usecase"))
			h.log.ErrorCtx(ctx, err, "get_operation: usecase error")
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	h.mc.Inc(series.Success())
	writeJSON(w, http.StatusOK, response{
		ExternalID:  op.ExternalID,
		Type:        string(op.Type),
		Status:      string(op.Status),
		UserID:      op.UserID,
		Amount:      op.Amount,
		Description: op.Description,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		return
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
