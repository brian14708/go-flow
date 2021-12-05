package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddParallel(t *testing.T) {
	assert.Panics(t, func() {
		OrderedParallel(nil)
	})
}

func TestAddOrderedParallel(t *testing.T) {
	ppl := New(nil)
	in := make(chan int, 64)
	out := make(chan int64, 64)
	ppl.Add("in", (<-chan int)(in))
	ppl.Add("calc", OrderedParallel(repeat(
		func(ctx context.Context, in <-chan int, out chan<- int64) error {
			for i := range in {
				if i%2 == 0 {
					time.Sleep(time.Millisecond)
				}
				out <- int64(i)
			}
			return nil
		}, 4)),
	)
	ppl.Add("out", (chan<- int64)(out))
	for i := 0; i < 10; i++ {
		in <- i
	}
	close(in)

	assert.NoError(t, ppl.Run(context.Background()))
	i := 0
	for o := range out {
		assert.Equal(t, int64(i), o)
		i++
	}
}

func TestAddOrderedParallelOutputMismatch(t *testing.T) {
	ppl := New(nil)
	idx := 0
	assert.Panics(t, func() {
		ppl.Add("calc", OrderedParallel(repeat(func() interface{} {
			idx++
			if idx%2 == 0 {
				return func(ctx context.Context, in <-chan int32, out chan<- int32) error {
					return nil
				}
			}
			return func(ctx context.Context, in <-chan int32) error {
				return nil
			}
		}, 4)))
	})
}
