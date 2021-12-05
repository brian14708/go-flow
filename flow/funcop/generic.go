package funcop

import (
	"context"
	"reflect"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/internal/token"
	"github.com/brian14708/go-flow/flow/port"
	"github.com/brian14708/go-flow/flowtype"
	"github.com/brian14708/go-flow/flowutil"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]interface{}, 0, 2)
		return &buf
	},
}

type generic struct {
	in      <-chan interface{}
	out     chan<- interface{}
	inType  reflect.Type
	outType reflect.Type
	outChan interface {
		DropMessage(interface{})
	}
	callFunc flowtype.CallFunc
	hasCtx   bool

	opt options
}

func (g *generic) call(buf []interface{}) []interface{} {
	return g.callFunc(buf, buf[:0])
}

func (g *generic) Ports() (in, out flow.PortMap) {
	if g.inType != nil {
		in = port.MakeMap(
			"in", port.TemplatePort(&g.in, g.inType),
		)
	}
	if g.outType != nil {
		out = port.MakeMap(
			"out", port.TemplatePort(&g.out, g.outType),
		)
	}
	return
}

func (g *generic) Description() string {
	return g.opt.String()
}

func (g *generic) forEachElem(
	ctx context.Context,
	exec func(val interface{}, token *token.Token, tokenCtx context.Context) error,
) error {
	if _, fctx := flow.FromContext(ctx); fctx != nil {
		g.outChan = fctx.GetChan("out")
	}

	if g.in == nil {
		return nil
	}

	if g.opt.parallel == 1 {
		chk := flowutil.NewContextChecker(ctx)
		for val := range g.in {
			if err := exec(val, nil, nil); err != nil {
				return err
			}

			if !chk.Valid() {
				return chk.Err()
			}
		}
		return nil
	} else if !g.opt.ordered {
		tg, tctx := errgroup.WithContext(ctx)
		for i := 0; i < g.opt.parallel; i++ {
			tg.Go(func() error {
				for {
					var (
						val interface{}
						ok  bool
					)
					select {
					case val, ok = <-g.in:
					case <-tctx.Done():
					}
					if !ok {
						return nil
					}

					if err := exec(val, nil, nil); err != nil {
						return err
					}
				}
			})
		}
		return tg.Wait()
	} else {
		q := token.NewTokenQueue(g.opt.parallel, g.opt.ordered)
		tg, tctx := errgroup.WithContext(ctx)

		for {
			token, err := q.Acquire(tctx)
			if err != nil {
				break
			}

			var (
				val interface{}
				ok  bool
			)
			select {
			case val, ok = <-g.in:
			case <-tctx.Done():
			}
			if !ok {
				break
			}

			tg.Go(func() error {
				return exec(val, token, tctx)
			})
		}

		return tg.Wait()
	}
}

func (g *generic) init(in, out reflect.Type, fn interface{}, hasCtx bool, opts []FuncOption) {
	g.callFunc = flowtype.FuncCaller(fn)
	g.hasCtx = hasCtx
	g.opt = options{
		parallel: 1,
	}
	for _, opt := range opts {
		opt(&g.opt)
	}
	if in != nil {
		g.inType = in
	}
	if out != nil {
		g.outType = out
	}
}
