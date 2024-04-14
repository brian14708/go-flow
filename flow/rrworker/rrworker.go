package rrworker

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/node"
	"github.com/brian14708/go-flow/flowtype"
)

var errWorkerClosed = errors.New("rrworker closed")

// request-reply worker.
type RRWorker struct {
	g       *flow.Graph
	outType reflect.Type

	mu   sync.RWMutex
	ch   reflect.Value
	send flowtype.SendFunc
}

var typeTaskInterface = reflect.TypeOf((*TaskInterface)(nil)).Elem()

func New(b flow.GraphBuilder, opts ...Option) (*RRWorker, error) {
	g, inPorts, outPorts, err := b.BuildGraph()
	if err != nil {
		return nil, err
	}
	if len(inPorts) == 0 {
		return nil, errors.New("rrworker must have input")
	}
	if len(outPorts) != 0 {
		return nil, errors.New("rrworker must have not have output")
	}

	o := options{
		chanSize: -1,
	}
	for _, opt := range opts {
		opt.apply(&o)
	}

	t, err := g.PortType(inPorts[0])
	if err != nil {
		return nil, err
	}
	if !t.Implements(typeTaskInterface) {
		return nil, fmt.Errorf("worker input type `%s' must implement pipeline.TaskInterface", t)
	}

	var outType reflect.Type
	for i := 0; i < t.NumMethod(); i++ {
		if m := t.Method(i); m.Name == "SetResult" {
			if m.Type.NumIn() != 2 || m.Type.NumOut() != 0 {
				return nil, fmt.Errorf(
					"SetResult method type `%s', expected `func (*Task) (T)'",
					m.Type,
				)
			}
			outType = m.Type.In(1)
			break
		}
	}

	ch := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, t), 0)
	n, err := node.NewChanNode(ch.Convert(reflect.ChanOf(reflect.RecvDir, t)).Interface())
	if err != nil {
		return nil, err
	}
	if err := g.AddNode("entrypoint", n); err != nil {
		return nil, err
	}
	var connOpts []flow.ConnectOption
	if o.chanSize >= 0 {
		connOpts = append(connOpts, flow.WithChanSize(o.chanSize))
	}
	err = g.Connect(
		[]string{"entrypoint:out"},
		inPorts,
		connOpts...,
	)
	if err != nil {
		return nil, err
	}

	return &RRWorker{
		g:       g,
		ch:      ch,
		send:    flowtype.ChanSender(ch.Interface()),
		outType: outType,
	}, nil
}

func (w *RRWorker) Type() (req, rep reflect.Type) {
	return w.ch.Type().Elem(), w.outType
}

func (w *RRWorker) Run(ctx context.Context) error {
	go func() {
		if ctx.Done() != nil {
			<-ctx.Done()
			w.mu.Lock()
			w.ch.Close()
			w.send = nil
			w.mu.Unlock()
		}
	}()
	err := w.g.Run(valueOnlyContext{ctx})
	if err != nil {
		// must panic, might be pending tasks in-queue causing deadlock
		panic("worker `" + w.g.ID() + "' exited with error: " + err.Error())
	}
	return ctx.Err()
}

func (w *RRWorker) Graph() *flow.Graph {
	return w.g
}

func (w *RRWorker) Submit(ctx context.Context, t TaskInterface, cb ResultCallback) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.send == nil {
		return errWorkerClosed
	}

	t.setCallback(cb)
	if !w.send(t, ctx.Done(), true) {
		return ctx.Err()
	}
	return nil
}

func (w *RRWorker) SubmitWait(ctx context.Context, t TaskInterface) (interface{}, error) {
	type result struct {
		v interface{}
		e error
	}
	ch := make(chan result, 1)
	err := w.Submit(ctx, t, func(val interface{}, err error) {
		ch <- result{v: val, e: err}
		close(ch)
	})
	if err != nil {
		return nil, err
	}
	select {
	case r := <-ch:
		return r.v, r.e
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type valueOnlyContext struct{ context.Context }

func (valueOnlyContext) Deadline() (deadline time.Time, ok bool) { return }
func (valueOnlyContext) Done() <-chan struct{}                   { return nil }
func (valueOnlyContext) Err() error                              { return nil }
