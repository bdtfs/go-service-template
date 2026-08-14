package model

import "time"

// OperationType enumerates the kinds of business operation this service tracks.
type OperationType string

const (
	OperationTypePayment OperationType = "payment"
	OperationTypeRefund  OperationType = "refund"
)

func (t OperationType) Valid() bool {
	switch t {
	case OperationTypePayment, OperationTypeRefund:
		return true
	default:
		return false
	}
}

// OperationStatus is the lifecycle state of an Operation.
type OperationStatus string

const (
	StatusInProgress OperationStatus = "in_progress"
	StatusSuccess    OperationStatus = "success"
	StatusFailed     OperationStatus = "failed"
)

// IsTerminal reports whether the status can no longer change.
func (s OperationStatus) IsTerminal() bool {
	return s == StatusSuccess || s == StatusFailed
}

// Operation is the aggregate root of the reference domain. ExternalID is the
// caller-supplied idempotency key.
type Operation struct {
	ExternalID  string
	Type        OperationType
	Status      OperationStatus
	UserID      int64
	Amount      int64
	Description string
	CreatedAt   time.Time
}
