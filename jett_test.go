package jett

import (
	"fmt"
	"github.com/micro-go/lock"
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
	runningWg        sync.WaitGroup
	running          bool
	mutex            sync.Mutex
	want             int // How many items I want before I unblock.
	have             int // How many items I have.
	waiting          []chan BlockCmd
	stage            BlockStage
	accumulatingWg   sync.WaitGroup
	accumulatingCond *mutexCond
	runningCond      *mutexCond
}

func newBlock(want int) *Block {
	accumulatingCond := newMutexCond()
	runningCond := newMutexCond()
	return &Block{want: want, stage: StageAccumulating, accumulatingCond: accumulatingCond, runningCond: runningCond}
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
	b.accumulatingCond.c.Broadcast()
	//	b.wg.Wait()
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

	//	if b.handled(f) {
	//		fmt.Println("handle() handled")
	//		return nil
	//	}

	// Wait for everyone to accumulate
	go func() {
		fmt.Println("finish ACCUM WG 1")
		b.accumulatingWg.Done()
	}()
	b.runningCond.wait()

	f()

	return nil
}

func (b *Block) handled(f RunFunc) bool {
	defer lock.Locker(&b.mutex).Unlock()
	b.have += 1
	fmt.Println("BLOCK have", b.have, "want", b.want)
	if b.have >= b.want {
		fmt.Println("block FLUSH")
		// Flush
		b.accumulatingCond.c.Broadcast()
		/*
			for _, c := range b.waiting {
				c <- CmdRun
			}
		*/
		f()
		return true
	}
	return false
}

// init() performs necessary initialization dynamically.
func (b *Block) init() {
	b.running = true
}

func (b *Block) addWaitingChannel(c chan BlockCmd) {
	defer lock.Locker(&b.mutex).Unlock()
	b.waiting = append(b.waiting, c)
}

func (b *Block) waitingChannelFinished(c chan BlockCmd) {
}

// ------------------------------------------------------------
// TYPES

type BlockCmd uint32

const (
	CmdRun BlockCmd = 1 << iota
	CacheGetter
)

type BlockStage uint32

const (
	StageAccumulating BlockStage = 1 << iota
	StageReadyToRun
	StageRunning
	StageDone
)

// ------------------------------------------------------------
// CONST and VAR

var (
	stdopts = Opts{MinWorkers: 1, MaxWorkers: 109, QueueSize: 256}
)
