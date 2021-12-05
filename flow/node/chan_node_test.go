package node

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flowtype/testutil"
)

func TestMakeChanNode(t *testing.T) {
	testutil.TestWithNilRegistry(t, func(t *testing.T) {
		ch := make(chan int)
		_, err := NewChanNode(&ch)
		assert.Error(t, err)
		_, err = NewChanNode(ch)
		assert.Error(t, err)

		t.Run("send", func(t *testing.T) {
			ch := make(chan int, 5)
			send, err := NewChanNode((chan<- int)(ch))
			assert.NoError(t, err)
			in, out := send.Ports()
			assert.Equal(t, 1, len(in))
			assert.Equal(t, 0, len(out))
			assert.Equal(t, reflect.TypeOf((chan<- int)(ch)).String(), send.(interface {
				NodeType() string
			}).NodeType())
			tmp := make(chan int, 5)
			*in["in"].(*<-chan int) = tmp
			tmp <- 5
			close(tmp)
			assert.NoError(t, send.Run(context.Background()))
			assert.Equal(t, 5, <-ch)
			assert.Equal(t, 0, <-ch)
		})

		t.Run("recv", func(t *testing.T) {
			ch := make(chan int)
			recv, err := NewChanNode((<-chan int)(ch))
			assert.NoError(t, err)
			in, out := recv.Ports()
			assert.Equal(t, 1, len(out))
			assert.Equal(t, 0, len(in))
			assert.Equal(t, reflect.TypeOf((<-chan int)(ch)).String(), recv.(interface {
				NodeType() string
			}).NodeType())
			close(ch)
			assert.NoError(t, recv.Run(context.Background()))
		})
	})
}

func TestChanNodeContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ch := make(chan int)
	recv, err := NewChanNode((<-chan int)(ch))
	assert.NoError(t, err)

	assert.Equal(t, ctx.Err(), recv.Run(ctx))
}
