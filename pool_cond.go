package jett

import (
	"github.com/micro-go/lock"
	"sync"
	"sync/atomic"
)

// ------------------------------------------------------------
// POOL-COND

func newCondWith(opts Opts) Pool {
	opts = opts.scrub()
	cwg := &countedWaitGroup{}
	cond := newMutexCond()
	p := &poolCond{opts: opts, running: true, cwg: cwg, cond: cond}
	// Start fixed routines
	for i := 0; i < opts.MinWorkers; i++ {
		p.startStaticWorker()
	}
	return p
}

type poolCond struct {
	opts      Opts
	q         queue
	running   bool
	cwg       *countedWaitGroup
	unhandled int64 // Number of messages that are in the system, queued or processing
	cond      *mutexCond
}

func (p *poolCond) Run(f RunFunc) error {
	if f == nil {
		return ErrBadRequest
	}
	atomic.AddInt64(&p.unhandled, 1)
	p.q.push(f)
	p.cond.signal()

	// If we have more unhandled items then workers
	// to handle them and we aren't at the worker limit, add one.
	unhandled := atomic.LoadInt64(&p.unhandled)
	workers := p.cwg.count()
	if unhandled > int64(workers) && int(workers) < p.opts.MaxWorkers {
		p.startDynamicWorker()
	}

	return nil
}

func (p *poolCond) Close() error {
	p.running = false
	p.cond.c.Broadcast()
	p.cwg.wait()
	return nil
}

func (p *poolCond) startStaticWorker() {
	p.cwg.add()

	go staticCondWorker(newCondWorkerArgs(p))
}

func (p *poolCond) startDynamicWorker() {
	p.cwg.add()

	go dynamicCondWorker(newCondWorkerArgs(p))
}

func staticCondWorker(args condWorkerArgs) {
	defer args.cwg.done()

	for *args.running == true {
		workerRunAll(args)
		args.cond.wait()
	}
}

func dynamicCondWorker(args condWorkerArgs) {
	defer args.cwg.done()

	workerRunAll(args)
}

func workerRunAll(args condWorkerArgs) {
	f := args.q.pop()
	for f != nil {
		f()
		atomic.AddInt64(args.unhandled, -1)
		f = args.q.pop()
	}
}

// ------------------------------------------------------------
// NODE

type node struct {
	next *node
	f    RunFunc
}

// ------------------------------------------------------------
// QUEUE

type queue struct {
	mutex sync.Mutex
	first *node
	last  *node
}

func (q *queue) push(f RunFunc) {
	defer lock.Locker(&q.mutex).Unlock()
	n := &node{f: f}
	if q.first == nil {
		q.first = n
		q.last = n
	} else {
		q.last.next = n
		q.last = n
	}
}

func (q *queue) pop() RunFunc {
	defer lock.Locker(&q.mutex).Unlock()
	if q.first == nil {
		return nil
	}
	f := q.first
	q.first = q.first.next
	if q.first == nil {
		q.last = nil
	} else if q.first.next == nil {
		q.last = q.first
	}
	f.next = nil
	return f.f
}

// ------------------------------------------------------------
// COND-WORKER-ARGS

// condWorkerArgs provides arguments to the worker routines.
// Not strictly necessary, but prevents making assumptions
// about the contents of the pool struct.
type condWorkerArgs struct {
	running   *bool
	q         *queue
	unhandled *int64
	cwg       *countedWaitGroup
	cond      *mutexCond
}

func newCondWorkerArgs(p *poolCond) condWorkerArgs {
	return condWorkerArgs{&p.running, &p.q, &p.unhandled, p.cwg, p.cond}
}
