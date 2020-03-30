package jett

import (
	"context"
)

// ------------------------------------------------------------
// RUN-FUNC

// RunFunc defines runnable behaviour clients can send to the pool.
type RunFunc func(context.Context) error

// ------------------------------------------------------------
// NEW-CONTEXT-FUNC

// NewContextFunc generates a new context.
type NewContextFunc func() (context.Context, error)
