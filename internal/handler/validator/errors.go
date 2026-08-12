package validator

import "errors"

var (
	ErrRequestEmpty      = errors.New("request body is empty")
	ErrInvalidPositive   = errors.New("value must be positive")
	ErrInvalidNonNegRule = errors.New("value must not be negative")
	ErrEmptyString       = errors.New("value must not be empty")
	ErrTooLong           = errors.New("value is too long")
)
