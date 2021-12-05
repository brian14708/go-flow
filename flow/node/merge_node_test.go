package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flowtype/testutil"
)

func TestMergeNode(t *testing.T) {
	testutil.TestWithNilRegistry(t, func(t *testing.T) {
		_, err := NewMergeNode(func(chan<- int, ...int) {}, 0)
		assert.Error(t, err)
		_, err = NewMergeNode(func(int, float32) {}, 2)
		assert.Error(t, err)
		_, err = NewMergeNode(func(int, ...int) {}, 2)
		assert.Error(t, err)

		m, err := NewMergeNode(func(out chan<- int32, n ...int) {
			sum := 0
			for _, n := range n {
				sum += n
			}
			if sum%2 == 0 {
				out <- int32(sum)
			}
		}, 3)
		assert.NoError(t, err)

		inCh := []chan interface{}{
			make(chan interface{}, 1),
			make(chan interface{}, 2),
			make(chan interface{}, 3),
		}
		outCh := make(chan int32)
		{
			in, out := m.Ports()
			*in["in_0"].(*flow.Port).Ref.(*<-chan interface{}) = inCh[0]
			*in["in_1"].(*flow.Port).Ref.(*<-chan interface{}) = inCh[1]
			*in["in_2"].(*flow.Port).Ref.(*<-chan interface{}) = inCh[2]
			*out["out"].(*chan<- int32) = outCh
		}

		go func() {
			assert.NoError(t, m.Run(context.Background()))
			close(outCh)
		}()

		go func() {
			inCh[0] <- 1
			inCh[1] <- 2
			inCh[2] <- 3

			inCh[0] <- 0
			inCh[1] <- 1
			inCh[2] <- 2

			inCh[0] <- 3
			inCh[1] <- 4
			inCh[2] <- 5

			close(inCh[0])
			close(inCh[1])
			close(inCh[2])
		}()

		assert.Equal(t, int32(6), <-outCh)
		assert.Equal(t, int32(12), <-outCh)
		assert.Equal(t, int32(0), <-outCh)
	})
}

func TestMergeGraph(t *testing.T) {
	g, err := flow.NewGraph(nil)
	assert.NoError(t, err)

	in1, in2, out := make(chan int, 8), make(chan int, 8), make(chan int, 8)
	inN1, _ := NewChanNode((<-chan int)(in1))
	inN2, _ := NewChanNode((<-chan int)(in2))
	outN, _ := NewChanNode((chan<- int)(out))
	m, _ := NewMergeNode(func(i chan<- int, o ...int) { i <- o[0] + o[1] }, 2)
	_ = g.AddNode("m", m)
	_ = g.AddNode("in1", inN1)
	_ = g.AddNode("in2", inN2)
	_ = g.AddNode("out", outN)
	_ = g.Connect([]string{"in1:out"}, []string{"m:in_0"})
	_ = g.Connect([]string{"in2:out"}, []string{"m:in_1"})
	_ = g.Connect([]string{"m:out"}, []string{"out:in"})

	in1 <- 10
	in2 <- 11
	close(in1)
	close(in2)
	assert.NoError(t, g.Run(context.Background()))
	assert.Equal(t, 21, <-out)
}
