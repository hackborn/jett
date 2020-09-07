package jett

import (
	"fmt"
)

// ------------------------------------------------------------
// POOL-CHAN

func newChanWith(opts Opts) Pool {
	opts = opts.scrub()
	queue := make(chan RunFunc, opts.QueueSize)
	done := make(chan struct{})
	cwg := &countedWaitGroup{}
	p := &poolChan{opts: opts, queue: queue, done: done, cwg: cwg}
	// Start fixed routines
	for i := 0; i < opts.MinWorkers; i++ {
		p.startWorker()
	}
	return p
}

type poolChan struct {
	opts  Opts
	queue chan RunFunc
	done  chan struct{}
	cwg   *countedWaitGroup
}

func (p *poolChan) Run(f RunFunc) error {
	if f == nil {
		return ErrBadRequest
	}
	p.queue <- f
	// If the queue is backed up and we have some
	// headroom, spin up a new routine.
	//	fmt.Println("Run() queue len", len(p.queue), "workers", p.cwg.count())
	if len(p.queue) > int(p.cwg.count()) {
		//		fmt.Println("Run() startWorker")
		p.startWorker()
	}
	return nil
}

func (p *poolChan) Close() error {
	close(p.done)
	p.cwg.wait()
	return nil
}

func (p *poolChan) startWorker() {
	p.cwg.add()
	go chanWorker(newChanWorkerArgs(p))
}

func chanWorker(args chanWorkerArgs) {
	defer args.cwg.done()

	ctx, err := args.opts.runWorkerInit()
	if err != nil {
		fmt.Println(newCantInitWorker(err))
		return
	}

	for {
		select {
		case <-args.done:
			return
		case f, more := <-args.queue:
			if more {
				err = f(ctx)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

// ------------------------------------------------------------
// CHAN-WORKER-ARGS

// chanWorkerArgs provides arguments to the worker routines.
// Not strictly necessary, but prevents making assumptions
// about the contents of the pool struct.
type chanWorkerArgs struct {
	opts  Opts
	queue <-chan RunFunc
	done  chan struct{}
	cwg   *countedWaitGroup
}

func newChanWorkerArgs(p *poolChan) chanWorkerArgs {
	return chanWorkerArgs{p.opts, p.queue, p.done, p.cwg}
}
