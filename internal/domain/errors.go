package domain

import "errors"

// ErrConflict is returned when a repository detects a version conflict during an update.
var ErrConflict = errors.New("conflict")
