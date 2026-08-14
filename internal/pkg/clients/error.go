package clients

import "fmt"

// ResponseError is returned when a downstream service replies with a non-2xx
// status. Code carries the HTTP status (or a domain code) for the caller to
// branch on.
type ResponseError struct {
	code int
}

func NewResponseError(code int) ResponseError {
	return ResponseError{code: code}
}

func (e ResponseError) Error() string { return fmt.Sprintf("response code %d", e.code) }

func (e ResponseError) Code() int { return e.code }

// ValidationError is returned when the request is rejected before it leaves the
// client (bad input) or the downstream reports a validation failure.
type ValidationError struct {
	message string
}

func NewValidationError(msg string) ValidationError {
	return ValidationError{message: msg}
}

func (e ValidationError) Error() string { return fmt.Sprintf("validation: %s", e.message) }
