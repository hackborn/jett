package jett

import (
	"fmt"
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
	fmt.Println("SIGNAL")
	p.cond.broadcast()

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
	fmt.Println("Close 1")
	p.stopAll()
	fmt.Println("Close 2")
	p.cwg.wait()
	fmt.Println("Close 3")
	return nil
}

func (p *poolCond) stopAll() {
	defer lock.Locker(&p.cond.m).Unlock()
	p.running = false
	p.cond.broadcast()
}

func (p *poolCond) startStaticWorker() {
	p.cwg.add()

	go staticCondWorker(newCondWorkerArgs(p), int(p.cwg.count()))
}

func (p *poolCond) startDynamicWorker() {
	p.cwg.add()

	go dynamicCondWorker(newCondWorkerArgs(p), int(p.cwg.count()))
}

func staticCondWorker(args condWorkerArgs, id int) {
	defer fmt.Println("static worker DONE", id)
	defer args.cwg.done()

	fmt.Println("STATIC WAIT 0 id", id)
	for args.isRunning() {
		workerRunAll(args, id)
		fmt.Println("STATIC WAIT 1")

		args.cond.m.Lock()
		if *args.running == true {
			args.cond.c.Wait()
		}
		args.cond.m.Unlock()

		fmt.Println("STATIC WAIT 2")
	}
	fmt.Println("STATIC WAIT DONE")

}

func dynamicCondWorker(args condWorkerArgs, id int) {
	defer fmt.Println("dynamic worker DONE", id)
	defer args.cwg.done()

	workerRunAll(args, id)
}

func workerRunAll(args condWorkerArgs, id int) {
	fmt.Println("workerRun", id, "queue len", args.q.len())
	f := args.q.pop()
	for f != nil {
		fmt.Println("worker", id, "queue len", args.q.len())
		f()
		atomic.AddInt64(args.unhandled, -1)

		// Bail if the pool is done running
		// and we're configured for that.
		if args.opts.CloseImmediately && *args.running == false {
			fmt.Println("worker return")
			return
		}

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

func (q *queue) len() int {
	defer lock.Locker(&q.mutex).Unlock()
	c := 0
	n := q.first
	for n != nil {
		c++
		n = n.next
	}
	return c
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
	opts      Opts
	running   *bool
	q         *queue
	unhandled *int64
	cwg       *countedWaitGroup
	cond      *mutexCond
}

func newCondWorkerArgs(p *poolCond) condWorkerArgs {
	return condWorkerArgs{p.opts, &p.running, &p.q, &p.unhandled, p.cwg, p.cond}
}

func (a condWorkerArgs) isRunning() bool {
	defer lock.Locker(&a.cond.m).Unlock()
	return *a.running == false
}
