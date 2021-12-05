package node

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/internal/funcx"
	"github.com/brian14708/go-flow/flow/port"
	"github.com/brian14708/go-flow/flowtype"
	"github.com/brian14708/go-flow/flowutil"
)

type mergeNode struct {
	srcs   []<-chan interface{}
	ch     interface{} // *<-chan outType
	fn     interface{}
	inType reflect.Type
}

func (m *mergeNode) Ports() (in, out flow.PortMap) {
	in = port.MakeMap()
	for i := range m.srcs {
		in[fmt.Sprintf("in_%d", i)] =
			port.TemplatePort(&m.srcs[i], m.inType)
	}
	out = port.MakeMap(
		"out", m.ch,
	)
	return
}

func (m *mergeNode) Run(ctx context.Context) error {
	fn := flowtype.FuncCaller(m.fn)
	chk := flowutil.NewContextChecker(ctx)
	vals := make([]interface{}, len(m.srcs)+1)
	vals[0] = reflect.ValueOf(m.ch).Elem().Interface()
	for {
		for i, src := range m.srcs {
			val, ok := <-src
			if !ok {
				return nil
			}
			vals[i+1] = val
		}
		fn(vals, nil)

		if !chk.Valid() {
			return chk.Err()
		}
	}
}

var mergeFnTemplate = funcx.MustNewTemplate(
	(func(chan<- funcx.T0, ...funcx.T1))(nil),
)

func NewMergeNode(fn interface{}, n int) (flow.Node, error) {
	if n <= 0 {
		return nil, errors.New("merge node must have input channel")
	}
	_, types, err := mergeFnTemplate.MatchValue(fn)
	if err != nil {
		return nil, err
	}
	defer types.Free()

	m := &mergeNode{
		srcs:   make([]<-chan interface{}, n),
		ch:     newChanPtr(reflect.SendDir, types.Get(funcx.T0{})),
		inType: types.Get(funcx.T1{}),
		fn:     fn,
	}
	return m, nil
}
