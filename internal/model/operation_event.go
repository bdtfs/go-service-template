package model

// OperationCreated is the domain event emitted after an operation is stored.
type OperationCreated struct {
	ExternalID string
	UserID     int64
	Amount     int64
}
