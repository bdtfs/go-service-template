package operation

import (
	"context"
	"fmt"

	"github.com/bdtfs/go-service-template/internal/model"
	"github.com/bdtfs/go-service-template/internal/pkg/clients/notifier"
)

// CreateOperationIn is the input to CreateOperation. ExternalID is optional; a
// UUID is generated when it is empty.
type CreateOperationIn struct {
	ExternalID  string
	Type        model.OperationType
	UserID      int64
	Amount      int64
	Description string
}

// CreateOperation persists a new in-progress operation and notifies downstream.
// It is idempotent on ExternalID and atomic: the operation is only considered
// created if the notification is enqueued in the same transaction.
func (uc *UseCase) CreateOperation(ctx context.Context, in CreateOperationIn) (model.Operation, error) {
	ctx, series := uc.s.WithOperation(ctx, "create_operation")

	if !in.Type.Valid() {
		uc.mc.Inc(series.Error("invalid_type"))
		return model.Operation{}, ErrInvalidType
	}
	if in.Amount <= 0 {
		uc.mc.Inc(series.Error("invalid_amount"))
		return model.Operation{}, ErrInvalidAmount
	}

	externalID := in.ExternalID
	if externalID == "" {
		externalID = uc.uuid().String()
	}

	op := model.Operation{
		ExternalID:  externalID,
		Type:        in.Type,
		Status:      model.StatusInProgress,
		UserID:      in.UserID,
		Amount:      in.Amount,
		Description: in.Description,
	}

	err := uc.trm.Do(ctx, func(ctx context.Context) error {
		if err := uc.st.SaveOperation(ctx, &op); err != nil {
			uc.mc.Inc(series.Error("save_operation"))
			return fmt.Errorf("save operation: %w", err)
		}

		if err := uc.notifier.NotifyOperationCreated(ctx, notifier.OperationCreatedIn{
			ExternalID: op.ExternalID,
			UserID:     op.UserID,
			Amount:     op.Amount,
		}); err != nil {
			uc.mc.Inc(series.Error("notify"))
			return fmt.Errorf("notify operation created: %w", err)
		}

		return nil
	})
	if err != nil {
		uc.log.ErrorCtx(ctx, err, "create operation failed")
		return model.Operation{}, err
	}

	uc.mc.Inc(series.Success())
	return op, nil
}
