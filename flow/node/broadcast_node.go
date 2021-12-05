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

type broadcastNode struct {
	ch    <-chan interface{}
	dests []chan<- interface{}
	typ   reflect.Type
	fn    interface{}
}

func (l *broadcastNode) Ports() (in, out flow.PortMap) {
	in = port.MakeMap(
		"in", port.TemplatePort(&l.ch, l.typ),
	)
	out = port.MakeMap()
	for i := range l.dests {
		out[fmt.Sprintf("out_%d", i)] =
			port.TemplatePort(&l.dests[i], l.typ)
	}
	return
}

func (l *broadcastNode) Run(ctx context.Context) error {
	fn := flowtype.FuncCaller(l.fn)
	chk := flowutil.NewContextChecker(ctx)

	var args [2]interface{}
	for val := range l.ch {
		args[0] = val
		for _, f := range l.dests[1:] {
			f <- fn(args[:1], args[:0])[0]
		}
		l.dests[0] <- val

		if !chk.Valid() {
			return chk.Err()
		}
	}
	return nil
}

var broadcastFnTemplate = funcx.MustNewTemplate(
	(func(funcx.T0) funcx.T0)(nil),
)

func NewBroadcastNode(fn interface{}, n int) (flow.Node, error) {
	if n <= 0 {
		return nil, errors.New("broadcast must have output channel")
	}

	_, types, err := broadcastFnTemplate.MatchValue(fn)
	if err != nil {
		return nil, err
	}
	defer types.Free()

	l := &broadcastNode{
		fn:    fn,
		typ:   types.Get(funcx.T0{}),
		dests: make([]chan<- interface{}, n),
	}
	return l, nil
}
