package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow"
)

func TestSelectNode(t *testing.T) {
	_, err := NewSelectNode(0)
	assert.Error(t, err)

	s, err := NewSelectNode(3)
	assert.NoError(t, err)

	idxCh := make(chan int, 10)
	inCh := []chan flow.AnyMessage{
		make(chan flow.AnyMessage, 1),
		make(chan flow.AnyMessage, 2),
		make(chan flow.AnyMessage, 3),
	}
	outCh := make(chan flow.AnyMessage)
	{
		in, out := s.Ports()
		*in["in_idx"].(*<-chan int) = idxCh
		*in["in_0"].(*<-chan flow.AnyMessage) = inCh[0]
		*in["in_1"].(*<-chan flow.AnyMessage) = inCh[1]
		*in["in_2"].(*<-chan flow.AnyMessage) = inCh[2]
		*out["out"].(*chan<- flow.AnyMessage) = outCh
	}

	go func() {
		assert.NoError(t, s.Run(context.Background()))
		close(outCh)
	}()

	inCh[0] <- 5
	inCh[1] <- 6
	inCh[2] <- 7

	idxCh <- 1
	idxCh <- 0
	idxCh <- 2
	close(idxCh)
	assert.Equal(t, 6, <-outCh)
	assert.Equal(t, 5, <-outCh)
	assert.Equal(t, 7, <-outCh)
	assert.Nil(t, <-outCh)
}

func TestSelectNodeError(t *testing.T) {
	s, err := NewSelectNode(1)
	assert.NoError(t, err)

	idxCh := make(chan int, 10)
	inCh := make(chan flow.AnyMessage, 1)
	outCh := make(chan flow.AnyMessage)
	{
		in, out := s.Ports()
		*in["in_idx"].(*<-chan int) = idxCh
		*in["in_0"].(*<-chan flow.AnyMessage) = inCh
		*out["out"].(*chan<- flow.AnyMessage) = outCh
	}
	close(inCh)
	idxCh <- 0

	assert.Panics(t, func() {
		_ = s.Run(context.Background())
	})
}
