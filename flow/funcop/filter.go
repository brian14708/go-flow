package funcop

import (
	"context"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/internal/funcx"
	"github.com/brian14708/go-flow/flow/internal/token"
)

type filter struct {
	generic
}

func (f *filter) Run(ctx context.Context) error {
	return f.forEachElem(ctx, func(val interface{}, token *token.Token, tokenCtx context.Context) error {
		bufp := bufPool.Get().(*[]interface{})
		buf := *bufp
		defer func() {
			*bufp = buf[:0]
			bufPool.Put(bufp)
		}()

		if f.hasCtx {
			buf = append(buf, ctx, val)
		} else {
			buf = append(buf, val)
		}

		ret := f.call(buf)
		var (
			retErr error
			keep   = ret[0].(bool)
		)
		if f.hasCtx {
			if err := ret[len(ret)-1]; err != nil {
				retErr = err.(error)
				keep = false
			}
		}

		if token != nil {
			if err := token.WaitSerialize(tokenCtx); err != nil {
				if f.outChan != nil && keep {
					f.outChan.DropMessage(val)
				}
				return err
			}
		}

		if retErr != nil {
			return retErr
		}
		if keep {
			f.out <- val
		}
		if token != nil {
			token.Done()
		}
		return nil
	})
}

func (f *filter) NodeType() string {
	return "Filter"
}

var filterFnTemplate = funcx.MustNewTemplate(
	(func(context.Context, funcx.T0) (bool, error))(nil),
	(func(funcx.T0) bool)(nil),
)

func newFilter(fn interface{}, opts ...FuncOption) (flow.Node, error) {
	idx, types, err := filterFnTemplate.MatchValue(fn)
	if err != nil {
		return nil, err
	}
	defer types.Free()

	f := &filter{}
	f.init(
		types.Get(funcx.T0{}), types.Get(funcx.T0{}), fn,
		idx == 0, opts,
	)
	return f, nil
}

func Filter(fn interface{}, opts ...FuncOption) flow.Node {
	f, err := newFilter(fn, opts...)
	if err != nil {
		panic("fail to create filter op: " + err.Error())
	}
	return f
}
