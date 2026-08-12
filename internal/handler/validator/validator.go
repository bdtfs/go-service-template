// Package validator provides small, composable request-validation rules. Each
// rule is a closure returning an error; Check runs them in order and returns the
// first failure. Rules join a specific field error onto a generic cause so both
// the machine-readable sentinel and a human message survive errors.Is/As.
package validator

import "errors"

type Rule func() error

// Check runs rules in order and returns the first non-nil error.
func Check(rules ...Rule) error {
	for _, rule := range rules {
		if err := rule(); err != nil {
			return err
		}
	}
	return nil
}

// Request fails when the decoded request pointer is nil.
func Request[T any](in *T) error {
	if in == nil {
		return ErrRequestEmpty
	}
	return nil
}

// Optional applies next only when v is non-nil.
func Optional[T any](v *T, next Rule) Rule {
	return func() error {
		if v == nil {
			return nil
		}
		return next()
	}
}

// RulePositiveInt requires n > 0.
func RulePositiveInt(n int64, fieldErr error) Rule {
	return func() error {
		if n <= 0 {
			return errors.Join(ErrInvalidPositive, fieldErr)
		}
		return nil
	}
}

// RuleNonEmpty requires a non-empty string.
func RuleNonEmpty(s string, fieldErr error) Rule {
	return func() error {
		if s == "" {
			return errors.Join(ErrEmptyString, fieldErr)
		}
		return nil
	}
}

// RuleMaxLen requires len(s) <= max.
func RuleMaxLen(s string, max int, fieldErr error) Rule {
	return func() error {
		if len(s) > max {
			return errors.Join(ErrTooLong, fieldErr)
		}
		return nil
	}
}
