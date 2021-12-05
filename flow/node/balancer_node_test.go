package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow"
)

func setupBalancer(balancer flow.Node) (chan flow.AnyMessage, []chan flow.AnyMessage, chan int) {
	inCh := make(chan flow.AnyMessage)
	idxCh := make(chan int, 10)
	outCh := []chan flow.AnyMessage{
		make(chan flow.AnyMessage, 1),
		make(chan flow.AnyMessage, 2),
		make(chan flow.AnyMessage, 3),
	}
	{
		in, out := balancer.Ports()
		*in["in"].(*<-chan flow.AnyMessage) = inCh
		*out["out_idx"].(*chan<- int) = idxCh
		*out["out_0"].(*chan<- flow.AnyMessage) = outCh[0]
		*out["out_1"].(*chan<- flow.AnyMessage) = outCh[1]
		*out["out_2"].(*chan<- flow.AnyMessage) = outCh[2]
	}
	go func() {
		if err := balancer.Run(context.Background()); err != nil {
			panic(err)
		}
		close(outCh[0])
		close(outCh[1])
		close(outCh[2])
		close(idxCh)
	}()
	return inCh, outCh, idxCh
}

func TestBalancer(t *testing.T) {
	balancer, err := NewBalancerNode(3)
	assert.NoError(t, err)
	inCh, outCh, idxCh := setupBalancer(balancer)
	for i := 0; i < 7; i++ {
		inCh <- i
	}
	close(inCh)

	testcases := [...]struct {
		ch     chan flow.AnyMessage
		chInt  chan int
		result []int
	}{
		{outCh[0], nil, []int{0, 6}},
		{outCh[1], nil, []int{1, 4}},
		{outCh[2], nil, []int{2, 3, 5}},
		{nil, idxCh, []int{0, 1, 2, 2, 1, 2, 0}},
	}
	for _, test := range testcases {
		var tmp []int
		if test.ch != nil {
			for i := range test.ch {
				tmp = append(tmp, i.(int))
			}
		}
		if test.chInt != nil {
			for i := range test.chInt {
				tmp = append(tmp, i)
			}
		}
		assert.Equal(t, test.result, tmp)
	}
}

func BenchmarkBalancer(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	balancer, _ := NewBalancerNode(3)
	inCh, outCh, idxCh := setupBalancer(balancer)
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
		<-outCh[<-idxCh]
	}
}
