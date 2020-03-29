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

func newChanWith(opts Opts) Pool {
	opts = opts.scrub()
	queue := make(chan func() error, opts.QueueSize)
	done := make(chan struct{})
	cwg := &countedWaitGroup{}
	p := &poolChan{queue: queue, done: done, cwg: cwg}
	// Start fixed routines
	for i := 0; i < opts.MinWorkers; i++ {
		p.startWorker()
	}
	return p
}

func newCondWith(opts Opts) Pool {
	opts = opts.scrub()
	done := make(chan struct{})
	cwg := &countedWaitGroup{}
	cond := newMutexCond()
	p := &poolCond{done: done, running: true, cwg: cwg, cond: cond}
	// Start fixed routines
	for i := 0; i < opts.MinWorkers; i++ {
		p.startStaticWorker()
	}
	return p
}
