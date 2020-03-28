package jett

// ------------------------------------------------------------
// OPTS

// Opts provides options when creating new pools.
type Opts struct {
	MinWorkers int // Minimum number of go routines.
	MaxWorkers int // Maximum number of go routines.
	QueueSize  int // Number of items in the queue before we block.
}

func NewDefaultOpts() Opts {
	return Opts{MinWorkers: 1, MaxWorkers: 10, QueueSize: 256}
}

// scrub() returns the options with any illegal
// values replaced with valid values.
func (o Opts) scrub() Opts {
	if o.MinWorkers < 1 {
		o.MinWorkers = 1
	}
	if o.MaxWorkers < o.MinWorkers {
		o.MaxWorkers = o.MinWorkers
	}
	if o.QueueSize < 1 {
		o.QueueSize = 1
	}
	return o
}
