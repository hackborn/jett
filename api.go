package jett

// ------------------------------------------------------------
// POOL

// RunFunc defines runnable behaviour clients can send to the pool.
type RunFunc func() error

// Pool defines a worker pool for running operations.
type Pool interface {
	Run(RunFunc) error
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
