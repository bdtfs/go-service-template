package operation

import (
	"context"
	"errors"
	"fmt"

	"github.com/bdtfs/go-service-template/internal/model"
	storagepkg "github.com/bdtfs/go-service-template/internal/pkg/storage"
)

// GetOperation returns an operation by external id, mapping the storage
// not-found sentinel to the domain-level ErrOperationNotFound.
func (uc *UseCase) GetOperation(ctx context.Context, externalID string) (model.Operation, error) {
	ctx, series := uc.s.WithOperation(ctx, "get_operation")

	op, err := uc.st.GetOperation(ctx, externalID)
	if err != nil {
		if errors.Is(err, storagepkg.ErrEntityNotFound) {
			uc.mc.Inc(series.Error("not_found"))
			return model.Operation{}, ErrOperationNotFound
		}
		uc.mc.Inc(series.Error("get_operation"))
		return model.Operation{}, fmt.Errorf("get operation: %w", err)
	}

	uc.mc.Inc(series.Success())
	return *op, nil
}
