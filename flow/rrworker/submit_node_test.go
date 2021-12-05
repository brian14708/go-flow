package rrworker

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/node"
)

type IntTask struct {
	Task
	v int
}

func (t *IntTask) SetResult(v int) {
	t.SetResultAny(v)
}

func TestSubmitNode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var w *RRWorker
	{
		g, err := flow.NewGraph(nil)
		assert.NoError(t, err)

		n, _ := node.NewFuncNode(func(a <-chan *IntTask) {
			r1, r2, r3, r4 := <-a, <-a, <-a, <-a
			r4.SetResult(4)
			r2.SetResult(2)
			r1.SetResult(1)
			r3.SetError(errors.New("A"))
		})
		assert.NoError(t, g.AddNode("work", n))

		w, err = New(flow.PartialGraph(g, []string{"work:in"}, nil))
		assert.NoError(t, err)
		req, rep := w.Type()
		assert.Equal(t, reflect.TypeOf((*IntTask)(nil)), req)
		assert.Equal(t, reflect.TypeOf((*int)(nil)).Elem(), rep)

		go func() { _ = w.Run(ctx) }()
	}

	assert.Panics(t, func() {
		SubmitNode(w,
			WithTaskPreparer(func(int) *Task {
				return nil
			}),
		)
	})

	// start main graph
	{
		g, err := flow.NewGraph(nil)
		assert.NoError(t, err)
		in, out := make(chan int), make(chan int)
		n, _ := node.NewChanNode((<-chan int)(in))
		assert.NoError(t, g.AddNode("in", n))
		n, _ = node.NewChanNode((chan<- int)(out))
		assert.NoError(t, g.AddNode("out", n))

		var errCnt int
		_ = g.AddNode("work", SubmitNode(w,
			WithTaskPreparer(func(i int) *IntTask {
				return &IntTask{v: i}
			}),
			WithErrorHandler(func(err error) error {
				errCnt++
				return err
			}),
			WithParallel(10, true),
		))
		_ = g.Connect([]string{"in:out"}, []string{"work:in"})
		_ = g.Connect([]string{"work:out"}, []string{"out:in"})

		errCh := make(chan error, 1)
		go func() {
			errCh <- g.Run(ctx)
		}()
		in <- 0
		in <- 0
		in <- 0
		in <- 0

		assert.Equal(t, 1, <-out)
		assert.Equal(t, 2, <-out)
		// r4 gets dropped becasue, r3 cause an error
		assert.Equal(t, 0, <-out)
		assert.Error(t, <-errCh)
		assert.Equal(t, 1, errCnt)
	}
}
