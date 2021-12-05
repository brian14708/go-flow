package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flowtype/testutil"
)

func setupBroadcast(b flow.Node) (chan interface{}, []chan interface{}) {
	inCh := make(chan interface{})
	outCh := []chan interface{}{
		make(chan interface{}, 1),
		make(chan interface{}, 2),
		make(chan interface{}, 3),
	}
	{
		in, out := b.Ports()
		*in["in"].(*flow.Port).Ref.(*<-chan interface{}) = inCh
		*out["out_0"].(*flow.Port).Ref.(*chan<- interface{}) = outCh[0]
		*out["out_1"].(*flow.Port).Ref.(*chan<- interface{}) = outCh[1]
		*out["out_2"].(*flow.Port).Ref.(*chan<- interface{}) = outCh[2]
	}
	go func() {
		if err := b.Run(context.Background()); err != nil {
			panic(err)
		}
		close(outCh[0])
		close(outCh[1])
		close(outCh[2])
	}()
	return inCh, outCh
}

func TestBroadcast(t *testing.T) {
	testutil.TestWithNilRegistry(t, func(t *testing.T) {
		_, err := NewBroadcastNode(func(i int) int { return i }, 0)
		assert.Error(t, err)
		_, err = NewBroadcastNode(func(i int32) int { return 0 }, 3)
		assert.Error(t, err)

		called := 0
		node, err := NewBroadcastNode(func(i int) int {
			called++
			return i
		}, 3)
		assert.NoError(t, err)

		inCh, outCh := setupBroadcast(node)
		go func() {
			for i := 0; i < 5; i++ {
				inCh <- i
			}
			close(inCh)
		}()

		for i := 0; i < 5; i++ {
			for _, o := range outCh {
				assert.Equal(t, i, <-o)
			}
		}
		assert.Equal(t, 5*(len(outCh)-1), called)
	})
}

func BenchmarkBroadcast(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	node, err := NewBroadcastNode(func(i int) int { return i }, 3)
	assert.NoError(b, err)

	inCh, outCh := setupBroadcast(node)
	go func() {
		for i := 0; ; i++ {
			select {
			case inCh <- i:
			case <-ctx.Done():
				close(inCh)
				return
			}
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, o := range outCh {
			<-o
		}
	}
}

func TestBroadcastGraph(t *testing.T) {
	g, err := flow.NewGraph(nil)
	assert.NoError(t, err)

	in, out := make(chan int, 8), make(chan int, 8)
	inN, _ := NewChanNode((<-chan int)(in))
	outN, _ := NewChanNode((chan<- int)(out))
	b, _ := NewBroadcastNode(func(i int) int { return i }, 1)
	_ = g.AddNode("b", b)
	_ = g.AddNode("in", inN)
	_ = g.AddNode("out", outN)
	_ = g.Connect([]string{"in:out"}, []string{"b:in"})
	_ = g.Connect([]string{"b:out_0"}, []string{"out:in"})

	in <- 10
	close(in)
	assert.NoError(t, g.Run(context.Background()))
	assert.Equal(t, 10, <-out)
}
