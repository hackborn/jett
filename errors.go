package jett

import (
	"fmt"
)

// ------------------------------------------------------------
// WRAPPERS

func newCantInitWorker(err error) error {
	return fmt.Errorf("Can't init worker: %w", err)
}

// ------------------------------------------------------------
// CONST and VAR

const (
	badRequestMsg = "Forbidden"
)

var (
	ErrBadRequest = fmt.Errorf(badRequestMsg)
)
