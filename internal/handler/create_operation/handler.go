// Package create_operation handles the create-operation HTTP transport.
package create_operation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bdtfs/go-service-template/internal/deps"
	"github.com/bdtfs/go-service-template/internal/handler/validator"
	"github.com/bdtfs/go-service-template/internal/model"
	"github.com/bdtfs/go-service-template/internal/pkg/usecase/operation"
	"github.com/bdtfs/go-service-template/pkg/metrics"
)

var (
	errInvalidUserID = errors.New("user_id must be positive")
	errInvalidAmount = errors.New("amount must be positive")
	errInvalidType   = errors.New("type is required")
)

//go:generate ../../../bin/mockery --name useCase
type useCase interface {
	CreateOperation(ctx context.Context, in operation.CreateOperationIn) (model.Operation, error)
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
		s:   metrics.NewSeries(metrics.SeriesTypeApiHandler, "create_operation"),
		uc:  uc,
	}
}

type request struct {
	ExternalID  string `json:"external_id"`
	Type        string `json:"type"`
	UserID      int64  `json:"user_id"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

type response struct {
	ExternalID string `json:"external_id"`
	Status     string `json:"status"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, series := h.s.WithOperation(r.Context(), "handle")

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.mc.Inc(series.Error("decode"))
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := validator.Check(
		validator.RuleNonEmpty(req.Type, errInvalidType),
		validator.RulePositiveInt(req.UserID, errInvalidUserID),
		validator.RulePositiveInt(req.Amount, errInvalidAmount),
	); err != nil {
		h.mc.Inc(series.Error("validation"))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	op, err := h.uc.CreateOperation(ctx, operation.CreateOperationIn{
		ExternalID:  req.ExternalID,
		Type:        model.OperationType(req.Type),
		UserID:      req.UserID,
		Amount:      req.Amount,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, operation.ErrInvalidType), errors.Is(err, operation.ErrInvalidAmount):
			h.mc.Inc(series.Error("domain_validation"))
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			h.mc.Inc(series.Error("usecase"))
			h.log.ErrorCtx(ctx, err, "create_operation: usecase error")
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	h.mc.Inc(series.Success())
	writeJSON(w, http.StatusCreated, response{ExternalID: op.ExternalID, Status: string(op.Status)})
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
