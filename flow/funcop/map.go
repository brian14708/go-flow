package funcop

import (
	"context"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/internal/funcx"
	"github.com/brian14708/go-flow/flow/internal/token"
)

type mapper struct {
	generic
}

func (m *mapper) Run(ctx context.Context) error {
	return m.forEachElem(ctx, func(val interface{}, token *token.Token, tokenCtx context.Context) error {
		bufp := bufPool.Get().(*[]interface{})
		buf := *bufp
		defer func() {
			*bufp = buf[:0]
			bufPool.Put(bufp)
		}()

		if m.hasCtx {
			buf = append(buf, ctx, val)
		} else {
			buf = append(buf, val)
		}

		ret := m.call(buf)
		var retErr error
		if m.hasCtx {
			err := ret[len(ret)-1]
			ret = ret[:len(ret)-1]
			if err != nil {
				retErr = err.(error)
				ret = nil
			}
		}

		if token != nil {
			if err := token.WaitSerialize(tokenCtx); err != nil {
				if m.outChan != nil && len(ret) == 1 {
					m.outChan.DropMessage(ret[0])
				}
				return err
			}
		}
		if retErr != nil {
			return retErr
		}
		if len(ret) == 1 {
			m.out <- ret[0]
		}
		if token != nil {
			token.Done()
		}
		return nil
	})
}

func (m *mapper) NodeType() string {
	return "Mapper"
}

var mapFnTemplate = funcx.MustNewTemplate(
	(func(context.Context, funcx.T0) (funcx.T1, error))(nil),
	(func(context.Context, funcx.T0) error)(nil),
	(func(funcx.T0) funcx.T1)(nil),
	(func(funcx.T0))(nil),
)

func newMap(fn interface{}, opts ...FuncOption) (flow.Node, error) {
	idx, types, err := mapFnTemplate.MatchValue(fn)
	if err != nil {
		return nil, err
	}
	defer types.Free()

	m := &mapper{}
	m.init(
		types.Get(funcx.T0{}), types.Get(funcx.T1{}), fn,
		idx <= 1, opts,
	)
	return m, nil
}

func Map(fn interface{}, opts ...FuncOption) flow.Node {
	m, err := newMap(fn, opts...)
	if err != nil {
		panic("fail to create map op: " + err.Error())
	}
	return m
}
