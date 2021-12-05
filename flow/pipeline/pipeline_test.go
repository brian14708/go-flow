package pipeline

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/funcop"
)

func TestPipeline(t *testing.T) {
	ppl := New(nil)

	in := make(chan int, 1)
	ppl.Add("", (<-chan int)(in))

	ppl.Add("g2", func(ppl *Pipeline) {
		ppl.Add("", new(Split))
		ppl.SplitOutput([]string{"out3"}).IgnoreOutput()

		ppl.Merge(
			ppl.SplitOutput("out1").
				Add("a", func(ctx context.Context, in <-chan int, out chan<- float32) error {
					for i := range in {
						out <- float32(i) * 12.5
					}
					return nil
				}).
				Add("b", func(ctx context.Context, in <-chan float32, out chan<- float32) error {
					for i := range in {
						out <- i * 2
					}
					return nil
				}),
			ppl.SplitOutput("out2"),
		)
	})

	ppl.Add("g3", func(ctx context.Context, in <-chan float32, out chan<- int) error {
		var sum int
		for i := range in {
			sum += int(i)
		}
		out <- sum
		return nil
	})

	ppl.Add("g4", func(_ context.Context, in <-chan flow.AnyMessage, out chan<- flow.AnyMessage) error {
		for i := range in {
			out <- i
		}
		return nil
	})

	result := make(chan int, 1)
	ppl.Add("output", (chan<- int)(result))

	in <- 10
	close(in)

	assert.NoError(t, ppl.Run(context.Background()))
	assert.Equal(t, <-result, 260)

	assert.Equal(t, 7, len(ppl.Graph().Topology().Nodes))
}

func TestPipelinePorts(t *testing.T) {
	ppl := New(nil)
	in, out := ppl.Ports()
	assert.Equal(t, 0, len(in))
	assert.Equal(t, 0, len(out))

	ppl.Add("A", func(<-chan int, chan<- float32) {})
	in, out = ppl.Ports()
	assert.Equal(t, 1, len(in))
	assert.Equal(t, 1, len(out))
}

func TestPipelineOutputType(t *testing.T) {
	ppl := New(nil)
	ppl.Add("g1", (<-chan int)(make(chan int)))
	assert.Equal(t, reflect.TypeOf(0), ppl.outputType())
	ppl.Add("g2", (chan<- int)(make(chan int)))
	assert.Nil(t, ppl.outputType()) // no output

	ppl = New(nil)
	ppl.Add("s", new(Split))
	assert.Panics(t, func() { ppl.outputType() }) // multi type output
}

func TestInputType(t *testing.T) {
	assert.Equal(t, reflect.TypeOf(0), inputType((chan<- int)(make(chan int))))
	assert.Nil(t, inputType((<-chan int)(make(chan int))))
	assert.Panics(t, func() { inputType(new(Join)) })
}

func TestPipelineDuplicateNode(t *testing.T) {
	ppl := New(nil)
	n := new(NoopIntNode)
	ppl.Add("a", n)
	assert.Panics(t, func() {
		ppl.Add("b", n)
	})
}

func TestFromGraph(t *testing.T) {
	g, _ := flow.NewGraph(nil)
	FromGraph(flow.PartialGraph(g, nil, nil))
}

func TestPipelineMultiOutput(t *testing.T) {
	ppl := New(nil)
	ppl.Add("A", new(Split))
	ppl.SplitOutput([]string{"out2"}).IgnoreOutput()

	out1 := ppl.SplitOutput("A:out1")
	out3 := ppl.SplitOutput("out3")

	out1.Add("a", func() interface{} { return new(NoopIntNode) })
	out1.Add("b", func() interface{} { return new(NoopIntNode) })
	out3.Add("c", func() interface{} { return new(NoopIntNode) })
	assert.Panics(t, func() {
		// bad input
		New(nil).Add("sum", new(SumIntNode),
			SideInput(nil, []string{"in3"}),
		)
	})
	out1.Add("sum", new(SumIntNode),
		SideInput(out3, []string{"in2"}),
	)

	ppl.Merge(out1, out3)

	ppl.Add("D", new(NoopIntNode))
}

func TestPipelineDangling(t *testing.T) {
	{
		ppl := New(nil)
		ch := make(chan int)
		ppl.Add("in", (<-chan int)(ch))
		assert.Panics(t, func() {
			_ = ppl.Run(context.Background())
		})
	}
	{
		ppl := New(nil)
		ch := make(chan int)
		ppl.Add("out", (chan<- int)(ch))
		assert.Panics(t, func() {
			_ = ppl.Run(context.Background())
		})
	}
	assert.Panics(t, func() {
		ppl := New(nil)
		ch := make(chan int)
		ppl.Add("in1", (chan<- int)(ch))
		ppl.Add("in2", (chan<- int)(ch))
	})
	assert.Panics(t, func() {
		ppl := New(nil)
		ch := make(chan int)
		ppl.Add("out1", (<-chan int)(ch))
		ppl.Add("out2", (<-chan int)(ch))
	})
}

func TestPipelineMerge(t *testing.T) {
	ppl := New(nil)
	ch := make(chan int)
	ppl.Add("in", (<-chan int)(ch))

	// merge with unrelated pipeline
	assert.Panics(t, func() { ppl.Merge(New(nil)) })

	// merge with empty pipline
	ppl.Merge(ppl.SubPipeline(""))

	// merge with input
	tmp := ppl.SubPipeline("")
	tmp.Add("danglinginput", func(context.Context, <-chan int, chan<- int) error {
		return nil
	})
	ppl.Merge(tmp).IgnoreInput()

	// self merge
	ppl.Merge(ppl)

	// merge with another channel
	ch2 := make(chan int)
	ppl.Merge(
		ppl.SubPipeline("").
			Add("in2", (<-chan int)(ch2)),
	)

	out := make(chan int)
	ppl.Add("out", (chan<- int)(out))

	go func() {
		assert.NoError(t, tmp.Run(context.Background()))
	}()

	ch <- 1
	close(ch)
	assert.Equal(t, 1, <-out)

	ch2 <- 2
	close(ch2)
	assert.Equal(t, 2, <-out)

	<-out
}

func TestPipelineString(t *testing.T) {
	ppl := New(nil)
	assert.Equal(t, "(nil)", ppl.String())

	tmp := New(nil)
	tmp.IgnoreInput()

	tmp = New(nil)
	tmp.IgnoreOutput()
	assert.Equal(t, "(nil)", tmp.String())

	ch := make(chan int)
	ppl.Add("test2", (chan<- int)(ch))
	ppl.Add("test", (<-chan int)(ch))
	assert.NotEqual(t, "(nil)", ppl.String())
	ppl.IgnoreOutput().IgnoreInput()
	assert.Equal(t, "input=[] output=[]", ppl.String())
}

func TestPipelineDiscard(t *testing.T) {
	ch := make(chan int)
	ppl := New(nil)
	ppl.Add("test2", (chan<- int)(ch))
	ppl.Add("test", (<-chan int)(ch))
	assert.Panics(t, func() {
		ppl.IgnoreInput("xx")
	})
	assert.Panics(t, func() {
		ppl.IgnoreOutput("xx")
	})

	ppl.IgnoreInput("in")
	ppl.IgnoreOutput("out")
}

func newSequentialPipeline(in, out chan int) *Pipeline {
	ppl := New(nil)
	ppl.Add("in", (<-chan int)(in))
	for i := 0; i < 64; i++ {
		ppl.Add("", func(
			in <-chan int,
			out chan<- int,
		) {
			for i := range in {
				out <- i + 2
			}
		})
	}
	ppl.Add("out", (chan<- int)(out))
	return ppl
}

func runSequentialPipeline(ctx context.Context) (chan<- int, <-chan int) {
	in, out := make(chan int, 32), make(chan int, 32)
	ppl := newSequentialPipeline(in, out)
	go func() {
		if err := ppl.Run(ctx); err != context.Canceled {
			panic(err)
		}
	}()
	return in, out
}

func newSequentialDynamicPipeline(in, out chan int) *Pipeline {
	ppl := New(nil)
	ppl.Add("in", (<-chan int)(in))
	for i := 0; i < 64; i++ {
		if i%2 == 0 {
			ppl.Add("", func(
				in <-chan flow.AnyMessage,
				out chan<- flow.AnyMessage,
			) {
				for i := range in {
					out <- i.(int) + 2
				}
			})
		} else {
			ppl.Add("", funcop.Map(func(in int) int {
				return i + 2
			}))
		}
	}
	ppl.Add("out", (chan<- int)(out))
	return ppl
}

func runSequentialDynamicPipeline(ctx context.Context) (chan<- int, <-chan int) {
	in, out := make(chan int, 32), make(chan int, 32)
	ppl := newSequentialDynamicPipeline(in, out)
	go func() {
		if err := ppl.Run(ctx); err != context.Canceled {
			panic(err)
		}
	}()
	return in, out
}

func BenchmarkPipelineLatency(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in, out := runSequentialPipeline(ctx)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		in <- 1
		<-out
	}
}

func BenchmarkPipelineThroughput(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in, out := runSequentialPipeline(ctx)
	b.ResetTimer()

	go func() {
		for {
			select {
			case in <- 1:
			case <-ctx.Done():
				return
			}
		}
	}()
	for i := 0; i < b.N; i++ {
		<-out
	}
}

func BenchmarkPipelineBuild(b *testing.B) {
	in, out := make(chan int), make(chan int)
	for i := 0; i < b.N; i++ {
		newSequentialPipeline(in, out)
	}
}

func BenchmarkDynamicPipelineBuild(b *testing.B) {
	in, out := make(chan int), make(chan int)
	for i := 0; i < b.N; i++ {
		newSequentialDynamicPipeline(in, out)
	}
}

func BenchmarkDynamicPipelineLatency(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in, out := runSequentialDynamicPipeline(ctx)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		in <- 1
		<-out
	}
}

func BenchmarkDynamicPipelineThroughput(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in, out := runSequentialDynamicPipeline(ctx)
	b.ResetTimer()

	go func() {
		for {
			select {
			case in <- 1:
			case <-ctx.Done():
				return
			}
		}
	}()
	for i := 0; i < b.N; i++ {
		<-out
	}
}

func inputType(n interface{}) reflect.Type {
	p := New(nil).Add("tmp", n)
	in := p.last.in
	if len(in) == 0 {
		return nil
	}

	elemType, err := p.g.PortType(in[0])
	if err != nil {
		panic(fmt.Sprintf("cannot get input type for `%s': %s", in[0], err.Error()))
	}
	for _, i := range in {
		t, err := p.g.PortType(i)
		if err != nil {
			panic(fmt.Sprintf("cannot get input type for `%s': %s", i, err.Error()))
		}
		if t != elemType {
			panic("multiple input types for pipeline")
		}
	}
	return elemType
}
