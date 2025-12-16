package domain

import "errors"

// ErrConflict is returned when a repository detects a version conflict during an update.
var ErrConflict = errors.New("conflict")

// ErrConstraintViolation is returned when a database constraint is violated (e.g., unique constraint).
var ErrConstraintViolation = errors.New("constraint violation")
