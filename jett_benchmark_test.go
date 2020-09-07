package jett

import (
	"context"
	"sync"
	"testing"
)

// ------------------------------------------------------------
// BENCHMARK COND SUSTAINED

func BenchmarkSustainedSmall1(t *testing.B) {
	sustainedSmall(t, 1, 10, 10000)
}

func BenchmarkSustainedSmall2(t *testing.B) {
	sustainedSmall(t, 1, 100, 10000)
}

func BenchmarkSustainedSmall3(t *testing.B) {
	sustainedSmall(t, 1, 1000, 10000)
}

func BenchmarkSustainedMedium1(t *testing.B) {
	sustainedSmall(t, 1, 10, 100000)
}

func BenchmarkSustainedMedium2(t *testing.B) {
	sustainedSmall(t, 1, 100, 100000)
}

func BenchmarkSustainedMedium3(t *testing.B) {
	sustainedSmall(t, 1, 1000, 100000)
}

func BenchmarkSustainedLarge1(t *testing.B) {
	sustainedSmall(t, 1, 10, 1000000)
}

func BenchmarkSustainedLarge2(t *testing.B) {
	sustainedSmall(t, 1, 100, 1000000)
}

func BenchmarkSustainedLarge3(t *testing.B) {
	sustainedSmall(t, 1, 1000, 1000000)
}

func sustainedSmall(t *testing.B, min, max, runners int) {
	opts := Opts{MinWorkers: min, MaxWorkers: max}
	p := NewWith(opts)
	defer p.Close()

	wait := &sync.WaitGroup{}
	r := &benchmarkRunner{wait}
	wait.Add(runners)
	for i := 0; i < runners; i++ {
		p.Run(r.Run)
	}
	wait.Wait()
}

// ------------------------------------------------------------
// BENCHMARK CHAN SUSTAINED

func BenchmarkChanSustainedSmall1(t *testing.B) {
	chanSustained(t, 1, 10, 10000)
}

func BenchmarkChanSustainedSmall2(t *testing.B) {
	chanSustained(t, 1, 100, 10000)
}

func BenchmarkChanSustainedSmall3(t *testing.B) {
	chanSustained(t, 1, 1000, 10000)
}

func BenchmarkChanSustainedMedium1(t *testing.B) {
	chanSustained(t, 1, 10, 100000)
}

func BenchmarkChanSustainedMedium2(t *testing.B) {
	chanSustained(t, 1, 100, 100000)
}

func BenchmarkChanSustainedMedium3(t *testing.B) {
	chanSustained(t, 1, 1000, 100000)
}

func BenchmarkChanSustainedLarge1(t *testing.B) {
	chanSustained(t, 1, 10, 1000000)
}

func BenchmarkChanSustainedLarge2(t *testing.B) {
	chanSustained(t, 1, 100, 1000000)
}

func BenchmarkChanSustainedLarge3(t *testing.B) {
	chanSustained(t, 1, 1000, 1000000)
}

func chanSustained(t *testing.B, min, max, runners int) {
	opts := Opts{MinWorkers: min, MaxWorkers: max}
	p := newChanWith(opts)
	defer p.Close()

	wait := &sync.WaitGroup{}
	r := &benchmarkRunner{wait}
	wait.Add(runners)
	for i := 0; i < runners; i++ {
		p.Run(r.Run)
	}
	wait.Wait()
}

// ------------------------------------------------------------
// BENCHMARK-RUNNER

type benchmarkRunner struct {
	wait *sync.WaitGroup
}

func newBenchmarkRunner(wait *sync.WaitGroup) *benchmarkRunner {
	wait.Add(1)
	return &benchmarkRunner{wait}
}

func (r *benchmarkRunner) Run(ctx context.Context) error {
	r.wait.Done()
	return nil
}
