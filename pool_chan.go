package jett

import (
	"fmt"
)

// ------------------------------------------------------------
// POOL-CHAN

type poolChan struct {
	queue chan func() error
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
	fmt.Println("Run() queue len", len(p.queue), "workers", p.cwg.count())
	if len(p.queue) > int(p.cwg.count()) {
		fmt.Println("Run() startWorker")
		//		p.startWorker()
	}
	return nil
}

func (p *poolChan) Close() error {
	close(p.done)
	p.cwg.wait()
	return nil
}

func (p *poolChan) startWorker() {
	fmt.Println("start worker")
	p.cwg.add()
	go chanWorker(newChanWorkerArgs(p))
}

func chanWorker(args chanWorkerArgs) {
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
// CHAN-WORKER-ARGS

// chanWorkerArgs provides arguments to the worker routines.
// Not strictly necessary, but prevents making assumptions
// about the contents of the pool struct.
type chanWorkerArgs struct {
	queue <-chan func() error
	done  chan struct{}
	cwg   *countedWaitGroup
}

func newChanWorkerArgs(p *poolChan) chanWorkerArgs {
	return chanWorkerArgs{p.queue, p.done, p.cwg}
}
