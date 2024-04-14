package flow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"golang.org/x/time/rate"

	"github.com/brian14708/go-flow/flow/channel"
)

func buildGraph(g *Graph) func(t *testing.T) {
	out := make(chan int, 2)

	if err := g.AddNode("doubler", FnNode(func(_ context.Context, in <-chan int, out chan<- int) error {
		for i := range in {
			out <- i * 16
		}
		return nil
	})); err != nil {
		panic(err.Error())
	}
	if err := g.AddNode("g1", GeneratorNode(1)); err != nil {
		panic(err.Error())
	}
	if err := g.AddNode("g2", GeneratorNode(2)); err != nil {
		panic(err.Error())
	}
	if err := g.AddNode("sum1", SumNode(out)); err != nil {
		panic(err.Error())
	}
	if err := g.AddNode("sum2", SumNode(out)); err != nil {
		panic(err.Error())
	}
	if err := g.Connect(
		[]string{"g1:out", "g2:out"},
		[]string{"doubler:in"},
		WithChanSize(32),
	); err != nil {
		panic(err.Error())
	}
	if err := g.Connect(
		[]string{"doubler:out"},
		[]string{"sum1:in", "sum2:in"},
		WithChanSize(0),
	); err != nil {
		panic(err.Error())
	}

	return func(t *testing.T) {
		assert.Equal(t, 240, <-out+<-out)
	}
}

func TestBuildGraph(t *testing.T) {
	g, err := NewGraph(nil)
	assert.NoError(t, err)
	checker := buildGraph(g)
	assert.NoError(t, g.Run(context.Background()))
	assert.Error(t, g.Run(context.Background()))
	checker(t)
}

func TestGraphCycle(t *testing.T) {
	out := make(chan int, 1)

	g, err := NewGraph(nil)
	assert.NoError(t, err)
	assert.NoError(t, g.AddNode("g1", GeneratorNode(1)))
	assert.NoError(t, g.AddNode("sum1", SumNode(out)))
	assert.NoError(t, g.Connect([]string{"g1:out"}, []string{"sum1:in"}))
	assert.NoError(t, g.Connect([]string{"sum1:out"}, []string{"g1:in"}))
	assert.NoError(t, g.Run(context.Background()))
	assert.Equal(t, 5, <-out)
}

func TestGraphID(t *testing.T) {
	g1, err := NewGraph(nil)
	assert.NoError(t, err)
	g2, err := NewGraph(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, g1.ID(), g2.ID())
}

func TestGraphAddNode(t *testing.T) {
	g, err := NewGraph(nil)
	assert.NoError(t, err)

	node := GeneratorNode(0)
	assert.NoError(t, g.AddNode("g", node))
	assert.Error(t, g.AddNode("$@!", GeneratorNode(0)))
	assert.Error(t, g.AddNode("g", GeneratorNode(0)))
	assert.Error(t, g.AddNode("duplicate", node))

	bad := GeneratorNode(0)
	bad.(*fnNode).out = make(chan int)
	assert.Error(t, g.AddNode("badport", bad))
}

func TestGraphError(t *testing.T) {
	err := errors.New("error")

	g, _ := NewGraph(nil)
	assert.NoError(t, g.AddNode("w", WaitNode()))
	assert.NoError(t, g.AddNode("e", FnNode(func(context.Context, <-chan int, chan<- int) error {
		return err
	})))

	assert.ErrorIs(t, g.Run(context.Background()), err)
}

func TestGraphMultiError(t *testing.T) {
	err := errors.New("error")
	err2 := errors.New("error2")

	g, _ := NewGraph(nil)
	assert.NoError(t, g.AddNode("w", WaitNode()))
	assert.NoError(t, g.AddNode("e1", FnNode(func(context.Context, <-chan int, chan<- int) error {
		return err
	})))
	assert.NoError(t, g.AddNode("e2", FnNode(func(context.Context, <-chan int, chan<- int) error {
		return err2
	})))

	gerr, ok := g.Run(context.Background()).(*GraphError)
	assert.True(t, ok)

	all := map[string]error{}
	for gerr != nil {
		all[gerr.NodeName()] = gerr.Unwrap()
		gerr = gerr.Next()
	}

	assert.Equal(t, 2, len(all))
	assert.Equal(t, err, all["e1"])
	assert.Equal(t, err2, all["e2"])
}

func TestGraphContextError(t *testing.T) {
	g, err := NewGraph(nil)
	assert.NoError(t, err)
	assert.NoError(t, g.AddNode("w", WaitNode()))

	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
	defer cancel()
	assert.Equal(t, context.DeadlineExceeded, g.Run(ctx))
}

func TestGraphBackgroundError(t *testing.T) {
	err := errors.New("error")
	g, _ := NewGraph(nil)
	g.background = append(g.background, func(context.Context) error {
		return err
	})
	assert.NoError(t, g.AddNode("w", WaitNode()))
	assert.ErrorIs(t, g.Run(context.Background()), err)
}

func TestGraphErrorStall(t *testing.T) {
	cnt := atomic.NewInt32(0)
	g, err := NewGraph(&GraphOptions{
		DefaultConnectOptions: []ConnectOption{
			channel.WithDropHandler(func(i interface{}) {
				cnt.Inc()
			}),
			channel.WithDrainRate(rate.Every(10 * time.Millisecond)),
			channel.WithSize(1),
		},
	})
	assert.NoError(t, err)
	assert.NoError(t, g.AddNode("g", GeneratorNode(1)))
	assert.NoError(t, g.AddNode("w", WaitNode()))
	assert.NoError(t, g.Connect([]string{"g:out"}, []string{"w:in"}))

	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
	defer cancel()
	assert.Equal(t, context.DeadlineExceeded, g.Run(ctx))
	assert.Equal(t, 5, int(cnt.Load()))
}

func TestGraphEarlyExit(t *testing.T) {
	cnt := atomic.NewInt32(0)
	g, err := NewGraph(&GraphOptions{
		DefaultConnectOptions: []ConnectOption{
			channel.WithDropHandler(func(i interface{}) {
				cnt.Inc()
			}),
			channel.WithDrainRate(rate.Every(10 * time.Millisecond)),
			channel.WithSize(1),
		},
	})
	assert.NoError(t, err)
	assert.NoError(t, g.AddNode("g", GeneratorNode(1)))
	assert.NoError(t, g.AddNode("w", WaitNode()))
	assert.NoError(t, g.AddNode("r", ReturnNode()))
	assert.NoError(t, g.Connect([]string{"g:out"}, []string{"w:in"}))
	assert.NoError(t, g.Connect([]string{"w:out"}, []string{"r:in"}))

	assert.NoError(t, g.Run(context.Background()))
	assert.Equal(t, 5, int(cnt.Load()))
}

func BenchmarkGraphBuild(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g, _ := NewGraph(nil)
			_ = buildGraph(g)
			_ = g.Run(context.Background())
		}
	})
}
