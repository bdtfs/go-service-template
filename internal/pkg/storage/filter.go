package storage

import (
	"time"

	"github.com/bdtfs/go-service-template/internal/model"
)

// OperationsFilter narrows a GetOperationList query. All fields are optional;
// nil fields are not applied.
type OperationsFilter struct {
	UserID   *int64
	Type     *model.OperationType
	DateFrom *time.Time
	DateTo   *time.Time
}
