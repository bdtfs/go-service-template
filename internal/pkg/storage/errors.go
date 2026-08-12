package storage

import "errors"

var (
	ErrEntityNotFound = errors.New("entity not found")
	ErrEntityExists   = errors.New("entity exists")
)
