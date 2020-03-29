package jett

// ------------------------------------------------------------
// POOL

// RunFunc defines runnable behaviour clients can send to the pool.
type RunFunc func() error

type Pool interface {
	Run(RunFunc) error
	Close() error
}

// New() answers a new thread pool with the default options.
func New() Pool {
	return NewWith(NewDefaultOpts())
}

// New() answers a new thread pool
func NewWith(opts Opts) Pool {
	return newCondWith(opts)
}
