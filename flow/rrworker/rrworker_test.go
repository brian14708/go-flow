package rrworker

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/node"
)

func TestRRWorker(t *testing.T) {
	g, err := flow.NewGraph(nil)
	assert.NoError(t, err)

	type ATask struct {
		Task
	}
	n, _ := node.NewFuncNode(func(_ context.Context, a <-chan *ATask) error {
		(<-a).SetResultAny(123)
		(<-a).SetError(errors.New("error"))
		return nil
	})
	assert.NoError(t, g.AddNode("work", n))

	w, err := New(flow.PartialGraph(g, []string{"work:in"}, nil))
	assert.NoError(t, err)
	req, rep := w.Type()
	assert.Equal(t, reflect.TypeOf((*ATask)(nil)), req)
	assert.Equal(t, nil, rep)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_ = w.Run(ctx)
		wg.Done()
	}()

	v, e := w.SubmitWait(ctx, &ATask{})
	assert.Equal(t, 123, v)
	assert.NoError(t, e)
	_, e = w.SubmitWait(ctx, &ATask{})
	assert.Error(t, e)

	cancel()
	wg.Wait()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		_, err = w.SubmitWait(ctx, &ATask{})
		cancel()
		if err == errWorkerClosed {
			break
		}
	}
}

type BadTask struct{ Task }

func (*BadTask) SetResult(int, int) {}

func TestBadType(t *testing.T) {
	g, err := flow.NewGraph(nil)
	assert.NoError(t, err)

	n, _ := node.NewFuncNode(func(<-chan *Task) {})
	assert.NoError(t, g.AddNode("task", n))
	_, err = New(flow.PartialGraph(g, []string{"task:in"}, nil))
	assert.NoError(t, err)

	n, _ = node.NewFuncNode(func(<-chan *BadTask) {})
	assert.NoError(t, g.AddNode("badtask", n))
	_, err = New(flow.PartialGraph(g, []string{"badtask:in"}, nil))
	assert.Error(t, err)

	n, _ = node.NewFuncNode(func(<-chan BadTask) {})
	assert.NoError(t, g.AddNode("badtask_v", n))
	_, err = New(flow.PartialGraph(g, []string{"badtask_v:in"}, nil))
	assert.Error(t, err)
}
