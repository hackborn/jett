package jett

import (
	"sync"
	"sync/atomic"
)

// countedWaitGroup tracks the number of items in the
// wait group. Be super nice if you could just ask the wait group.
type countedWaitGroup struct {
	_wg    sync.WaitGroup
	_count int32
}

func (c *countedWaitGroup) count() int32 {
	return atomic.LoadInt32(&c._count)
}

func (c *countedWaitGroup) add() {
	c._wg.Add(1)
	atomic.AddInt32(&c._count, 1)
}

func (c *countedWaitGroup) done() {
	c._wg.Done()
	atomic.AddInt32(&c._count, -1)
}

func (c *countedWaitGroup) wait() {
	c._wg.Wait()
}
