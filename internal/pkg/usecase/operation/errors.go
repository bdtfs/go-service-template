package operation

import "errors"

var (
	ErrOperationNotFound = errors.New("operation not found")
	ErrInvalidType       = errors.New("invalid operation type")
	ErrInvalidAmount     = errors.New("amount must be positive")
)
