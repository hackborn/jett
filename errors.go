package jett

import (
	"errors"
)

// ------------------------------------------------------------
// CONST and VAR

const (
	badRequestMsg = "Forbidden"
)

var (
	ErrBadRequest = errors.New(badRequestMsg)
)
