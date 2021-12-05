package funcop

import (
	"context"
	"reflect"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/internal/funcx"
	"github.com/brian14708/go-flow/flow/internal/token"
)

type flatMapper struct {
	generic
}

func (m *flatMapper) Run(ctx context.Context) error {
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
			if err := ret[len(ret)-1]; err != nil {
				retErr = err.(error)
				ret[0] = nil
			}
		}

		if token != nil {
			if err := token.Wait(tokenCtx); err != nil {
				if m.outChan != nil && ret[0] != nil {
					out := reflect.ValueOf(ret[0])
					for i := 0; i < out.Len(); i++ {
						m.outChan.DropMessage(out.Index(i).Interface())
					}
				}
				return err
			}
		}
		if retErr != nil {
			return retErr
		}
		if ret[0] != nil {
			out := reflect.ValueOf(ret[0])
			for i := 0; i < out.Len(); i++ {
				m.out <- out.Index(i).Interface()
			}
		}
		if token != nil {
			token.Release()
		}

		return nil
	})
}

func (m *flatMapper) NodeType() string {
	return "FlatMapper"
}

var flatMapFnTemplate = funcx.MustNewTemplate(
	(func(context.Context, funcx.T0) ([]funcx.T1, error))(nil),
	(func(funcx.T0) []funcx.T1)(nil),
)

func newFlatMap(fn interface{}, opts ...FuncOption) (flow.Node, error) {
	idx, types, err := flatMapFnTemplate.MatchValue(fn)
	if err != nil {
		return nil, err
	}
	defer types.Free()

	m := &flatMapper{}
	m.init(
		types.Get(funcx.T0{}), types.Get(funcx.T1{}), fn,
		idx == 0, opts,
	)
	return m, nil
}

func FlatMap(fn interface{}, opts ...FuncOption) flow.Node {
	m, err := newFlatMap(fn, opts...)
	if err != nil {
		panic("fail to create flat map op: " + err.Error())
	}
	return m
}
