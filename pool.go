package jett

import (
	"fmt"
)

// ------------------------------------------------------------
// POOL

type Pool interface {
	Run(func() error) error
	Close() error
}

// New() answers a new thread pool with the default options.
func New() Pool {
	return NewWith(NewDefaultOpts())
}

// New() answers a new thread pool
func NewWith(opts Opts) Pool {
	opts = opts.scrub()
	queue := make(chan func() error, opts.QueueSize)
	done := make(chan struct{})
	cwg := &countedWaitGroup{}
	p := &pool{queue: queue, done: done, cwg: cwg}
	// Start fixed routines
	for i := 0; i < opts.MinWorkers; i++ {
		p.startWorker()
	}
	return p
}

// ------------------------------------------------------------
// POOL IMPLEMENTATION

type pool struct {
	queue chan func() error
	done  chan struct{}
	cwg   *countedWaitGroup
}

func (p *pool) Run(f func() error) error {
	if f == nil {
		return ErrBadRequest
	}
	p.queue <- f
	// If the queue is backed up and we have some
	// headroom, spin up a new routine.
	fmt.Println("queue len", len(p.queue), "workers", p.cwg.count())
	if len(p.queue) > int(p.cwg.count()) {
		p.startWorker()
	}
	return nil
}

func (p *pool) Close() error {
	close(p.done)
	p.cwg.wait()
	return nil
}

func (p *pool) startWorker() {
	fmt.Println("start worker")
	p.cwg.add()
	go worker(newWorkerArgs(p))
}

func worker(args workerArgs) {
	defer args.cwg.done()

	for {
		select {
		case <-args.done:
			return
		case f, more := <-args.queue:
			if more {
				f()
				//				fmt.Println("msg", msg)
			}
		}
	}
}

// ------------------------------------------------------------
// WORKER-ARGS

// workerArgs provides arguments to the worker routines.
// Not strictly necessary, but prevents making assumptions
// about the contents of the pool struct.
type workerArgs struct {
	queue <-chan func() error
	done  chan struct{}
	cwg   *countedWaitGroup
}

func newWorkerArgs(p *pool) workerArgs {
	return workerArgs{p.queue, p.done, p.cwg}
}
