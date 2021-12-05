package rrworker

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/internal/funcx"
	"github.com/brian14708/go-flow/flow/internal/token"
	"github.com/brian14708/go-flow/flow/port"
	"github.com/brian14708/go-flow/flowtype"
)

type submitNode struct {
	in  <-chan interface{}
	out chan<- interface{}
	w   *RRWorker

	inType       reflect.Type
	taskPreparer flowtype.CallFunc
	errHandler   func(error) error
	queue        *token.TokenQueue

	opt submitOptions
}

func (c *submitNode) Ports() (in, out flow.PortMap) {
	in = port.MakeMap(
		"in", port.TemplatePort(&c.in, c.inType),
	)

	var rep reflect.Type
	if _, rep = c.w.Type(); rep == nil {
		rep = reflect.TypeOf((*flow.AnyMessage)(nil)).Elem()
	}
	out = port.MakeMap(
		"out", port.TemplatePort(&c.out, rep),
	)
	return
}

func (c *submitNode) Run(ctx context.Context) error {
	var ch flow.Chan
	if _, fctx := flow.FromContext(ctx); fctx != nil {
		ch = fctx.GetChan("in")
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	var (
		outErr  error
		errOnce sync.Once
	)
	setError := func(err error) {
		errOnce.Do(func() {
			outErr = err
			cancel()
		})
	}

	var buf [1]interface{}
	for {
		token, err := c.queue.Acquire(ctx)
		if err != nil {
			setError(err)
			break
		}

		var (
			msg interface{}
			ok  bool
		)
		select {
		case msg, ok = <-c.in:
		case <-ctx.Done():
		}
		if !ok {
			break
		}

		if c.taskPreparer != nil {
			buf[0] = msg
			msg = c.taskPreparer(buf[:], buf[:0])[0]
		}

		wg.Add(1)
		err = c.w.Submit(
			context.Background(), // no need to cancel if error
			msg.(TaskInterface),
			func(v interface{}, err error) {
				go func() {
					defer wg.Done()
					if waitErr := token.Wait(ctx); waitErr != nil {
						if ch != nil && err == nil {
							ch.DropMessage(v)
						}
						setError(waitErr)
						return
					}
					if err != nil {
						if c.errHandler != nil {
							err = c.errHandler(err)
						}
						if err != nil {
							setError(err)
							return
						}
					} else {
						c.out <- v
					}
					token.Release()
				}()
			},
		)
		if err != nil {
			setError(err)
			wg.Done()
			break
		}
	}
	wg.Wait()
	return outErr
}

func (c *submitNode) NodeType() string {
	return "RRWorkerSubmit"
}

func (c *submitNode) Description() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "RRWorker: %s\n", c.w.g.ID())
	fmt.Fprintf(buf, "Parallel: %d\n", c.opt.parallel)
	fmt.Fprintf(buf, "Ordered: %v", c.opt.ordered)
	return buf.String()
}

var taskPreparerFnTemplate = funcx.MustNewTemplate(
	(func(funcx.T0) funcx.T1)(nil),
)

func NewSubmitNode(w *RRWorker, opts ...SubmitNodeOption) (flow.Node, error) {
	o := submitOptions{
		parallel: 1,
	}
	for _, opt := range opts {
		opt(&o)
	}

	req, _ := w.Type()
	s := &submitNode{
		w: w,

		inType:     req,
		errHandler: o.errHandler,
		queue:      token.NewTokenQueue(o.parallel, o.ordered),
		opt:        o,
	}

	if o.taskPreparer != nil {
		_, types, err := taskPreparerFnTemplate.MatchValue(o.taskPreparer)
		if err != nil {
			return nil, fmt.Errorf("invalid task maker: %w", err)
		}
		defer types.Free()
		if typ := types.Get(funcx.T1{}); s.inType != typ {
			return nil, fmt.Errorf("invalid task maker output `%s', expected `%s'", typ, s.inType)
		}
		s.inType = types.Get(funcx.T0{})
		s.taskPreparer = flowtype.FuncCaller(o.taskPreparer)
	}
	return s, nil
}

func SubmitNode(w *RRWorker, opts ...SubmitNodeOption) flow.Node {
	n, err := NewSubmitNode(w, opts...)
	if err != nil {
		panic("fail to create worker submit op: " + err.Error())
	}
	return n
}

//

type submitOptions struct {
	taskPreparer interface{}
	errHandler   func(error) error
	parallel     int
	ordered      bool
}

type SubmitNodeOption func(*submitOptions)

func WithTaskPreparer(fn interface{}) SubmitNodeOption {
	return func(o *submitOptions) {
		o.taskPreparer = fn
	}
}

func WithErrorHandler(fn func(error) error) SubmitNodeOption {
	return func(o *submitOptions) {
		o.errHandler = fn
	}
}

func WithParallel(p int, ordered bool) SubmitNodeOption {
	if p < 1 {
		panic("invalid parallel value")
	}
	return func(o *submitOptions) {
		o.parallel = p
		o.ordered = ordered
	}
}
