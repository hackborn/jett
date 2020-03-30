package jett

import (
	"context"
)

// ------------------------------------------------------------
// OPTS

// Opts provides options when creating new pools.
type Opts struct {
	MinWorkers       int            // Minimum number of go routines.
	MaxWorkers       int            // Maximum number of go routines.
	QueueSize        int            // Number of items in the queue before we block.
	CloseImmediately bool           // If true, calling Close() will return as soon as possible, potentially leaving messages unprocessed. If false, process all messages before returning.
	WorkerInit       NewContextFunc // Optional. If not nil, generate a new context on each worker routine that gets passed to all RunFuncs that are performed on that routine. Must be thread safe.
}

func NewDefaultOpts() Opts {
	return Opts{MinWorkers: 1, MaxWorkers: 10, QueueSize: 256, CloseImmediately: false}
}

// scrub() returns the options with any illegal
// values replaced with valid values.
func (o Opts) scrub() Opts {
	if o.MinWorkers < 1 {
		o.MinWorkers = 1
	}
	if o.MaxWorkers < o.MinWorkers {
		o.MaxWorkers = o.MinWorkers
	}
	if o.QueueSize < 1 {
		o.QueueSize = 1
	}
	return o
}

func (o Opts) runWorkerInit() (context.Context, error) {
	if o.WorkerInit == nil {
		return nil, nil
	}
	return o.WorkerInit()
}
