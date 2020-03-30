package jett

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// ------------------------------------------------------------
// TEST-POOL

func TestPool(t *testing.T) {
	cases := []struct {
		opts       Opts
		operations int // Number of operations to run
		block      *Block
	}{
		{stdopts, 10, newBlock(10)},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			runTestPool(t, tc.opts, tc.operations, tc.block)
		})
	}
}

func runTestPool(t *testing.T, opts Opts, operations int, block *Block) {
	//	defer block.Close()
	p := NewWith(opts)
	defer p.Close()

	var have int32 = 0

	for i := 0; i < operations; i++ {
		f := func() error {
			atomic.AddInt32(&have, 1)
			return nil
		}
		p.Run(block.Handler(f))
	}
	// Wait for the block
	fmt.Println("BLOCK 1")
	block.Accumulate()
	fmt.Println("BLOCK 2")
	block.Run()
	fmt.Println("BLOCK 3")
	if have != int32(operations) {
		fmt.Println("mismatch, have", have, "want", operations)
		t.Fatal()
	}
}

// ------------------------------------------------------------
// BLOCK

// Block blocks on messages.
type Block struct {
	runningWg      sync.WaitGroup
	running        bool
	mutex          sync.Mutex
	want           int // How many items I want before I unblock.
	have           int // How many items I have.
	accumulatingWg sync.WaitGroup
	runningCond    *mutexCond
}

func newBlock(want int) *Block {
	runningCond := newMutexCond()
	return &Block{want: want, runningCond: runningCond}
}

func (b *Block) Handler(inner RunFunc) RunFunc {
	b.init()
	b.runningWg.Add(1)
	b.accumulatingWg.Add(1)

	fmt.Println("Add Handler")
	wrapper := func() error {
		return b.handle(inner)
	}
	return wrapper
}

func (b *Block) Close() error {
	fmt.Println("....block close??")
	b.running = false
	return nil
}

func (b *Block) Accumulate() {
	fmt.Println("ACCUMULATE 1")
	b.accumulatingWg.Wait()
	fmt.Println("ACCUMULATE 2")
}

func (b *Block) Run() {
	fmt.Println("Run() 1")
	b.runningCond.c.Broadcast()
	fmt.Println("Run() 2")
	b.runningWg.Wait()
	fmt.Println("Run() 3")
}

func (b *Block) handle(f RunFunc) error {
	fmt.Println("handle()")
	defer b.runningWg.Done()
	defer fmt.Println("handle() DONE")
	if !b.running {
		fmt.Println("finish ACCUM WG ??")
		b.accumulatingWg.Done()
		return nil
	}

	// Wait for everyone to accumulate
	go func() {
		fmt.Println("finish ACCUM WG 1")
		b.accumulatingWg.Done()
	}()
	b.runningCond.wait()

	f()

	return nil
}

// init() performs necessary initialization dynamically.
func (b *Block) init() {
	b.running = true
}

// ------------------------------------------------------------
// TYPES

type BlockCmd uint32

const (
	CmdRun BlockCmd = 1 << iota
	CacheGetter
)

// ------------------------------------------------------------
// CONST and VAR

var (
	stdopts = Opts{MinWorkers: 1, MaxWorkers: 109, QueueSize: 256}
)
