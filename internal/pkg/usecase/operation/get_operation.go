package operation

import (
	"context"
	"errors"
	"fmt"

	"github.com/bdtfs/go-service-template/internal/model"
)

// GetOperation returns an operation by external id, mapping the model-level
// absence signal to the use-case error.
func (uc *UseCase) GetOperation(ctx context.Context, externalID string) (model.Operation, error) {
	ctx, series := uc.s.WithOperation(ctx, "get_operation")

	op, err := uc.st.GetOperation(ctx, externalID)
	if err != nil {
		if errors.Is(err, model.ErrOperationNotFound) {
			uc.mc.Inc(series.Error("not_found"))
			return model.Operation{}, ErrOperationNotFound
		}
		uc.mc.Inc(series.Error("get_operation"))
		return model.Operation{}, fmt.Errorf("get operation: %w", err)
	}

	uc.mc.Inc(series.Success())
	return *op, nil
}
