package jett

// ------------------------------------------------------------
// POOL

// Pool defines a worker pool for running operations.
type Pool interface {
	// Schedule the RunFunc on one of the worker routines.
	// The func will be passed a Context only if an Opts.WorkerInit
	// was supplied when the pool was created.
	Run(RunFunc) error

	// Close the pool. Either wait for pending operations to
	// finish or discard them, based on the Opts.
	Close() error
}

// New() answers a new thread pool with the default options.
func New() Pool {
	return NewWith(NewDefaultOpts())
}

// New() answers a new thread pool with the desired options.
func NewWith(opts Opts) Pool {
	return newCondWith(opts)
}
