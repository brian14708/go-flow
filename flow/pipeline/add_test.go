package pipeline

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow"
)

func TestPipelineAdd(t *testing.T) {
	// bad block type
	assert.Panics(t, func() { New(nil).Add("a", "bad type") })
	assert.Panics(t, func() {
		n := new(SumIntNode)
		New(nil).Add("a", n).Add("b", n)
	})
}

func TestPipelineConnection(t *testing.T) {
	assert.Panics(t, func() {
		// dangling connection options
		New(nil).Add("a", func(context.Context, chan<- int) error { return nil },
			flow.WithChanSize(12),
		)
	})

	cnt := 0
	ppl := New(nil)
	ppl.Add("a", func(context.Context, chan<- int) error { return nil })
	ppl.Add("b",
		func(ctx context.Context, in <-chan int) error {
			cnt = cap(in)
			return nil
		},
		flow.WithChanSize(123),
	)
	assert.NoError(t, ppl.Run(context.Background()))
	assert.Equal(t, 123, cnt)
}

func TestPipelineMatchOutput(t *testing.T) {
	{
		ppl := New(nil)
		ppl.Add("in", (<-chan int)(make(chan int)))
		ppl.Add("split", new(Split))
		ppl.SplitOutput([]string{"out1"}).IgnoreOutput()
		ppl.SplitOutput("out2").IgnoreOutput()
		ppl.Add("out", (chan<- int)(make(chan int)))
	}
	assert.Panics(t, func() {
		New(nil).Add("split", new(Split)).SplitOutput("xx")
	})
}

func TestPipelineSideInput(t *testing.T) {
	assert.Panics(t, func() {
		SideInput(nil, 123) // bad type
	})
	{
		ppl := New(nil)
		ppl2 := ppl.SubPipeline("").
			Add("in", (<-chan int)(make(chan int)))

		ppl.Add("sum", new(SumIntNode),
			SideInput(ppl2, []string{"in2"}),
		)
	}
	{
		ppl := New(nil)
		ppl.Add("sum", new(SumIntNode),
			SideInput(nil, []string{"in2"}),
		)
	}
	assert.Panics(t, func() {
		// different graph
		ppl := New(nil)
		ppl.Add("sum", new(SumIntNode),
			SideInput(New(nil), []string{"in2"}),
		)
	})
}

func repeat(t interface{}, n int) []interface{} {
	ret := make([]interface{}, n)
	for i := 0; i < n; i++ {
		ret[i] = t
	}
	return ret
}

func TestAddParallelBlock(t *testing.T) {
	ppl := New(nil)
	in := make(chan int, 64)
	out := make(chan int, 64)
	ppl.Add("in", (<-chan int)(in))
	ppl.Add("",
		repeat(func(ctx context.Context, in <-chan int, out chan<- int) error {
			for i := range in {
				out <- i * 2
			}
			return nil
		}, 10),
	)
	ppl.Add("out", (chan<- int)(out))
	for i := 0; i < 10; i++ {
		in <- i
	}
	close(in)

	assert.NoError(t, ppl.Run(context.Background()))

	results := make([]int, 0, len(out))
	for o := range out {
		results = append(results, o)
	}
	sort.Ints(results)
	for i := 0; i < 10; i++ {
		assert.Equal(t, i*2, results[i])
	}
}
