package jett

import (
	"sync"
	"sync/atomic"
)

// ------------------------------------------------------------
// MUTEX-COND

// mutexCond combines a Cond with its locker.
type mutexCond struct {
	m sync.Mutex
	c *sync.Cond
}

func newMutexCond() *mutexCond {
	mc := &mutexCond{}
	mc.c = sync.NewCond(&mc.m)
	return mc
}

func (m *mutexCond) broadcast() {
	m.c.Broadcast()
}

func (m *mutexCond) signal() {
	m.c.Signal()
}

func (m *mutexCond) wait() {
	m.m.Lock()
	m.c.Wait()
	m.m.Unlock()
}

// ------------------------------------------------------------
// COUNTED-WAIT-GROUP

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
