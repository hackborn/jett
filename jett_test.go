package jett

import (
	"fmt"
	"testing"
)

// ------------------------------------------------------------
// TEST-POOL

func TestPool(t *testing.T) {
	cases := []struct {
		opts Opts
	}{
		{stdopts},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			p := NewWith(tc.opts)
			defer p.Close()

			for i := 0; i < 10; i++ {
				v := i
				f := func() error {
					fmt.Println("msg ", v)
					return nil
				}
				p.Run(f)
			}
			t.Fatal()
		})
	}
}

// ------------------------------------------------------------
// CONST and VAR

var (
	stdopts = Opts{MinWorkers: 1, MaxWorkers: 10}
)
