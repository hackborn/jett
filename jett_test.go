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
	block.Wait()
	fmt.Println("BLOCK 2")
	if have != int32(operations) {
		fmt.Println("mismatch, have", have, "want", operations)
		t.Fatal()
	}
}

// ------------------------------------------------------------
// BLOCK

// Block blocks on messages.
type Block struct {
	done    chan struct{}
	wg      sync.WaitGroup
	running bool
	mutex   sync.Mutex
	want    int // How many items I want before I unblock.
	have    int // How many items I have.
	waiting []chan BlockCmd
}

func newBlock(want int) *Block {
	return &Block{want: want}
}

func (b *Block) Handler(inner RunFunc) RunFunc {
	b.init()
	b.wg.Add(1)
	fmt.Println("Add Handler")
	wrapper := func() error {
		return b.handle(inner)
	}
	return wrapper
}

func (b *Block) Close() error {
	b.running = false
	close(b.done)
	b.wg.Wait()
	return nil
}

func (b *Block) Wait() {
	b.wg.Wait()
}

func (b *Block) handle(f RunFunc) error {
	fmt.Println("handle()")
	defer fmt.Println("handle() DONE")
	defer b.wg.Done()
	if !b.running {
		return nil
	}

	if b.handled(f) {
		fmt.Println("handle() handled")
		return nil
	}

	// Register waiter
	c := make(chan BlockCmd)
	b.addWaitingChannel(c)
	defer b.waitingChannelFinished(c)

	// Wait
	fmt.Println("handle() BLOCK")
	for {
		select {
		case <-b.done:
			return nil
		case cmd, more := <-c:
			if more && cmd == CmdRun {
				fmt.Println("waiter called")
				f()
			}
			return nil
		}
	}
	return nil
}

func (b *Block) handled(f RunFunc) bool {
	defer lock.Locker(&b.mutex).Unlock()
	b.have += 1
	fmt.Println("BLOCK have", b.have, "want", b.want)
	if b.have >= b.want {
		fmt.Println("block FLUSH")
		// Flush
		for _, c := range b.waiting {
			c <- CmdRun
		}
		f()
		return true
	}
	return false
}

// init() performs necessary initialization dynamically.
func (b *Block) init() {
	if b.done == nil {
		b.done = make(chan struct{})
		b.running = true
	}
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

// ------------------------------------------------------------
// CONST and VAR

var (
	stdopts = Opts{MinWorkers: 1, MaxWorkers: 10, QueueSize: 256}

	blockAll = newBlock(99999)
)

const (
	CmdRun BlockCmd = 1 << iota
	CacheGetter
)
